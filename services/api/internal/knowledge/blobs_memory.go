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
	bodies  map[string][]byte
	Expires int64
}

// NewMemoryBlobs returns an empty blob stand-in.
func NewMemoryBlobs() *MemoryBlobs {
	return &MemoryBlobs{objects: map[string]int64{}, bodies: map[string][]byte{}, Expires: DefaultPresignSeconds}
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

// Remove drops an object. Tests use it. It is not an HTTP route.
func (m *MemoryBlobs) Remove(objectKey string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.objects, objectKey)
	delete(m.bodies, objectKey)
}

// Put stores object bytes for a worker test. It is not an HTTP route.
func (m *MemoryBlobs) Put(objectKey string, body []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	copied := append([]byte(nil), body...)
	m.objects[objectKey] = int64(len(copied))
	m.bodies[objectKey] = copied
}

// Stat implements Blobs.
func (m *MemoryBlobs) Stat(_ context.Context, objectKey string) (int64, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	size, ok := m.objects[objectKey]
	return size, ok, nil
}

// Read implements Blobs. A size-only object is an unread storage error, not a missing key.
func (m *MemoryBlobs) Read(_ context.Context, objectKey string) ([]byte, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	body, ok := m.bodies[objectKey]
	if !ok {
		if _, sized := m.objects[objectKey]; sized {
			return nil, false, errStorage
		}
		return nil, false, nil
	}
	return append([]byte(nil), body...), true, nil
}
