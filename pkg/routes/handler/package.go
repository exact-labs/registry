package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"just/pkg/helpers"
	"just/pkg/parse"
	"just/pkg/types"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	pb_types "github.com/pocketbase/pocketbase/tools/types"
)

func PackageIndex(app core.App, c echo.Context) error {
	package_name, err := parse.EncodeName(c.PathParam("name_version"))
	collection, err := app.Dao().FindCollectionByNameOrId(package_name)
	records, err := app.Dao().FindRecordsByExpr(package_name, dbx.HashExp{"visibility": "public"})

	latest := records[len(records)-1]
	original := records[0]

	times := make(map[string]pb_types.DateTime)
	pkgs := make(map[string]types.VersionInfo)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, &types.ErrorResponse{Status: http.StatusInternalServerError, Error: err})
	}

	for _, record := range records {
		dependencies := make(map[string]string)
		filename := record.GetString("tarball")
		filePath := record.BaseFilesPath() + "/" + filename

		_ = json.Unmarshal([]byte(record.GetString("dependencies")), &dependencies)

		fs, err := app.NewFilesystem()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, &types.ErrorResponse{Status: http.StatusInternalServerError, Error: err})
		}
		defer fs.Close()

		attribute, err := fs.Attributes(filePath)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, &types.ErrorResponse{Status: http.StatusInternalServerError, Error: err})
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
				Tarball:   fmt.Sprintf("https://r.justjs.dev/%s/_/%s/%s.tgz", c.PathParam("name_version"), record.GetString("version"), c.PathParam("name_version")),
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
		return c.JSON(http.StatusInternalServerError, &types.ErrorResponse{Status: http.StatusInternalServerError, Error: err})
	}
	defer fs.Close()

	attribute, err := fs.Attributes(filePath)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, &types.ErrorResponse{Status: http.StatusInternalServerError, Error: err})
	}

	return c.JSON(http.StatusOK, &types.PackageInfo{
		Name:        c.PathParam("name_version"),
		Id:          collection.Id,
		Description: latest.GetString("description"),
		Versions:    pkgs,
		Times:       times,
		Dist: types.DistInfo{
			Version:   latest.GetString("version"),
			Integrity: fmt.Sprintf("MD5_%x", attribute.MD5),
			Tarball:   fmt.Sprintf("https://r.justjs.dev/%s/_/%s.tgz", c.PathParam("name_version"), c.PathParam("name_version")),
			Size:      attribute.Size,
		},
		License: latest.GetString("license"),
	})
}

func PackageVersion(app core.App, c echo.Context, split []string) error {
	package_name, err := parse.EncodeName(split[0])
	package_version := split[1]
	dependencies := make(map[string]string)

	records, err := app.Dao().FindRecordsByExpr(package_name, dbx.HashExp{"visibility": "public", "version": package_version})
	_ = json.Unmarshal([]byte(records[0].GetString("dependencies")), &dependencies)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, &types.ErrorResponse{Status: http.StatusInternalServerError, Error: err})
	}

	filename := records[0].GetString("tarball")
	filePath := records[0].BaseFilesPath() + "/" + filename

	fs, err := app.NewFilesystem()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, &types.ErrorResponse{Status: http.StatusInternalServerError, Error: err})
	}
	defer fs.Close()

	attribute, err := fs.Attributes(filePath)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, &types.ErrorResponse{Status: http.StatusInternalServerError, Error: err})
	}

	return c.JSON(http.StatusOK, &types.VersionInfo{
		Id:           records[0].Id,
		Access:       records[0].GetStringSlice("access"),
		Version:      records[0].GetString("version"),
		Published:    records[0].Created,
		Description:  records[0].GetString("description"),
		Author:       records[0].GetString("author"),
		License:      helpers.PackageHasLicense(records[0]),
		Private:      helpers.PackagePrivacyStatus(records[0]),
		Dependencies: dependencies,
		Dist: types.DistInfo{
			Version:   records[0].GetString("version"),
			Integrity: fmt.Sprintf("MD5_%x", attribute.MD5),
			Tarball:   fmt.Sprintf("https://r.justjs.dev/%s/_/%s/%s.tgz", split[0], records[0].GetString("version"), split[0]),
			Size:      attribute.Size,
		},
	})
}
