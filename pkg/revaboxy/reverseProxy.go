package revaboxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

const DefaultName = "default"

const CookieName = "revaboxy"

const HeaderName = "revaboxy"

type Version struct {
	Name       string
	URL        *url.URL
	Percentage float64
}

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

type SettingChanger func(s *settings)

func UseLogger(l Logger) SettingChanger {
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
		cookie, _ := req.Cookie(fmt.Sprintf("%s-name", CookieName))

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
		name := r.Request.Header.Get(fmt.Sprintf("%s-name", HeaderName))
		existingCookie, _ := r.Request.Cookie(fmt.Sprintf("%s-name", CookieName))

		if name != "" && (existingCookie == nil || versions.get(existingCookie.Value) != nil) {
			newCookie := &http.Cookie{
				Name:     fmt.Sprintf("%s-name", CookieName),
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
		name := r.Header.Get(fmt.Sprintf("%s-name", HeaderName))
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

	req.Header.Add(fmt.Sprintf("%s-name", HeaderName), targetVersion.Name)
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
