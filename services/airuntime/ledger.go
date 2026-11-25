package airuntime

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

// Entry is one usage row. It has no prompt, source text, or unit price.
type Entry struct {
	ID              string
	SceneCode       string
	Subclass        string
	Tier            string
	CacheHit        bool
	Success         bool
	RetryIndex      int
	Tokens          int
	MetricKind      string
	MetricValue     int
	PackID          string
	JobID           string
	TemplateVersion string
	Caller          string
	Fingerprint     string
	At              time.Time
}

// Query filters ledger rows. Empty fields are not filtered.
// CacheHit is applied only when CacheHitSet is true.
type Query struct {
	SceneCode   string
	Subclass    string
	JobID       string
	CacheHit    bool
	CacheHitSet bool
}

// Recorder stores usage rows. Append does not update an existing row.
type Recorder interface {
	Append(ctx context.Context, entry Entry) error
	List(ctx context.Context, query Query) ([]Entry, error)
}

// MemoryRecorder keeps rows in this process. Tests and single-process runs use it.
type MemoryRecorder struct {
	mu   sync.Mutex
	rows []Entry
}

// NewMemoryRecorder returns an empty ledger.
func NewMemoryRecorder() *MemoryRecorder {
	return &MemoryRecorder{}
}

// Append implements Recorder.
func (m *MemoryRecorder) Append(_ context.Context, entry Entry) error {
	if err := validateEntry(entry); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rows = append(m.rows, entry)
	return nil
}

// List implements Recorder.
func (m *MemoryRecorder) List(_ context.Context, query Query) ([]Entry, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]Entry, 0, len(m.rows))
	for _, row := range m.rows {
		if !matchEntry(row, query) {
			continue
		}
		out = append(out, row)
	}
	return out, nil
}

func (r *Runtime) append(ctx context.Context, entry Entry) error {
	if entry.Caller == "" {
		entry.Caller = CallerLearner
	}
	if entry.At.IsZero() {
		entry.At = time.Now().UTC()
	}
	if entry.ID == "" {
		entry.ID = newEntryID()
	}
	if err := validateEntry(entry); err != nil {
		return err
	}
	recorder := r.currentRecorder()
	if recorder == nil {
		return ErrLedger
	}
	if err := recorder.Append(ctx, entry); err != nil {
		return ErrLedger
	}
	return nil
}

func (r *Runtime) currentRecorder() Recorder {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.recorder
}

func (r *Runtime) nextIndex(ctx context.Context, jobID, scene string) (int, error) {
	recorder := r.currentRecorder()
	if recorder == nil {
		return 0, ErrLedger
	}
	rows, err := recorder.List(ctx, Query{JobID: jobID, SceneCode: scene})
	if err != nil {
		return 0, ErrLedger
	}
	return len(rows), nil
}

func validateEntry(entry Entry) error {
	if entry.SceneCode != SceneS01 && entry.SceneCode != SceneS02 {
		return ErrUsage
	}
	if entry.Subclass == "" || entry.Subclass == "其他" {
		return ErrUsage
	}
	if entry.Caller != CallerLearner {
		return ErrUsage
	}
	if entry.Tier == "" || entry.PackID == "" || entry.JobID == "" {
		return ErrUsage
	}
	if entry.RetryIndex < 0 || entry.Tokens < 0 || entry.MetricValue < 0 {
		return ErrUsage
	}
	return nil
}

func matchEntry(row Entry, query Query) bool {
	if query.SceneCode != "" && row.SceneCode != query.SceneCode {
		return false
	}
	if query.Subclass != "" && row.Subclass != query.Subclass {
		return false
	}
	if query.JobID != "" && row.JobID != query.JobID {
		return false
	}
	if query.CacheHitSet && row.CacheHit != query.CacheHit {
		return false
	}
	return true
}

func newEntryID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "led_unknown"
	}
	return "led_" + hex.EncodeToString(buf)
}
