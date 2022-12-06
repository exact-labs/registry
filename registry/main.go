package main

import (
	"log"
	"mime/multipart"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/forms"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/models/schema"
	"github.com/pocketbase/pocketbase/tools/types"
)

type OkResponse struct {
	Status  int64                  `json:"status"`
	Message map[string]interface{} `json:"message"`
}

type ErrorResponse struct {
	Status int64 `json:"status"`
	Error  error `json:"error"`
}

func create_package(app core.App, package_name string) error {
	exists, _ := app.Dao().FindCollectionByNameOrId(package_name)

	if exists != nil {
		return nil
	} else {
		collection := &models.Collection{}
		form := forms.NewCollectionUpsert(app, collection)
		form.Name = package_name
		form.Type = models.CollectionTypeBase
		form.ListRule = types.Pointer("")
		form.ViewRule = types.Pointer("")
		form.CreateRule = types.Pointer("")
		form.UpdateRule = types.Pointer("@request.auth.id = access.id")
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
			Name:     "de_listed",
			Type:     schema.FieldTypeBool,
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

type UploadedFile struct {
	name   string
	header *multipart.FileHeader
}

func create_version(app core.App, c echo.Context, package_name string) error {
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
				package_name := c.PathParam("name")

				if err := create_package(app, package_name); err != nil {
					return c.JSON(http.StatusInternalServerError, &ErrorResponse{Status: http.StatusInternalServerError, Error: err})
				}

				if err := create_version(app, c, package_name); err != nil {
					return c.JSON(http.StatusInternalServerError, &ErrorResponse{Status: http.StatusInternalServerError, Error: err})
				}

				return c.JSON(http.StatusOK, &OkResponse{Status: http.StatusOK, Message: map[string]interface{}{"created": package_name}})
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
