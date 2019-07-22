package session

import (
	"fmt"

	"github.com/jinzhu/gorm"
)

// close on error (no message)
//
// As a precaution, do not attempt to perform complex or multiple
// database operations — or rather perhaps, never load and attempt
// to work with more than one loaded (.Open) database connection
func dbinit() (*gorm.DB, bool) {
	db, err := gorm.Open(datasys, datasource)
	result := false
	if err != nil {
		db.Close()
		result = true
	}
	return db, result
}

// no close on error (no message)
//
// As a precaution, do not attempt to perform complex or multiple
// database operations — or rather perhaps, never load and attempt
// to work with more than one loaded (.Open) database connection
func dbinik() (*gorm.DB, bool) {
	db, err := gorm.Open(datasys, datasource)
	result := false
	if err != nil {
		result = true
	}
	return db, result
}

// closes the database and prints requested status on error.
//
// As a precaution, do not attempt to perform complex or multiple
// database operations — or rather perhaps, never load and attempt
// to work with more than one loaded (.Open) database connection
func iniC(format string, msg ...interface{}) (*gorm.DB, bool) {
	return inik(true, format, msg...)
}

// keep error
//
// As a precaution, do not attempt to perform complex or multiple
// database operations — or rather perhaps, never load and attempt
// to work with more than one loaded (.Open) database connection
func iniK(format string, msg ...interface{}) (*gorm.DB, bool) {
	return inik(false, format, msg...)
}

// closes the database and prints requested status on error.
//
// As a precaution, do not attempt to perform complex or multiple
// database operations — or rather perhaps, never load and attempt
// to work with more than one loaded (.Open) database connection
func inik(closeOnError bool, format string, msg ...interface{}) (*gorm.DB, bool) {
	db, e := dbinik()
	if e {
		if format != "" {
			fmt.Printf(format, msg...)
		}
		if closeOnError {
			db.Close()
		}
	}
	return db, e
}
