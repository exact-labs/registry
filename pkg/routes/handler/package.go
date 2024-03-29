package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
   "strings"

	"registry/pkg/helpers"
	"registry/pkg/parse"
	"registry/pkg/response"
	"registry/pkg/types"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	pb_types "github.com/pocketbase/pocketbase/tools/types"
)

func PackageIndex(app core.App, c echo.Context) error {
	package_name, err := parse.EncodeName(c.PathParam("package"))
	if err != nil {
		return c.JSON(500, response.ErrorFromString(500, err.Error()))
	}

	collection, err := app.Dao().FindCollectionByNameOrId(package_name)
	if err != nil {
		return c.JSON(404, response.ErrorFromString(404, "package not found"))
	}

	records, err := app.Dao().FindRecordsByExpr(package_name, dbx.HashExp{"visibility": "public"})
	if err != nil {
		return c.JSON(500, response.ErrorFromString(500, err.Error()))
	}

	latest := records[len(records)-1]
	original := records[0]

	times := make(map[string]pb_types.DateTime)
	pkgs := make(map[string]types.VersionInfo)

	for _, record := range records {
		dependencies := make(map[string]string)
		filename := record.GetString("tarball")
		filePath := record.BaseFilesPath() + "/" + filename

		if err := json.Unmarshal([]byte(record.GetString("dependencies")), &dependencies); err != nil {
			return c.JSON(500, response.ErrorFromString(500, err.Error()))
		}

		fs, err := app.NewFilesystem()
		if err != nil {
			return c.JSON(500, response.ErrorFromString(500, err.Error()))
		}
		defer fs.Close()

		attribute, err := fs.Attributes(filePath)
		if err != nil {
			return c.JSON(500, response.ErrorFromString(500, err.Error()))
		}

		pkgs[record.GetString("version")] = types.VersionInfo{
			Id:           record.Id,
			Access:       record.GetStringSlice("access"),
			Version:      record.GetString("version"),
			Published:    record.Created,
			Description:  record.GetString("description"),
			Author:       record.GetString("author"),
			License:      helpers.PackageHasLicense(record),
			Private:      helpers.PackagePrivacyStatus(record),
			Dependencies: dependencies,
			Dist: types.DistInfo{
				Version:   record.GetString("version"),
				Integrity: fmt.Sprintf("MD5_%x", attribute.MD5),
				Tarball:   fmt.Sprintf("%s/%s/_/%s/%s.tgz", helpers.TarPath(), c.PathParam("package"), record.GetString("version"), c.PathParam("package")),
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
		return c.JSON(500, response.ErrorFromString(500, err.Error()))
	}
	defer fs.Close()

	attribute, err := fs.Attributes(filePath)
	if err != nil {
		return c.JSON(500, response.ErrorFromString(500, err.Error()))
	}

	return c.JSON(http.StatusOK, &types.PackageInfo{
		Name:        c.PathParam("package"),
		Id:          collection.Id,
		Description: latest.GetString("description"),
		Versions:    pkgs,
		Times:       times,
		Dist: types.DistInfo{
			Version:   latest.GetString("version"),
			Integrity: fmt.Sprintf("MD5_%x", attribute.MD5),
			Tarball:   fmt.Sprintf("%s/%s/_/%s.tgz", helpers.TarPath(), c.PathParam("package"), c.PathParam("package")),
			Size:      attribute.Size,
		},
		License: latest.GetString("license"),
	})
}

func PackageVersion(app core.App, c echo.Context) error {
	packageVersion := parse.GetSemVer(c.PathParam("package"))   
	packageName := strings.ReplaceAll(c.PathParam("package"), fmt.Sprintf("@%s", packageVersion), "")
	dependencies := make(map[string]string)

	encodedName, err := parse.EncodeName(packageName)
	if err != nil {
		return c.JSON(500, response.ErrorFromString(500, err.Error()))
	}

	records, err := app.Dao().FindRecordsByExpr(encodedName, dbx.HashExp{"visibility": "public", "version": packageVersion})
	if len(records) == 0 {
		return c.JSON(404, response.ErrorFromString(404, "package or version not found"))
	}

	record := records[0]
	filename := record.GetString("tarball")
	filePath := record.BaseFilesPath() + "/" + filename

	if err := json.Unmarshal([]byte(records[0].GetString("dependencies")), &dependencies); err != nil {
		return c.JSON(500, response.ErrorFromString(500, err.Error()))
	}

	fs, err := app.NewFilesystem()
	if err != nil {
		return c.JSON(500, response.ErrorFromString(500, err.Error()))
	}
	defer fs.Close()

	attribute, err := fs.Attributes(filePath)
	if err != nil {
		return c.JSON(500, response.ErrorFromString(500, err.Error()))
	}

	return c.JSON(http.StatusOK, &types.VersionInfo{
		Id:           record.Id,
		Access:       record.GetStringSlice("access"),
		Version:      record.GetString("version"),
		Published:    record.Created,
		Description:  record.GetString("description"),
		Author:       record.GetString("author"),
		License:      helpers.PackageHasLicense(record),
		Private:      helpers.PackagePrivacyStatus(record),
		Dependencies: dependencies,
		Dist: types.DistInfo{
			Version:   record.GetString("version"),
			Integrity: fmt.Sprintf("MD5_%x", attribute.MD5),
			Tarball:   fmt.Sprintf("%s/%s/_/%s/%s.tgz", helpers.TarPath(), packageName, packageVersion, packageName),
			Size:      attribute.Size,
		},
	})
}
