package revaboxy

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func mustURLParse(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}
	return u
}

type testRoundTripper struct {
	hostAnswer map[string]string
}

func (rt *testRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	answer, ok := rt.hostAnswer[req.URL.Host]
	if !ok {
		return nil, errors.New("could not find host")
	}
	return &http.Response{
		Header:     make(http.Header),
		Request:    req,
		Body:       ioutil.NopCloser(strings.NewReader(answer)),
		StatusCode: http.StatusOK,
	}, nil
}

func TestReverseProxyWithCookies(t *testing.T) {
	// Setup the server
	url1, _ := url.Parse("http://url1.test/path1")
	url2, _ := url.Parse("http://url2.test/path2")

	cookieJar, _ := cookiejar.New(nil)

	proxy, err := New(
		[]Version{
			{
				Name:        DefaultName,
				URL:         url1,
				Probability: 0,
			},
			{
				Name:        "test2",
				URL:         url2,
				Probability: 0.5,
			},
		},
		WithTransport(&testRoundTripper{
			hostAnswer: map[string]string{
				"url1.test": "url1-data",
				"url2.test": "url2-data",
			},
		}),
		WithCookieName("custom"),
	)
	if err != nil {
		t.Fatal("could not create proxy", err)
	}

	// Make the first request
	recorder := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "http://baseurl.com/basepath", nil)
	if err != nil {
		t.Fatal("could not create request", err)
	}
	proxy.ServeHTTP(recorder, req)

	if expected, real := http.StatusOK, recorder.Code; expected != real {
		t.Fatalf("expected status code %v, got %v", expected, real)
	}

	firstBody, err := ioutil.ReadAll(recorder.Body)
	if err != nil {
		t.Fatal("could not read body")
	}

	cookieJar.SetCookies(req.URL, recorder.Result().Cookies())

	// Makes sure that subsequent requests uses the same path
	for i := 0; i < 10; i++ {
		recorder := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodGet, "http://baseurl.com/basepath", nil)
		for _, c := range cookieJar.Cookies(req.URL) {
			req.AddCookie(c)
		}
		if err != nil {
			t.Fatal("could not create request", err)
		}
		proxy.ServeHTTP(recorder, req)
		if expected, real := http.StatusOK, recorder.Code; expected != real {
			t.Fatalf("expected status code %v, got %v", expected, real)
		}
		body, err := ioutil.ReadAll(recorder.Body)
		if err != nil {
			t.Fatal("could not read body")
		}

		if expected, real := firstBody, body; !bytes.Equal(expected, real) {
			t.Fatalf("expected body to be equal to the first body (%s) but got %s", expected, real)
		}

		cookieJar.SetCookies(req.URL, recorder.Result().Cookies())
	}
}

func TestReverseProxyWithUnavailableVersion(t *testing.T) {
	// Setup the server
	defaultURL, _ := url.Parse("http://defaulturl.test/path1")
	failURL, _ := url.Parse("http://faulurl.test/path1")

	proxy, err := New(
		[]Version{
			{
				Name:        DefaultName,
				URL:         defaultURL,
				Probability: 0,
			},
			{
				Name:        "shouldfail",
				URL:         failURL,
				Probability: 0.999,
			},
		},
		WithTransport(&testRoundTripper{
			hostAnswer: map[string]string{
				"defaulturl.test": "defaulturl-data",
			},
		}),
	)
	if err != nil {
		t.Fatal("could not create proxy", err)
	}

	// Make the first request
	recorder := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "http://faulurl.test/path1", nil)
	if err != nil {
		t.Fatal("could not create request", err)
	}
	proxy.ServeHTTP(recorder, req)

	if expected, real := http.StatusOK, recorder.Code; expected != real {
		t.Fatalf("expected status code %v, got %v", expected, real)
	}
}

