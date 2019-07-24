// +build sessembed

package session

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
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
	// ServiceConf provides some configuration elements.
	// Its used to store key names for form values here.
	ServiceConf struct {
		FormSession
		KeyResponse        string
		AdvanceOnKeepYear  int
		AdvanceOnKeepMonth int
		AdvanceOnKeepDay   int
		UnsafeURI          []string
		CheckURIHandler    UnsafeURIHandler
	}
	// Service is not required with exception to this little demo ;)
	Service struct {
		Conf       ServiceConf
		AppID      string
		Host       string
		Port       string
		DataSystem string
		DataSource string
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
	service = Service{
		Conf: ServiceConf{
			KeyResponse:        KeyGinSessionValid,
			AdvanceOnKeepYear:  defaultAdvanceYear,
			AdvanceOnKeepMonth: defaultAdvanceMonth,
			AdvanceOnKeepDay:   defaultAdvanceDay,
			UnsafeURI:          wrapup(strings.Split("index,this,that", ",")...),
			CheckURIHandler:    unsafeURIHandlerRx,
			FormSession:        FormSession{User: formUser, Pass: formPass, Keep: formKeep}},
	}
	// SessionConfiguration is our live configuration.
	// It stores default form element names and a key that
	// will be made available to all http.Request responses
	// that are marked unsafe (to be checked for a valid sesison).
)

// AddDate uses ServiceConf defaults to push an expiration date forward.
// The `value` interface can of type `time.Time` or `session.Session`.
// If the supplied value is not valid, we'll return a given `time.Now()`.
func (s *ServiceConf) AddDate(value interface{}) time.Time {
	var result time.Time
	switch t := value.(type) {
	case *time.Time:
	case time.Time:
		result = t.AddDate(s.AdvanceOnKeepYear, s.AdvanceOnKeepMonth, s.AdvanceOnKeepDay)
		break
	case *Session:
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
		User: r.FormValue(service.Conf.User),
		Pass: r.FormValue(service.Conf.Pass),
		Keep: r.FormValue(service.Conf.Keep),
	}
}
func (f *FormSession) hasUser() bool { return f.User != "" }
func (f *FormSession) hasPass() bool { return f.Pass != "" }
func (f *FormSession) hasKeep() bool { return f.Keep != "" && (f.Keep == "true" || f.Keep == "1") }

// OverrideSessionConfig I'm not sure why I put this here.
// We could just explicitly set it or a variable within.
func (s *Service) OverrideSessionConfig(c ServiceConf) {
	s.Conf = c
}

// OverrideSessionConfig2 lets us explicitly override all
// Session Service values.
func (s *Service) OverrideSessionConfig2(
	advY, advM, advD int,
	frmUser, frmPass, frmKeep string,
	rKey string,
	uriHandler UnsafeURIHandler,
	unsafeURI ...string) {
	if advY != -1 {
		s.Conf.AdvanceOnKeepYear = advY
	}
	if advM != -1 {
		s.Conf.AdvanceOnKeepMonth = advM
	}
	if advD != -1 {
		s.Conf.AdvanceOnKeepDay = advD
	}
	s.Conf.KeyResponse = rKey
	s.Conf.UnsafeURI = unsafeURI
	s.Conf.CheckURIHandler = uriHandler
	s.Conf.FormSession = FormSession{User: "user", Pass: "pass", Keep: "keep"}
}

// UnsafeURIHandlerRx uses a simple regular expression to validate
// wether or not the URI is unsafe.
func unsafeURIHandlerRx(uri, unsafe string) bool {
	regexp.MatchString(fmt.Sprintf(baseMatchFmt, unsafe), uri)
	return strings.Contains(uri, unsafe)
}

func (s *Service) isunsafe(input string) (bool, string) {
	for _, unsafe := range service.Conf.UnsafeURI {
		if s.Conf.CheckURIHandler(input, unsafe) {
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
