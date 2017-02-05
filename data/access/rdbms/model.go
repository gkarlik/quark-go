package rdbms

import (
	"github.com/gkarlik/quark-go/system"
)

// DbContext represents relational database access mechanism. It is used to insert, modify, delete and query the data.
// Additionally it can create database transactions.
type DbContext interface {
	BeginTransaction() DbTransaction
	IsInTransaction() bool

	system.Disposer
}

// DbTransaction represents transaction in relational database.
type DbTransaction interface {
	Context() DbContext

	Rollback() error
	Commit() error
}

// Entity is object which represents relational data stored in database.
type Entity struct{}

// Repository "Mediates between the domain and data mapping layers using a collection-like interface for accessing domain objects." (Martin Fowler).
// Repository should represent DDD "Aggregate". This means that each repository method should preserve aggregate consistency.
// To preserve consistency between several Aggregates use DbTransaction (BeginTransaction() method from DbContext).
type Repository interface {
	Context() DbContext

	First(where ...interface{}) (interface{}, error)
	Find(where ...interface{}) ([]interface{}, error)

	Save(entity interface{}) error
	Delete(entity interface{}, where ...interface{}) error
}
