// Package sessionstore provides types and methods for managing encryption
// sessions. An encryption session encapsulates an algorithm, a key and an
// expiry time. Note that this package is not responsible for cleaning up
// expired sessions; that responsibility lies upstream in the data layer.
package sessionstore

import (
	"atostechtest/internal/datastore"
	"errors"
	"log/slog"
	"time"
)

var (
	// ErrDatabaseError indicates a problem communicating with the database.
	ErrDatabaseError = errors.New("database error")

	// ErrSessionExpired is returned when a session exists but has expired.
	ErrSessionExpired = errors.New("session expired")

	// ErrSessionNotFound is returned when a session does not exist.
	ErrSessionNotFound = errors.New("session not found")
)

// Session encapsulates a session object.
type Session struct {
	AlgorithmName string
	Key           string
	ExpiresAt     time.Time
}

// Store manages the creation and retrieval of session objects. The zero value
// is not ready to use, create new instances with the NewStore() function.
type Store struct {
	db            datastore.DB
	stopChan      chan bool
	logger        *slog.Logger
	maxSessionAge time.Duration
}

// New takes an implementation of the datastore.DB interface and returns a
// pointer to Store.
func New(db datastore.DB, maxSessionAge time.Duration) *Store {
	logger := slog.Default().With("component", "sessionstore")
	s := &Store{
		db:            db,
		stopChan:      make(chan bool),
		maxSessionAge: maxSessionAge,
		logger:        logger,
	}

	return s
}

// NewSession attempts to create a new session with a given algorithm and key
// in the underlying data store. If there are any issues communicating with
// database an ErrDatabaseError is returned. On successful session
// creation a session ID is returned.
func (s *Store) NewSession(algorithm, key string) (string, error) {
	id, err := s.db.WriteSession(algorithm, key)
	if err != nil {
		return "", ErrDatabaseError
	}

	return id, nil
}

// GetSession takes a session ID and performs a session look up in the
// underlying data layer. In the absence of a valid session this method always
// returns an error. If the session is not found an ErrSessionNotFound is
// returned. If the session has expired an ErrSessionExpired is returned. If
// there are any issues communicating with database an ErrDatabaseError is
// returned.
func (s *Store) GetSession(id string) (*Session, error) {
	session, err := s.db.ReadSession(id)
	if err != nil {
		s.logger.Error("retrieving session from data store",
			"id", id,
			"err", err)
		return nil, ErrDatabaseError
	}
	if session == nil {
		return nil, ErrSessionNotFound
	}
	expiresAt := session.CreatedAt.Add(s.maxSessionAge)
	if expiresAt.Before(time.Now()) {
		return nil, ErrSessionExpired
	}
	return &Session{
		AlgorithmName: session.AlgorithmName,
		Key:           session.Key,
		ExpiresAt:     expiresAt,
	}, nil
}
