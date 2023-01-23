package response

import "just/pkg/types"

func ErrorFromString(status int64, error string) types.Response {
	return types.Response{Status: status, Message: map[string]interface{}{
		"error": error,
	}}
}

func Error(status int64, error error) types.ErrorResponse {
	return types.ErrorResponse{Status: status, Error: error}
}
