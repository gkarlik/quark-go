package document

import (
	"github.com/gkarlik/quark-go/system"
)

// DbContext represents document database access mechanism. It is used to access collections.
type DbContext interface {
	GetCollection(name string) Collection

	system.Disposer
}

// Document is object which represents data stored in nosql document database.
type Document struct{}

// Collection represents nosql document database collection.
type Collection interface {
	Context() DbContext

	Insert(doc interface{}) error
	Update(selector interface{}, doc interface{}) error
	Remove(doc interface{}) error

	Find(selector interface{}) ([]interface{}, error)
}
