
A session package intended to provide some foundation which may be cusomized and
implemented, perhaps useful in [github.com/gin-gonic/gin] middleware aside any other heuristic it can be wired to.

This package provides a secure logon session by utilizing a sqlite3 database via [GORM],
so easily conforms to other data-systems.

----

**limitations**

- freshly brewed.
- [*todo/feature*] One session on one client (browser session / IP) is allowed per User once initial session is created.  (can easily be modified)  
  Will likely fix this soon.

----

**getting started**

this is the source code as found in the [server example](./examples/srv).

```golang
package main

import (
	"fmt"
	"net/http"

	_ "github.com/jinzhu/gorm/dialects/sqlite"
	_ "github.com/mattn/go-sqlite3"

	"github.com/gin-gonic/gin"
	"github.com/tfwio/session"
)

var (
	service = session.Service{
		AppID:              "sessions_demo",
		Port:               ":5500",    // Port is used for
		CookieHTTPOnly:     true,       // hymmm
		CookieSecure:       false,      // we want to see em in the browser
		KeySessionIsValid:  "is-valid", // default: session.isValid
		AdvanceOnKeepYear:  0,
		AdvanceOnKeepMonth: 6,
		AdvanceOnKeepDay:   0,
		// if regexp matches (our default check/handler), the httpResponse is aborted
		// with a simple message.
		//
		// you could just use a string array.
		URIEnforce: session.WrapURIExpression("^/index/?$,^/this/?$,^/that"),
		// if one of these matches, we pass (using gin) `gin.Client.Set("is-valid", true)`
		// so that in our handler we can check its status (`c.Get("is-valid")`) and handle accordingly.
		// The **actual** key used ("is-valid") is stored to `Service.KeySessionIsValid`.
		URICheck:        []string{},
		URIMatchHandler: nil, // use default
		URIAbortHandler: nil, // use default
		// defaults the requestHandlers use to look up form values.
		FormSession:     session.FormSession{User: "user", Pass: "pass", Keep: "keep"},
	}
)

//
// we still need to provide some demo forms, however you can just use
// get or post variables to view the xhr/json response(s).
//
// http://localhost:5500/login/?user=admin&pass=password
// http://localhost:5500/login/?user=admin&pass=password&keep=true
// http://localhost:5500/register/?user=admin&pass=password
// http://localhost:5500/register/?user=admin&pass=password&keep=true
// http://localhost:5500/stat/
// http://localhost:5500/logout/
//

func main() {

	// optional; ensure absolute (working) data source path for sqlite3
	// configuration.DataSource, _ = filepath.Abs(configuration.DataSource)

	gin.SetMode(gin.ReleaseMode)
	engine := gin.Default()

	session.SetupService(&service, engine, "sqlite3", "./ormus.db", -1, -1)
	// at this point you can override the crypto settings
	// session.OverrideCrypto(...)

	// this "index" is defined in service.URIEnforce,
	// so you must be logged in to view it.
	engine.GET("/index/", func(g *gin.Context) {
		g.String(http.StatusOK, "Hello")
	})
	fauxHost := fmt.Sprintf("127.0.0.1%s", service.Port)
	fmt.Printf("using host: \"%s\"\n", fauxHost)
	engine.Run(fauxHost)
}
```

**dataset**

users table: `users: id name salt hash`

sessions table: `sessions: id userid sessid host created expires cli-key keep-alive`

* [host] value stores what is provided to the cookie name.  
* [cli-key] is provided the client IP.

**response handlers**

current http response handlers:  
`/login/` `/logout/` `/stat/` `/register/`  
*!unregister*

**middleware service configs**

Regular expressions are used to validate URI path for two basic heuristics.

- `Service.URICheck []string`: Regular expressions supplied here will push a boolean
  value into `gin.Context.Set(key,value)` and `.Get` dictionary indicating wether
  the response is valid.  A key "lookup" (`ctx.Get("lookup")`) value of false tells us
  that checking for a valid session wasn't required.  If true, then the (deault)
  "is-valid" key will report weather or not we have a valid session.
- `Service.URIEnforce []string`: Regular expressions supplied here will, if we have
  a valid session, continue to serve content.  If there is no valid session then
  it will (by default settings) abort the httpRequest and report a simple string message.

If no regexp string(s) is supplied to `Service.URICheck` or `Service.URIEnforce`
(i.e. `len(x) == 0`) then no checks are performed and you've just rendered this
service useless ;)

`Service.MatchExpHandler` default:
```
// DefaultMatchExpHandler uses a simple regular expression to validate
// wether or not the URI session is to be validated.
func DefaultMatchExpHandler(uri, expression string) bool {
	if match, err := regexp.MatchString(expression, uri); err == nil {
		return match
	}
	return false
}
```

`Service.URIAbortHandler` default:
```
// DefaultURIAbortHandler is the default abort handler.
// It simply prints "authorization required" and serves "unauthorized" http
// response 401 aborting further processing.
func DefaultURIAbortHandler(ctx *gin.Context, ename string) {
	ctx.String(http.StatusUnauthorized, "authorization required")
	ctx.Abort()
}
```



[GORM]:                         https://github.com/jinzhu/gorm
[github.com/gin-gonic/gin]:     https://github.com/gin-gonic/gin

