package main

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/forms"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/models/schema"
)

type OkResponse struct {
	Status  int64                  `json:"status"`
	Message map[string]interface{} `json:"message"`
}

type ErrorResponse struct {
	Status int64 `json:"status"`
	Error  error `json:"error"`
}

func update_package(app core.App, package_name string, record_id string, de_listed bool) error {
	record, err := app.Dao().FindRecordById(package_name, record_id)
	if err != nil {
		return err
	}

	form := forms.NewRecordUpsert(app, record)

	form.LoadData(map[string]any{
		"de_listed": de_listed,
	})

	if err := form.Submit(); err != nil {
		return err
	}

	return nil
}

func create_package(app core.App, c echo.Context) error {
	package_name := c.FormValue("name")
	exists, _ := app.Dao().FindCollectionByNameOrId(package_name)

	if exists != nil {
		return nil
	} else {
		collection := &models.Collection{}
		form := forms.NewCollectionUpsert(app, collection)
		form.Name = package_name
		form.Type = models.CollectionTypeBase
		form.ListRule = nil
		form.ViewRule = nil
		form.CreateRule = nil
		form.UpdateRule = nil
		form.DeleteRule = nil

		form.Schema.AddField(&schema.SchemaField{
			Name:     "access",
			Type:     schema.FieldTypeRelation,
			Required: true,
			Unique:   false,
			Options: &schema.RelationOptions{
				CollectionId:  "_pb_users_auth_",
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
            Values: []string{"public", "private"},
         },
		})

		form.Schema.AddField(&schema.SchemaField{
			Name:     "description",
			Type:     schema.FieldTypeText,
			Required: false,
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

func create_version(app core.App, c echo.Context) error {
	package_name := c.FormValue("name")
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

func main() {
	app := pocketbase.New()

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		e.Router.AddRoute(echo.Route{
			Method: http.MethodPost,
			Path:   "/api/package/create/:name",
			Handler: func(c echo.Context) error {
				if err := create_package(app, c); err != nil {
					return c.JSON(http.StatusInternalServerError, &ErrorResponse{Status: http.StatusInternalServerError, Error: err})
				}

				if err := create_version(app, c); err != nil {
					return c.JSON(http.StatusInternalServerError, &ErrorResponse{Status: http.StatusInternalServerError, Error: err})
				}

				return c.JSON(http.StatusOK, &OkResponse{Status: http.StatusOK, Message: map[string]interface{}{"created": c.FormValue("name")}})
			},
			Middlewares: []echo.MiddlewareFunc{
				apis.ActivityLogger(app),
				apis.RequireRecordAuth("users"),
			},
		})

		return nil
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
