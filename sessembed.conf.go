package session

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
)

type (
	// URIMatchHandler is used to see if an incoming URI is one that requires a valid session.
	//
	// param uri is the uri being checked.
	//
	// param unsafe is the node in our array of unsafe uri-strings.
	URIMatchHandler func(string, string) bool
	// URIAbortHandler is used to customize how we abort a requestHandler
	// when UriEnforce is postured to abort when a user is not logged in.
	//
	// The string parameter is the regular expression or validation input that
	// was used to URI-check in order service the abort.
	URIAbortHandler func(*gin.Context, string)
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
		AppID               string
		Port                string
		CookieSecure        bool
		CookieHTTPOnly      bool
		KeySessionIsValid   string
		KeySessionIsChecked string
		AdvanceOnKeepYear   int
		AdvanceOnKeepMonth  int
		AdvanceOnKeepDay    int
		// supply a uri-path token such as "/json/" to check.
		// We supply a `KeySessionIsValid` for the responseHandler
		// to utilize to handle the secure content manually.
		URICheck []string
		// Unlike URICheck, we'll abort a response for any URI
		// path provided to this list if user is not logged in.
		URIEnforce      []string
		VerboseCheck    bool
		URIMatchHandler URIMatchHandler
		URIAbortHandler URIAbortHandler
	}
)

const (
	// defaultKeySessionIsValid is used in middleware to provide
	// a boolean value indicating wether or not session has
	// been checked and is valid.
	defaultKeySessionIsValid   = "is-valid"
	defaultKeySessionIsChecked = "is-checked"
	actionLogin                = "login"
	actionLogout               = "logout"
	actionRegister             = "register"
	actionStatus               = "status"
	actionUnregister           = "unregister" // not implemented yet
	baseMatchFmt               = "^%s"
)

var (
	service *Service
	// SessionConfiguration is our live configuration.
	// It stores default form element names and a key that
	// will be made available to all http.Request responses
	// that are marked unsafe (to be checked for a valid sesison).
)

// DefaultService creates/returns a default session service configuration
// with no URIEnforce or URICheck definitions.
//
// This construct can be further configured, then supplied to
// the call to SetupService.
func DefaultService() *Service {
	return &Service{
		AppID:               "session",
		Port:                ":5500",
		CookieSecure:        false,
		CookieHTTPOnly:      true,
		AdvanceOnKeepYear:   0,
		AdvanceOnKeepMonth:  6,
		AdvanceOnKeepDay:    0,
		KeySessionIsValid:   defaultKeySessionIsValid,
		KeySessionIsChecked: defaultKeySessionIsChecked,
		URIEnforce:          []string{},
		URICheck:            []string{},
		// this is identical to default uri-handler (set URIMatchHandler to nil for default)
		VerboseCheck: false,
		FormSession:  FormSession{User: "user", Pass: "pass", Keep: "keep"},
	}
}

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
func SetupService(value *Service, engine *gin.Engine, dbsys, dbsrc string, saltSize, hashSize int) {
	service = value
	if engine != nil {

		service.attachRoutesAndMiddleware(engine)

		if service.URIMatchHandler == nil {
			fmt.Fprintln(os.Stderr, "<session:URIMatchHandler> callback was nil; using default regexp validation.")
			service.URIMatchHandler = DefaultURIMatchHandler
		}
		if service.URIAbortHandler == nil {
			fmt.Fprintln(os.Stderr, "<session:URIMatchHandler> callback was nil; using default abort handler.")
			service.URIAbortHandler = DefaultURIAbortHandler
		}
	}
	SetDefaults(dbsys, dbsrc, saltSize, hashSize)
}

// DefaultURIMatchHandler uses a simple regular expression to validate
// wether or not the URI session is to be validated.
func DefaultURIMatchHandler(uri, expression string) bool {
	if match, err := regexp.MatchString(expression, uri); err == nil {
		return match
	}
	return false
}

// DefaultURIAbortHandler is the default abort handler.
// It simply prints "authorization required" and serves "unauthorized" http
// response 401 aborting further processing.
func DefaultURIAbortHandler(ctx *gin.Context, ename string) {
	ctx.String(http.StatusUnauthorized, "authorization required")
	ctx.Abort()
}

func (s *Service) isunsafe(input string, inputs ...string) (bool, string) {
	for _, unsafe := range inputs {
		if s.URIMatchHandler(input, unsafe) {
			return true, unsafe
		}
	}
	return false, ""
}
