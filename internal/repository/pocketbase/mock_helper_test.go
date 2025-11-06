package pocketbase

import (
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// MockDBHelper is a simple mock for db.DBHelperInterface
type MockDBHelper struct {
	GetOneRowFn  func(query string, params dbx.Params) (dbx.NullStringMap, error)
	GetAllRowsFn func(query string, params dbx.Params) ([]dbx.NullStringMap, error)
	ExecFn       func(query string, params dbx.Params) error
	CountFn      func(query string, params dbx.Params) (int, error)
	ExistsFn     func(query string, params dbx.Params) (bool, error)
	AppFn        func() *pocketbase.PocketBase
}

func (m *MockDBHelper) GetOneRow(query string, params dbx.Params) (dbx.NullStringMap, error) {
	if m.GetOneRowFn != nil {
		return m.GetOneRowFn(query, params)
	}
	return nil, nil
}

func (m *MockDBHelper) GetAllRows(query string, params dbx.Params) ([]dbx.NullStringMap, error) {
	if m.GetAllRowsFn != nil {
		return m.GetAllRowsFn(query, params)
	}
	return nil, nil
}

func (m *MockDBHelper) Exec(query string, params dbx.Params) error {
	if m.ExecFn != nil {
		return m.ExecFn(query, params)
	}
	return nil
}

func (m *MockDBHelper) Count(query string, params dbx.Params) (int, error) {
	if m.CountFn != nil {
		return m.CountFn(query, params)
	}
	return 0, nil
}

func (m *MockDBHelper) Exists(query string, params dbx.Params) (bool, error) {
	if m.ExistsFn != nil {
		return m.ExistsFn(query, params)
	}
	return false, nil
}



func (m *MockDBHelper) App() *pocketbase.PocketBase {
	if m.AppFn != nil {
		return m.AppFn()
	}
	// Trả về nil vì không cần mock app cho test hiện tại
	return nil
}

// MockApp mocks PocketBase app for ORM operations
type MockApp struct {
	FindCollectionByNameOrIdFn func(nameOrId string) (*core.Collection, error)
	SaveFn                     func(record *core.Record) error
}

func (m *MockApp) FindCollectionByNameOrId(nameOrId string) (*core.Collection, error) {
	if m.FindCollectionByNameOrIdFn != nil {
		return m.FindCollectionByNameOrIdFn(nameOrId)
	}
	// Trả về mock collection
	collection := core.NewBaseCollection("reminders")
	return collection, nil
}

func (m *MockApp) Save(record *core.Record) error {
	if m.SaveFn != nil {
		return m.SaveFn(record)
	}
	return nil
}