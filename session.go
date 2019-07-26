// Package session works with GORM to enable `http.Cookie`s,
// `Session` and `User` capability — along with github.com/gin-gonic/gin
package session

import (
	"time"
)

// Session represents users who are logged in.
type Session struct {
	ID        int64     `gorm:"auto_increment;unique_index;primary_key;column:id"`
	UserID    int64     `gorm:"column:user_id"` // [users].[id]
	SessID    string    `gorm:"not null;column:sessid"`
	Host      string    `gorm:"column:host"` // running multiple server instance/port(s)?
	Created   time.Time `gorm:"not null;column:created"`
	Expires   time.Time `gorm:"not null;column:expires"`
	Client    string    `gorm:"not null;column:cli-key"` // .Request.RemoteAddr
	KeepAlive bool      `gorm:"column:keep-alive"`
}

// TableName Set User's table name to be `users`
func (Session) TableName() string {
	return "sessions"
}

// IsValid returns if the session is expired.
// If `Session.ID` == 0, then it just returns `false`.
//
// This is only valid for when we've looked up the session
// through using a client-browser-cookie prior.
func (s *Session) IsValid() bool {
	if s.ID == 0 {
		return false
	}
	return time.Now().Before(s.Expires)
}

// Refresh will update the `Session.Expires` date AND
// the `SessID` with new values.
//
// Note that this does not store a http.Cookie.
//
// if save is true, the record is updated in [database].[sessions] table.
func (s *Session) Refresh(save bool) {
	s.Created = time.Now()
	s.Expires = service.AddDate(s)
	s.SessID = toUBase64(NewSaltString(defaultSaltSize))
	if save {
		s.Save()
	}
}

// GetUser gets a user by the UserID stored in the Session.
func (s *Session) GetUser() (User, bool) {
	u := User{}
	if u.ByID(s.UserID) {
		return u, true
	}
	return u, false
}

// Destroy will update the `Session.Expires` date AND
// the `SessID` with new, EXPIRED  values.
func (s *Session) Destroy(andSave bool) {
	s.Expires = time.Now()
	if andSave {
		s.Save()
	}
}

// EnsureTableSessions creates table [sessions] if not exist.
func EnsureTableSessions() {
	var s Session
	db, _ := iniK("error(ensure-table-sessions) loading db; (expected)\n")
	// if !e {
	defer db.Close()
	if !db.HasTable(s) {
		db.CreateTable(s)
	}
	// }
}

// Save session data to db.
func (s *Session) Save() bool {
	db, err := iniC("error(validate-session) loading database\n")
	if err {
		return false
	}
	defer db.Close()
	db.Save(s)
	return db.RowsAffected > 0
}

// ListSessions returns a list of all sessions.
//
// The method first fetches a list of User elements
// then reports the Sessions with user-data (name).
func ListSessions() ([]Session, int) {
	sessions := []Session{}
	db, err := iniC("error(session-cli-list) loading db\n")
	defer db.Close()
	if !err {
		db.Find(&sessions)
	}
	return sessions, len(sessions)
}
