package knowledge

// Allowed content types for this slice. A patch is required to add another.
var mediaBySource = map[string]map[string]struct{}{
	"camera":     setOf("image/jpeg", "image/png", "image/webp", "image/heic", "image/heif"),
	"paste_note": setOf("text/plain"),
	"file": setOf(
		"application/pdf",
		"application/vnd.ms-powerpoint",
		"application/vnd.openxmlformats-officedocument.presentationml.presentation",
		"application/msword",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"image/jpeg", "image/png", "image/webp", "image/heic", "image/heif",
		"audio/mpeg", "audio/mp4", "audio/x-m4a", "audio/aac", "audio/wav", "audio/x-wav", "audio/wave",
	),
}

func setOf(values ...string) map[string]struct{} {
	out := make(map[string]struct{}, len(values))
	for _, value := range values {
		out[value] = struct{}{}
	}
	return out
}

func mediaError(source, contentType string) *Error {
	if source == "" || contentType == "" {
		return codeErr("VALIDATION_MISSING_FIELD")
	}
	allowed, ok := mediaBySource[source]
	if !ok {
		return codeErr("VALIDATION_INVALID_SOURCE")
	}
	if _, ok := allowed[contentType]; !ok {
		return codeErr("VALIDATION_UNSUPPORTED_MEDIA")
	}
	return nil
}
