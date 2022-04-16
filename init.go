package session

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	newLogger = logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second,   // Slow SQL threshold
			LogLevel:                  logger.Silent, // Log level
			IgnoreRecordNotFoundError: true,          // Ignore ErrRecordNotFound error for logger
			Colorful:                  true,          // Disable color
		})
)

//
// https://github.com/glebarez/sqlite
//

// close on error (no message)
//
// As a precaution, do not attempt to perform complex or multiple
// database operations — or rather perhaps, never load and attempt
// to work with more than one loaded (.Open) database connection
//
// returns true on error
func dbinit() (*gorm.DB, bool) {
	db, err := gorm.Open(sqlite.Open(datasource), &gorm.Config{Logger: newLogger})
	result := false
	if err != nil {
		//db.Close()
		result = true
	}
	return db, result
}

// no close on error (no message)
//
// As a precaution, do not attempt to perform complex or multiple
// database operations — or rather perhaps, never load and attempt
// to work with more than one loaded (.Open) database connection
//
// returns true on error
func dbinik() (*gorm.DB, bool) {
	db, err := gorm.Open(sqlite.Open(datasource), &gorm.Config{Logger: newLogger})
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
//
// returns true on error
func iniC(format string, msg ...interface{}) (*gorm.DB, bool) {
	return inik(true, format, msg...)
}

// keep error
//
// As a precaution, do not attempt to perform complex or multiple
// database operations — or rather perhaps, never load and attempt
// to work with more than one loaded (.Open) database connection
//
// returns true on error
func iniK(format string, msg ...interface{}) (*gorm.DB, bool) {
	return inik(false, format, msg...)
}

// closes the database and prints requested status on error.
//
// As a precaution, do not attempt to perform complex or multiple
// database operations — or rather perhaps, never load and attempt
// to work with more than one loaded (.Open) database connection
//
// returns true on error
func inik(closeOnError bool, format string, msg ...interface{}) (*gorm.DB, bool) {
	db, e := dbinik()
	result := false
	if e {
		if format != "" {
			fmt.Printf("well then: "+format, msg...)
		}
		fmt.Printf("error: %v\n", e)
		result = true
		// if closeOnError {
		// 	db.Close()
		// }
	}
	return db, result
}
