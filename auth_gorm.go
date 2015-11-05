package goal

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

// validateCols columns are valid
func validateCols(usernameCol string, passwordCol string, user interface{}) error {
	// validateCols column names
	scope := db.NewScope(user)
	cols := []string{usernameCol, passwordCol}
	for _, col := range cols {
		if !scope.HasColumn(col) {
			errorMsg := fmt.Sprintf("Column %s does not exist", col)
			return errors.New(errorMsg)
		}
	}

	return nil
}

// RegisterWithPassword checks if username exists and
// sets password with bcrypt algorithm
// Client can provides extra data to be saved into database for user
func RegisterWithPassword(
	w http.ResponseWriter, request *http.Request,
	username string, usernameCol string,
	password string, passwordCol string, extra []byte) error {

	user, err := getUserResource()
	if err != nil {
		return err
	}

	err = validateCols(usernameCol, passwordCol, &user)
	if err != nil {
		return err
	}

	// Search db, if a username is already defined, return error
	qry := fmt.Sprintf("%s = ?", usernameCol)
	db.Where(qry, username).First(&user)
	if user != nil {
		errorMsg := fmt.Sprintf("Username is already registered: %s", username)
		return errors.New(errorMsg)
	}

	// validateCols extra data
	if extra != nil {
		err := json.Unmarshal(extra, &user)
		if err != nil {
			return err
		}
	}

	// Since user was populated with extra data, we need to
	// setup new scope
	scope := db.NewScope(&user)

	// Save a new record to db
	scope.SetColumn(usernameCol, username)

	// Hashing the password with the default cost of 10
	hashedPw, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	scope.SetColumn(passwordCol, hashedPw)
	err = scope.DB().New().Create(scope.Value).Error

	// Set current session
	SetUserSession(w, request)

	return nil
}

// LoginWithPassword checks if username and password correct
// and set user into session
func LoginWithPassword(
	w http.ResponseWriter, request *http.Request,
	username string, usernameCol string,
	password string, passwordCol string) error {

	user, err := getUserResource()
	if err != nil {
		return err
	}

	err = validateCols(usernameCol, passwordCol, user)
	if err != nil {
		return err
	}

	// Search db, if a username is not found, return error
	qry := fmt.Sprintf("%s = ?", usernameCol)
	err = db.Where(qry, username).First(user).Error
	if err != nil {
		return err
	}

	if user == nil {
		errorMsg := fmt.Sprintf("Username not found: %s", username)
		return errors.New(errorMsg)
	}

	// Make sure the password is correct
	scope := db.NewScope(user)
	hashedPw, ok := scope.Get(passwordCol)
	if !ok {
		errorMsg := fmt.Sprintf("Unable to get value from column: %s", passwordCol)
		return errors.New(errorMsg)
	}

	var hashed string
	hashed, ok = hashedPw.(string)
	if !ok {
		return errors.New("Password is not valid string")
	}

	// Comparing the password with the hash
	err = bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password))
	if err != nil {
		return err
	}

	// Set current session
	SetUserSession(w, request)

	return nil
}

// HandleLogout let user logout from the system
func HandleLogout(w http.ResponseWriter, request *http.Request) {
	ClearUserSession(w, request)
}
