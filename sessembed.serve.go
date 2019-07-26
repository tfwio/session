package session

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// SessHost gets a simple string that is used in our sessions db
// using Configuration.appID as the root-name.
//
// Its also used as a foundation for cookie names.
func (s *Service) SessHost() string {
	return fmt.Sprintf("%s%s", s.AppID, strings.TrimLeft(s.Port, ":"))
}

// attachRoutesAndMiddleware is called to connect gin.Engine to middleware and
// /logout/, /login/, /register/ and /stat/ URI.
func (s *Service) attachRoutesAndMiddleware(engine *gin.Engine) {
	// fmt.Println("--> LOGON SESSIONS SUPPORTED")
	engine.Use(s.sessMiddleware)
	engine.Any("/logout/", s.serveLogout)
	engine.Any("/login/", s.serveLogin)
	engine.Any("/register/", s.serveRegister)
	engine.Any("/stat/", s.serveUserStatus)
}

func (s *Service) sessMiddleware(g *gin.Context) {

	var (
		enforce, check bool
		ename, cname   string
	)
	issecure := false
	if len(s.URICheck) > 0 {
		enforce, ename = s.isunsafe(g.Request.RequestURI, s.URICheck...)
	}
	if len(s.URIEnforce) > 0 {
		check, cname = s.isunsafe(g.Request.RequestURI, s.URIEnforce...)
	}
	lookup := enforce || check // do we need to check?
	if lookup {
		issecure := QueryCookieValidate(s.SessHost(), g)
		g.Set(s.KeyResponse, issecure)
	}
	if enforce && !issecure { // abort response.
		g.String(http.StatusForbidden, "ABORT(%s)!", ename)
		g.Abort()
	}
	fmt.Fprintf(os.Stderr,
		"check[%s]: %v, enforce[%s]: %v, result: %v, URI: %s\n",
		cname, check, ename, enforce, issecure, g.Request.RequestURI)
	// a flag to check on the status in our actual handler.
	// use `g.Get(<Key>)` from responseHandler
	g.Next() // (calling this probably isn't necessary)
}

const tfmt = "2016-01-02 03:04 PM"

// serveUserStatus serves JSON checking if a session exists,
// persists, and a user exists.
// `{status: true,  detail: "found", data: <username>}` if all checks out,
// `{status: false, detail: "exists"}` if not logged in and
// `{status: false, detail: "none"}` if no user was found.
//
// In addition to checking the user status, if the user is logged in
// and "keep" is set to true, then we'll extend the lifetime of the session
// thereby living up to the "Keep Alive" semantic.
//
// There may well be other ways of keeping sessions alive, however
// most Javascript/HTML applications will check the stats in order
// to supply access to Login, Logout and Register form/functions,
// so this seems like a decent semantic for now.
func (s *Service) serveUserStatus(g *gin.Context) {
	sh := s.SessHost()
	if sess, success := QueryCookie(sh, g); success {
		fmt.Fprintf(os.Stderr, "HOST: %s, CRD: %s, EXP: %s\n", sh, sess.Created.Format(tfmt), sess.Expires.Format(tfmt))
		u, success := sess.GetUser()
		isvalid := sess.IsValid()
		fmt.Fprintf(os.Stderr, "gotuser: %v, isvalid: %v\n", success, isvalid)
		if success && isvalid {
			if sess.KeepAlive {
				sess.Refresh(true)
				SetCookieExpires(g, sh, sess.SessID, sess.Expires)
			}
			g.JSON(http.StatusOK, &LogonModel{Action: actionStatus, Detail: "found", Status: true, Data: map[string]interface{}{"user": u.Name, "created": sess.Created, "expires": sess.Expires}})
		} else {
			g.JSON(http.StatusOK, &LogonModel{Action: actionStatus, Detail: "exists", Status: false})
		}
	} else {
		g.JSON(http.StatusOK, &LogonModel{Action: actionStatus, Detail: "none", Status: false})
	}
}

