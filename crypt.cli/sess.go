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
	"os"
	"path/filepath"
	"strings"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	_ "github.com/mattn/go-sqlite3"
	"github.com/tfwio/session"
)

const (
	defaulutDataset = "data.db"
)

var (
	fdb      = flag.String("db", defaulutDataset, "specify a database to use.")
	fSaltLen = flag.Int("s", -1, "provide default salt length.  -1 will allow an internally definded default size")
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
)

func testDatabase() {
	db, err := gorm.Open("sqlite3", *fdb)
	defer db.Close()
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

	session.SetDefaults(*fdb, "sqlite3", -1, -1)
	session.EnsureTableUsers()
	session.EnsureTableSessions()

	switch strings.ToLower(os.Args[1]) {
	case "create":
		// 1. create a user
		// 2. create session for user
		fCreate.Parse(os.Args[2:])
		if len(*fcPass) > 4 && len(*fcUser) > 3 {
			u := session.User{}
			if *fcSess {
				println("- Sesssion generation requested")
			}
			u.Create(*fcUser, *fcPass, *fSaltLen)
			fmt.Printf("%v\n", u)
			b, s := u.CreateSession32(nil, 2, "cli-example-app")
			fmt.Printf("success: %v; session=%v\n", b, s)
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
		fmt.Printf("--> '%s'\n  CRD: %s\n  EXP: %s\n  SID: %s\n",
			usermap[x.UserID].Name,
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
