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
	configuration = Configuration{
		appID:      "sessions_demo",
		Host:       "127.0.0.1",
		Port:       ":5500",
		DataSource: "./ormus.db3",
		DataSystem: "sqlite3",
	}
)

//
// we still need to provide some demo forms, however you can just use
// get or post variables such as
//
// http://localhost:5500/login/?user=admin&pass=password
// http://localhost:5500/register/?user=admin&pass=password
// http://localhost:5500/stat/
// http://localhost:5500/logout/
//

func main() {

	session.CookieDefaults(
		12,    // 12h expiration applies to cookies and sessions
		true,  // httpOnly
		false, // if true, cookies will not be visible to the client/user for example, in chrome.
	)

	// optional; ensure absolute (working) data source path for sqlite3
	// configuration.DataSource, _ = filepath.Abs(configuration.DataSource)
	session.SetDefaults(configuration.DataSource, configuration.DataSystem, -1, -1)

	gin.SetMode(gin.ReleaseMode)
	engine := gin.Default()

	configuration.initServerLogin(engine)

	// index is in _unSafeHandlers, so you must be logged in to view it.
	engine.GET("/index/", func(g *gin.Context) {
		g.String(http.StatusOK, "Hello")
	})

	fmt.Printf("using host: \"%s%s\"\n", configuration.Host, configuration.Port)
	engine.Run(fmt.Sprintf("%s%s", configuration.Host, configuration.Port))
}
