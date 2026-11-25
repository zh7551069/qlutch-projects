package airuntime

import (
	"context"
	"errors"
	"sync"
)

const (
	// SceneS01 is textbook parsing and knowledge-point mining.
	SceneS01 = "S01"
	// SceneS02 is background prefetch of a small question batch.
	SceneS02 = "S02"

	// SubclassOCR is image recognition. Paste notes must not use it.
	SubclassOCR = "S01-OCR"
	// SubclassASR is audio transcription. Paste notes must not use it.
	SubclassASR = "S01-ASR"
	// SubclassText is text or document mining. It is not OCR or ASR.
	SubclassText = "S01-TEXT"
	// SubclassPrefetch is the S02 background prefetch subclass from the A5 contract.
	SubclassPrefetch = "后台预生成"

	// TierLight, TierRecognize, and TierStandard are routing tiers.
	// Routing is by tier, not by vendor. Heavy is unused in this slice.
	TierLight       = "轻"
	TierRecognize   = "识别"
	TierStandard    = "标准"
	CallerLearner   = "learner"
	templateS01     = "s01-map.v1"
	templateS02     = "s02-prefetch.v1"
	maxTitleRunes   = 80
	prefetchNodes   = 2
	prefetchPerNode = 2
)

// ErrProviderNotConfigured is returned until a provider adapter is configured.
var ErrProviderNotConfigured = errors.New("未配置提供方")

// ErrSceneCodeRequired means a call arrived without the accounting scene code.
var ErrSceneCodeRequired = errors.New("缺少场景码")

// ErrSubclassRequired means the event has no real subclass.
// Such an event is not recorded as 其他.
var ErrSubclassRequired = errors.New("缺少子类")

// ErrNoTemplate means this scene has no effective template version.
var ErrNoTemplate = errors.New("无生效模板")

// ErrFragmentBoundary means a call would merge separate fragments, or S02
// received material bytes.
var ErrFragmentBoundary = errors.New("片段不能合并成一次请求")

// ErrNotEnterable means S01 did not produce a map a learner can enter.
var ErrNotEnterable = errors.New("地图不可进入")

// ErrLedger means the attempt was not recorded.
var ErrLedger = errors.New("账本写入失败")

// ErrSceneStopped means the scene entry is refusing calls.
var ErrSceneStopped = errors.New("场景不可用")

// ErrUsage means the ledger row is missing a scene, subclass, or caller.
var ErrUsage = errors.New("用量场景不正确")

// LedgerEntry is one model call. SceneCode is required.
// Fragments holds exactly one piece for S01. S02 leaves it empty.
// The entry is not a place to store a prompt or a textbook.
type LedgerEntry struct {
	SceneCode       string
	Subclass        string
	Tier            string
	TemplateVersion string
	Fragments       [][]byte
	Media           string
	NodeTitle       string
	QuestionLimit   int
}

// CallResult is the provider result for one fragment or one node.
type CallResult struct {
	Tokens    int
	Map       MapTree
	Questions []Question
}

// MapTree is a catalog → chapter → point map. Titles only.
type MapTree struct {
	Catalogs []Catalog
}

// Catalog is the top of an enterable map.
type Catalog struct {
	Title    string
	Chapters []Chapter
}

// Chapter sits under a catalog.
type Chapter struct {
	Title  string
	Points []Point
}

// Point is a knowledge point a learner can enter.
type Point struct {
	Title string
}

// Question is one prefetched stem. It does not carry source text.
type Question struct {
	NodeID string
	Stem   string
}

// ProviderRequest is one provider call. It carries one fragment or one node title.
type ProviderRequest struct {
	SceneCode       string
	Subclass        string
	Tier            string
	TemplateVersion string
	Media           string
	Fragment        []byte
	NodeTitle       string
	QuestionLimit   int
}

// ProviderResponse is the provider output for that one call.
type ProviderResponse struct {
	Tokens    int
	Map       MapTree
	Questions []Question
}

