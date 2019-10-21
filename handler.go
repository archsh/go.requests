package requests

import (
	"net/http"
	"strings"
)

func UploadHandlerFunc(root, temp, filename string, w http.ResponseWriter, r *http.Request) error {
	//fc := ParseChunk(r.URL.Query())
	if strings.ToUpper(r.Method) == "GET" || strings.ToUpper(r.Method) == "HEAD" {

	} else if strings.ToUpper(r.Method) == "POST" {

	} else {

	}
	return nil
}
