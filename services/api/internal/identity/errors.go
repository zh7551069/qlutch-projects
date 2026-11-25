package identity

import "net/http"

// Error is an API failure that uses a contract error code.
type Error struct {
	Code    string
	Message string
	Details map[string]any
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	return e.Code
}

var messages = map[string]string{
	"AUTH_REQUIRED":                  "需要登录",
	"AUTH_EXPIRED":                   "登录已失效",
	"AUTH_FORBIDDEN":                 "没有权限",
	"AUTH_WRONG_REALM":               "令牌域不匹配",
	"AUTH_ACCOUNT_NOT_FOUND":         "无此账号",
	"AUTH_ACCOUNT_EXISTS":            "账号已存在",
	"AUTH_OTP_INVALID":               "验证码不正确",
	"AUTH_OTP_EXPIRED":               "验证码已过期",
	"AUTH_OTP_MAX_ATTEMPTS":          "验证码尝试次数已用尽",
	"VALIDATION_INVALID_PHONE":       "手机号格式不正确",
	"VALIDATION_UNSUPPORTED_REGION":  "暂不支持该区号",
	"VALIDATION_UNSUPPORTED_LOCALE":  "暂不支持该语言",
	"VALIDATION_INVALID_PURPOSE":     "验证码用途不正确",
	"VALIDATION_INVALID_OTP_FORMAT":  "验证码需为 6 位数字",
	"VALIDATION_MISSING_FIELD":       "缺少必填字段",
	"VALIDATION_NICKNAME":            "昵称不正确",
	"VALIDATION_INVALID_STEP":        "未知的建档步骤",
	"VALIDATION_INVALID_STEP_ANSWER": "这一步的回答不正确",
	"VALIDATION_STEP_OUT_OF_ORDER":   "请先完成当前步骤",
	"IDEMPOTENCY_KEY_MISMATCH":       "幂等键与请求体不一致",
	"RATE_LIMITED":                   "操作过于频繁",
	"INTERNAL":                       "服务暂时不可用",
}

var statuses = map[string]int{
	"AUTH_REQUIRED":                  http.StatusUnauthorized,
	"AUTH_EXPIRED":                   http.StatusUnauthorized,
	"AUTH_OTP_INVALID":               http.StatusUnauthorized,
	"AUTH_OTP_EXPIRED":               http.StatusUnauthorized,
	"AUTH_FORBIDDEN":                 http.StatusForbidden,
	"AUTH_WRONG_REALM":               http.StatusForbidden,
	"AUTH_OTP_MAX_ATTEMPTS":          http.StatusForbidden,
	"AUTH_ACCOUNT_NOT_FOUND":         http.StatusNotFound,
	"AUTH_ACCOUNT_EXISTS":            http.StatusConflict,
	"IDEMPOTENCY_KEY_MISMATCH":       http.StatusConflict,
	"VALIDATION_NICKNAME":            http.StatusUnprocessableEntity,
	"VALIDATION_INVALID_STEP_ANSWER": http.StatusUnprocessableEntity,
	"VALIDATION_STEP_OUT_OF_ORDER":   http.StatusUnprocessableEntity,
	"VALIDATION_INVALID_PHONE":       http.StatusBadRequest,
	"VALIDATION_UNSUPPORTED_REGION":  http.StatusBadRequest,
	"VALIDATION_UNSUPPORTED_LOCALE":  http.StatusBadRequest,
	"VALIDATION_INVALID_PURPOSE":     http.StatusBadRequest,
	"VALIDATION_INVALID_OTP_FORMAT":  http.StatusBadRequest,
	"VALIDATION_MISSING_FIELD":       http.StatusBadRequest,
	"VALIDATION_INVALID_STEP":        http.StatusBadRequest,
	"RATE_LIMITED":                   http.StatusTooManyRequests,
	"INTERNAL":                       http.StatusInternalServerError,
}

func codeErr(code string) *Error {
	return &Error{Code: code, Message: messages[code]}
}

func rateErr(seconds int) *Error {
	if seconds < 1 {
		seconds = 1
	}
	return &Error{
		Code:    "RATE_LIMITED",
		Message: messages["RATE_LIMITED"],
		Details: map[string]any{"retry_after_seconds": seconds},
	}
}

func statusOf(code string) int {
	if status, ok := statuses[code]; ok {
		return status
	}
	return http.StatusInternalServerError
}
