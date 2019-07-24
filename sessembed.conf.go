package session

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
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
	// FormSession will collect form data.
	FormSession struct {
		User string
		Pass string
		Keep string
	}
	// Service is not required with exception to this little demo ;)
	Service struct {
		FormSession
		AppID              string
		Port               string
		CookieSecure       bool
		CookieHTTPOnly     bool
		KeyResponse        string
		AdvanceOnKeepYear  int
		AdvanceOnKeepMonth int
		AdvanceOnKeepDay   int
		UnsafeURI          []string
		CheckURIHandler    UnsafeURIHandler
	}
)

const (
	// defaultKeySessionIsValid is used in middleware to provide
	// a boolean value indicating wether or not session has
	// been checked and is valid.
	defaultKeySessionIsValid = "session.isValid"
	actionLogin              = "login"
	actionLogout             = "logout"
	actionRegister           = "register"
	actionStatus             = "status"
	formUser                 = "user"
	formPass                 = "pass"
	formKeep                 = "keep"
	defaultAdvanceYear       = 0
	defaultAdvanceMonth      = 6
	defaultAdvanceDay        = 0
	baseMatchFmt             = "^%s"
)

var (
	service = Service{
		AppID:              "sessions_demo",
		Port:               ":5500",
		CookieSecure:       false,
		CookieHTTPOnly:     true,
		AdvanceOnKeepYear:  defaultAdvanceYear,  // 0
		AdvanceOnKeepMonth: defaultAdvanceMonth, // 6
		AdvanceOnKeepDay:   defaultAdvanceDay,   // 0
		KeyResponse:        defaultKeySessionIsValid,
		UnsafeURI:          WrapURIPathString("index,this,that"),
		// this is identical to default uri-handler (set CheckURIHandler to nil for default)
		CheckURIHandler: func(uri, unsafe string) bool {
			regexp.MatchString(fmt.Sprintf(baseMatchFmt, unsafe), uri)
			return strings.Contains(uri, unsafe)
		},
		FormSession: FormSession{User: formUser, Pass: formPass, Keep: formKeep},
	}
	// SessionConfiguration is our live configuration.
	// It stores default form element names and a key that
	// will be made available to all http.Request responses
	// that are marked unsafe (to be checked for a valid sesison).
)

// AddDate uses ServiceConf defaults to push an expiration date forward.
// The `value` interface can of type `time.Time` or `session.Session`.
// If the supplied value is not valid, we'll return a given `time.Now()`.
func (s *Service) AddDate(value interface{}) time.Time {
	var result time.Time
	switch t := value.(type) {
	case *time.Time:
		result = t.AddDate(s.AdvanceOnKeepYear, s.AdvanceOnKeepMonth, s.AdvanceOnKeepDay)
		break
	case time.Time:
		result = t.AddDate(s.AdvanceOnKeepYear, s.AdvanceOnKeepMonth, s.AdvanceOnKeepDay)
		break
	case *Session:
		result = t.Created.AddDate(s.AdvanceOnKeepYear, s.AdvanceOnKeepMonth, s.AdvanceOnKeepDay)
		break
	case Session:
		result = t.Created.AddDate(s.AdvanceOnKeepYear, s.AdvanceOnKeepMonth, s.AdvanceOnKeepDay)
		break
	default:
		fmt.Fprintln(os.Stderr, "==============> ERROR: Expected time.Time or session.Session value for ServiceConf.AddDate()")
		result = time.Now()
		break
	}
	return result
}

// GetFormSession gets form values from http.Request
func GetFormSession(r *http.Request) FormSession {
	return FormSession{
		User: r.FormValue(service.User),
		Pass: r.FormValue(service.Pass),
		Keep: r.FormValue(service.Keep),
	}
}
func (f *FormSession) hasUser() bool { return f.User != "" }
func (f *FormSession) hasPass() bool { return f.Pass != "" }
func (f *FormSession) hasKeep() bool { return f.Keep != "" && (f.Keep == "true" || f.Keep == "1") }

// SetupService sets up session service.
//
// Set saltSize or hashSize to -1 to persist internal defaults.
func SetupService(value Service, engine *gin.Engine, dbsys, dbsrc string, saltSize, hashSize int) {
	service = value
	if engine != nil {
		service.attachRoutesAndMiddleware(engine)
		if service.CheckURIHandler == nil {
			service.CheckURIHandler = unsafeURIHandlerRx
		}
	}
	SetDefaults(dbsys, dbsrc, saltSize, hashSize)
}

// UnsafeURIHandlerRx uses a simple regular expression to validate
// wether or not the URI is unsafe.
func unsafeURIHandlerRx(uri, unsafe string) bool {
	regexp.MatchString(fmt.Sprintf(baseMatchFmt, unsafe), uri)
	return strings.Contains(uri, unsafe)
}

func (s *Service) isunsafe(input string) (bool, string) {
	for _, unsafe := range service.UnsafeURI {
		if s.CheckURIHandler(input, unsafe) {
			return true, unsafe
		}
	}
	return false, ""
}
