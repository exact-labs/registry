package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/dbx"
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
				Values:    []string{"public", "private"},
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

type DistInfo struct {
	Integrity    string `json:"integrity"`
	Tarball      string `json:"tarball"`
	FileCount    int64  `json:"fileCount"`
	UnpackedSize int64  `json:"unpackedSize"`
}

type VersionInfo struct {
	Id           string            `json:"_id"`
	Access       []string          `json:"_maintainers"`
	Version      string            `json:"version"`
	Published    types.DateTime    `json:"published"`
	Description  string            `json:"description"`
	Author       string            `json:"author"`
	License      string            `json:"license"`
	Private      bool              `json:"private"`
	Dependencies map[string]string `json:"dependencies"`
	Dist         DistInfo          `json:"dist"`
}

type PackageInfo struct {
	Id          string                    `json:"_id"`
	Name        string                    `json:"name"`
   License     string                    `json:"license"`
	Description string                    `json:"description"`
	Versions    map[string]VersionInfo    `json:"versions"`
	Times       map[string]types.DateTime `json:"times"`
	Dist        DistInfo                  `json:"dist"`
}

func isPrivate(record *models.Record) bool {
	if record.GetString("visibility") == "private" {
		return true
	}

	return false
}

func hasLicense(record *models.Record) string {
	if record.GetString("license") == "" {
		return "none"
	}

	return record.GetString("license")
}