func (s *Service) serveLogout(g *gin.Context) {
	sh := s.SessHost()
	// fmt.Println("==> LOGOUT ATTEMPT")
	sess, success := QueryCookie(sh, g)
	if success {
		// fmt.Printf("  ==> CLIENT COOKIE EXISTS; USER=%d\n", sess.UserID)
		SetCookieDestroy(g, sh, sess.SessID)
		if time.Now().Before(sess.Expires) {
			// fmt.Printf("  --> NOT EXPIRED; USER=%d\n", sess.UserID)
			g.JSON(http.StatusOK, &LogonModel{Action: actionLogout, Detail: "Session exists; logged out.", Status: true})
		} else {
			// fmt.Printf("  --> SESSION EXP: %v\n", sess.Expires)
			g.JSON(http.StatusOK, &LogonModel{Action: actionLogout, Detail: "User was logged out prior; Logout re-enforced.", Status: false})
		}
		sess.Expires = time.Now()
		sess.KeepAlive = false
		sess.Save()
	} else {
		// fmt.Printf("==> SESSION NOT EXIST; NOTHING TO DO")
		g.JSON(http.StatusOK, &LogonModel{Action: actionLogout, Detail: "Session not exist; nothing to do.", Status: false})
	}
}

func (s *Service) serveLogin(g *gin.Context) {

	// fmt.Println("==> LOGIN REQUEST")

	form := GetFormSession(g.Request)

	j := LogonModel{Action: actionLogin, Detail: "session creation failed.", Status: false}
	sh := s.SessHost()

	u := User{}
	if !u.ByName(form.User) {

		// println("  --> USER NOT FOUND!")
		j.Detail = "No user record."
		j.Status = false

	} else {
		// We have a valid user;
		sess, success := u.UserSession(sh, g)
		if success {
			// fmt.Println("  ==> FOUND USER, VALIDATING PW")
			if form.hasPass() {
				if u.ValidatePassword(form.Pass) {
					// fmt.Println("  ==> PW:GOOD")
					sess.Refresh(false)
					if form.hasKeep() {
						sess.KeepAlive = true
						SetCookieExpires(g, sh, sess.SessID, sess.Expires)
					} else {
						sess.KeepAlive = false
						SetCookieSessOnly(g, sh, sess.SessID)
					}
					sess.Save()
					SetCookieSessOnly(g, sh+"_xo", u.Name)
					j.Detail = "Logged in."
					j.Status = true
					j.Data = map[string]interface{}{"user": u.Name, "created": sess.Created, "expires": sess.Expires}
				} else {
					// fmt.Println("  ==> PW:FAIL")
					j.Detail = "Password did not match."
					j.Status = true
				}
			}
		} else {
			// There is no session for the user.
			// this use case shouldn't exist since a session is created when a user is created!
			// this should report success to spite the fact that its a failure.
			// fmt.Println("  ==> DESTROY SESSION")
			sess.Destroy(true)
			SetCookieDestroy(g, sh, sess.SessID)
			SetCookieDestroy(g, sh+"_xo", "")
			j.Detail = "Session destroyed."
			j.Status = true
		}
	}
	g.JSON(http.StatusOK, j)
}

func (s *Service) serveRegister(g *gin.Context) {

	j := LogonModel{Action: actionRegister, Detail: "user creation failed.", Status: false}

	form := GetFormSession(g.Request)

	u := User{}
	if e := u.Create(form.User, form.Pass); e != 0 {
		switch e {
		case -1:
			j.Detail = "Failed to load db."
			break
		case 1:
			j.Detail = "User record already exists."
			break
		case 2:
			j.Detail = "Chek Name and Pass length; should be >= 5 chars."
			break
		}
	} else {

		sh := s.SessHost()
		e, sess := u.CreateSession(g, sh, form.hasKeep())
		if !e {
			if form.hasKeep() {
				SetCookieExpires(g, sh, sess.SessID, sess.Expires)
			} else {
				SetCookieSessOnly(g, sh, sess.SessID)
			}
			SetCookieSessOnly(g, sh+"_xo", u.Name)
			j.Status = true
			j.Detail = "User and Session created."
		} else {
			j.Status = false
			j.Detail = "User created; session failed."
		}
	}
	g.JSON(http.StatusOK, j)
}
