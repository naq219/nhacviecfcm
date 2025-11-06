package db

import (
	"fmt"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// InTransaction runs a function in a database transaction.
// It automatically commits on nil error, or rolls back if fn returns error.
// Example:
//
//	err := db.InTransaction(app, func(tx *db.DBHelper) error {
//	    return tx.Exec("UPDATE users SET name={:n} WHERE id={:id}", dbx.Params{
//	        "n": "John", "id": 1,
//	    })
//	})
func InTransaction(app *pocketbase.PocketBase, fn func(tx *DBHelper) error) error {
	if app == nil {
		return fmt.Errorf("app cannot be nil")
	}

	return app.RunInTransaction(func(txApp core.App) error {
		// Convert the transaction app to DBHelper
		txHelper := &DBHelper{app: txApp.(*pocketbase.PocketBase)}
		return fn(txHelper)
	})
}
