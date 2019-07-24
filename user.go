package session

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
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
	defer db.Close()
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
func (u *User) ByName(name string) bool {
	fmt.Printf("--> looking for %s\n", name)

	db, err := iniC("error(user-by-name) loading database\n")
	result := false
	defer db.Close()
	if !err {
		db.Where("[user] = ?", name).First(u)
		if u.Name == name {
			result = true
		}
	}
	// fmt.Printf("!-> FOUND %s, %d\n", u.Name, u.ID)
	return result
}

// ByID gets a user by [id].
func (u *User) ByID(id int64) bool {
	db, err := iniC("error(user-by-id) loading database\n")
	defer db.Close()
	if !err {
		db.FirstOrInit(u, User{ID: id})
	}
	return u.ID == id
}

// CreateSession Save a session into the sessions table.
//
// The method is written to utilize gin-gonic/gin `*gin.Context` as
// its suggested input interface given that we can use it to retrieve
// the ClientIP() and store that value to our database in order to
// validate a given user-session.
//
// Currently, a user is limited to one session (connection) on one client.
//
// FIXME: we should be checking if there is a existing record in sessions table
// and re-using it for the user executing UPDATE as opposed to CREATE.
func (u *User) CreateSession(r interface{}, host string) (bool, Session) {

	t := time.Now()
	result := false
	sess := Session{
		Host:    host,
		UserID:  u.ID,
		SessID:  toUBase64(NewSaltString(defaultSaltSize)),
		Created: t,
		Expires: t.AddDate(0, defaultCookieAgeMonths, 0),
	}

	sess.Client = getClientString(r)

	db, err := iniC("error(create-session) loading database\n")

	defer db.Close()
	if !err {
		db.Create(&sess)
		if db.RowsAffected == 1 {
			result = true
		}
	}

	return result, sess
}

// Create attempts to create a user and returns success or failure.
// If a user allready exists results in failure.
//
// Returns
// (-1) `db.Open`,
// (2) `User.Name` or `User.Pass` < `len(5)`
// (1) `User.Name` exists
// (0) on success
func (u *User) Create(name string, pass string) int {

	if len(name) < 5 {
		return 2
	}
	if len(pass) < 5 {
		return 2
	}

	if u.ByName(name) {
		return 1
	}

	db, err := iniC("error(user-create): loading database\n")
	if err {
		return -1
	}

	bsalt := NewSaltCSRNG(defaultSaltSize)
	u.Name = name
	u.Salt = bytesToBase64(bsalt)
	u.Hash = bytesToBase64(GetPasswordHash(pass, bsalt))
	// fmt.Printf("--> %s, %s, %v\n", u.Name, pass, u.Salt)

	defer db.Close()
	db.Create(u)

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
func (u *User) ValidatePassword(pass string) bool {
	// fmt.Println("==> ValidatePassword")
	// open the database
	db, err := iniC("error(validate-password) loading database\n")
	if err {
		return false
	}

	result := false
	tempUser := User{Name: "really?"}
	db.LogMode(dataLogging)
	db.FirstOrInit(&tempUser, User{Name: u.Name})

	if db.RowsAffected == 0 && tempUser.Name != u.Name {
		db.Close()
		// fmt.Println("Record not found")
		return false
	}

	defer db.Close()
	if tempUser.Name != u.Name {
		// fmt.Printf("- no user found. %v\n", tempUser)
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
	db.LogMode(dataLogging)
	defer db.Close()
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
	defer db.Close()
	if !db.HasTable(u) {
		db.CreateTable(u)
	}
	// }
}
