package airuntime

import (
	"context"
	"unicode/utf8"
)

// S01Input is one map attempt. Body is split inside the runtime and is not stored.
type S01Input struct {
	ContentType string
	Body        []byte
	PackID      string
	JobID       string
}

// S01Output is an enterable map. CacheHit means the provider was not called again.
type S01Output struct {
	Map      MapTree
	CacheHit bool
}

// NodeRef is a map point selected for a small prefetch. Title is not the textbook.
type NodeRef struct {
	ID    string
	Title string
}

// S02Input asks for a small batch on the current node and the next one.
type S02Input struct {
	Nodes  []NodeRef
	PackID string
	JobID  string
}

// S02Output is the truncated prefetch. It does not change the import job.
type S02Output struct {
	Questions []Question
}

// RunS01 runs one S01 attempt through Call unless the same material and
// template version are already cached. Success and failure both append a row.
func (r *Runtime) RunS01(ctx context.Context, in S01Input) (S01Output, error) {
	if err := ctx.Err(); err != nil {
		return S01Output{}, err
	}
	subclass, tier, media, ok := route(in.ContentType)
	if !ok {
		return S01Output{}, ErrSubclassRequired
	}
	index, err := r.nextIndex(ctx, in.JobID, SceneS01)
	if err != nil {
		return S01Output{}, err
	}
	base := r.s01Base(in, subclass, tier, media, index)
	version := r.template(SceneS01)
	base.TemplateVersion = version
	if version == "" {
		return S01Output{}, r.failAttempt(ctx, base, ErrNoTemplate)
	}
	if r.isStopped(SceneS01) {
		return S01Output{}, r.failAttempt(ctx, base, ErrSceneStopped)
	}
	if tree, hit := r.cached(fingerprint(in.Body), version); hit {
		base.Success = true
		base.CacheHit = true
		base.Tokens = 0
		base.MetricValue = 0
		if err := r.append(ctx, base); err != nil {
			return S01Output{}, err
		}
		return S01Output{Map: tree, CacheHit: true}, nil
	}
	parts := split(media, in.Body)
	if len(parts) == 0 || !fragmentsStayApart(in.Body, parts) {
		return S01Output{}, r.failAttempt(ctx, base, ErrFragmentBoundary)
	}
	var tokens int
	var merged MapTree
	for _, part := range parts {
		res, callErr := r.Call(ctx, LedgerEntry{
			SceneCode:       SceneS01,
			Subclass:        subclass,
			Tier:            tier,
			TemplateVersion: version,
			Media:           media,
			Fragments:       [][]byte{part},
		})
		if callErr != nil {
			return S01Output{}, r.failAttempt(ctx, base, callErr)
		}
		tokens += res.Tokens
		merged = mergeTree(merged, res.Map)
	}
	base.MetricValue = metricValue(media, in.Body, len(parts))
	if !mapEnterable(merged) || leaksSource(merged, parts) {
		base.Tokens = tokens
		return S01Output{}, r.failAttempt(ctx, base, ErrNotEnterable)
	}
	base.Success = true
	base.Tokens = tokens
	if err := r.append(ctx, base); err != nil {
		return S01Output{}, err
	}
	r.remember(fingerprint(in.Body), version, merged)
	return S01Output{Map: merged}, nil
}

// FailS01 records an S01 attempt that stopped before a provider call.
// It does not invent OCR or ASR when the route is text.
func (r *Runtime) FailS01(ctx context.Context, in S01Input) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	subclass, tier, media, ok := route(in.ContentType)
	if !ok {
		return ErrSubclassRequired
	}
	index, err := r.nextIndex(ctx, in.JobID, SceneS01)
	if err != nil {
		return err
	}
	base := r.s01Base(in, subclass, tier, media, index)
	base.TemplateVersion = r.template(SceneS01)
	base.Success = false
	return r.append(ctx, base)
}

