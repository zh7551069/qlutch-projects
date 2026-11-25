package knowledge

import (
	"context"
	"errors"
	"time"
)

var (
	errConflict = errors.New("conflict")
	errNotFound = errors.New("not found")
	errStorage  = errors.New("object storage unavailable")
	errQueue    = errors.New("queue unavailable")
)

type pack struct {
	ID     string
	UserID string
}

type material struct {
	ID          string
	PackID      string
	UserID      string
	Source      string
	ContentType string
	ByteSize    int64
	Filename    string
	ObjectKey   string
}

type job struct {
	ID              string
	PackID          string
	MaterialID      string
	UserID          string
	Source          string
	Status          string
	ProgressMessage string
	FailureCode     string
	FailureMessage  string
}

type mapNode struct {
	ID       string
	PackID   string
	JobID    string
	ParentID string
	Kind     string
	Title    string
	Position int
}

type prefetchRow struct {
	ID       string
	PackID   string
	JobID    string
	NodeID   string
	Stem     string
	Position int
}

type idemRow struct {
	Hash       [32]byte
	JobID      string
	ErrCode    string
	ErrDetails map[string]any
	Expires    time.Time
}

// Store keeps packs, material metadata, jobs, skip records, and idempotency keys.
// Material bytes are not stored here.
type Store interface {
	InsertPack(ctx context.Context, item pack) error
	FindPack(ctx context.Context, id string) (*pack, error)
	InsertMaterial(ctx context.Context, item material) error
	FindMaterial(ctx context.Context, id string) (*material, error)
	AcceptJob(ctx context.Context, item job) error
	DeleteJob(ctx context.Context, id string) error
	FindJob(ctx context.Context, id string) (*job, error)
	UpdateJobStatus(ctx context.Context, id, status string) (*job, error)
	CountJobs(ctx context.Context, userID string) (int, error)
	RecordSkip(ctx context.Context, userID string, at time.Time) error
	SkipRecorded(ctx context.Context, userID string) (bool, error)
	SaveIdempotency(ctx context.Context, userID, op, key string, row idemRow) error
	FindIdempotency(ctx context.Context, userID, op, key string, now time.Time) (*idemRow, error)
	ClaimJob(ctx context.Context, id string) (item *job, settled bool, err error)
	FailJob(ctx context.Context, id, status, message string) error
	CommitMapReady(ctx context.Context, jobID, userID string, nodes []mapNode) error
	MapNodes(ctx context.Context, packID string) ([]mapNode, error)
	InsertPrefetch(ctx context.Context, rows []prefetchRow) error
	CountPrefetch(ctx context.Context, jobID string) (int, error)
}

// Blobs is an S3-compatible object store. Implementations must not log secrets or object bytes.
type Blobs interface {
	PresignPut(ctx context.Context, objectKey, contentType string) (uploadURL string, headers map[string]string, expiresIn int64, err error)
	Stat(ctx context.Context, objectKey string) (size int64, found bool, err error)
	Read(ctx context.Context, objectKey string) (body []byte, found bool, err error)
}

// Enqueuer accepts an opaque job payload. The payload is a job id, not material bytes.
type Enqueuer interface {
	Enqueue(ctx context.Context, payload []byte) error
}
