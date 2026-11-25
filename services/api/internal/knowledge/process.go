package knowledge

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"strings"

	"github.com/zh7551069/qlutch-projects/services/airuntime"
)

const (
	msgRetryable = "地图暂时没有生成，可以用同一份材料再试一次"
	msgTerminal  = "这份材料已经无法读取，请重新上传"
)

// SetRuntime installs the airuntime entry used by ProcessImport.
// It does not mark a job map_ready. Production main leaves the provider unset.
func (s *Service) SetRuntime(rt *airuntime.Runtime) {
	s.runtime = rt
}

// ProcessQueued reads the import payload enqueued by acceptance.
// The payload is a job id. Material bytes are not in the queue.
func (s *Service) ProcessQueued(ctx context.Context, payload []byte) error {
	var queued queuedJob
	if err := decodeQueued(payload, &queued); err != nil || queued.JobID == "" {
		return nil
	}
	return s.ProcessImport(ctx, queued.JobID)
}

// ProcessImport moves one accepted job through S01.
// The map is committed before status becomes map_ready. S02 runs only after that.
func (s *Service) ProcessImport(ctx context.Context, jobID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if s.runtime == nil {
		return errors.New("airuntime missing")
	}
	item, settled, err := s.store.ClaimJob(ctx, jobID)
	if err != nil || item == nil {
		return err
	}
	if settled {
		if item.Status != statusMapReady {
			return nil
		}
		if err := s.identity.CompleteFirstPack(ctx, item.UserID); err != nil {
			return err
		}
		return s.ensurePrefetch(ctx, item)
	}
	mat, err := s.store.FindMaterial(ctx, item.MaterialID)
	if err != nil {
		return err
	}
	if mat == nil {
		return s.settleUnread(ctx, item, nil, nil, true)
	}
	body, found, readErr := s.blobs.Read(ctx, mat.ObjectKey)
	if readErr != nil {
		return s.settleUnread(ctx, item, mat, nil, false)
	}
	if !found || len(body) == 0 || int64(len(body)) != mat.ByteSize {
		return s.settleUnread(ctx, item, mat, body, true)
	}
	out, err := s.runtime.RunS01(ctx, airuntime.S01Input{
		ContentType: mat.ContentType,
		Body:        body,
		PackID:      item.PackID,
		JobID:       item.ID,
	})
	if err != nil {
		if errors.Is(err, airuntime.ErrLedger) {
			return err
		}
		return s.store.FailJob(ctx, item.ID, statusFailedRetryable, msgRetryable)
	}
	nodes := flattenMap(out.Map, item)
	if !enterableNodes(nodes) {
		return s.store.FailJob(ctx, item.ID, statusFailedRetryable, msgRetryable)
	}
	if err := s.store.CommitMapReady(ctx, item.ID, item.UserID, nodes); err != nil {
		return err
	}
	if err := s.identity.CompleteFirstPack(ctx, item.UserID); err != nil {
		return err
	}
	return s.runPrefetch(ctx, item, nodes)
}

func (s *Service) settleUnread(ctx context.Context, item *job, mat *material, body []byte, terminal bool) error {
	in := airuntime.S01Input{PackID: item.PackID, JobID: item.ID}
	if mat != nil {
		in.ContentType = mat.ContentType
		in.Body = body
	}
	if mat != nil {
		if err := s.runtime.FailS01(ctx, in); err != nil && !errors.Is(err, airuntime.ErrSubclassRequired) {
			return err
		}
	}
	status := statusFailedRetryable
	message := msgRetryable
	if terminal {
		status = statusFailedTerminal
		message = msgTerminal
	}
	return s.store.FailJob(ctx, item.ID, status, message)
}

func (s *Service) ensurePrefetch(ctx context.Context, item *job) error {
	count, err := s.store.CountPrefetch(ctx, item.ID)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	nodes, err := s.store.MapNodes(ctx, item.PackID)
	if err != nil {
		return err
	}
	return s.runPrefetch(ctx, item, filterJob(nodes, item.ID))
}

