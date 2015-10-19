package goal

import (
	"encoding/json"
	"net/http"
	"reflect"
	"strings"
)

// To shorten the code, define a type
type simpleResponse func(http.ResponseWriter, *http.Request) (int, interface{})

func simpleStructName(resource interface{}) string {
	// Extract name of resource type
	cls := reflect.TypeOf(resource).String()
	arr := strings.Split(cls, ".")
	name := arr[len(arr)-1]
	return name
}

// Write response back to client
func renderJSON(rw http.ResponseWriter, request *http.Request, handler simpleResponse) {
	if handler == nil {
		rw.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	code, data := handler(rw, request)

	content, err := json.Marshal(data)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(code)
	rw.Write(content)
}