// RunS02 prefetches a small batch after the caller has committed map_ready.
// More nodes than the current and next one are refused. The cap is not a
// locked product number. One hundred is not the batch size.
func (r *Runtime) RunS02(ctx context.Context, in S02Input) (S02Output, error) {
	if err := ctx.Err(); err != nil {
		return S02Output{}, err
	}
	if len(in.Nodes) == 0 {
		return S02Output{}, nil
	}
	index, err := r.nextIndex(ctx, in.JobID, SceneS02)
	if err != nil {
		return S02Output{}, err
	}
	base := Entry{
		SceneCode:       SceneS02,
		Subclass:        SubclassPrefetch,
		Tier:            TierLight,
		RetryIndex:      index,
		PackID:          in.PackID,
		JobID:           in.JobID,
		Caller:          CallerLearner,
		TemplateVersion: r.template(SceneS02),
		MetricKind:      metricFragments,
	}
	if len(in.Nodes) > prefetchNodes || base.TemplateVersion == "" || r.isStopped(SceneS02) {
		base.MetricValue = len(in.Nodes)
		return S02Output{}, r.failAttempt(ctx, base, ErrFragmentBoundary)
	}
	var tokens int
	var questions []Question
	for _, node := range in.Nodes {
		res, callErr := r.Call(ctx, LedgerEntry{
			SceneCode:       SceneS02,
			Subclass:        SubclassPrefetch,
			Tier:            TierLight,
			TemplateVersion: base.TemplateVersion,
			NodeTitle:       node.Title,
			QuestionLimit:   prefetchPerNode,
		})
		if callErr != nil {
			base.Tokens = 0
			base.MetricValue = len(in.Nodes)
			return S02Output{}, r.failAttempt(ctx, base, callErr)
		}
		tokens += res.Tokens
		kept := 0
		for _, question := range res.Questions {
			if kept >= prefetchPerNode {
				break
			}
			stem := trimStem(question.Stem)
			if stem == "" {
				continue
			}
			questions = append(questions, Question{NodeID: node.ID, Stem: stem})
			kept++
		}
	}
	base.Success = true
	base.Tokens = tokens
	base.MetricValue = len(in.Nodes)
	if err := r.append(ctx, base); err != nil {
		return S02Output{}, err
	}
	return S02Output{Questions: questions}, nil
}

func (r *Runtime) s01Base(in S01Input, subclass, tier, media string, index int) Entry {
	return Entry{
		SceneCode:   SceneS01,
		Subclass:    subclass,
		Tier:        tier,
		RetryIndex:  index,
		PackID:      in.PackID,
		JobID:       in.JobID,
		Caller:      CallerLearner,
		Fingerprint: fingerprint(in.Body),
		MetricKind:  metricKind(media),
	}
}

func (r *Runtime) failAttempt(ctx context.Context, entry Entry, cause error) error {
	entry.Success = false
	entry.CacheHit = false
	if err := r.append(ctx, entry); err != nil {
		return err
	}
	if cause == nil {
		return ErrNotEnterable
	}
	return cause
}

func metricValue(media string, body []byte, parts int) int {
	switch media {
	case mediaImage:
		return 1
	case mediaAudio:
		return len(body)
	default:
		return parts
	}
}

func (r *Runtime) cached(fp, version string) (MapTree, bool) {
	if fp == "" || version == "" {
		return MapTree{}, false
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	tree, ok := r.reuse[fp+"\n"+version]
	return tree, ok
}

func (r *Runtime) remember(fp, version string, tree MapTree) {
	if fp == "" || version == "" {
		return
	}
	r.mu.Lock()
	r.reuse[fp+"\n"+version] = tree
	r.mu.Unlock()
}

func trimStem(stem string) string {
	if stem == "" || utf8.RuneCountInString(stem) > maxTitleRunes {
		return ""
	}
	for _, r := range stem {
		if r == '\n' || r == '\r' {
			return ""
		}
	}
	return stem
}
