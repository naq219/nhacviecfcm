package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/tools/types"
)

func init() {
	m.Register(func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("reminders")
		if err != nil {
			return err
		}

		collection.ListRule = types.Pointer("user_id = @request.auth.id")
		collection.ViewRule = types.Pointer("user_id = @request.auth.id")
		collection.CreateRule = types.Pointer("@request.auth.id != \"\"")
		collection.UpdateRule = types.Pointer("user_id = @request.auth.id")
		collection.DeleteRule = types.Pointer("user_id = @request.auth.id")

		return app.Save(collection)
	}, func(app core.App) error {
		// Rollback changes
		collection, err := app.FindCollectionByNameOrId("reminders")
		if err != nil {
			return err
		}

		collection.ListRule = nil
		collection.ViewRule = nil
		collection.CreateRule = nil
		collection.UpdateRule = nil
		collection.DeleteRule = nil

		return app.Save(collection)
	})
}