// Package main provides a CLI executable that at worked properly at some
// point but ATM, hasn't been tested since a number of changes had been
// implemented.
//
// Though not thourally documented by the cli's `help` command,
// there are three sub-commands that can be used to test out
// the implemented crypto validiation.
//
//
// 1. `sess.exe create -u <username> -p <password>` will create
//    a user in the table `users` and also create a session in the
//    `sessions` table.
//
// 2. `sess.exe validate -u <username> -p <password>` will check
//    and validate the provided password.
//
// 3. `sess.exe list` will list the entries in the sessions table and
//    provide the User.Name for each row.
package main

// tickers would help mitigate sessions
// https://gobyexample.com/tickers

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tfwio/session"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	defaulutDataset = "data.db"
)

var (
	fdb      = flag.String("db", defaulutDataset, "specify a database to use.")
	fSaltLen = flag.Int("s", -1, "provide default salt length.  -1 will allow an internally definded default size")
	fHashLen = flag.Int("h", -1, "provide default hash length.  -1 will allow an internally definded default size")
	//
	fList = flag.NewFlagSet("list", flag.ExitOnError)
	//
	fCreate = flag.NewFlagSet("create", flag.ExitOnError)
	fcUser  = fCreate.String("u", "admin", "speficy username.")
	fcPass  = fCreate.String("p", "", "for validation and creation of a login profile.")
	fcSess  = fCreate.Bool("sess", false, "Create a session for the user.")
	//
	fValidate = flag.NewFlagSet("validate", flag.ExitOnError)
	fvUser    = fValidate.String("u", "admin", "speficy username.")
	fvPass    = fValidate.String("p", "", "for validation and creation of a login profile.")
	//
	//fvalid   = flag.String("V", "", "for validation and creation of a login profile.")
	//fsalt    = flag.String("salt", "", "[optional] supply salt and hash to validate -V <pass> (or fallback to db).")
	//fhash    = flag.String("hash", "", "[optional] supply salt and hash to validate -V <pass> (or fallback to db).")
	newLogger = logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second,   // Slow SQL threshold
			LogLevel:                  logger.Silent, // Log level
			IgnoreRecordNotFoundError: true,          // Ignore ErrRecordNotFound error for logger
			Colorful:                  false,         // Disable color,
		})
)

func testDatabase() {
	db, err := gorm.Open(sqlite.Open(defaulutDataset), &gorm.Config{Logger: newLogger})
	// defer db.Close()
	if err != nil {
		fmt.Printf("error loading empty database: %v\n", db)
	} else {
		println("success: opened empty database.")
	}
}

func main() {

	if len(os.Args) == 1 {
		flag.PrintDefaults()
		println()
		fmt.Printf("%s create (subcommand) args:\n", AbsBase(os.Args[0]))
		println()
		fCreate.PrintDefaults()
		println()
		fmt.Printf("%s validate (subcommand) args:\n", AbsBase(os.Args[0]))
		println()
		fValidate.PrintDefaults()
		return
	}

	session.SetDefaults("sqlite3", *fdb, *fSaltLen, *fHashLen)

	switch strings.ToLower(os.Args[1]) {
	case "create":
		// 1. create a user
		// 2. create session for user
		fCreate.Parse(os.Args[2:])

		if len(*fcPass) > 4 && len(*fcUser) > 3 {
			u := session.User{}
			// s := session.Session{}
			if *fcSess {
				println("- session generation requested")
			}
			// attempt to create the user?
			if xr := u.Create_CheckErr(*fcUser, *fcPass); int(xr) == 0 {
				fmt.Printf("Create User: {\n  Name: %s\n  Hash: %s,\n  Salt: %s,\n  ID: %v\n}\n", u.Name, u.Hash, u.Salt, u.ID)
				if xr != 0 {
					fmt.Printf("error: %s\n", xr)
				}
				if success, s := u.CreateSession(nil, "cli-example-app", false); !success {
					fmt.Printf("{\n  success:         %v;\n  session-id:      %s;\n  session-client:  %s,\n  session-host:    %s\n  user-id:       %v\n}\n", !success, s.SessID, s.Client, s.Host, s.UserID)
				} else {
					fmt.Printf("!{\n  success:         %v;\n  session-id:      %s;\n  session-client:  %s,\n  session-host:    %s\n  user-id:       %v\n}\n", !success, s.SessID, s.Client, s.Host, s.UserID)
				}
			} else {
				fmt.Printf("- User \"%s\" exists\n", u.Name)
				if success, s := u.CreateSession(nil, "cli-example-app", false); !success {
					fmt.Printf("{\n  success:         %v;\n  session-id:      %s;\n  session-client:  %s,\n  session-host:    %s\n  user-id:       %v\n}\n", !success, s.SessID, s.Client, s.Host, s.UserID)
				} else {
					// fmt.Printf("!{\n  success:         %v;\n  session-id:      %s;\n  session-client:  %s,\n  session-host:    %s\n  user-id:       %v\n}\n", !success, s.SessID, s.Client, s.Host, s.UserID)
				}
			}
		} else {
			fmt.Printf("- username %s; pass %s\n", *fcUser, *fcPass)
			println("- username must be > len(3) chars long")
			println("- you must supply a password > len(4) chars long")
		}
	case "validate":
		fValidate.Parse(os.Args[2:])
		if len(*fvPass) > 4 && len(*fvUser) > 3 {
			u := session.User{Name: *fvUser}
			result := u.ValidatePassword(*fvPass)
			fmt.Printf("Result: %v \n", result)
		} else {
			fmt.Printf("- username %s; pass %s\n", *fvUser, *fvPass)
			println("- username must be > len(3) chars long")
			println("- you must supply a password > len(4) chars long")
		}
	case "list":
		List()
	default:
		println()
		flag.PrintDefaults()
	}

}

// List returns a  sessions for CLI.
// The method first fetches a list of User elements
// then reports the Sessions with user-data (name).
func List() {
	// list sessions
	usermap := session.UserGetList()
	sessions, count := session.ListSessions()
	fmt.Printf("--> found %d entries\n", count)
	for _, x := range sessions {
		fmt.Printf("--> '%s'\n  UserID: %v\n  CRD: %s\n  EXP: %s\n  SID: %s\n",
			usermap[x.UserID].Name,
			x.UserID,
			x.Created.Format("20060102_1504.005"),
			x.Expires.Format("20060102_1504.005"),
			x.SessID)
	}
}

// AbsBase returns `filepath.Base(path)` after converting to absolute representation of path; Ignores errors.
func AbsBase(path string) (dir string) {
	return filepath.Base(Abs(path))
}

// Abs returns an absolute representation of path; Ignores errors.
func Abs(path string) (dir string) {
	dir, _ = filepath.Abs(path)
	return dir
}
