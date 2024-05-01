package sessionstore

import (
	"asostechtest/internal/datastore"
	"errors"
	"testing"
	"time"
)

type mockDB struct {
	sessions map[string]*datastore.Session
}

func (m *mockDB) WriteSession(algorithm, key string) (string, error) {
	session := &datastore.Session{
		AlgorithmName: algorithm,
		Key:           key,
		CreatedAt:     time.Now(),
	}
	id := "mock_session_id"
	m.sessions[id] = session
	return id, nil
}

func (m *mockDB) ReadSession(id string) (*datastore.Session, error) {
	session, exists := m.sessions[id]
	if !exists {
		return nil, nil
	}
	return session, nil
}

func TestStore_NewSession(t *testing.T) {
	mockDB := &mockDB{sessions: make(map[string]*datastore.Session)}
	store := New(mockDB, time.Hour)
	algorithm := "mock_algorithm"
	key := "mock_key"

	sessionID, err := store.NewSession(algorithm, key)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if sessionID == "" {
		t.Error("expected non-empty session ID, got empty string")
	}

	session, exists := mockDB.sessions[sessionID]
	if !exists {
		t.Error("expected session to exist")
	}

	if session.AlgorithmName != algorithm || session.Key != key {
		t.Errorf("expected session with algorithm %q and key %q, got %v",
			algorithm,
			key,
			session)
	}
}

func TestStore_GetSession(t *testing.T) {
	mockDB := &mockDB{sessions: make(map[string]*datastore.Session)}
	store := New(mockDB, time.Hour)
	algorithm := "mock_algorithm"
	key := "mock_key"

	sessionID, _ := store.NewSession(algorithm, key)

	t.Run("Session does not exist", func(t *testing.T) {
		_, err := store.GetSession("non_existent_session_id")
		if !errors.Is(err, ErrSessionNotFound) {
			t.Errorf("expected ErrSessionNotFound, got %v", err)
		}
	})

	t.Run("Session has expired", func(t *testing.T) {
		mockDB.sessions[sessionID].CreatedAt = time.Now().Add(-time.Hour * 2)
		_, err := store.GetSession(sessionID)
		if !errors.Is(err, ErrSessionExpired) {
			t.Errorf("expected ErrSessionExpired, got %v", err)
		}
	})

	t.Run("Session is valid", func(t *testing.T) {
		mockDB.sessions[sessionID].CreatedAt = time.Now().Add(time.Hour * 2)
		s, err := store.GetSession(sessionID)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if s == nil {
			t.Errorf("expected session to not be nil")
		}
		if s.AlgorithmName != algorithm || s.Key != key {
			t.Errorf("expected session with algorithm %q and key %q, got %v",
				algorithm,
				key,
				s)
		}
	})
}
