package knowledge

import "time"

// Process defaults. A4-O2 is still open. These are not locked product numbers.
// See docs/contracts/knowledge-a4.md.
const (
	DefaultPresignSeconds int64 = 900
	DefaultMaxBytes       int64 = 32 << 20
	DefaultIdempotencyTTL       = 24 * time.Hour
)

const (
	maxIdempotencyKey = 200
	maxBodyBytes      = 1 << 16
	maxFilenameRunes  = 255
)

const (
	statusAccepted        = "accepted"
	statusRunning         = "running"
	statusMapReady        = "map_ready"
	statusFailedRetryable = "failed_retryable"
	statusFailedTerminal  = "failed_terminal"
	opImport              = "import"
	opSkip                = "skip"
)
