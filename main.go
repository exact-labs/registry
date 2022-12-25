package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"just/registry/helpers"

	"github.com/labstack/echo/v5"
	"golang.org/x/exp/slices"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/forms"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/models/schema"
	"github.com/pocketbase/pocketbase/tools/search"
	"github.com/pocketbase/pocketbase/tools/types"
)

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
				MaxSize:   20485760,
				MimeTypes: []string{"application/gzip"},
			},
		})

		if err := form.Submit(); err != nil {
			return err
		}
	}

	return nil
}

func check_auth(app core.App, c echo.Context, package_name string) bool {
	exists, _ := app.Dao().FindCollectionByNameOrId(package_name)
	if exists == nil {
		return true
	}

	records, err := app.Dao().FindRecordsByExpr(c.FormValue("name"), dbx.HashExp{"visibility": "public"})
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

func create_version(app core.App, c echo.Context) error {
	package_name := c.FormValue("name")
	if check_auth(app, c, package_name) == false {
		return errors.New("the authorized record model is not allowed to perform this action")
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

func package_index(app core.App, c echo.Context, split []string) error {
	package_name := split[0]
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
				Version:   record.GetString("version"),
				Integrity: fmt.Sprintf("MD5_%x", attribute.MD5),
				Tarball:   fmt.Sprintf("https://r.justjs.dev/%s/_/%s/%s.tgz", package_name, record.GetString("version"), package_name),
				Size:      attribute.Size,
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
			Version:   latest.GetString("version"),
			Integrity: fmt.Sprintf("MD5_%x", attribute.MD5),
			Tarball:   fmt.Sprintf("https://r.justjs.dev/%s/_/%s.tgz", package_name, package_name),
			Size:      attribute.Size,
		},
		License: latest.GetString("license"),
	})
}

func package_version(app core.App, c echo.Context, split []string) error {
	package_name := split[0]
	package_version := split[1]
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
			Version:   records[0].GetString("version"),
			Integrity: fmt.Sprintf("MD5_%x", attribute.MD5),
			Tarball:   fmt.Sprintf("https://r.justjs.dev/%s/_/%s/%s.tgz", package_name, records[0].GetString("version"), package_name),
			Size:      attribute.Size,
		},
	})
}

