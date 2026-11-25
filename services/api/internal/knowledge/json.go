package knowledge

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"unicode/utf8"
)

func decodeObject(raw []byte, dest any) *Error {
	if len(raw) > maxBodyBytes {
		return codeErr("VALIDATION_MISSING_FIELD")
	}
	dec := json.NewDecoder(strings.NewReader(string(raw)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dest); err != nil {
		return codeErr("VALIDATION_MISSING_FIELD")
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return codeErr("VALIDATION_MISSING_FIELD")
	}
	return nil
}

func canonicalJSON(raw []byte) ([]byte, error) {
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" || len(raw) > maxBodyBytes {
		return nil, errors.New("body")
	}
	dec := json.NewDecoder(strings.NewReader(trimmed))
	dec.UseNumber()
	var value any
	if err := dec.Decode(&value); err != nil {
		return nil, err
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return nil, err
	}
	if _, ok := value.(map[string]any); !ok {
		return nil, errors.New("object")
	}
	return json.Marshal(value)
}

func bodyHash(canon []byte) [32]byte {
	return sha256.Sum256(canon)
}

func optionalObject(raw []byte) []byte {
	if strings.TrimSpace(string(raw)) == "" {
		return []byte("{}")
	}
	return raw
}

func validIdempotencyKey(key string) bool {
	if key == "" || len(key) > maxIdempotencyKey {
		return false
	}
	for _, r := range key {
		if r < 0x21 || r > 0x7e {
			return false
		}
	}
	return true
}

func validFilename(name string) bool {
	if name == "" || utf8.RuneCountInString(name) > maxFilenameRunes {
		return false
	}
	if strings.ContainsAny(name, `/\\`) || strings.Contains(name, "..") {
		return false
	}
	for _, r := range name {
		if r < 0x20 || r == 0x7f {
			return false
		}
	}
	return true
}

func knownStatus(status string) bool {
	switch status {
	case statusAccepted, statusRunning, statusMapReady, statusFailedRetryable, statusFailedTerminal:
		return true
	default:
		return false
	}
}
