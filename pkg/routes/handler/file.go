package handler

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"just/pkg/helpers"
	"just/pkg/parse"
	"just/pkg/types"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

func GetFile(app core.App, c echo.Context, split []string, file_name string, raw_name string, user_agent string) error {
	add_mod := strings.NewReplacer(`from './`, fmt.Sprintf(`from './%s/`, raw_name), `from "./`, fmt.Sprintf(`from "./%s/`, raw_name))
	regex := regexp.MustCompile("curl")

	if len(split) == 1 {
		encoded_split, _ := parse.EncodeName(split[0])
		records, err := app.Dao().FindRecordsByExpr(encoded_split, dbx.HashExp{"visibility": "public"})
		filePath := fmt.Sprintf("packages/storage/%s/%s", records[len(records)-1].BaseFilesPath(), records[len(records)-1].GetString("tarball"))

		if err != nil {
			return c.JSON(http.StatusInternalServerError, &types.ErrorResponse{Status: http.StatusInternalServerError, Error: err})
		}

		file, err := helpers.ReadTar(filePath)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, &types.ErrorResponse{Status: http.StatusInternalServerError, Error: err})
		}

		if file_name == "index_file" {
			if regex.MatchString(user_agent) {
				return c.String(http.StatusOK, add_mod.Replace(string(file[records[len(records)-1].GetString("index")].Data)))
			} else {
				return c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%v/%v", split[0], records[len(records)-1].GetString("index")))
			}
		} else {
			return c.String(http.StatusOK, add_mod.Replace(string(file[file_name].Data)))
		}
	} else {
		encoded_split, _ := parse.EncodeName(split[0])
		records, err := app.Dao().FindRecordsByExpr(encoded_split, dbx.HashExp{"version": split[1]})
		filePath := fmt.Sprintf("packages/storage/%s/%s", records[len(records)-1].BaseFilesPath(), records[len(records)-1].GetString("tarball"))

		if err != nil {
			return c.JSON(http.StatusInternalServerError, &types.ErrorResponse{Status: http.StatusInternalServerError, Error: err})
		}

		file, err := helpers.ReadTar(filePath)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, &types.ErrorResponse{Status: http.StatusInternalServerError, Error: err})
		}

		if file_name == "index_file" {
			if regex.MatchString(user_agent) {
				return c.String(http.StatusOK, add_mod.Replace(string(file[records[len(records)-1].GetString("index")].Data)))
			} else {
				return c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%v/%v", raw_name, records[len(records)-1].GetString("index")))
			}
		} else {
			return c.String(http.StatusOK, add_mod.Replace(string(file[file_name].Data)))
		}
	}
}
