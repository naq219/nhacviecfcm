package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		// Get reminders collection
		collection, err := app.FindCollectionByNameOrId("reminders")
		if err != nil {
			return err
		}

		// Add created field
		collection.Fields.Add(&core.DateField{
			Name:     "created",
			Required: true,
		})

		// Add updated field  
		collection.Fields.Add(&core.DateField{
			Name:     "updated",
			Required: true,
		})

		return app.Save(collection)
	}, func(app core.App) error {
		// Rollback - remove created and updated fields
		collection, err := app.FindCollectionByNameOrId("reminders")
		if err != nil {
			return err
		}

		// Remove fields
		collection.Fields.RemoveById("created")
		collection.Fields.RemoveById("updated")

		return app.Save(collection)
	})
}