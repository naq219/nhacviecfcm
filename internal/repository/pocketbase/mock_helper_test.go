package pocketbase

import (
	"github.com/pocketbase/dbx"
)

// MockDBHelper is a simple mock for db.DBHelperInterface
type MockDBHelper struct {
	GetOneRowFn  func(query string, params dbx.Params) (dbx.NullStringMap, error)
	GetAllRowsFn func(query string, params dbx.Params) ([]dbx.NullStringMap, error)
	ExecFn       func(query string, params dbx.Params) error
	CountFn      func(query string, params dbx.Params) (int, error)
	ExistsFn     func(query string, params dbx.Params) (bool, error)
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