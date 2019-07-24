package main

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	_ "github.com/jinzhu/gorm/dialects/sqlite"
	_ "github.com/mattn/go-sqlite3"

	"github.com/gin-gonic/gin"
	"github.com/tfwio/session"
)

var (
	sessSvc = session.Service{
		AppID:              "sessions_demo",
		Port:               ":5500",
		DataSystem:         "sqlite3",
		DataSource:         "./ormus.db3",
		CookieHTTPOnly:     true,                       // hymmm
		CookieSecure:       false,                      // we want to see em in the browser
		KeyResponse:        session.KeyGinSessionValid, // gin-session-isValid
		AdvanceOnKeepYear:  0,                          // 0
		AdvanceOnKeepMonth: 6,                          // 6
		AdvanceOnKeepDay:   0,                          // 0
		UnsafeURI:          []string{"/index/", "/this/", "/that"},
		// CheckURIHandler:    UnsafeURIHandlerRx,
		CheckURIHandler: func(uri, unsafe string) bool {
			regexp.MatchString(fmt.Sprintf("^%s", unsafe), uri)
			return strings.Contains(uri, unsafe)
		},
		// expected form GET/POST params: "user", "pass" and "keep"
		FormSession: session.FormSession{User: "user", Pass: "pass", Keep: "keep"},
	}
)

//
// we still need to provide some demo forms, however you can just use
// get or post variables such as
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

	session.SetupService(sessSvc, engine)

	// index is in _unSafeHandlers, so you must be logged in to view it.
	engine.GET("/index/", func(g *gin.Context) {
		g.String(http.StatusOK, "Hello")
	})
	fauxHost := fmt.Sprintf("127.0.0.1%s", sessSvc.Port)
	fmt.Printf("using host: \"%s\"\n", fauxHost)
	engine.Run(fauxHost)
}
