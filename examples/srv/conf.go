package main

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/tfwio/session"
)

type (
	// UnsafeURIHandler is used to see if an incoming URI is one that requires a valid session.
	//
	// param uri is the uri being checked.
	//
	// param unsafe is the node in our array of unsafe uri-strings.
	UnsafeURIHandler func(uri, unsafe string) bool
	// LogonModel responds to a login action such as "/login/" or (perhaps) "/login-refresh/"
	LogonModel struct {
		Action string      `json:"action"`
		Status bool        `json:"status"`
		Detail string      `json:"detail"`
		Data   interface{} `json:"data,omitempty"`
	}
	// Configuration is not required with exception to this little demo ;)
	Configuration struct {
		appID      string
		Host       string
		Port       string
		DataSystem string
		DataSource string
	}
	// FormSession will collect form data.
	FormSession struct {
		User string
		Pass string
		Keep string
	}
	// SessConfig provides some configuration elements.
	// Its used to store key names for form values here.
	SessConfig struct {
		FormSession
		KeyResponse        string
		AdvanceOnKeepYear  int
		AdvanceOnKeepMonth int
		AdvanceOnKeepDay   int
		UnsafeURI          []string
		CheckURIHandler    UnsafeURIHandler
	}
)

const (
	// KeyGinSessionValid is used in middleware to provide
	// a boolean value indicating wether or not session has
	// been checked and is valid.
	KeyGinSessionValid  = "gin-session-isValid"
	actionLogin         = "login"
	actionLogout        = "logout"
	actionRegister      = "register"
	actionStatus        = "status"
	formUser            = "user"
	formPass            = "pass"
	formKeep            = "keep"
	defaultAdvanceYear  = 0
	defaultAdvanceMonth = 6
	defaultAdvanceDay   = 0
	baseMatchFmt        = "^%s"
)

var (
	// SessionConfiguration is our live configuration.
	// It stores default form element names and a key that
	// will be made available to all http.Request responses
	// that are marked unsafe (to be checked for a valid sesison).
	sessConfig = SessConfig{
		KeyResponse:        KeyGinSessionValid,
		AdvanceOnKeepYear:  defaultAdvanceYear,
		AdvanceOnKeepMonth: defaultAdvanceMonth,
		AdvanceOnKeepDay:   defaultAdvanceDay,
		UnsafeURI:          wrapup(strings.Split("index,this,that", ",")...),
		CheckURIHandler:    UnsafeURIHandlerRx,
		FormSession:        FormSession{User: formUser, Pass: formPass, Keep: formKeep}}
)

// AddDate uses SessConfig defaults to push an expiration date forward.
// The `value` interface can of type `time.Time` or `session.Session`.
// If the supplied value is not valid, we'll return a given `time.Now()`.
func (s *SessConfig) AddDate(value interface{}) time.Time {
	var result time.Time
	switch t := value.(type) {
	case *time.Time:
	case time.Time:
		result = t.AddDate(s.AdvanceOnKeepYear, s.AdvanceOnKeepMonth, s.AdvanceOnKeepDay)
		break
	case *session.Session:
	case session.Session:
		result = t.Created.AddDate(s.AdvanceOnKeepYear, s.AdvanceOnKeepMonth, s.AdvanceOnKeepDay)
		break
	default:
		fmt.Fprintln(os.Stderr, "==============> ERROR: Expected time.Time or session.Session value for SessConfig.AddDate()")
		result = time.Now()
		break
	}
	return result
}

// GetFormSession gets form values from http.Request
func GetFormSession(r *http.Request) FormSession {
	return FormSession{
		User: r.FormValue(sessConfig.User),
		Pass: r.FormValue(sessConfig.Pass),
		Keep: r.FormValue(sessConfig.Keep),
	}
}
func (f *FormSession) hasUser() bool { return f.User != "" }
func (f *FormSession) hasPass() bool { return f.Pass != "" }
func (f *FormSession) hasKeep() bool { return f.Keep != "" && (f.Keep == "true" || f.Keep == "1") }

// OverrideSessionConfig I'm not sure why I put this here.
// We could just explicitly set it or a variable within.
func OverrideSessionConfig(c SessConfig) {
	sessConfig = c
}

// UnsafeURIHandlerRx uses a simple regular expression to validate
// wether or not the URI is unsafe.
func UnsafeURIHandlerRx(uri, unsafe string) bool {
	regexp.MatchString(fmt.Sprintf(baseMatchFmt, unsafe), uri)
	return strings.Contains(uri, unsafe)
}

func isunsafe(input string) (bool, string) {
	for _, unsafe := range sessConfig.UnsafeURI {
		if sessConfig.CheckURIHandler(input, unsafe) {
			return true, unsafe
		}
	}
	return false, ""
}

func wrapup(inputs ...string) []string {
	data := inputs
	for i, handler := range data {
		data[i] = strings.TrimRight(WReapLeft("/", handler), "/")
		// println(data[i])
	}
	return data
}
