package revaboxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

// DefaultName is the name of the default version
// The version with this name will take up the rest of the percentage if any remain
// when adding all percentages together
const DefaultName = "default"

const cookieName = "revaboxy"

const headerName = "revaboxy"

// Version is one of the versions used in the A/B/C... test
type Version struct {
	// The name of the a/b testing version
	Name string
	// The URL to the root of the target
	URL *url.URL
	// The percentage from 0-1 of this version being used
	Percentage float64
}

// Logger is the logger interface used with revaboxy
type Logger interface {
	Log(string)
}

type settings struct {
	logger Logger
}

func (s *settings) Log(log string) {
	if s.logger != nil {
		s.logger.Log(log)
	}
}

// SettingChanger changes the revaboxy settings
type SettingChanger func(s *settings)

// WithLogger sets the logger to be used
func WithLogger(l Logger) SettingChanger {
	return func(s *settings) {
		s.logger = l
	}
}

// New ...
func New(vv []Version, settingChangers ...SettingChanger) (*httputil.ReverseProxy, error) {
	settings := &settings{}
	for _, s := range settingChangers {
		s(settings)
	}

	// Add all versions
	versions := versions{}
	for _, v := range vv {
		versions.add(v)
	}
	if err := versions.valid(); err != nil {
		return nil, err
	}

	// The director changes the request. If the user has already been assigned a version, that one will be used.
	// Otherwise a random version will be assigned to the user
	director := func(req *http.Request) {
		cookie, _ := req.Cookie(fmt.Sprintf("%s-name", cookieName))

		if cookie != nil {
			version, ok := versions[cookie.Value]
			if ok {
				settings.Log(fmt.Sprintf("using previus used version %s", version.Name))
				modifyRequest(req, version)
			} else {
				settings.Log(fmt.Sprintf("could not use previus version %s and using a random version instead", cookie.Value))
				modifyRequest(req, versions.getRandomVersion())
			}
		} else {
			settings.Log(fmt.Sprintf("new request, using random version"))
			modifyRequest(req, versions.getRandomVersion())
		}
	}

	// Add a cookie to the response that
	modifyResponse := func(r *http.Response) error {
		name := r.Request.Header.Get(fmt.Sprintf("%s-name", headerName))
		existingCookie, _ := r.Request.Cookie(fmt.Sprintf("%s-name", cookieName))

		if name != "" && (existingCookie == nil || versions.get(existingCookie.Value) != nil) {
			newCookie := &http.Cookie{
				Name:     fmt.Sprintf("%s-name", cookieName),
				Value:    name,
				Path:     "/",
				Expires:  time.Now().Add(time.Hour * 24 * 3),
				HttpOnly: true,
			}
			r.Header.Add("Set-Cookie", newCookie.String())
		}

		return nil
	}

	// Make sure a failed request (by not reaching the host) to a version that is not
	// the default one is redirected to the default one
	defaultReverseProxy := httputil.NewSingleHostReverseProxy(versions[DefaultName].URL)
	errorHandler := func(w http.ResponseWriter, r *http.Request, err error) {
		name := r.Header.Get(fmt.Sprintf("%s-name", headerName))
		if name != "" && name != DefaultName {
			settings.Log(fmt.Sprintf("could not connect to %s, using default instead", name))
			defaultReverseProxy.ServeHTTP(w, r)
			return
		}

		w.WriteHeader(http.StatusBadGateway)
	}

	return &httputil.ReverseProxy{
		Director:       director,
		ModifyResponse: modifyResponse,
		ErrorHandler:   errorHandler,
	}, nil
}

func modifyRequest(req *http.Request, targetVersion *Version) {
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

	req.Header.Add(fmt.Sprintf("%s-name", headerName), targetVersion.Name)
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
