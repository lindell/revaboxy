package revaboxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

// Revaboxy creates an A/B test between different urls.
// It does also save which test i run on the client through cookies and serve the same version on subsequent requests
// The revaboxy handler should be created with New
type Revaboxy struct {
	reverseProxy *httputil.ReverseProxy
}

// DefaultName is the name of the default version
// The version with this name will take up the rest of the probability if any remain
// when adding all probabilitys together
const DefaultName = "default"

// Version is one of the versions used in the A/B/C... test
type Version struct {
	// The name of the a/b testing version
	Name string
	// The URL to the root of the target
	URL *url.URL
	// The probability from 0-1 of this version being used
	Probability float64
}

// Logger is the logger interface used with revaboxy
type Logger interface {
	Printf(string, ...interface{})
}

type settings struct {
	logger     Logger
	headerName string

	cookieName   string
	cookieExpiry time.Duration

	roundTripper http.RoundTripper
}

// Setting changes the revaboxy settings
type Setting func(s *settings)

// WithLogger sets the logger to be used
func WithLogger(l Logger) Setting {
	return func(s *settings) {
		s.logger = l
	}
}

// WithTransport sets the http transport to be used
func WithTransport(rt http.RoundTripper) Setting {
	return func(s *settings) {
		s.roundTripper = rt
	}
}

// WithHeaderName sets the name of the header that is set for the downsteam service to use
// The header will contain the name of the version used
// If the value is not set with this setting, it will default to the default, "revaboxy-name"
func WithHeaderName(headerName string) Setting {
	return func(s *settings) {
		s.headerName = headerName
	}
}

// WithCookieName sets the name of the cookie that where the selected version is stored
// If the value is not set with this setting, it will default to the default, "revaboxy-name"
func WithCookieName(cookieName string) Setting {
	return func(s *settings) {
		s.cookieName = cookieName
	}
}

// WithCookieExpiry sets the expiry time of the client cookie, default is 3 days
func WithCookieExpiry(expiry time.Duration) Setting {
	return func(s *settings) {
		s.cookieExpiry = expiry
	}
}

// New creates a revaboxy client. Versions required but, any number of additional settings may be provided
func New(vv []Version, settingChangers ...Setting) (*Revaboxy, error) {
	// Default values
	settings := &settings{
		logger:       &nopLogger{},
		headerName:   "Revaboxy-Name",
		cookieName:   "revaboxy-name",
		cookieExpiry: time.Hour * 24 * 7,
		roundTripper: http.DefaultTransport,
	}
	// Apply all settings
	for _, s := range settingChangers {
		s(settings)
	}

	logger := settings.logger

	// Add all versions
	versions := versions{}
	for _, v := range vv {
		err := versions.add(v)
		if err != nil {
			return nil, err
		}
	}
	if err := versions.valid(); err != nil {
		return nil, err
	}

	// The director changes the request. If the user has already been assigned a version, that one will be used.
	// Otherwise a random version will be assigned to the user
	director := func(req *http.Request) {
		cookie, _ := req.Cookie(settings.cookieName)

		if cookie != nil {
			version, ok := versions[cookie.Value]
			if ok {
				logger.Printf("using previus used version %s", version.Name)
				modifyRequest(settings, req, version)
			} else {
				logger.Printf("could not use previus version %s and using a random version instead", cookie.Value)
				modifyRequest(settings, req, versions.getRandomVersion())
			}
		} else {
			logger.Printf("new request, using random version")
			modifyRequest(settings, req, versions.getRandomVersion())
		}
	}

	// Add a cookie to the response that
	modifyResponse := func(r *http.Response) error {
		name := r.Request.Header.Get(settings.headerName)
		existingCookie, _ := r.Request.Cookie(settings.cookieName)

		if name != "" && (existingCookie == nil || versions.get(existingCookie.Value) != nil) {
			newCookie := &http.Cookie{
				Name:    settings.cookieName,
				Value:   name,
				Path:    "/",
				Expires: time.Now().Add(settings.cookieExpiry),
			}
			r.Header.Add("Set-Cookie", newCookie.String())
		}

		return nil
	}

	// Make sure a failed request (by not reaching the host) to a version that is not
	// the default one is redirected to the default one
	defaultReverseProxy := httputil.NewSingleHostReverseProxy(versions[DefaultName].URL)
	defaultReverseProxy.Transport = settings.roundTripper
	errorHandler := func(w http.ResponseWriter, r *http.Request, err error) {
		name := r.Header.Get(settings.headerName)
		if name != "" && name != DefaultName {
			logger.Printf("could not connect to %s, using default instead", name)
			defaultReverseProxy.ServeHTTP(w, r)
			return
		}

		w.WriteHeader(http.StatusBadGateway)
	}

	return &Revaboxy{
		reverseProxy: &httputil.ReverseProxy{
			Director:       director,
			ModifyResponse: modifyResponse,
			ErrorHandler:   errorHandler,
			Transport:      settings.roundTripper,
		},
	}, nil
}

func modifyRequest(s *settings, req *http.Request, targetVersion *Version) {
	url := targetVersion.URL
	targetQuery := url.RawQuery

	req.URL.Scheme = url.Scheme
	req.URL.Host = url.Host
	req.URL.Path = singleJoiningSlash(url.Path, req.URL.Path)
	if targetQuery == "" || req.URL.RawQuery == "" {
		req.URL.RawQuery = targetQuery + req.URL.RawQuery
	} else {
		req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
	}
	if _, ok := req.Header["User-Agent"]; !ok {
		// explicitly disable User-Agent so it's not set to default value
		req.Header.Set("User-Agent", "")
	}

	req.Header.Add(s.headerName, targetVersion.Name)
}

func (revaboxy *Revaboxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	revaboxy.reverseProxy.ServeHTTP(w, r)
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

type nopLogger struct{}

func (m *nopLogger) Printf(string, ...interface{}) {}
