package identity

import (
	"context"
	"errors"
	"sync"
)

// memoryStore is the in-process account store used when PostgreSQL is not configured.
type memoryStore struct {
	mu           sync.Mutex
	byID         map[string]*user
	byPhone      map[string]*user
	sessions     map[string]*session
	accessIndex  map[[32]byte]string
	refreshIndex map[[32]byte]string
}

func newMemoryStore() *memoryStore {
	return &memoryStore{
		byID:         map[string]*user{},
		byPhone:      map[string]*user{},
		sessions:     map[string]*session{},
		accessIndex:  map[[32]byte]string{},
		refreshIndex: map[[32]byte]string{},
	}
}

func (m *memoryStore) FindByPhone(_ context.Context, region, national string) (*user, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return cloneUser(m.byPhone[phoneKey(region, national)]), nil
}

func (m *memoryStore) FindByID(_ context.Context, id string) (*user, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return cloneUser(m.byID[id]), nil
}

func (m *memoryStore) Insert(_ context.Context, account *user) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := phoneKey(account.PhoneRegion, account.PhoneNational)
	if _, ok := m.byPhone[key]; ok {
		return errAccountExists
	}
	if _, ok := m.byID[account.ID]; ok {
		return errAccountExists
	}
	cloned := cloneUser(account)
	m.byID[account.ID] = cloned
	m.byPhone[key] = cloned
	return nil
}

func (m *memoryStore) DeleteAccount(_ context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	account := m.byID[id]
	if account != nil {
		delete(m.byPhone, phoneKey(account.PhoneRegion, account.PhoneNational))
		delete(m.byID, id)
	}
	for sid, sess := range m.sessions {
		if sess.UserID != id {
			continue
		}
		delete(m.accessIndex, sess.AccessHash)
		delete(m.refreshIndex, sess.RefreshHash)
		delete(m.sessions, sid)
	}
	return nil
}

func (m *memoryStore) SaveProfile(_ context.Context, id, locale, regionPref string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	account := m.byID[id]
	if account == nil {
		return errors.New("account missing")
	}
	account.Locale = locale
	account.RegionPref = regionPref
	return nil
}

func (m *memoryStore) SaveAnswer(_ context.Context, id, step, answer string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	account := m.byID[id]
	if account == nil {
		return errors.New("account missing")
	}
	if account.Answers == nil {
		account.Answers = map[string]string{}
	}
	account.Answers[step] = answer
	return nil
}

func (m *memoryStore) SaveFirstPack(_ context.Context, id string, has bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	account := m.byID[id]
	if account == nil {
		return errors.New("account missing")
	}
	account.HasFirstPack = has
	return nil
}

func (m *memoryStore) InsertSession(_ context.Context, sess *session) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cloned := cloneSession(sess)
	m.sessions[cloned.ID] = cloned
	m.accessIndex[cloned.AccessHash] = cloned.ID
	m.refreshIndex[cloned.RefreshHash] = cloned.ID
	return nil
}

func (m *memoryStore) DeleteSession(_ context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	sess := m.sessions[id]
	if sess == nil {
		return nil
	}
	delete(m.accessIndex, sess.AccessHash)
	delete(m.refreshIndex, sess.RefreshHash)
	delete(m.sessions, id)
	return nil
}

func (m *memoryStore) FindSessionByAccess(_ context.Context, hash [32]byte) (*session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	id, ok := m.accessIndex[hash]
	if !ok {
		return nil, nil
	}
	return cloneSession(m.sessions[id]), nil
}

func (m *memoryStore) FindSessionByRefresh(_ context.Context, hash [32]byte) (*session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	id, ok := m.refreshIndex[hash]
	if !ok {
		return nil, nil
	}
	return cloneSession(m.sessions[id]), nil
}
