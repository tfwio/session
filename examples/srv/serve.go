package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tfwio/session"
)

// SessHost gets a simple string that is used in our sessions db
// using Configuration.appID as the root-name.
//
// Its also used as a foundation for cookie names.
func (c *Configuration) SessHost() string {
	return fmt.Sprintf("%s%s", c.appID, strings.TrimLeft(c.Port, ":"))
}

func (c *Configuration) initServerLogin(engine *gin.Engine) {
	// fmt.Println("--> LOGON SESSIONS SUPPORTED")
	engine.Use(c.sessMiddleware)
	engine.Any("/logout/", c.serveLogout)
	engine.Any("/login/", c.serveLogin)
	engine.Any("/register/", c.serveRegister)
	engine.Any("/stat/", c.serveUserStatus)
}

func (c *Configuration) sessMiddleware(g *gin.Context) {
	yn := false
	if result, name := isunsafe(g.Request.RequestURI); result {
		yn = session.QueryCookieValidate(c.SessHost(), g)
		// from here we could perhaps abort a response.
		if !yn {
			g.String(http.StatusForbidden, "ABORT(%s)!", name)
			g.Abort()
		}
	}
	// a flag to check on the status in our actual handler.
	// use `g.Get(<Key>)` from responseHandler
	g.Set(sessConfig.KeyResponse, yn)
	g.Next() // (calling this probably isn't necessary)
}

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
func (c *Configuration) serveUserStatus(g *gin.Context) {
	sh := c.SessHost()
	if sess, success := session.QueryCookie(sh, g); success {
		if u, success := sess.GetUser(); success && sess.IsValid() {
			if sess.KeepAlive {
				sess.Refresh(true)
				session.SetCookieExpires(g, sh, sess.SessID, sess.Expires)
			}
			g.JSON(http.StatusOK, &LogonModel{Action: actionStatus, Detail: "found", Status: true, Data: map[string]interface{}{"user": u.Name, "created": sess.Created, "expires": sess.Expires}})
		} else {
			g.JSON(http.StatusOK, &LogonModel{Action: actionStatus, Detail: "exists", Status: false})
		}
	} else {
		g.JSON(http.StatusOK, &LogonModel{Action: actionStatus, Detail: "none", Status: false})
	}
}

func (c *Configuration) serveLogout(g *gin.Context) {
	sh := c.SessHost()
	// fmt.Println("==> LOGOUT ATTEMPT")
	sess, success := session.QueryCookie(sh, g)
	if success {
		// fmt.Printf("  ==> CLIENT COOKIE EXISTS; USER=%d\n", sess.UserID)
		session.SetCookieDestroy(g, sh, sess.SessID)
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

func (c *Configuration) serveLogin(g *gin.Context) {

	// fmt.Println("==> LOGIN REQUEST")

	form := GetFormSession(g.Request)

	j := LogonModel{Action: actionLogin, Detail: "session creation failed.", Status: false}
	sh := c.SessHost()

	u := session.User{}
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
						session.SetCookieExpires(g, sh, sess.SessID, sess.Expires)
					} else {
						sess.KeepAlive = false
						session.SetCookieSessOnly(g, sh, sess.SessID)
					}
					sess.Save()
					session.SetCookieSessOnly(g, sh+"_xo", u.Name)
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
			session.SetCookieDestroy(g, sh, sess.SessID)
			session.SetCookieDestroy(g, sh+"_xo", "")
			j.Detail = "Session destroyed."
			j.Status = true
		}
	}
	g.JSON(http.StatusOK, j)
}

func (c *Configuration) serveRegister(g *gin.Context) {

	j := LogonModel{Action: actionRegister, Detail: "user creation failed.", Status: false}

	form := GetFormSession(g.Request)

	u := session.User{}
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

		sh := c.SessHost()
		e, sess := u.CreateSession(g, sh, form.hasKeep())
		if !e {
			if form.hasKeep() {
				session.SetCookieExpires(g, sh, sess.SessID, sess.Expires)
			} else {
				session.SetCookieSessOnly(g, sh, sess.SessID)
			}
			session.SetCookieSessOnly(g, sh+"_xo", u.Name)
			j.Status = true
			j.Detail = "User and Session created."
		} else {
			j.Status = false
			j.Detail = "User created; session failed."
		}
	}
	g.JSON(http.StatusOK, j)
}
