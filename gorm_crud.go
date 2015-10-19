// gorm_handlers provides basic methods to interact with
// database using GORM. https://github.com/jinzhu/gorm

package goal

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"

	_ "github.com/go-sql-driver/mysql" // Driver for mysql
	_ "github.com/lib/pq"              // Driver for postgres
	_ "github.com/mattn/go-sqlite3"    // Driver for sqlite
)

// Global variable to interact with database
var db *gorm.DB

// InitGormDb initializes global variable db
func InitGormDb(newDb *gorm.DB) {
	db = newDb
}

// DB returns global variable db
func DB() *gorm.DB {
	return db
}

// Read provides basic implementation to retrieve object
// based on request parameters
func Read(model interface{}, request *http.Request) (int, interface{}) {
	if db == nil {
		panic("Database is not initialized yet")
	}

	// Get assumes url requests always has "id" parameters
	vars := mux.Vars(request)

	// Retrieve id parameter, if error return 400 HTTP error code
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		return 400, nil
	}

	// Retrieve from database
	if db.First(model, id).Error != nil {
		return 500, nil
	}

	return 200, model
}

// Create provides basic implementation to create a record
// into the database
func Create(model interface{}, request *http.Request) (int, interface{}) {
	if db == nil {
		panic("Database is not initialized yet")
	}

	// Parse request body into model
	decoder := json.NewDecoder(request.Body)
	err := decoder.Decode(model)
	if err != nil {
		fmt.Println(err)
		return 500, nil
	}

	// Save to database
	if db.Create(model).Error != nil {
		return 500, nil
	}

	return 200, model
}

// Update provides basic implementation to update a record
// inside database
func Update(model interface{}, request *http.Request) (int, interface{}) {
	if db == nil {
		panic("Database is not initialized yet")
	}

	// Get assumes url requests always has "id" parameters
	vars := mux.Vars(request)

	// Retrieve id parameter, if error return 400 HTTP error code
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		return 400, nil
	}

	// Retrieve from database
	if db.First(model, id).Error != nil {
		return 500, nil
	}

	// Parse request body into model
	decoder := json.NewDecoder(request.Body)
	err = decoder.Decode(model)
	if err != nil {
		fmt.Println(err)
		return 500, nil
	}

	// Save to database
	if db.Save(model).Error != nil {
		return 500, nil
	}

	return 200, model
}

// Delete provides basic implementation to delete a record inside
// a database
func Delete(model interface{}, request *http.Request) (int, interface{}) {
	if db == nil {
		panic("Database is not initialized yet")
	}

	// Get assumes url requests always has "id" parameters
	vars := mux.Vars(request)

	// Retrieve id parameter, if error return 400 HTTP error code
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		return 400, nil
	}

	// Delete record, if failed show 500 error code
	if db.Delete(model, id).Error != nil {
		return 500, nil
	}

	return 200, nil
}
