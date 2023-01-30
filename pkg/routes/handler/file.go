package handler

import (
   "os"
	"fmt"
	"regexp"
	"strings"

	"registry/pkg/helpers"
	"registry/pkg/parse"
	"registry/pkg/response"

	"github.com/evanw/esbuild/pkg/api"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/models"
)

func PackageError(info string) string {
	return fmt.Sprintf(`/* r.justjs.dev - error */
throw new Error("[r.justjs.dev] " + "%s");
export default null;
`, info)
}

func IndexFile(name string, version string, index string, defaultExport bool) string {
	if defaultExport {
		return fmt.Sprintf(`/* r.justjs.dev - %[2]s@%[3]s */
export * from "/%[1]s/%[2]s/%[3]s/es2022/%s";
export { default } from "/%[1]s/%[2]s/%[3]s/es2022/%s";
`, os.Getenv("JUST_VERSION"), name, version, index)
	} else {
		return fmt.Sprintf(`/* r.justjs.dev - %[2]s@%[3]s */
export * from "/%[1]s/%[2]s/%[3]s/es2022/%s";
`, os.Getenv("JUST_VERSION"), name, version, index)
	}
}

func HasDefaultExport(record *models.Record) (bool, error) {
	hasExport := regexp.MustCompile(`export default | as default}`).MatchString
	filePath := fmt.Sprintf("packages/storage/%s/%s", record.BaseFilesPath(), record.GetString("tarball"))
	file, err := helpers.ReadFromTar(record.GetString("index"), filePath)
	if err != nil {
		return false, err
	}

	if hasExport(string(file)) {
		return true, nil
	}

	return false, nil
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
			return c.String(200, PackageError(fmt.Sprintf(`ImportError: %s@%s can only be used as local package`, packageName, packageVersion)))
		}

		defaultExport, err := HasDefaultExport(record)
		if err != nil {
			return c.JSON(500, response.ErrorFromString(500, err.Error()))
		}

		return c.String(200, IndexFile(packageName, packageVersion, record.GetString("index"), defaultExport))
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
			return c.String(200, PackageError(fmt.Sprintf(`ImportError: %s@%s can only be used as local package`, packageName, record.GetString("version"))))
		}

		defaultExport, err := HasDefaultExport(record)
		if err != nil {
			return c.JSON(500, response.ErrorFromString(500, err.Error()))
		}

		return c.String(200, IndexFile(packageName, record.GetString("version"), record.GetString("index"), defaultExport))
	}
}

func GetFile(app core.App, c echo.Context) error {
	packageName := c.PathParam("package")
	packageVersion := c.PathParam("version")
	esVersion := c.PathParam("esm")
	fileName := c.PathParam("*")

	var esTarget = api.DefaultTarget
	switch esVersion {
	case "es2022":
		esTarget = api.ES2022
	case "es2021":
		esTarget = api.ES2021
	case "es2020":
		esTarget = api.ES2020
	case "es2019":
		esTarget = api.ES2019
	case "es2018":
		esTarget = api.ES2018
	case "es2017":
		esTarget = api.ES2017
	case "es2016":
		esTarget = api.ES2016
	case "es2015":
		esTarget = api.ES2015
	case "es6":
		esTarget = api.ES2015
	default:
		return c.String(200, PackageError(fmt.Sprintf("BuildError: target %s cannot be used for %s", esVersion, c.PathParam("package"))))
	}

	add_mod := strings.NewReplacer(`from"./`, fmt.Sprintf(`from"/%s/%s/%s/%s/`, os.Getenv("JUST_VERSION"), packageName, packageVersion, esVersion))
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
	banner := map[string]string{"js": fmt.Sprintf("/* r.justjs.dev - esbuild bundle(%s@%s) %s production */", packageName, packageVersion, esVersion)}

	file, err := helpers.ReadFromTar(fileName, filePath)
	if err != nil {
		return c.String(200, PackageError(fmt.Sprintf("resovleESModule: open /vfs/%s/%s/%s/%s: no such file or directory", encodedName, packageName, packageVersion, fileName)))
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
		Target:            esTarget,
		Format:            api.FormatESModule,
		LogLevel:          api.LogLevelSilent,
		Platform:          api.PlatformBrowser,
	})

	return c.String(200, add_mod.Replace(string(result.OutputFiles[0].Contents)))
}
