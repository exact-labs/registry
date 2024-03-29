package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"

	"registry/pkg/create"
	"registry/pkg/parse"
	"registry/pkg/response"
	"registry/pkg/routes/handler"
	"registry/pkg/templates"
	"registry/pkg/types"

	"github.com/labstack/echo/v5"
	"github.com/mileusna/useragent"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/tools/search"
)

func Router(app core.App) error {
	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		e.Router.GET("/api/:ver/templates/*", apis.StaticDirectoryHandler(os.DirFS(templates.Dir()), false))

		e.Router.AddRoute(echo.Route{
			Method: http.MethodGet,
			Path:   "/",
			Handler: func(c echo.Context) error {
				return c.Redirect(http.StatusMovedPermanently, "https://justjs.dev/docs/registry")
			},
			Middlewares: []echo.MiddlewareFunc{
				apis.ActivityLogger(app),
			},
		})

		e.Router.AddRoute(echo.Route{
			Method: http.MethodGet,
			Path:   "/:package",
			Handler: func(c echo.Context) error {
				checkAgent := regexp.MustCompile(`Wget/|curl|^$`).MatchString
				userAgent := useragent.Parse(c.Request().UserAgent()).String

				if checkAgent(userAgent) {
					return handler.GetIndex(app, c)
				} else {
					if parse.HasSemVersion(c.PathParam("package")) {
						return handler.PackageVersion(app, c)
					} else {
						return handler.PackageIndex(app, c)
					}
				}
			},
			Middlewares: []echo.MiddlewareFunc{
				apis.ActivityLogger(app),
			},
		})

		e.Router.AddRoute(echo.Route{
			Method: http.MethodGet,
			Path:   "/:package/mod",
			Handler: func(c echo.Context) error {
				return handler.GetIndex(app, c)
			},
			Middlewares: []echo.MiddlewareFunc{
				apis.ActivityLogger(app),
			},
		})

		e.Router.AddRoute(echo.Route{
			Method: http.MethodGet,
			Path:   "/:name/_/:version/:archive",
			Handler: func(c echo.Context) error {
				fs, err := app.NewFilesystem()
				package_name, _ := parse.EncodeName(c.PathParam("name"))
				package_version := c.PathParam("version")
				records, err := app.Dao().FindRecordsByExpr(package_name, dbx.HashExp{"visibility": "public", "version": package_version})
				filePath := fmt.Sprintf("%s/%s", records[0].BaseFilesPath(), records[0].GetString("tarball"))
				servedName := fmt.Sprintf("%s-%s.tgz", c.PathParam("name"), records[0].GetString("version"))

				if err != nil {
               return c.JSON(500, response.ErrorFromString(500, err.Error()))
				}

				if err != nil {
               return c.JSON(500, response.ErrorFromString(500, err.Error()))
				}
				defer fs.Close()

				if err := fs.Serve(c.Response(), c.Request(), filePath, servedName); err != nil {
               return c.JSON(500, response.ErrorFromString(500, err.Error()))
				}

				return nil
			},
			Middlewares: []echo.MiddlewareFunc{
				apis.ActivityLogger(app),
			},
		})

		e.Router.AddRoute(echo.Route{
			Method: http.MethodGet,
			Path:   "/:name/_/:archive",
			Handler: func(c echo.Context) error {
				fs, err := app.NewFilesystem()
				package_name, _ := parse.EncodeName(c.PathParam("name"))
				records, err := app.Dao().FindRecordsByExpr(package_name, dbx.HashExp{"visibility": "public"})
				filePath := fmt.Sprintf("%s/%s", records[len(records)-1].BaseFilesPath(), records[len(records)-1].GetString("tarball"))
				servedName := fmt.Sprintf("%s-%s.tgz", c.PathParam("name"), records[len(records)-1].GetString("version"))

				if err != nil {
               return c.JSON(500, response.ErrorFromString(500, err.Error()))
				}

				if err != nil {
               return c.JSON(500, response.ErrorFromString(500, err.Error()))
				}
				defer fs.Close()

				if err := fs.Serve(c.Response(), c.Request(), filePath, servedName); err != nil {
               return c.JSON(500, response.ErrorFromString(500, err.Error()))
				}

				return nil
			},
			Middlewares: []echo.MiddlewareFunc{
				apis.ActivityLogger(app),
			},
		})

		e.Router.AddRoute(echo.Route{
			Method: http.MethodGet,
			Path:   fmt.Sprintf("/%s/:package/:version/:esm/*", os.Getenv("JUST_VERSION")),
			Handler: func(c echo.Context) error {
				return handler.GetFile(app, c)
			},
			Middlewares: []echo.MiddlewareFunc{
				apis.ActivityLogger(app),
			},
		})

      e.Router.AddRoute(echo.Route{
         Method: http.MethodGet,
         Path:   fmt.Sprintf("/source/:package/:version/*"),
         Handler: func(c echo.Context) error {
            return handler.GetSource(app, c)
         },
         Middlewares: []echo.MiddlewareFunc{
            apis.ActivityLogger(app),
         },
      })

		e.Router.AddRoute(echo.Route{
			Method: http.MethodPost,
			Path:   "/api/:ver/create",
			Handler: func(c echo.Context) error {
				if err := create.Package(app, c); err != nil {
					return c.JSON(500, response.ErrorFromString(500, err.Error()))
				}

				if err := create.Version(app, c); err != nil {
					return c.JSON(500, response.ErrorFromString(500, err.Error()))
				}

				return c.JSON(http.StatusOK, &types.Response{Status: http.StatusOK, Message: map[string]interface{}{"created": c.FormValue("name")}})
			},
			Middlewares: []echo.MiddlewareFunc{
				apis.ActivityLogger(app),
				apis.RequireAdminOrRecordAuth("just_auth_system"),
			},
		})

		e.Router.AddRoute(echo.Route{
			Method: http.MethodGet,
			Path:   "/maintainers/:name",
			Handler: func(c echo.Context) error {
				encoded_name, _ := parse.EncodeName(c.PathParam("name"))
				records, err := app.Dao().FindRecordsByExpr(encoded_name, dbx.HashExp{"visibility": "public"})
				if err != nil {
					return c.JSON(404, response.ErrorFromString(404, "package not found"))
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

		e.Router.AddRoute(echo.Route{
			Method: http.MethodGet,
			Path:   "/api/:ver/dependencies/:name",
			Handler: func(c echo.Context) error {
				encoded_name, _ := parse.EncodeName(c.PathParam("name"))
				records, err := app.Dao().FindRecordsByExpr(encoded_name, dbx.HashExp{"visibility": "public"})
				if err != nil {
					return c.JSON(http.StatusNotFound, &types.Response{Status: http.StatusNotFound, Message: map[string]interface{}{
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
					Filter([]search.FilterData{"id!='_pb_users_auth_'"}).
					ParseAndExec(c.QueryString(), &collections)

				if err != nil {
					return c.JSON(500, response.ErrorFromString(500, err.Error()))
				}

				for _, collection := range collections {
					pkgs[parse.OriginalName(collection.Name)] = map[string]interface{}{
						"id":      collection.Id,
						"b62":     collection.Name,
						"created": collection.Created,
						"updated": collection.Updated,
					}
				}

				return c.JSON(http.StatusOK, &types.Result{
					Page:       result.Page,
					PerPage:    result.PerPage,
					TotalItems: result.TotalItems,
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

	return nil
}
