package db

import (
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// ORMDBHelperInterface defines the interface for ORM operations
// This allows for easy mocking in tests
type ORMDBHelperInterface interface {
	FindCollectionByNameOrId(nameOrId string) (*core.Collection, error)
	Save(record *core.Record) error
}

// ORMDBHelper provides ORM operations using PocketBase app
type ORMDBHelper struct {
	app *pocketbase.PocketBase
}

// NewORMDBHelper returns a helper bound to the current PocketBase app
func NewORMDBHelper(app *pocketbase.PocketBase) *ORMDBHelper {
	return &ORMDBHelper{app: app}
}

// FindCollectionByNameOrId finds a collection by name or ID
func (h *ORMDBHelper) FindCollectionByNameOrId(nameOrId string) (*core.Collection, error) {
	return h.app.FindCollectionByNameOrId(nameOrId)
}

// Save saves a record using PocketBase ORM
func (h *ORMDBHelper) Save(record *core.Record) error {
	return h.app.Save(record)
}