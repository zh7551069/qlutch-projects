package knowledge

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
	"AUTH_REQUIRED":                    "需要登录",
	"AUTH_EXPIRED":                     "登录已失效",
	"AUTH_FORBIDDEN":                   "没有权限",
	"AUTH_WRONG_REALM":                 "令牌域不匹配",
	"VALIDATION_MISSING_FIELD":         "缺少必填字段",
	"VALIDATION_INVALID_SOURCE":        "导入来源不正确",
	"VALIDATION_UNSUPPORTED_MEDIA":     "不支持该文件类型",
	"VALIDATION_MEDIA_TOO_LARGE":       "文件超过大小限制",
	"VALIDATION_MATERIAL_NOT_UPLOADED": "材料尚未上传完成",
	"NOT_FOUND":                        "找不到对应内容",
	"CONFLICT":                         "已有进行中的导入",
	"IDEMPOTENCY_KEY_MISMATCH":         "幂等键与请求体不一致",
	"RATE_LIMITED":                     "操作过于频繁",
	"INTERNAL":                         "服务暂时不可用",
}

var statuses = map[string]int{
	"AUTH_REQUIRED":                    http.StatusUnauthorized,
	"AUTH_EXPIRED":                     http.StatusUnauthorized,
	"AUTH_FORBIDDEN":                   http.StatusForbidden,
	"AUTH_WRONG_REALM":                 http.StatusForbidden,
	"VALIDATION_MISSING_FIELD":         http.StatusBadRequest,
	"VALIDATION_INVALID_SOURCE":        http.StatusBadRequest,
	"VALIDATION_UNSUPPORTED_MEDIA":     http.StatusBadRequest,
	"VALIDATION_MEDIA_TOO_LARGE":       http.StatusBadRequest,
	"VALIDATION_MATERIAL_NOT_UPLOADED": http.StatusUnprocessableEntity,
	"NOT_FOUND":                        http.StatusNotFound,
	"CONFLICT":                         http.StatusConflict,
	"IDEMPOTENCY_KEY_MISMATCH":         http.StatusConflict,
	"RATE_LIMITED":                     http.StatusTooManyRequests,
	"INTERNAL":                         http.StatusInternalServerError,
}

func codeErr(code string) *Error {
	return &Error{Code: code, Message: messages[code]}
}

func tooLarge(maxBytes int64) *Error {
	return &Error{
		Code:    "VALIDATION_MEDIA_TOO_LARGE",
		Message: messages["VALIDATION_MEDIA_TOO_LARGE"],
		Details: map[string]any{"max_bytes": maxBytes},
	}
}

func statusOf(code string) int {
	if status, ok := statuses[code]; ok {
		return status
	}
	return http.StatusInternalServerError
}
