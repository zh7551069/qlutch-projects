package airuntime

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"unicode/utf8"
)

const (
	maxTextFragment     = 4000
	maxDocumentFragment = 4096
	mediaText           = "text"
	mediaDocument       = "document"
	mediaImage          = "image"
	mediaAudio          = "audio"
	metricFragments     = "fragments"
	metricPages         = "pages"
	metricBytes         = "bytes"
)

func route(contentType string) (subclass, tier, media string, ok bool) {
	switch contentType {
	case "image/jpeg", "image/png", "image/webp", "image/heic", "image/heif":
		return SubclassOCR, TierRecognize, mediaImage, true
	case "audio/mpeg", "audio/mp4", "audio/x-m4a", "audio/aac", "audio/wav", "audio/x-wav", "audio/wave":
		return SubclassASR, TierRecognize, mediaAudio, true
	case "text/plain":
		return SubclassText, TierLight, mediaText, true
	case "application/pdf",
		"application/vnd.ms-powerpoint",
		"application/vnd.openxmlformats-officedocument.presentationml.presentation",
		"application/msword",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document":
		return SubclassText, TierStandard, mediaDocument, true
	default:
		return "", "", "", false
	}
}

func metricKind(media string) string {
	switch media {
	case mediaImage:
		return metricPages
	case mediaAudio:
		return metricBytes
	default:
		return metricFragments
	}
}

func fingerprint(body []byte) string {
	if len(body) == 0 {
		return ""
	}
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}

func split(media string, body []byte) [][]byte {
	switch media {
	case mediaText:
		return splitText(body)
	case mediaDocument:
		return splitBytes(body, maxDocumentFragment)
	case mediaImage, mediaAudio:
		return splitBytes(body, len(body))
	default:
		return nil
	}
}

func splitText(body []byte) [][]byte {
	text := strings.ReplaceAll(string(body), "\r\n", "\n")
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	blocks := strings.Split(text, "\n\n")
	var parts [][]byte
	for _, block := range blocks {
		block = strings.TrimSpace(block)
		if block == "" {
			continue
		}
		parts = append(parts, splitLong(block, maxTextFragment)...)
	}
	return parts
}

func splitLong(text string, limit int) [][]byte {
	raw := []byte(text)
	if len(raw) <= limit {
		return [][]byte{append([]byte(nil), raw...)}
	}
	return splitBytes(raw, limit)
}

func splitBytes(body []byte, size int) [][]byte {
	if len(body) == 0 || size < 1 {
		return nil
	}
	var parts [][]byte
	for start := 0; start < len(body); start += size {
		end := start + size
		if end > len(body) {
			end = len(body)
		}
		part := make([]byte, end-start)
		copy(part, body[start:end])
		parts = append(parts, part)
	}
	return parts
}

func fragmentsStayApart(body []byte, parts [][]byte) bool {
	if len(parts) < 2 {
		return true
	}
	for _, part := range parts {
		if bytes.Equal(part, body) {
			return false
		}
	}
	return true
}

func mergeTree(dst, src MapTree) MapTree {
	dst.Catalogs = append(dst.Catalogs, src.Catalogs...)
	return dst
}

func mapEnterable(tree MapTree) bool {
	for _, catalog := range tree.Catalogs {
		if badTitle(catalog.Title) {
			continue
		}
		for _, chapter := range catalog.Chapters {
			if badTitle(chapter.Title) {
				continue
			}
			for _, point := range chapter.Points {
				if !badTitle(point.Title) {
					return true
				}
			}
		}
	}
	return false
}

func badTitle(title string) bool {
	title = strings.TrimSpace(title)
	if title == "" || strings.ContainsAny(title, "\r\n") {
		return true
	}
	return utf8.RuneCountInString(title) > maxTitleRunes
}

func leaksSource(tree MapTree, parts [][]byte) bool {
	var titles []string
	for _, catalog := range tree.Catalogs {
		titles = append(titles, catalog.Title)
		for _, chapter := range catalog.Chapters {
			titles = append(titles, chapter.Title)
			for _, point := range chapter.Points {
				titles = append(titles, point.Title)
			}
		}
	}
	for _, title := range titles {
		for _, part := range parts {
			if len(part) > 24 && strings.Contains(title, string(part)) {
				return true
			}
		}
	}
	return false
}
