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
		AppID:      "sessions_demo",
		Host:       "127.0.0.1",
		Port:       ":5500",
		DataSource: "./ormus.db3",
		DataSystem: "sqlite3",
		Conf: session.ServiceConf{
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
			FormSession: session.FormSession{User: "user", Pass: "pass", Keep: "keep"}},
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

	session.CookieDefaults(
		12,    // 12h expiration [db].[sessions] table, not so much cookies.
		true,  // httpOnly
		false, // if true, cookies will not be visible to the client/user for example, in chrome.
	)

	// optional; ensure absolute (working) data source path for sqlite3
	// configuration.DataSource, _ = filepath.Abs(configuration.DataSource)
	session.SetDefaults(sessSvc.DataSystem, sessSvc.DataSource, -1, -1)

	gin.SetMode(gin.ReleaseMode)
	engine := gin.Default()

	sessSvc.AttachRoutesAndMiddleware(engine)

	// index is in _unSafeHandlers, so you must be logged in to view it.
	engine.GET("/index/", func(g *gin.Context) {
		g.String(http.StatusOK, "Hello")
	})

	fmt.Printf("using host: \"%s%s\"\n", sessSvc.Host, sessSvc.Port)
	engine.Run(fmt.Sprintf("%s%s", sessSvc.Host, sessSvc.Port))
}
