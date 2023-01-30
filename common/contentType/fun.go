package contentType

import (
	"strings"
)

type contentType string

const (
	ApplicationJson contentType = "application/json"
	Form            contentType = "application/x-www-form-urlencoded"
	MultipartForm   contentType = "multipart/form-data"
)

// Check the Content-Type of request
// if the content-type of request is not matched requireType then return false,
// else return true
func Check(contentType string, requireType contentType) bool {
	return strings.Contains(strings.ToLower(contentType), string(requireType))
}
