// +build ignore


// Package main provides a CLI executable that at worked properly at some
// point but ATM, hasn't been tested since a number of changes had been
// implemented.
package main

// tickers would help mitigate sessions
// https://gobyexample.com/tickers

import (
	"flag"
	"fmt"
	"os"
	"strings"

	
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	_ "github.com/mattn/go-sqlite3"
	"github.com/tfwio/session"
)

var (
	fdb      = flag.String("db", "data/ormus.db", "specify a database to use.")
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
		fmt.Printf("%s create (subcommand) args:\n", util.AbsBase(os.Args[0]))
		println()
		fCreate.PrintDefaults()
		println()
		fmt.Printf("%s validate (subcommand) args:\n", util.AbsBase(os.Args[0]))
		println()
		fValidate.PrintDefaults()
		return
	}

	ormus.SetDefaults(*fdb, "sqlite3", -1)
	ormus.EnsureTableUsers()
	ormus.EnsureTableSessions()

	switch strings.ToLower(os.Args[1]) {
	case "create":
		// 1. create a user
		// 2. create session for user
		fCreate.Parse(os.Args[2:])
		if len(*fcPass) > 4 && len(*fcUser) > 3 {
			u := ormus.User{}
			if *fcSess {
				println("- Sesssion generation requested")
			}
			u.Create(*fcUser, *fcPass, *fSaltLen)
			fmt.Printf("%v\n", u)
			b, s := u.CreateSession32(nil, 2, "tfw.io:CLI")
			fmt.Printf("success: %v; session=%v\n", b, s)
		} else {
			fmt.Printf("- username %s; pass %s\n", *fcUser, *fcPass)
			println("- username must be > len(3) chars long")
			println("- you must supply a password > len(4) chars long")
		}
	case "validate":
		fValidate.Parse(os.Args[2:])
		if len(*fvPass) > 4 && len(*fvUser) > 3 {
			u := ormus.User{Name: *fvUser}
			result := u.ValidatePassword(*fvPass)
			fmt.Printf("Result: %v \n", result)
		} else {
			fmt.Printf("- username %s; pass %s\n", *fvUser, *fvPass)
			println("- username must be > len(3) chars long")
			println("- you must supply a password > len(4) chars long")
		}
	case "list":
		ormus.CLIList()
	default:
		println()
		flag.PrintDefaults()
	}

}
