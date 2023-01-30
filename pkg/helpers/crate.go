package helpers

import (
	"encoding/json"
	"net/http"
	"strings"
)

type response struct {
	Crate crate `json:"crate"`
}

type crate struct {
	Version string `json:"newest_version"`
}

func GetJustVersion() (string, error) {
	data := response{}
	client := &http.Client{}
	url := "https://crates.io/api/v1/crates/justjs"

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", "JustRegistry/IndexFinder (Crate Version Scanner)")
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	err = json.NewDecoder(res.Body).Decode(&data)
	if err != nil {
		return "", err
	}

	return strings.ReplaceAll(data.Crate.Version, ".", ""), nil
}
