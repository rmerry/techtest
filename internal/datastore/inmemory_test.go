package datastore

import (
	"testing"
	"time"
)

func TestWriteSession(t *testing.T) {
	db := NewInMemory(time.Hour)

	algorithm := "AES"
	key := "secret_key"

	sessionID, err := db.WriteSession(algorithm, key)
	if err != nil {
		t.Errorf("unexpected error writing session: %v", err)
	}

	if sessionID == "" {
		t.Error("expected non-empty session ID, got empty string")
	}

	session, _ := db.ReadSession(sessionID)
	if session == nil {
		t.Error("expected session to exist in memory, but it was not found")
	}
	if session.AlgorithmName != algorithm || session.Key != key {
		t.Errorf("expected session with algorithm %q and key %q, got %v", algorithm, key, session)
	}

	db.Close()
}

func TestReadSession(t *testing.T) {
	db := NewInMemory(time.Hour)

	algorithm := "AES"
	key := "secret_key"

	sessionID, _ := db.WriteSession(algorithm, key)

	t.Run("Session found", func(t *testing.T) {
		session, err := db.ReadSession(sessionID)
		if err != nil {
			t.Errorf("unexpected error reading session: %v", err)
		}
		if session == nil {
			t.Error("expected session to be found, but it was not found")
		}
		if session.AlgorithmName != algorithm || session.Key != key {
			t.Errorf("expected session with algorithm %q and key %q, got %v", algorithm, key, session)
		}
	})

	t.Run("Session does not exist", func(t *testing.T) {
		nonExistentID := "non_existent_session_id"
		session, err := db.ReadSession(nonExistentID)
		if err != nil {
			t.Errorf("unexpected error reading session: %v", err)
		}
		if session != nil {
			t.Errorf("expected session with ID %q not to be found, but it was found", nonExistentID)
		}
	})

	db.Close()
}

// Using timers in my tests, yeah I'm not a big fan but here we go!:
func TestSessionCleanUp(t *testing.T) {
	expiryPollInterval = time.Second // Set expiry poll interval to 1 second
	db := NewInMemory(time.Second)   // Set max session age to 1 second

	sessionID, _ := db.WriteSession("AES", "key")
	time.Sleep(2 * time.Second)

	// Housekeeping routine should have removed the session by now.
	session, _ := db.ReadSession(sessionID)
	if session != nil {
		t.Error("expected session to be cleaned up, but it still exists")
	}

	db.Close()
}
