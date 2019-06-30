package revaboxy

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

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

	cookies := recorder.Result().Cookies()

	// Makes sure that subsequent requests uses the same path
	for i := 0; i < 10; i++ {
		recorder := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodGet, "http://baseurl.com/basepath", nil)
		for _, c := range cookies {
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

		cookies = recorder.Result().Cookies()
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
