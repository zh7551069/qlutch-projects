package knowledge

import (
	"context"
	"strings"
	"sync"
)

const memoryUploadPrefix = "http://127.0.0.1/upload/"

// MemoryBlobs records uploaded sizes in this process. Tests use it. It is not an HTTP route.
type MemoryBlobs struct {
	mu      sync.Mutex
	objects map[string]int64
	Expires int64
}

// NewMemoryBlobs returns an empty blob stand-in.
func NewMemoryBlobs() *MemoryBlobs {
	return &MemoryBlobs{objects: map[string]int64{}, Expires: DefaultPresignSeconds}
}

// PresignPut implements Blobs. The URL carries the object key so a test can complete the upload.
func (m *MemoryBlobs) PresignPut(_ context.Context, objectKey, contentType string) (string, map[string]string, int64, error) {
	if objectKey == "" || contentType == "" {
		return "", nil, 0, errStorage
	}
	expires := m.Expires
	if expires < 1 {
		expires = DefaultPresignSeconds
	}
	return memoryUploadPrefix + objectKey, map[string]string{"Content-Type": contentType}, expires, nil
}

// Complete records a successful PUT of size bytes for an upload URL from this stand-in.
func (m *MemoryBlobs) Complete(uploadURL string, size int64) {
	key := strings.TrimPrefix(uploadURL, memoryUploadPrefix)
	m.mu.Lock()
	defer m.mu.Unlock()
	m.objects[key] = size
}

// Stat implements Blobs.
func (m *MemoryBlobs) Stat(_ context.Context, objectKey string) (int64, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	size, ok := m.objects[objectKey]
	return size, ok, nil
}