func (s *Service) runPrefetch(ctx context.Context, item *job, nodes []mapNode) error {
	selected := selectPoints(nodes)
	if len(selected) == 0 {
		return nil
	}
	out, err := s.runtime.RunS02(ctx, airuntime.S02Input{
		Nodes:  selected,
		PackID: item.PackID,
		JobID:  item.ID,
	})
	if err != nil || len(out.Questions) == 0 {
		return nil
	}
	rows := make([]prefetchRow, 0, len(out.Questions))
	for i, question := range out.Questions {
		if question.Stem == "" || question.NodeID == "" {
			continue
		}
		rows = append(rows, prefetchRow{
			ID:       newID("pre_"),
			PackID:   item.PackID,
			JobID:    item.ID,
			NodeID:   question.NodeID,
			Stem:     question.Stem,
			Position: i,
		})
	}
	if len(rows) == 0 {
		return nil
	}
	_ = s.store.InsertPrefetch(ctx, rows)
	return nil
}

// TestingMapEnterable reports whether the pack has a catalog, chapter, and point.
// It is not an HTTP route and does not change status.
func (s *Service) TestingMapEnterable(ctx context.Context, packID string) (bool, error) {
	nodes, err := s.store.MapNodes(ctx, packID)
	if err != nil {
		return false, err
	}
	return enterableNodes(nodes), nil
}

// TestingPrefetchCount returns stored S02 stems for a job. It is not an HTTP route.
func (s *Service) TestingPrefetchCount(ctx context.Context, jobID string) (int, error) {
	return s.store.CountPrefetch(ctx, jobID)
}

// TestingMaterialKey returns the server-side object key. It is not an HTTP route.
func (s *Service) TestingMaterialKey(ctx context.Context, materialID string) (string, error) {
	mat, err := s.store.FindMaterial(ctx, materialID)
	if err != nil || mat == nil {
		return "", err
	}
	return mat.ObjectKey, nil
}

func decodeQueued(payload []byte, dest *queuedJob) error {
	if len(payload) == 0 || len(payload) > maxBodyBytes {
		return errors.New("payload")
	}
	return json.Unmarshal(payload, dest)
}

func flattenMap(tree airuntime.MapTree, item *job) []mapNode {
	var nodes []mapNode
	position := 0
	for _, catalog := range tree.Catalogs {
		catalogID := newID("node_")
		nodes = append(nodes, mapNode{
			ID: catalogID, PackID: item.PackID, JobID: item.ID,
			Kind: "catalog", Title: strings.TrimSpace(catalog.Title), Position: position,
		})
		position++
		for _, chapter := range catalog.Chapters {
			chapterID := newID("node_")
			nodes = append(nodes, mapNode{
				ID: chapterID, PackID: item.PackID, JobID: item.ID, ParentID: catalogID,
				Kind: "chapter", Title: strings.TrimSpace(chapter.Title), Position: position,
			})
			position++
			for _, point := range chapter.Points {
				nodes = append(nodes, mapNode{
					ID: newID("node_"), PackID: item.PackID, JobID: item.ID, ParentID: chapterID,
					Kind: "point", Title: strings.TrimSpace(point.Title), Position: position,
				})
				position++
			}
		}
	}
	return nodes
}

func enterableNodes(nodes []mapNode) bool {
	children := map[string][]mapNode{}
	var catalogs []mapNode
	for _, node := range nodes {
		title := strings.TrimSpace(node.Title)
		if title == "" || strings.ContainsAny(title, "\r\n") || len([]rune(title)) > 80 {
			return false
		}
		switch node.Kind {
		case "catalog":
			catalogs = append(catalogs, node)
		case "chapter", "point":
		default:
			return false
		}
		children[node.ParentID] = append(children[node.ParentID], node)
	}
	for _, catalog := range catalogs {
		for _, chapter := range children[catalog.ID] {
			if chapter.Kind != "chapter" {
				continue
			}
			for _, point := range children[chapter.ID] {
				if point.Kind == "point" {
					return true
				}
			}
		}
	}
	return false
}

// selectPoints keeps the current point and the next one.
// A5-O2 is open: this is not a locked batch size, and it is not a full map.
func selectPoints(nodes []mapNode) []airuntime.NodeRef {
	var points []mapNode
	for _, node := range nodes {
		if node.Kind == "point" {
			points = append(points, node)
		}
	}
	sort.Slice(points, func(i, j int) bool {
		if points[i].Position == points[j].Position {
			return points[i].ID < points[j].ID
		}
		return points[i].Position < points[j].Position
	})
	var out []airuntime.NodeRef
	for _, point := range points {
		if len(out) == 2 {
			break
		}
		out = append(out, airuntime.NodeRef{ID: point.ID, Title: point.Title})
	}
	return out
}

func filterJob(nodes []mapNode, jobID string) []mapNode {
	var out []mapNode
	for _, node := range nodes {
		if node.JobID == jobID {
			out = append(out, node)
		}
	}
	return out
}