func main() {
	app := pocketbase.New()

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		e.Router.AddRoute(echo.Route{
			Method: http.MethodGet,
			Path:   "/:name",
			Handler: func(c echo.Context) error {
				package_name := c.PathParam("name")
				collection, err := app.Dao().FindCollectionByNameOrId(package_name)
				records, err := app.Dao().FindRecordsByExpr(package_name, dbx.HashExp{"visibility": "public"})

				latest := records[len(records)-1]
				original := records[0]

				times := make(map[string]types.DateTime)
				pkgs := make(map[string]VersionInfo)

				if err != nil {
					return c.JSON(http.StatusInternalServerError, &ErrorResponse{Status: http.StatusInternalServerError, Error: err})
				}

				for _, record := range records {
					dependencies := make(map[string]string)
					filename := record.GetString("tarball")
					filePath := record.BaseFilesPath() + "/" + filename

					_ = json.Unmarshal([]byte(record.GetString("dependencies")), &dependencies)

					fs, err := app.NewFilesystem()
					if err != nil {
						return c.JSON(http.StatusInternalServerError, &ErrorResponse{Status: http.StatusInternalServerError, Error: err})
					}
					defer fs.Close()

					attribute, err := fs.Attributes(filePath)
					if err != nil {
						return c.JSON(http.StatusInternalServerError, &ErrorResponse{Status: http.StatusInternalServerError, Error: err})
					}

					pkgs[record.GetString("version")] = VersionInfo{
						Id:           record.Id,
						Access:       record.GetStringSlice("access"),
						Version:      record.GetString("version"),
						Published:    record.Created,
						Description:  record.GetString("description"),
						Author:       record.GetString("author"),
						License:      hasLicense(record),
						Private:      isPrivate(record),
						Dependencies: dependencies,
						Dist: DistInfo{
							Integrity:    fmt.Sprintf("MD5_%x", attribute.MD5),
							Tarball:      fmt.Sprintf("https://r.justjs.dev/std/_/%s/%s.tgz", record.GetString("version"), package_name),
							FileCount:    1,
							UnpackedSize: attribute.Size,
						},
					}
				}

				for _, record := range records {
					times[record.GetString("version")] = record.Created
				}

				times["created"] = original.Created
				times["updated"] = latest.Updated

				filename := latest.GetString("tarball")
				filePath := latest.BaseFilesPath() + "/" + filename

				fs, err := app.NewFilesystem()
				if err != nil {
					return c.JSON(http.StatusInternalServerError, &ErrorResponse{Status: http.StatusInternalServerError, Error: err})
				}
				defer fs.Close()

				attribute, err := fs.Attributes(filePath)
				if err != nil {
					return c.JSON(http.StatusInternalServerError, &ErrorResponse{Status: http.StatusInternalServerError, Error: err})
				}

				return c.JSON(http.StatusOK, &PackageInfo{
					Name:        package_name,
					Id:          collection.Id,
					Description: latest.GetString("description"),
					Versions:    pkgs,
					Times:       times,
					Dist: DistInfo{
						Integrity:    fmt.Sprintf("MD5_%x", attribute.MD5),
						Tarball:      fmt.Sprintf("https://r.justjs.dev/std/_/%s.tgz", package_name),
						FileCount:    1,
						UnpackedSize: attribute.Size,
					},
					License: latest.GetString("license"),
				})
			},
			Middlewares: []echo.MiddlewareFunc{
				apis.ActivityLogger(app),
			},
		})

		return nil
	})

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		e.Router.AddRoute(echo.Route{
			Method: http.MethodGet,
			Path:   "/:name/:version",
			Handler: func(c echo.Context) error {
				package_name := c.PathParam("name")
				package_version := c.PathParam("version")

				dependencies := make(map[string]string)

				records, err := app.Dao().FindRecordsByExpr(package_name, dbx.HashExp{"version": package_version})
				_ = json.Unmarshal([]byte(records[0].GetString("dependencies")), &dependencies)

				if err != nil {
					return c.JSON(http.StatusInternalServerError, &ErrorResponse{Status: http.StatusInternalServerError, Error: err})
				}

				filename := records[0].GetString("tarball")
				filePath := records[0].BaseFilesPath() + "/" + filename

				fs, err := app.NewFilesystem()
				if err != nil {
					return c.JSON(http.StatusInternalServerError, &ErrorResponse{Status: http.StatusInternalServerError, Error: err})
				}
				defer fs.Close()

				attribute, err := fs.Attributes(filePath)
				if err != nil {
					return c.JSON(http.StatusInternalServerError, &ErrorResponse{Status: http.StatusInternalServerError, Error: err})
				}

				return c.JSON(http.StatusOK, &VersionInfo{
					Id:           records[0].Id,
					Access:       records[0].GetStringSlice("access"),
					Version:      records[0].GetString("version"),
					Published:    records[0].Created,
					Description:  records[0].GetString("description"),
					Author:       records[0].GetString("author"),
					License:      hasLicense(records[0]),
					Private:      isPrivate(records[0]),
					Dependencies: dependencies,
					Dist: DistInfo{
						Integrity:    fmt.Sprintf("MD5_%x", attribute.MD5),
						Tarball:      fmt.Sprintf("https://r.justjs.dev/std/_/%s/%s.tgz", records[0].GetString("version"), package_name),
						FileCount:    1,
						UnpackedSize: attribute.Size,
					},
				})
			},
			Middlewares: []echo.MiddlewareFunc{
				apis.ActivityLogger(app),
			},
		})

		return nil
	})

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		e.Router.AddRoute(echo.Route{
			Method: http.MethodGet,
			Path:   "/:name/_/:version/:archive",
			Handler: func(c echo.Context) error {
				fs, err := app.NewFilesystem()
				package_name := c.PathParam("name")
				package_version := c.PathParam("version")

				records, err := app.Dao().FindRecordsByExpr(package_name, dbx.HashExp{"version": package_version})
				filePath := fmt.Sprintf("%s/%s", records[0].BaseFilesPath(), records[0].GetString("tarball"))
				servedName := fmt.Sprintf("%s-%s.tgz", package_name, records[0].GetString("version"))

				if err != nil {
					return c.JSON(http.StatusInternalServerError, &ErrorResponse{Status: http.StatusInternalServerError, Error: err})
				}

				if err != nil {
					return c.JSON(http.StatusInternalServerError, &ErrorResponse{Status: http.StatusInternalServerError, Error: err})
				}
				defer fs.Close()

				if err := fs.Serve(c.Response(), c.Request(), filePath, servedName); err != nil {
					return c.JSON(http.StatusInternalServerError, &ErrorResponse{Status: http.StatusInternalServerError, Error: err})
				}

				return nil
			},
			Middlewares: []echo.MiddlewareFunc{
				apis.ActivityLogger(app),
			},
		})

		return nil
	})

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		e.Router.AddRoute(echo.Route{
			Method: http.MethodGet,
			Path:   "/:name/_/:archive",
			Handler: func(c echo.Context) error {
				fs, err := app.NewFilesystem()
				package_name := c.PathParam("name")

				records, err := app.Dao().FindRecordsByExpr(package_name, dbx.HashExp{"visibility": "public"})
				filePath := fmt.Sprintf("%s/%s", records[len(records)-1].BaseFilesPath(), records[len(records)-1].GetString("tarball"))
				servedName := fmt.Sprintf("%s-%s.tgz", package_name, records[len(records)-1].GetString("version"))

				if err != nil {
					return c.JSON(http.StatusInternalServerError, &ErrorResponse{Status: http.StatusInternalServerError, Error: err})
				}

				if err != nil {
					return c.JSON(http.StatusInternalServerError, &ErrorResponse{Status: http.StatusInternalServerError, Error: err})
				}
				defer fs.Close()

				if err := fs.Serve(c.Response(), c.Request(), filePath, servedName); err != nil {
					return c.JSON(http.StatusInternalServerError, &ErrorResponse{Status: http.StatusInternalServerError, Error: err})
				}

				return nil
			},
			Middlewares: []echo.MiddlewareFunc{
				apis.ActivityLogger(app),
			},
		})

		return nil
	})

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		e.Router.AddRoute(echo.Route{
			Method: http.MethodPost,
			Path:   "/create",
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
