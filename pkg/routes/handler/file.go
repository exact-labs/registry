package handler

import (
	"fmt"
	"strings"

	"registry/pkg/helpers"
	"registry/pkg/parse"
	"registry/pkg/response"

	"github.com/evanw/esbuild/pkg/api"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

func FileNotFound(info string) string {
	return fmt.Sprintf(`/* r.justjs.dev - error */
throw new Error("[r.justjs.dev] " + "resovleESModule: open %s: no such file or directory");
export default null;
`, info)
}

func PackageOnlyError(info string) string {
   return fmt.Sprintf(`/* r.justjs.dev - error */
throw new Error("[r.justjs.dev] " + "ImportError: %s can only be used as local package");
export default null;
`, info)
}

func IndexFile(name string, version string, index string) string {
   return fmt.Sprintf(`/* r.justjs.dev - %s@%s */
export * from "/v052/%[1]s/%[2]s/es2022/%s";
`, name, version, index)
}

func GetIndex(app core.App, c echo.Context) error {
	if parse.HasSemVersion(c.PathParam("package")) {
		packageVersion := parse.GetSemVer(c.PathParam("package"))
		packageName := strings.ReplaceAll(c.PathParam("package"), fmt.Sprintf("@%s", packageVersion), "")
		encodedName, err := parse.EncodeName(packageName)
		if err != nil {
			return c.JSON(500, response.ErrorFromString(500, err.Error()))
		}
      
      records, err := app.Dao().FindRecordsByExpr(encodedName, dbx.HashExp{"visibility": "public", "version": packageVersion})
      if err != nil {
         return c.JSON(500, response.ErrorFromString(500, err.Error()))
      }
      
      record := records[0]
      if record.GetString("group") == "local" {
         return c.String(404, PackageOnlyError(fmt.Sprintf(`%s@%s/%s`, packageName, packageVersion)))
      }
      
		return c.String(200, IndexFile(packageName, packageVersion, record.GetString("index")))
	} else {
      packageName := c.PathParam("package")
		encodedName, err := parse.EncodeName(packageName)
		if err != nil {
			return c.JSON(500, response.ErrorFromString(500, err.Error()))
		}
      
		records, err := app.Dao().FindRecordsByExpr(encodedName, dbx.HashExp{"visibility": "public"})
		if err != nil {
			return c.JSON(500, response.ErrorFromString(500, err.Error()))
		}

		record := records[len(records)-1]
      if record.GetString("group") == "local" {
         return c.String(404, PackageOnlyError(fmt.Sprintf(`%s@%s/%s`, packageName, record.GetString("version"))))
      }
      
      return c.String(200, IndexFile(packageName, record.GetString("version"), record.GetString("index")))
	}
}

func GetFile(app core.App, c echo.Context) error {
	packageName := c.PathParam("package")
	packageVersion := c.PathParam("version")
	esVersion := c.PathParam("esm")
	fileName := c.PathParam("*")

	add_mod := strings.NewReplacer(`from"./`, fmt.Sprintf(`from"/v052/%s/%s/%s/`, packageName, packageVersion, esVersion))
	encodedName, err := parse.EncodeName(packageName)
	if err != nil {
		return c.JSON(500, response.ErrorFromString(500, err.Error()))
	}

	records, err := app.Dao().FindRecordsByExpr(encodedName, dbx.HashExp{"visibility": "public", "version": packageVersion})
	if err != nil {
		return c.JSON(500, response.ErrorFromString(500, err.Error()))
	}

	record := records[len(records)-1]
	filePath := fmt.Sprintf("packages/storage/%s/%s", record.BaseFilesPath(), record.GetString("tarball"))
	banner := map[string]string{"js": fmt.Sprintf("/* r.justjs.dev - esbuild bundle(%s@%s) es2022 production */", packageName, packageVersion)}

	file, err := helpers.ReadFromTar(fileName, filePath)
	if err != nil {
		return c.String(404, FileNotFound(fmt.Sprintf("/vfs/%s/%s/%s/%s", encodedName, packageName, packageVersion, fileName)))
	}

	contents := &api.StdinOptions{
		Contents:   string(file),
		Sourcefile: fileName,
	}

	loader := map[string]api.Loader{
		".wasm":  api.LoaderDataURL,
		".svg":   api.LoaderDataURL,
		".png":   api.LoaderDataURL,
		".webp":  api.LoaderDataURL,
		".ttf":   api.LoaderDataURL,
		".eot":   api.LoaderDataURL,
		".woff":  api.LoaderDataURL,
		".woff2": api.LoaderDataURL,
	}

	result := api.Build(api.BuildOptions{
		Loader:            loader,
		Stdin:             contents,
		EntryPoints:       nil,
		MinifyWhitespace:  true,
		MinifyIdentifiers: true,
		MinifySyntax:      true,
		KeepNames:         true,
		Write:             false,
		Bundle:            false,
		Banner:            banner,
		Target:            api.ES2022,
		Format:            api.FormatESModule,
		LogLevel:          api.LogLevelSilent,
		Platform:          api.PlatformBrowser,
	})

	return c.String(200, add_mod.Replace(string(result.OutputFiles[0].Contents)))
}
