// Package datastore is responsible for all data layer interactions and
// housekeeping.
package datastore

import (
	"time"
)

// Session encapsulates a session object at the data layer.
type Session struct {
	AlgorithmName string
	Key           string
	CreatedAt     time.Time
}

// DB is the core datastore interface. All implementations herein should
// conform to this interface.
type DB interface {
	ReadSession(id string) (*Session, error)
	WriteSession(algorithm, key string) (string, error)
}