// Provider is the adapter slot. The production process does not install a
// stand-in that marks a job ready. A missing provider is an unconfigured call.
type Provider interface {
	Complete(ctx context.Context, req ProviderRequest) (ProviderResponse, error)
}

// Runtime is the unified model entry used by the API and worker processes.
type Runtime struct {
	mu          sync.Mutex
	provider    Provider
	recorder    Recorder
	s01Template string
	s02Template string
	stopped     map[string]bool
	reuse       map[string]MapTree
}

// New returns the production runtime. It has a template version and no provider.
func New() *Runtime {
	return &Runtime{
		s01Template: templateS01,
		s02Template: templateS02,
		reuse:       map[string]MapTree{},
	}
}

// SetProvider installs the adapter used by Call. Production main does not call this.
func (r *Runtime) SetProvider(provider Provider) {
	r.mu.Lock()
	r.provider = provider
	r.mu.Unlock()
}

// SetRecorder installs append-only usage storage. A nil recorder cannot succeed a call.
func (r *Runtime) SetRecorder(recorder Recorder) {
	r.mu.Lock()
	r.recorder = recorder
	r.mu.Unlock()
}

// HasRecorder reports whether usage rows can be stored.
func (r *Runtime) HasRecorder() bool {
	return r.currentRecorder() != nil
}

func (r *Runtime) setTemplate(scene, version string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if scene == SceneS02 {
		r.s02Template = version
		return
	}
	r.s01Template = version
}

// StopScene refuses later attempts for that scene and records them as failures.
func (r *Runtime) StopScene(scene string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.stopped == nil {
		r.stopped = map[string]bool{}
	}
	r.stopped[scene] = true
}

// Status reports the current provider state.
func (r *Runtime) Status() string {
	if r.currentProvider() == nil {
		return ErrProviderNotConfigured.Error()
	}
	return "ready"
}

// Call is the only model entry. S01 sends exactly one fragment.
// S02 sends a node title and no material bytes. No cache or price is applied here.
func (r *Runtime) Call(ctx context.Context, entry LedgerEntry) (CallResult, error) {
	if err := ctx.Err(); err != nil {
		return CallResult{}, err
	}
	if entry.SceneCode == "" {
		return CallResult{}, ErrSceneCodeRequired
	}
	provider := r.currentProvider()
	if provider == nil {
		return CallResult{}, ErrProviderNotConfigured
	}
	if entry.Subclass == "" {
		return CallResult{}, ErrSubclassRequired
	}
	req := ProviderRequest{
		SceneCode:       entry.SceneCode,
		Subclass:        entry.Subclass,
		Tier:            entry.Tier,
		TemplateVersion: entry.TemplateVersion,
		Media:           entry.Media,
		NodeTitle:       entry.NodeTitle,
		QuestionLimit:   entry.QuestionLimit,
	}
	if entry.SceneCode == SceneS02 {
		if entry.NodeTitle == "" || len(entry.Fragments) != 0 {
			return CallResult{}, ErrFragmentBoundary
		}
		res, err := provider.Complete(ctx, req)
		if err != nil {
			return CallResult{}, err
		}
		return CallResult{Tokens: res.Tokens, Questions: res.Questions}, nil
	}
	if len(entry.Fragments) != 1 {
		return CallResult{}, ErrFragmentBoundary
	}
	req.Fragment = append([]byte(nil), entry.Fragments[0]...)
	res, err := provider.Complete(ctx, req)
	if err != nil {
		return CallResult{}, err
	}
	return CallResult{Tokens: res.Tokens, Map: res.Map}, nil
}

func (r *Runtime) currentProvider() Provider {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.provider
}

func (r *Runtime) template(scene string) string {
	r.mu.Lock()
	defer r.mu.Unlock()
	if scene == SceneS02 {
		return r.s02Template
	}
	return r.s01Template
}

func (r *Runtime) isStopped(scene string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.stopped[scene]
}
