package identity

import "time"

// Documented defaults. These are implementation choices, not contract locks.
// See docs/contracts/identity-a2.md.
const (
	DefaultOTPTTL           = 5 * time.Minute
	DefaultOTPMaxAttempts   = 5
	DefaultSendCooldown     = 60 * time.Second
	DefaultSendBudget       = 5
	DefaultSendWindow       = 10 * time.Minute
	DefaultSendIPBudget     = 30
	DefaultVerifyBudget     = 10
	DefaultVerifyWindow     = 10 * time.Minute
	DefaultVerifyIPBudget   = 40
	DefaultAccessTTL        = 15 * time.Minute
	DefaultRefreshTTL       = 30 * 24 * time.Hour
	DefaultIdempotencyTTL   = 24 * time.Hour
	DefaultNicknameMaxRunes = 32
	DefaultAnswerMaxBytes   = 4096
	DefaultPhoneMaxDigits   = 32
	DefaultIdempotencyKey   = 200
	maxBodyBytes            = 1 << 16
)

const (
	prefixAccess   = "qlu_a_"
	prefixRefresh  = "qlu_r_"
	prefixInternal = "qlu_i_"
	defaultLocale  = "zh-Hans"
)

var stepOrder = []string{
	"nickname",
	"age",
	"study_or_work",
	"identity",
	"how_to_address",
	"exam_track",
	"daily_time",
}
