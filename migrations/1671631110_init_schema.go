package migrations

import (
	"log"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/tools/types"
	"golang.org/x/crypto/bcrypt"
)

func init() {
	m.Register(func(app core.App) error {

		hashed, _ := bcrypt.GenerateFromPassword([]byte("123"), bcrypt.DefaultCost)
		_, err := app.DB().Insert("_superusers", map[string]any{
			"id":       "admin_default",
			"email":    "a@a.a",
			"password": string(hashed),
		}).Execute()
		if err != nil {
			log.Println("Error creating admin user:", err)
			return err
		}

		// Create musers collection (auth type)
		musersCollection := core.NewAuthCollection("musers")
		musersCollection.ListRule = types.Pointer("id = @request.auth.id")
		musersCollection.ViewRule = types.Pointer("id = @request.auth.id")
		musersCollection.CreateRule = types.Pointer("")
		musersCollection.DeleteRule = types.Pointer("id = @request.auth.id")
		musersCollection.UpdateRule = types.Pointer("id = @request.auth.id")

		// Add custom fields to musers
		musersCollection.Fields.Add(&core.TextField{
			Name:     "fcm_token",
			Required: false,
		})
		musersCollection.Fields.Add(&core.BoolField{
			Name:     "is_fcm_active",
			Required: false,
		})
		musersCollection.Fields.Add(&core.TextField{
			Name:     "fcm_error",
			Required: false,
		})
		musersCollection.Fields.Add(&core.AutodateField{
			Name:     "created",
			OnCreate: true,
		})
		musersCollection.Fields.Add(&core.AutodateField{
			Name:     "updated",
			OnCreate: true,
			OnUpdate: true,
		})

		if err := app.Save(musersCollection); err != nil {
			return err
		}

		// Create reminders collection
		remindersCollection := core.NewBaseCollection("reminders")

		// Add fields to reminders
		remindersCollection.Fields.Add(&core.RelationField{
			Name:          "user_id",
			Required:      true,
			MaxSelect:     1,
			CascadeDelete: true,
			CollectionId:  musersCollection.Id,
		})
		remindersCollection.Fields.Add(&core.TextField{
			Name:     "title",
			Required: true,
		})
		remindersCollection.Fields.Add(&core.TextField{
			Name:     "description",
			Required: false,
		})
		remindersCollection.Fields.Add(&core.SelectField{
			Name:      "type",
			Required:  true,
			MaxSelect: 1,
			Values:    []string{"one_time", "recurring"},
		})
		remindersCollection.Fields.Add(&core.SelectField{
			Name:      "calendar_type",
			Required:  true,
			MaxSelect: 1,
			Values:    []string{"solar", "lunar"},
		})
		remindersCollection.Fields.Add(&core.DateField{
			Name:     "next_trigger_at",
			Required: true,
		})
		remindersCollection.Fields.Add(&core.TextField{
			Name:     "trigger_time_of_day",
			Required: false,
		})
		remindersCollection.Fields.Add(&core.JSONField{
			Name:     "recurrence_pattern",
			Required: false,
		})
		remindersCollection.Fields.Add(&core.SelectField{
			Name:      "repeat_strategy",
			Required:  true,
			MaxSelect: 1,
			Values:    []string{"none", "retry_until_complete"},
		})
		remindersCollection.Fields.Add(&core.NumberField{
			Name:     "retry_interval_sec",
			Required: false,
		})
		remindersCollection.Fields.Add(&core.NumberField{
			Name:     "max_retries",
			Required: false,
		})
		remindersCollection.Fields.Add(&core.NumberField{
			Name:     "retry_count",
			Required: false,
		})
		remindersCollection.Fields.Add(&core.SelectField{
			Name:      "status",
			Required:  true,
			MaxSelect: 1,
			Values:    []string{"active", "completed", "paused"},
		})
		remindersCollection.Fields.Add(&core.DateField{
			Name: "snooze_until",
		})
		remindersCollection.Fields.Add(&core.DateField{
			Name:     "last_completed_at",
			Required: false,
		})
		remindersCollection.Fields.Add(&core.DateField{
			Name: "last_sent_at",
		})
		remindersCollection.Fields.Add(&core.AutodateField{
			Name:     "created",
			OnCreate: true,
		})
		remindersCollection.Fields.Add(&core.AutodateField{
			Name:     "updated",
			OnCreate: true,
			OnUpdate: true,
		})

		if err := app.Save(remindersCollection); err != nil {
			return err
		}

		// Add optional DateFields sau khi tạo collection để tránh bug Pocketbase
		// alterQueries := []string{
		// 	`ALTER TABLE reminders ADD COLUMN last_completed_at TEXT DEFAULT NULL`,
		// 	`ALTER TABLE reminders ADD COLUMN last_sent_at TEXT DEFAULT NULL`,
		// 	`ALTER TABLE reminders ADD COLUMN snooze_until TEXT DEFAULT NULL`,
		// }

		// for _, query := range alterQueries {
		// 	q := app.DB().NewQuery(query)
		// 	if _, err := q.Execute(); err != nil {
		// 		return fmt.Errorf("failed to add optional date fields: %w", err)
		// 	}
		// }

		// Create system_status collection
		systemStatusCollection := core.NewBaseCollection("system_status")

		systemStatusCollection.Fields.Add(&core.NumberField{
			Name:     "mid",
			Required: true,
		})
		systemStatusCollection.Fields.Add(&core.BoolField{
			Name:     "worker_enabled",
			Required: false,
		})
		systemStatusCollection.Fields.Add(&core.TextField{
			Name:     "last_error",
			Required: false,
		})
		systemStatusCollection.Fields.Add(&core.AutodateField{
			Name:     "created",
			OnCreate: true,
		})
		systemStatusCollection.Fields.Add(&core.AutodateField{
			Name:     "updated",
			OnCreate: true,
			OnUpdate: true,
		})

		if err := app.Save(systemStatusCollection); err != nil {
			return err
		}

		// Insert initial system status data
		systemStatusRecord := core.NewRecord(systemStatusCollection)
		systemStatusRecord.Set("mid", 1)
		systemStatusRecord.Set("worker_enabled", false)
		systemStatusRecord.Set("last_error", "")
		systemStatusRecord.Set("updated", types.NowDateTime())

		if err := app.Save(systemStatusRecord); err != nil {
			return err
		}

		return nil
	}, func(app core.App) error {
		// down queries - delete collections in reverse order
		if collection, _ := app.FindCollectionByNameOrId("system_status"); collection != nil {
			if err := app.Delete(collection); err != nil {
				return err
			}
		}

		if collection, _ := app.FindCollectionByNameOrId("reminders"); collection != nil {
			if err := app.Delete(collection); err != nil {
				return err
			}
		}

		if collection, _ := app.FindCollectionByNameOrId("musers"); collection != nil {
			if err := app.Delete(collection); err != nil {
				return err
			}
		}

		return nil
	})
}