func main() {
	_, isUsingGoRun := helpers.InspectRuntime()

	app := pocketbase.NewWithConfig(&pocketbase.Config{
		DefaultDataDir: "packages",
		DefaultDebug:   isUsingGoRun,
	})

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		e.Router.AddRoute(echo.Route{
			Method: http.MethodGet,
			Path:   "/:name_version",
			Handler: func(c echo.Context) error {
				split := strings.Split(c.PathParam("name_version"), "@")

				if len(split) == 1 {
					return package_index(app, c, split)
				} else {
					return package_version(app, c, split)
				}
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
			Method: http.MethodGet,
			Path:   "/:name_version/*",
			Handler: func(c echo.Context) error {
				split := strings.Split(c.PathParam("name_version"), "@")
				file_name := c.PathParam("*")

				if len(split) == 1 {
					records, err := app.Dao().FindRecordsByExpr(split[0], dbx.HashExp{"visibility": "public"})
					filePath := fmt.Sprintf("packages/storage/%s/%s", records[len(records)-1].BaseFilesPath(), records[len(records)-1].GetString("tarball"))

					if err != nil {
						return c.JSON(http.StatusInternalServerError, &ErrorResponse{Status: http.StatusInternalServerError, Error: err})
					}

					file, err := helpers.ReadTar(filePath)
					if err != nil {
						return c.JSON(http.StatusInternalServerError, &ErrorResponse{Status: http.StatusInternalServerError, Error: err})
					}

					return c.String(http.StatusOK, string(file[file_name].Data))
				} else {
					records, err := app.Dao().FindRecordsByExpr(split[0], dbx.HashExp{"version": split[1]})
					filePath := fmt.Sprintf("packages/storage/%s/%s", records[len(records)-1].BaseFilesPath(), records[len(records)-1].GetString("tarball"))

					if err != nil {
						return c.JSON(http.StatusInternalServerError, &ErrorResponse{Status: http.StatusInternalServerError, Error: err})
					}

					file, err := helpers.ReadTar(filePath)
					if err != nil {
						return c.JSON(http.StatusInternalServerError, &ErrorResponse{Status: http.StatusInternalServerError, Error: err})
					}

					return c.String(http.StatusOK, string(file[file_name].Data))
				}
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
					return c.JSON(http.StatusInternalServerError, &Response{Status: http.StatusInternalServerError, Message: map[string]interface{}{
						"error": err.Error(),
					}})
				}

				return c.JSON(http.StatusOK, &Response{Status: http.StatusOK, Message: map[string]interface{}{"created": c.FormValue("name")}})
			},
			Middlewares: []echo.MiddlewareFunc{
				apis.ActivityLogger(app),
				apis.RequireAdminOrRecordAuth("just_auth_system"),
			},
		})

		return nil
	})

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		e.Router.AddRoute(echo.Route{
			Method: http.MethodGet,
			Path:   "/maintainers/:name",
			Handler: func(c echo.Context) error {
				records, err := app.Dao().FindRecordsByExpr(c.PathParam("name"), dbx.HashExp{"visibility": "public"})
				if err != nil {
					return c.JSON(http.StatusNotFound, &Response{Status: http.StatusNotFound, Message: map[string]interface{}{
						"error": "package or file not found",
					}})
				}

				if c.QueryParam("type") == "expanded" {
					apis.EnrichRecord(c, app.Dao(), records[len(records)-1], "access")
					return c.JSON(http.StatusOK, records[len(records)-1].Expand())
				} else {
					return c.JSON(http.StatusOK, records[len(records)-1].GetStringSlice("access"))
				}
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
			Path:   "/dependencies/:name",
			Handler: func(c echo.Context) error {
				records, err := app.Dao().FindRecordsByExpr(c.PathParam("name"), dbx.HashExp{"visibility": "public"})
				if err != nil {
					return c.JSON(http.StatusNotFound, &Response{Status: http.StatusNotFound, Message: map[string]interface{}{
						"error": "package or file not found",
					}})
				}

				packages := make(map[string][]string)
				for _, record := range records {
					dep_list := make(map[string]string)
					urls := []string{}

					_ = json.Unmarshal([]byte(record.GetString("dependencies")), &dep_list)
					for dep_name := range dep_list {
						dep, err := app.Dao().FindRecordsByExpr(dep_name, dbx.HashExp{"visibility": "public"})
						if err == nil {
							urls = append(urls, fmt.Sprintf("https://r.justjs.dev/%s/_/%s/%s.tgz", dep_name, dep[0].GetString("version"), dep_name))
						}
					}
					packages[record.GetString("version")] = urls
				}

				return c.JSON(http.StatusOK, packages)
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
			Path:   "/packages",
			Handler: func(c echo.Context) error {
				fieldResolver := search.NewSimpleFieldResolver(
					"id", "created", "updated", "name", "system", "type",
				)

				collections := []*models.Collection{}
				pkgs := make(map[string]interface{})

				result, err := search.NewProvider(fieldResolver).
					Query(app.Dao().CollectionQuery()).
					ParseAndExec(c.QueryString(), &collections)

				if err != nil {
					return c.JSON(http.StatusInternalServerError, &ErrorResponse{Status: http.StatusInternalServerError, Error: err})
				}

				for _, collection := range collections {
					pkgs[collection.Name] = map[string]interface{}{
						"id":      collection.Id,
						"created": collection.Created,
						"updated": collection.Updated,
					}
				}

				delete(pkgs, "just_auth_system")

				return c.JSON(http.StatusOK, &Result{
					Page:       result.Page,
					PerPage:    result.PerPage,
					TotalItems: result.TotalItems - 1,
					TotalPages: result.TotalPages,
					Packages:   pkgs,
				})

			},
			Middlewares: []echo.MiddlewareFunc{
				apis.ActivityLogger(app),
			},
		})

		return nil
	})
   
   if err := copyTemplates(); err != nil {
      log.Panic(err)
   }
   
	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		e.Router.GET("/templates/*", apis.StaticDirectoryHandler(os.DirFS(templatesDir()), false))
		return nil
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
