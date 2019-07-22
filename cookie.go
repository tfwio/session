package session

import (
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	cookieSecure   = false
	cookieHTTPOnly = true
	cookieAgeSec   = ConstCookieAge48H
	cookieAgeHrs   = 12
)

//
const (
	ConstCookieAge48H = /*60*60*/ 3600 * 48
	ConstCookieAge24H = /*60*60*/ 3600 * 24
	ConstCookieAge12H = /*60*60*/ 3600 * 12
)

// SetCookieMaxAge will set a cookie with our default settings.
//
// Will use `maxAge` in seconds to tell the browser
// to destroy the cookie as oppsed to using `SetCookieExpires`.
// See `CookieDefaults` in order to override default settings.
//
// Note: *Like `github.com/gogonic/gin`, we are applying `url.QueryEscape`
// `value` stored to the cookie so be sure to UnEscape the value when retrieved.*
func SetCookieMaxAge(cli *gin.Context, name string, value string, maxAge int) {
	http.SetCookie(cli.Writer, &http.Cookie{
		Name:     name,
		Value:    url.QueryEscape(value),
		MaxAge:   maxAge,
		Path:     "/",
		Secure:   cookieSecure,
		HttpOnly: cookieHTTPOnly,
	})
}

// SetCookieSessOnly will set a cookie with our default settings.
// Will expire with the browser session.
//
// See `CookieDefaults` in order to override default settings.
//
// Note: *Like `github.com/gogonic/gin`, we are applying `url.QueryEscape`
// `value` stored to the cookie so be sure to UnEscape the value when retrieved.*
func SetCookieSessOnly(cli *gin.Context, name string, value string) {
	http.SetCookie(cli.Writer, &http.Cookie{
		Name:     name,
		Value:    url.QueryEscape(value),
		Path:     "/",
		Secure:   cookieSecure,
		HttpOnly: cookieHTTPOnly,
	})
}

// SetCookieExpires will set a cookie with our default settings.
//
// Will use Expires (as opposed to `SetCookieMaxAge`).
// See `CookieDefaults` in order to override default settings.
//
// Note: *Like `github.com/gogonic/gin`, we are applying `url.QueryEscape`
// `value` stored to the cookie so be sure to UnEscape the value when retrieved.*
func SetCookieExpires(cli *gin.Context, name string, value string, expire time.Time) {
	http.SetCookie(cli.Writer, &http.Cookie{
		Name:     name,
		Value:    url.QueryEscape(value),
		Expires:  expire,
		Path:     "/",
		Secure:   cookieSecure,
		HttpOnly: cookieHTTPOnly,
	})
}

// CookieDefaults sets default cookie expire age and security.
func CookieDefaults(ageSec int, httpOnly bool, isSecure bool) {
	cookieAgeSec = ageSec
	cookieHTTPOnly = httpOnly
	cookieSecure = isSecure
}

// getCookie does what it says.  if there is an error the returned value is `nil`.
func getCookie(cname string, client *gin.Context) *http.Cookie {
	var result *http.Cookie
	if xid, e := client.Request.Cookie(cname); e == nil {
		result = xid
	}
	return result
}

// getCookieValue returns a string value if present, or an empty string.
func getCookieValue(cname string, client *gin.Context) string {
	cookie := getCookie(cname, client)
	cookieValue := ""
	if cookie != nil {
		if sessid, x := url.QueryUnescape(cookie.Value); x == nil {
			cookieValue = sessid
		}
	}
	return cookieValue
}

// cookieValue takes in a `*http.Cookie` and attempts to return
// a string value.  If no value (error), then we'll return an empty string.
func cookieValue(cookie *http.Cookie) string {
	cookieValue := ""
	if cookie != nil {
		if sessid, x := url.QueryUnescape(cookie.Value); x == nil {
			cookieValue = sessid
		}
	}
	return cookieValue
}

// QueryCookieValidate checks against a provided salt and hash.
// BUT FIRST, it checks for a valid session?
//
// - check if we have a session cookie
//
// - if so then...
func QueryCookieValidate(cookieName string, client *gin.Context) bool {

	clistr := getClientString(client)
	cookie := getCookie(cookieName, client)
	sessid := cookieValue(cookie)

	if sessid == "" {
		return false
	}

	result := false

	db, err := iniC("error(validate-session) loading database\n")
	if err {
		return false
	}
	db.LogMode(dataLogging)
	sess := Session{}
	defer db.Close()
	db.First(&sess, "[cli-key] = ? AND [host] = ? AND [sessid] = ?", clistr, cookieName, sessid)
	// fmt.Printf("SESS\nsess: %s\ncook: %s\n", sess.SessID, sessid)
	// fmt.Printf("EXPR\nsess: %v\ncook: %v\n", sess.Expires, cookie.Expires)

	if sess.SessID == sessid {
		result = time.Now().Before(sess.Expires)
		// fmt.Printf("==> SESSION IS VALID\n")
	}

	return result
}

// QueryCookie looks in `sessions` table for a matching `sess_id`
// and returns the matching `Session` if found or an empty session.
// (bool) Success value tells us if a match was found.
//
// THIS DOES NOT VALIDATE THE SESSION! IT JUST LOOKS
// FOR A SESSION ON THE GIVEN HOST!
//
// If a matching session results, may be used to determine or lookup
// the owning User.
//
// - Returns `false` on error (with an empty session).
//
// - Returns `true` on success with a Session out of our database.
func QueryCookie(host string, client *gin.Context) (Session, bool) {
	// println("==> QueryCookie")
	clistr := getClientString(client)
	cookiesess := getCookieValue(host, client)

	sess := Session{}
	if cookiesess == "" {
		return sess, false
	}
	db, err := iniC("error(validate-session) loading database\n")
	if err {
		return sess, false
	}
	db.LogMode(dataLogging)
	defer db.Close()
	db.First(&sess, "[cli-key] = ? AND [host] = ? AND [sessid] = ?", clistr, host, cookiesess)
	// fmt.Printf("  --> SESSID MATCH: %v\n", sess.SessID == cookiesess)
	return sess, sess.SessID == cookiesess
}
