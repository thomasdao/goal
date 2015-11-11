package goal

import (
	"encoding/json"
	"net/http"
	"reflect"
)

// To shorten the code, define a type
type simpleResponse func(http.ResponseWriter, *http.Request) (int, interface{}, error)

// TableName returns table name for the resource
func TableName(resource interface{}) string {
	// Extract name of resource type
	name := db.NewScope(resource).TableName()
	return name
}

// Write response back to client
func renderJSON(rw http.ResponseWriter, request *http.Request, handler simpleResponse) {
	if handler == nil {
		http.Error(rw, http.ErrNotSupported.Error(), http.StatusMethodNotAllowed)
		return
	}

	code, data, err := handler(rw, request)

	if err != nil {
		http.Error(rw, err.Error(), code)
		return
	}

	var content []byte
	content, err = json.Marshal(data)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(code)
	rw.Header().Set("Content-Type", "application/json")
	rw.Write(content)
}

// RegisterModel initializes default routes for a model
func RegisterModel(resource interface{}) {
	db.AutoMigrate(resource)
	sharedAPI.AddDefaultCrudPaths(resource)
	sharedAPI.AddDefaultQueryPath(resource)
}

// dynamicSlice creates a slice with element with resource type
// Copied from http://stackoverflow.com/a/25386460/622510 (Thanks @nemo)
func dynamicSlice(resource interface{}) interface{} {
	rType := reflect.TypeOf(resource)

	// Create a slice to begin with
	slice := reflect.MakeSlice(reflect.SliceOf(rType), 0, 0)

	// Create a pointer to a slice value and set it to the slice
	x := reflect.New(slice.Type())
	x.Elem().Set(slice)

	return x.Interface()
}
