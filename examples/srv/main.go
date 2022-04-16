package main

import (
	"fmt"
	"net/http"

	_ "gorm.io/driver/sqlite"

	"github.com/gin-gonic/gin"
	"github.com/tfwio/session"
)

var (
	service = session.Service{
		AppID:               "sessions_demo",
		Port:                ":5500",      // Port is used for
		CookieHTTPOnly:      true,         // hymmm
		CookieSecure:        false,        // we want to see em in the browser
		KeySessionIsValid:   "is-valid",   // is also what is used by default
		KeySessionIsChecked: "is-checked", // is also what is used by default
		AdvanceOnKeepYear:   0,
		AdvanceOnKeepMonth:  6,
		AdvanceOnKeepDay:    0,
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
		FormSession: session.FormSession{User: "user", Pass: "pass", Keep: "keep"},
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
