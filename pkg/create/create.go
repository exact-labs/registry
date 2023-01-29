package create

import (
	"errors"
	"fmt"

	"registry/pkg/parse"
   "golang.org/x/exp/slices"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/forms"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/models/schema"
	"github.com/pocketbase/pocketbase/tools/types"
)

func Package(app core.App, c echo.Context) error {
	package_name, err := parse.EncodeName(c.FormValue("name"))
	if err != nil {
		return err
	}

	exists, _ := app.Dao().FindCollectionByNameOrId(package_name)
	auth, err := app.Dao().FindCollectionByNameOrId("just_auth_system")
	if err != nil {
		return err
	}

	if exists != nil {
		return nil
	} else {
		collection := &models.Collection{}
		form := forms.NewCollectionUpsert(app, collection)
		form.Name = package_name
		form.Type = models.CollectionTypeBase
		form.ListRule = types.Pointer("@request.auth.id = access.id")
		form.ViewRule = types.Pointer("@request.auth.id = access.id")
		form.CreateRule = nil
		form.UpdateRule = nil
		form.DeleteRule = nil

		form.Schema.AddField(&schema.SchemaField{
			Name:     "access",
			Type:     schema.FieldTypeRelation,
			Required: true,
			Unique:   false,
			Options: &schema.RelationOptions{
				CollectionId:  auth.Id,
				CascadeDelete: false,
			},
		})

		form.Schema.AddField(&schema.SchemaField{
			Name:     "visibility",
			Type:     schema.FieldTypeSelect,
			Required: true,
			Unique:   false,
			Options: &schema.SelectOptions{
				MaxSelect: 1,
				Values:    []string{"public", "private"},
			},
		})

		form.Schema.AddField(&schema.SchemaField{
			Name:     "group",
			Type:     schema.FieldTypeSelect,
			Required: true,
			Unique:   false,
			Options: &schema.SelectOptions{
				MaxSelect: 1,
				Values:    []string{"local", "net", "both"},
			},
		})

		form.Schema.AddField(&schema.SchemaField{
			Name:     "description",
			Type:     schema.FieldTypeText,
			Required: false,
			Unique:   false,
		})

		form.Schema.AddField(&schema.SchemaField{
			Name:     "index",
			Type:     schema.FieldTypeText,
			Required: true,
			Unique:   false,
		})

		form.Schema.AddField(&schema.SchemaField{
			Name:     "author",
			Type:     schema.FieldTypeText,
			Required: true,
			Unique:   false,
		})

		form.Schema.AddField(&schema.SchemaField{
			Name:     "url",
			Type:     schema.FieldTypeText,
			Required: false,
			Unique:   false,
		})

		form.Schema.AddField(&schema.SchemaField{
			Name:     "repository",
			Type:     schema.FieldTypeText,
			Required: false,
			Unique:   false,
		})

		form.Schema.AddField(&schema.SchemaField{
			Name:     "license",
			Type:     schema.FieldTypeText,
			Required: false,
			Unique:   false,
		})

		form.Schema.AddField(&schema.SchemaField{
			Name:     "dependencies",
			Type:     schema.FieldTypeJson,
			Required: false,
			Unique:   false,
		})

		form.Schema.AddField(&schema.SchemaField{
			Name:     "version",
			Type:     schema.FieldTypeText,
			Required: true,
			Unique:   true,
			Options: &schema.TextOptions{
				Pattern: `^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(?:-((?:0|[1-9][0-9]*|[0-9]*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9][0-9]*|[0-9]*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`,
			},
		})

		form.Schema.AddField(&schema.SchemaField{
			Name:     "tarball",
			Type:     schema.FieldTypeFile,
			Required: true,
			Unique:   false,
			Options: &schema.FileOptions{
				MaxSelect: 1,
				MaxSize:   10485760,
				MimeTypes: []string{"application/gzip"},
			},
		})

		if err := form.Submit(); err != nil {
			return err
		}
	}

	return nil
}

func CheckAuth(app core.App, c echo.Context, package_name string) bool {
	exists, _ := app.Dao().FindCollectionByNameOrId(package_name)
	if exists == nil {
		return true
	}

	records, err := app.Dao().FindRecordsByExpr(package_name, dbx.HashExp{"visibility": "public"})
	if err != nil {
		return false
	}

	admin, _ := c.Get(apis.ContextAdminKey).(*models.Admin)
	user, _ := c.Get(apis.ContextAuthRecordKey).(*models.Record)

	if admin != nil {
		return true
	}

	if len(records) == 0 {
		return true
	}

	return slices.Contains(records[len(records)-1].GetStringSlice("access"), user.Id)
}

func Version(app core.App, c echo.Context) error {
	package_name, err := parse.EncodeName(c.FormValue("name"))
	if err != nil {
		return err
	}

	if CheckAuth(app, c, package_name) == false {
		return errors.New(fmt.Sprintf("You do not have permission to publish '%s'. Are you logged in as the correct user?", c.FormValue("name")))
	}

	collection, err := app.Dao().FindCollectionByNameOrId(package_name)
	if err != nil {
		return err
	}

	record := models.NewRecord(collection)
	form := forms.NewRecordUpsert(app, record)

	if err := form.LoadRequest(c.Request(), ""); err != nil {
		return err
	}

	if err := form.Submit(); err != nil {
		return err
	}

	return nil
}
