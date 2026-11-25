package knowledge

import (
	"context"
	"sync"
	"time"
)

type memoryStore struct {
	mu        sync.Mutex
	packs     map[string]*pack
	materials map[string]*material
	jobs      map[string]*job
	skips     map[string]time.Time
	idem      map[string]*idemRow
}

func newMemoryStore() *memoryStore {
	return &memoryStore{
		packs:     map[string]*pack{},
		materials: map[string]*material{},
		jobs:      map[string]*job{},
		skips:     map[string]time.Time{},
		idem:      map[string]*idemRow{},
	}
}

func (m *memoryStore) InsertPack(_ context.Context, item pack) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.packs[item.ID]; ok {
		return errConflict
	}
	m.packs[item.ID] = clonePack(&item)
	return nil
}

func (m *memoryStore) FindPack(_ context.Context, id string) (*pack, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return clonePack(m.packs[id]), nil
}

func (m *memoryStore) InsertMaterial(_ context.Context, item material) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.materials[item.ID]; ok {
		return errConflict
	}
	if m.packs[item.PackID] == nil {
		return errNotFound
	}
	m.materials[item.ID] = cloneMaterial(&item)
	return nil
}

func (m *memoryStore) FindMaterial(_ context.Context, id string) (*material, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return cloneMaterial(m.materials[id]), nil
}

func (m *memoryStore) AcceptJob(_ context.Context, item job) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	owner := m.packs[item.PackID]
	if owner == nil || owner.UserID != item.UserID {
		return errNotFound
	}
	mat := m.materials[item.MaterialID]
	if mat == nil || mat.UserID != item.UserID || mat.PackID != item.PackID {
		return errNotFound
	}
	for _, existing := range m.jobs {
		if existing.PackID == item.PackID && activeStatus(existing.Status) {
			return errConflict
		}
	}
	if _, ok := m.jobs[item.ID]; ok {
		return errConflict
	}
	m.jobs[item.ID] = cloneJob(&item)
	return nil
}

func (m *memoryStore) DeleteJob(_ context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.jobs, id)
	return nil
}

func (m *memoryStore) FindJob(_ context.Context, id string) (*job, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return cloneJob(m.jobs[id]), nil
}

func (m *memoryStore) UpdateJobStatus(_ context.Context, id, status string) (*job, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	item := m.jobs[id]
	if item == nil {
		return nil, errNotFound
	}
	if activeStatus(status) {
		for _, existing := range m.jobs {
			if existing.ID != id && existing.PackID == item.PackID && activeStatus(existing.Status) {
				return nil, errConflict
			}
		}
	}
	item.Status = status
	return cloneJob(item), nil
}

func (m *memoryStore) CountJobs(_ context.Context, userID string) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	n := 0
	for _, item := range m.jobs {
		if item.UserID == userID {
			n++
		}
	}
	return n, nil
}

func (m *memoryStore) RecordSkip(_ context.Context, userID string, at time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.skips[userID]; ok {
		return nil
	}
	m.skips[userID] = at
	return nil
}

func (m *memoryStore) SkipRecorded(_ context.Context, userID string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.skips[userID]
	return ok, nil
}

func (m *memoryStore) SaveIdempotency(_ context.Context, userID, op, key string, row idemRow) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.idem[idemKey(userID, op, key)] = cloneIdem(&row)
	return nil
}

func (m *memoryStore) FindIdempotency(_ context.Context, userID, op, key string, now time.Time) (*idemRow, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	id := idemKey(userID, op, key)
	row := m.idem[id]
	if row == nil {
		return nil, nil
	}
	if !row.Expires.After(now) {
		delete(m.idem, id)
		return nil, nil
	}
	return cloneIdem(row), nil
}

func idemKey(userID, op, key string) string { return userID + "\n" + op + "\n" + key }

func activeStatus(status string) bool {
	return status == statusAccepted || status == statusRunning
}

func clonePack(in *pack) *pack {
	if in == nil {
		return nil
	}
	out := *in
	return &out
}

func cloneMaterial(in *material) *material {
	if in == nil {
		return nil
	}
	out := *in
	return &out
}

func cloneJob(in *job) *job {
	if in == nil {
		return nil
	}
	out := *in
	return &out
}

func cloneIdem(in *idemRow) *idemRow {
	if in == nil {
		return nil
	}
	out := *in
	if in.ErrDetails != nil {
		out.ErrDetails = map[string]any{}
		for key, value := range in.ErrDetails {
			out.ErrDetails[key] = value
		}
	}
	return &out
}
