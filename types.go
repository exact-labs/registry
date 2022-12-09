package main

import "github.com/pocketbase/pocketbase/tools/types"

type Response struct {
	Status  int64                  `json:"status"`
	Message map[string]interface{} `json:"message"`
}

type ErrorResponse struct {
	Status int64 `json:"status"`
	Error  error `json:"error"`
}

type DistInfo struct {
	Version   string `json:"version"`
	Integrity string `json:"integrity"`
	Tarball   string `json:"tarball"`
	Size      int64  `json:"size"`
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

type Result struct {
	Page       int `json:"page"`
	PerPage    int `json:"perPage"`
	TotalItems int `json:"totalItems"`
	TotalPages int `json:"totalPages"`
	Packages   any `json:"packages"`
}
