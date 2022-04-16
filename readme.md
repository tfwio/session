
A session package intended to provide some foundation which may be cusomized and
implemented, perhaps useful in [github.com/gin-gonic/gin] middleware aside any other heuristic it can be wired to.

This package provides a secure logon session by utilizing a sqlite3 database via [GORM],
so easily conforms to other data-systems.

----

2022-04-16  
UPDATED to conform to quite a few updates to [GORM]

-  update gorm package references (to `gorm.io/driver/sqlite` and `gorm.io/gorm`)
- there were a few code semantic updates such as how gorm now handles logging and lack of need to `.Close()` a given iteration memory space(?).


----

**LIMITATIONS**

- freshly brewed.
- theres no Unregister function or routeHandler!

----

**GET STARTED**

See: [server example](./examples/srv).


**dataset**

users table: `users: id name salt hash`

sessions table: `sessions: id userid sessid host created expires cli-key keep-alive`

* [host] value stores what is provided to the cookie name such as `<appname><port>`.  
* [cli-key] is provided the client IP in base64.

**response handlers**

current http response handlers:  
`/login/` `/logout/` `/stat/` `/register/`  
*!unregister*

**middleware service configs**

Regular expressions are used to validate URI path for two basic heuristics.
There are two "Keys" that are configured in the enum type `Service`, namely
`Service.KeySessionIsValid` and `Service.KeySessionIsChecked` which correspond
to the following regular expression input `[]string` arrays:

- `Service.URICheck`: Regular expressions supplied here will push a boolean
  value into `gin.Context.Set(key,value)` and `.Get` dictionary indicating wether
  the response is valid.  A key "lookup" (`ctx.Get("lookup")`) value of false tells us
  that checking for a valid session wasn't required.  If true, then the (deault)
  "is-valid" key will report weather or not we have a valid session.
- `Service.URIEnforce`: Regular expressions supplied here will, if we have
  a valid session, continue to serve content.  If there is no valid session then
  it will (by default settings) abort the httpRequest and report a simple string message.

If no regexp string(s) is supplied to `Service.URICheck` or `Service.URIEnforce`
(i.e. `len(x) == 0`) then no checks are performed and you've just rendered this
service useless ;)

`Service.MatchExpHandler` default:
```golang
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
```golang
// DefaultURIAbortHandler is the default abort handler.
// It simply prints "authorization required" and serves "unauthorized" http
// response 401 aborting further processing.
func DefaultURIAbortHandler(ctx *gin.Context, ename string) {
	ctx.String(http.StatusUnauthorized, "authorization required")
	ctx.Abort()
}
```



[GORM]:                         https://gorm.io/
[github.com/gin-gonic/gin]:     https://github.com/gin-gonic/gin

