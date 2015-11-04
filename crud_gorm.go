// gorm_handlers provides basic methods to interact with
// database using GORM. https://github.com/jinzhu/gorm

package goal

import (
	"encoding/json"
	"fmt"
	"net/http"

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
func Read(resource interface{}, request *http.Request) (int, interface{}, error) {
	if db == nil {
		panic("Database is not initialized yet")
	}

	// Get assumes url requests always has "id" parameters
	vars := mux.Vars(request)

	// Retrieve id parameter
	id := vars["id"]

	// Attempt to retrieve from redis first, if not exist, retrieve from
	// database and cache it
	var err error
	if Pool() != nil {
		name := TableName(resource)
		redisKey := DefaultRedisKey(name, id)
		err = RedisGet(redisKey, resource)
		if err == nil && resource != nil {
			// Check if resource is authorized
			err = CanPerform(resource, request, true)
			if err != nil {
				return 403, nil, err
			}

			return 200, resource, nil
		}
	}

	// Retrieve from database
	err = db.First(resource, id).Error
	if err != nil {
		return 500, nil, err
	}

	// Save to redis
	if Pool() != nil {
		key := RedisKey(resource)
		RedisSet(key, resource)
	}

	// Check if resource is authorized
	err = CanPerform(resource, request, true)
	if err != nil {
		return 403, nil, err
	}

	return 200, resource, nil
}

// Create provides basic implementation to create a record
// into the database
func Create(resource interface{}, request *http.Request) (int, interface{}, error) {
	if db == nil {
		panic("Database is not initialized yet")
	}

	// Parse request body into resource
	decoder := json.NewDecoder(request.Body)
	err := decoder.Decode(resource)
	if err != nil {
		fmt.Println(err)
		return 500, nil, err
	}

	// Save to database
	err = db.Create(resource).Error
	if err != nil {
		return 500, nil, err
	}

	return 200, resource, nil
}

// Update provides basic implementation to update a record
// inside database
func Update(resource interface{}, request *http.Request) (int, interface{}, error) {
	if db == nil {
		panic("Database is not initialized yet")
	}

	// Get assumes url requests always has "id" parameters
	vars := mux.Vars(request)

	// Retrieve id parameter, if error return 400 HTTP error code
	id := vars["id"]

	// Retrieve from database
	err := db.First(resource, id).Error
	if err != nil {
		return 500, nil, err
	}

	// Check permission
	err = CanPerform(resource, request, false)
	if err != nil {
		return 403, nil, err
	}

	// Parse request body into resource
	decoder := json.NewDecoder(request.Body)
	err = decoder.Decode(resource)
	if err != nil {
		fmt.Println(err)
		return 500, nil, err
	}

	// Save to database
	err = db.Save(resource).Error
	if err != nil {
		return 500, nil, err
	}

	return 200, resource, err
}

// Delete provides basic implementation to delete a record inside
// a database
func Delete(resource interface{}, request *http.Request) (int, interface{}, error) {
	if db == nil {
		panic("Database is not initialized yet")
	}

	// Get assumes url requests always has "id" parameters
	vars := mux.Vars(request)

	// Retrieve id parameter, if error return 400 HTTP error code
	id := vars["id"]

	// Retrieve from database
	err := db.First(resource, id).Error
	if err != nil {
		return 500, nil, err
	}

	// Check permission
	err = CanPerform(resource, request, false)
	if err != nil {
		return 403, nil, err
	}

	// Delete record, if failed show 500 error code
	err = db.Delete(resource, id).Error
	if err != nil {
		return 500, nil, err
	}

	return 200, nil, nil
}