func TestNew(t *testing.T) {
	url, _ := url.Parse("http://an-url")

	tests := []struct {
		name     string
		versions []Version
		settings []Setting
		wantErr  bool
	}{
		{
			name: "valid",
			versions: []Version{
				{
					Name:        DefaultName,
					URL:         url,
					Probability: 0,
				},
				{
					Name:        "ab-test-1",
					URL:         url,
					Probability: 0.5,
				},
			},
			wantErr: false,
		},
		{
			name: "no default",
			versions: []Version{
				{
					Name:        "ab-test-1",
					URL:         url,
					Probability: 0.5,
				},
			},
			wantErr: true,
		},
		{
			name: "too much probability",
			versions: []Version{
				{
					Name:        DefaultName,
					URL:         url,
					Probability: 0.6,
				},
				{
					Name:        "ab-test-1",
					URL:         url,
					Probability: 0.5,
				},
			},
			wantErr: true,
		},
		{
			name: "multiple with same name",
			versions: []Version{
				{
					Name:        DefaultName,
					URL:         url,
					Probability: 0.3,
				},
				{
					Name:        "ab-test-1",
					URL:         url,
					Probability: 0.3,
				},
				{
					Name:        "ab-test-1",
					URL:         url,
					Probability: 0.4,
				},
			},
			wantErr: true,
		},
		{
			name: "not 100%",
			versions: []Version{
				{
					Name:        DefaultName,
					URL:         url,
					Probability: 0.3,
				},
				{
					Name:        "ab-test-1",
					URL:         url,
					Probability: 0.3,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.versions, tt.settings...)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_modifyRequestUrl(t *testing.T) {

	settings := &settings{
		headerName: "test",
	}

	tests := []struct {
		name    string
		reqURL  string
		version *Version
		wantURL *url.URL
	}{
		{
			name:   "no path",
			reqURL: "http://example.com/",
			version: &Version{
				URL: mustURLParse("http://test.com/"),
			},
			wantURL: mustURLParse("http://test.com/"),
		},
		{
			name:   "no path 2",
			reqURL: "http://example.com",
			version: &Version{
				URL: mustURLParse("http://test.com"),
			},
			wantURL: mustURLParse("http://test.com/"),
		},
		{
			name:   "req path",
			reqURL: "http://example.com/test",
			version: &Version{
				URL: mustURLParse("http://test.com"),
			},
			wantURL: mustURLParse("http://test.com/test"),
		},
		{
			name:   "version path",
			reqURL: "http://example.com/",
			version: &Version{
				URL: mustURLParse("http://test.com/test"),
			},
			wantURL: mustURLParse("http://test.com/test/"),
		},
		{
			name:   "both path",
			reqURL: "http://example.com/test1",
			version: &Version{
				URL: mustURLParse("http://test.com/test2"),
			},
			wantURL: mustURLParse("http://test.com/test2/test1"),
		},
		{
			name:   "both path 2",
			reqURL: "http://example.com/test1/",
			version: &Version{
				URL: mustURLParse("http://test.com/test2/"),
			},
			wantURL: mustURLParse("http://test.com/test2/test1/"),
		},
		{
			name:   "query",
			reqURL: "http://example.com/test1?test1=test1",
			version: &Version{
				URL: mustURLParse("http://test.com/test2?test2=test2"),
			},
			wantURL: mustURLParse("http://test.com/test2/test1?test2=test2&test1=test1"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, tt.reqURL, nil)
			if err != nil {
				t.Fatal("could not parse url", err)
			}

			modifyRequest(settings, req, tt.version)

			if *req.URL != *tt.wantURL {
				t.Errorf("modifyRequest url = %s, wantURL = %s", req.URL.String(), tt.wantURL.String())
				return
			}
		})
	}
}

type testLogger struct {
	logs int
}

func (l *testLogger) Printf(string, ...interface{}) {
	l.logs++
}

func Test_WithLogger(t *testing.T) {

	l := &testLogger{}
	rt := &testRoundTripper{
		hostAnswer: map[string]string{},
	}

	proxy, err := New(
		[]Version{
			{
				Name:        DefaultName,
				URL:         mustURLParse("http://example.com"),
				Probability: 0,
			},
			{
				Name:        "test",
				URL:         mustURLParse("http://example.com"),
				Probability: 1,
			},
		},
		WithTransport(rt),
		WithLogger(l),
	)
	if err != nil {
		t.Fatal("should not error when creating revaboxy")
	}

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
	proxy.ServeHTTP(rec, req)

	if real, expected := l.logs, 2; real != expected {
		t.Fatalf("expected %v logs, got %v", expected, real)
	}
}

type savingRoundtripper struct {
	req *http.Request
}

func (rt *savingRoundtripper) RoundTrip(req *http.Request) (*http.Response, error) {
	rt.req = &(*req)

	return &http.Response{
		Header:     make(http.Header),
		Request:    req,
		Body:       ioutil.NopCloser(strings.NewReader("test answer")),
		StatusCode: http.StatusOK,
	}, nil
}

func Test_WithHeader(t *testing.T) {
	rt := &savingRoundtripper{}

	const newHeaderName = "Test-Name"

	proxy, err := New(
		[]Version{
			{
				Name:        DefaultName,
				URL:         mustURLParse("http://example.com"),
				Probability: 1,
			},
		},
		WithHeaderName(newHeaderName),
		WithTransport(rt),
	)
	if err != nil {
		t.Fatal("should not error when creating revaboxy")
	}

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
	proxy.ServeHTTP(rec, req)

	if real, expected := rt.req.Header.Get(newHeaderName), DefaultName; real != expected {
		t.Fatalf(`expected "%s" header to be "%s", got "%s"`, newHeaderName, expected, real)
	}
}

func durationAbs(d time.Duration) time.Duration {
	if d < 0 {
		return -d
	}
	return d
}

func Test_WithCookieExpiry(t *testing.T) {
	rt := &savingRoundtripper{}

	expiryDuration := time.Hour * 4

	proxy, err := New(
		[]Version{
			{
				Name:        DefaultName,
				URL:         mustURLParse("http://example.com"),
				Probability: 1,
			},
		},
		WithCookieExpiry(expiryDuration),
		WithTransport(rt),
	)
	if err != nil {
		t.Fatal("should not error when creating revaboxy")
	}

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
	proxy.ServeHTTP(rec, req)

	cookies := rec.Result().Cookies()

	if real, expected := len(cookies), 1; real != expected {
		t.Fatalf(`expected %d cookies, got %d"`, real, expected)
	}

	cookie := cookies[0]
	diffDuration := cookie.Expires.Add(-expiryDuration).Sub(time.Now())
	if durationAbs(diffDuration) > time.Second {
		t.Fatalf(`expected cookie to have expiry set correctly`)
	}
}
