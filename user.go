package session

import (
	"errors"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// User structure
type User struct {
	ID   int64  `gorm:"auto_increment;unique_index;primary_key;column:id"`
	Name string `gorm:"size:27;column:user"`
	Salt string `gorm:"size:432;column:salt"`
	Hash string `gorm:"size:432;column:hash"`
}

// TableName Set User's table name to be `users`
func (User) TableName() string {
	return "users"
}

// UserGetList gets a map of all `User`s by ID.
func UserGetList() map[int64]User {
	var users []User
	usermap := make(map[int64]User)
	db, err := iniC("error(user-get-list) loading database\n")
	//defer db.Close()
	if !err {
		db.Find(&users)
		// fmt.Printf("- found %d entries\n", len(users))
		for _, x := range users {
			usermap[x.ID] = x
		}
	}
	return usermap
}

/* http://jinzhu.me/gorm/crud.html#query */

// ByName gets a user by [name].
// If `u` properties are set, then those are defaulted in FirstOrInit
//
// return true on success
func (u *User) ByName(name string) bool {
	// fmt.Printf("ByName(%s)\n", name)
	db, err := iniC("error(user-by-name) loading database\n")
	result := false
	//defer db.Close()
	if !err {
		if err1 := db.Where("[user] = ?", name).First(&u).Error; err1 != nil {
			if errors.Is(err1, gorm.ErrRecordNotFound) && false { // THIS PRINT ERROR IS IGNORED
				println("ERROR: record not found")
			}
		}
		if u.Name == name {
			result = true
		}
	}
	// fmt.Printf("!-> FOUND %s, %d, result-found-success: %v\n", u.Name, u.ID, result)
	return result
}

// ByID gets a user by [id].
func (u *User) ByID(id int64) bool {
	db, err := iniC("error(user-by-id) loading database\n")
	//defer db.Close()
	if !err {
		db.FirstOrInit(u, User{ID: id})
	}
	return u.ID == id
}

// CreateSession Save a session into the sessions table.
//
// (param: `r interface{}`) is to utilize gin-gonic/gin `*gin.Context` as
// its suggested input interface given that we can use it to retrieve
// the `ClientIP()` and store that value to our database in order to
// validate a given user-session.
//
// returns true on error
func (u *User) CreateSession(r interface{}, host string, keepAlive bool) (bool, Session) {

	t := time.Now()
	result := true

	if service == nil {
		service = DefaultService()
	}
	// println("-- setup session data (create-session)")
	// println("==================")
	// fmt.Printf("Host: %s, UserID: %v, keepAlive: %v, salt-size: %v, created: %s\n", host, u.ID, keepAlive, defaultSaltSize, t.Local().String())
	// println("---------")
	// fmt.Printf("Service: %v, r (interface): %v\n", service, r)
	sess := Session{
		Host:      host,
		UserID:    u.ID,
		KeepAlive: keepAlive,
		SessID:    toUBase64(NewSaltString(defaultSaltSize)),
		Created:   t,
		Expires:   service.AddDate(t),
	}

	// acceptable client is of type: gin.Context, nil and string
	// looks like we're giving it nil.
	// anyways, attempt to load the nil client's name.
	// host?
	sess.Client = getClientString(r)
	// fmt.Printf("My Client String: %s\n", sess.Client)
	// println("---------")

	// guess we're making sure the database exists again?

	//defer db.Close()
	var sx = Session{}
	if berry, xs := sx.HasSessionForUser(u); !berry {
		if errors.Is(xs, gorm.ErrRecordNotFound) { // lets just pretend this didn't happen
			println("--- ERROR: record not found")
			db, _ := iniC("error(user-create-session) loading database\n")
			println("--- CREATING SESSION")

			mrr := db.Create(&sess)
			if mrr.Error != nil {
				println("ERROR: error creating session data.", mrr.Error.Error())
			}

			if err := mrr.Error; err != nil {
				fmt.Printf("ERROR: %s\n", err.Error())
				println("Session creation ERROR", u.Name, u.ID)
			} else if mrr.RowsAffected == 1 {
				println("- SESSION CREATED.", u.Name, u.ID)
				result = false
			} else {
				fmt.Printf("so what now? %v %v\n", db.RowsAffected, sess.ID)
			}

		} else {
			println("--- ERROR: unknown error.")
		}
	} else {
		if sx.IsValid() {
			fmt.Printf("- Session EXPIRES %s!\n", sx.Expires)
		} else {
			fmt.Printf("- Session EXPIRED of %s!\n", sx.Expires)
		}
	}

	return result, sess
}

type UserErrorConst int

const (
	Perfection UserErrorConst = iota
	HasName
	LenName
	LenPass
	CheckDB UserErrorConst = -1
)

func (u UserErrorConst) String() string {
	switch u {
	case CheckDB:
		return "GORM_ERROR_Table-User: check database or db path"
	case Perfection:
		return "GORM_ERROR_Table-User: Perfect (as in no errors to report)"
	case HasName:
		return "GORM_ERROR_Table-User: \"name\" exists"
	case LenName:
		return "GORM_ERROR_Table-User: check name-length"
	case LenPass:
		return "GORM_ERROR_Table-User: check pass-length"
	default:
		return "GORM_ERROR_Table-User: how in the hell did you get here?"
	}
}

// Create attempts to create a user and returns success or failure.
// If a user allready exists results in failure.
//
// Returns
func (u *User) Create_CheckErr(name string, pass string) UserErrorConst {
	return UserErrorConst(u.Create(name, pass))
}

// Create attempts to create a user and returns success or failure.
// If a user allready exists results in failure.
//
// Returns
func (u *User) Create(name string, pass string) int {

	if len(name) < 5 {
		return int(LenName)
	}
	if len(pass) < 5 {
		return int(LenPass)
	}

	if u.ByName(name) {
		return int(HasName)
	}

	// make if no table
	db, err := iniC("error(user-create): loading database\n")
	if err {
		return int(CheckDB)
	}

	// salt salt hash hash
	bsalt := NewSaltCSRNG(defaultSaltSize)
	u.Name = name
	u.Salt = bytesToBase64(bsalt)
	u.Hash = bytesToBase64(GetPasswordHash(pass, bsalt))
	// fmt.Printf("--> %s, %s, %v\n", u.Name, pass, u.Salt)

	// defer db.Close()
	if err := db.Create(u).Error; err != nil {
		fmt.Printf("ERROR: %s\n", err.Error())
	}

	return 0
}

// validate checks against a provided salt and hash.
// This method does not actually look anything up from a database.
//
// Salt and Hash MUST BE PRESENT before calling!
func (u *User) validate(pass string) bool {
	result := CheckPassword(
		pass,
		fromBase64(u.Salt),
		fromBase64(u.Hash))
	return result
}

// ValidatePassword checks against a provided salt and hash.
// we use the user's [name] to find the table-record
// and then validate the password.
//
// return false on error
func (u *User) ValidatePassword(pass string) bool {
	// fmt.Println("==> ValidatePassword")
	// open the database
	db, err := iniC("error(validate-password) loading database\n")
	if err {
		return false
	}

	result := false
	tempUser := User{Name: "really?"}
	//db.LogMode(dataLogging)
	db.FirstOrInit(&tempUser, User{Name: u.Name})

	if db.RowsAffected == 0 && tempUser.Name != u.Name {
		//db.Close()
		// fmt.Println("Record not found")
		return false
	}
	fmt.Printf("User-Name: %s, id: %v\n", u.Name, u.ID)

	//defer db.Close()
	if tempUser.Name != u.Name {
		// fmt.Printf("- no user found. %v\n", tempUser)
		// may as well just return false here, right?
	} else {
		result = tempUser.validate(pass)
	}

	return result
}

// UserSession grabs a session from sessions table matching `user_id`, `host`
// and `cli-key` (is the app-id used to store SessID info).
//
// Nothing is validated, we just grab the `sessions.session` so that it
// can be reused and/or updated.
//
// returns (`Session`, `success` bool)
func (u *User) UserSession(host string, client *gin.Context) (Session, bool) {
	// fmt.Println("==> UserSession()")
	clistr := getClientString(client)
	sess := Session{}
	db, err := iniC("error(validate-session) loading database\n")
	if err {
		return sess, false
	}
	//db.LogMode(dataLogging)
	//defer db.Close()
	db.First(&sess, "[cli-key] = ? AND [host] = ? AND [user_id] = ?", clistr, host, u.ID)
	// fmt.Printf("  --> MATCH: %v\n", sess.UserID == u.ID)
	return sess, sess.UserID == u.ID
}

// ValidateSessionByUserID checks to see if a session exists in the database
// provided the `User.ID` of the current `User` record.
//
// It also checks if the session is expired.
//
// - returns `true` if the Session is valid and has not expired.
//
// - returns `false` if `User.ID` is NOT set or the Session has expired.
func (u *User) ValidateSessionByUserID(host string, client *gin.Context) bool {
	println("==> ValidateSessionByUserID")
	if u.ID == 0 {
		println("  --> User ID == 0; aborting.")
		return false
	}

	sess, success := u.UserSession(host, client)
	if !success {
		return false
	}

	if time.Now().Before(sess.Expires) {
		// fmt.Println("  --> SESSION NOT EXPIRED")
		return true
	}
	// fmt.Println("  --> SESSION EXPIRED")

	return false
}

// EnsureTableUsers creates table [users] if not exist.
func EnsureTableUsers() {
	var u User
	db, _ := iniK("error(ensure-table-users) loading db (perhaps expected)\n")
	// if !e {
	//defer db.Close()
	if !db.Migrator().HasTable(u) {
		db.Migrator().CreateTable(u)
	}
	// }
}
