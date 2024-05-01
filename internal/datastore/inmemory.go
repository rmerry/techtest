package datastore

import (
	"sync"
	"time"

	"log/slog"

	"github.com/google/uuid"
)

var (
	// expiryPollInterval is used by the session house keeping routine, if
	// defines the frequency that the routine should trigger. Not constant to
	// allow for testing.
	expiryPollInterval = time.Second * 60
)

// InMemory is a volitile in-memory data store implementation that satisfies
// the DB interface. The zero value is not ready to be used, call the
// NewInMemory() function instead.
type InMemory struct {
	mu            *sync.Mutex
	data          map[string]*Session
	logger        *slog.Logger
	stopChan      chan bool
	maxSessionAge time.Duration
}

// NewInMemory takes a maxSessionAge and returns a new instance of InMemory.
// Calling this function also starts the session house keeping routine which
// periodically checks for and deletes expired sessions. Calling Close() will
// shutdown this routine and render the returned InMemory object unusable.
func NewInMemory(maxSessionAge time.Duration) *InMemory {
	logger := slog.Default().With("component", "datastore.InMemory")

	ims := &InMemory{
		mu:            &sync.Mutex{},
		data:          make(map[string]*Session),
		logger:        logger,
		maxSessionAge: maxSessionAge,
		stopChan:      make(chan bool),
	}

	go ims.sessionCleanUpFunc()

	return ims
}

// ReadSession takes a single session ID and performs a session lookup in the
// in-memory store. If a session is found with a matching session ID
// ReadSession returns a pointer to the session object; if no session is found
// nil is returned. The error will always be nil in this in-memory
// implementation.
func (db *InMemory) ReadSession(id string) (*Session, error) {
	db.mu.Lock()
	s, ok := db.data[id]
	db.mu.Unlock()
	if !ok {
		return nil, nil
	}
	return s, nil
}

// WriteSession takes an algorithm and key and creates a new unique session ID
// for them which it then stores in the in-memory data store. The newly created
// session ID is returned. The error will always be nil in this in-memory
// implementation.
func (db *InMemory) WriteSession(algorithm, key string) (string, error) {
	id := uuid.NewString()
	db.mu.Lock()
	db.data[id] = &Session{
		AlgorithmName: algorithm,
		Key:           key,
		CreatedAt:     time.Now().UTC(),
	}
	db.mu.Unlock()

	return id, nil
}

// Close attempts to gracefully stop the session housekeeping routine and clear
// down the in-memory store. Attempting to use the in-memory store after this
// call will result in undefined behaviour.
func (db *InMemory) Close() {
	db.stopChan <- true
	db.data = nil
	db.logger.Info("closed")
}

func (db *InMemory) sessionCleanUpFunc() {
loop:
	for {
		select {
		case <-db.stopChan:
			break loop
		case <-time.NewTicker(expiryPollInterval).C:
			db.logger.Info("running session cleanup")
			db.cleanUpExpiredSessions() // Blocking.
		}
	}

}

func (db *InMemory) cleanUpExpiredSessions() {
	startTime := time.Now()

	var deleted int
	db.mu.Lock()
	for k, v := range db.data {
		if v != nil && time.Now().Sub(v.CreatedAt) > db.maxSessionAge {
			delete(db.data, k)
			deleted += 1
		}
	}
	db.mu.Unlock()

	db.logger.Info("clean up completed",
		"deleted sessions", deleted,
		"duration (ms)", time.Now().Sub(startTime).Milliseconds())
}
