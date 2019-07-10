package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"syscall"

	"github.com/lindell/revaboxy/internal/time"

	"github.com/lindell/revaboxy/pkg/revaboxy"
)

func main() {
	host := envOrDefault("HOST", "")
	port := envOrDefault("PORT", "80")
	addr := host + ":" + port

	versions, err := versionsFromEnvVars()
	if err != nil {
		log.Fatal(err)
	}

	// Configuring settings
	settings := []revaboxy.Setting{}
	settings = append(settings, revaboxy.WithLogger(log.New(os.Stdout, "", log.Ldate|log.Ltime|log.LUTC)))
	if headerName, ok := syscall.Getenv("HEADER_NAME"); ok {
		settings = append(settings, revaboxy.WithHeaderName(headerName))
	}
	if cookieName, ok := syscall.Getenv("COOKIE_NAME"); ok {
		settings = append(settings, revaboxy.WithCookieName(cookieName))
	}
	if cookieExpiryStr, ok := syscall.Getenv("COOKIE_EXPIRY"); ok {
		cookieExpiry, err := time.ParseDuration(cookieExpiryStr)
		if err != nil {
			log.Fatal("could not parse cookie expiry", err)
		}
		settings = append(settings, revaboxy.WithCookieExpiry(cookieExpiry))
	}

	proxy, err := revaboxy.New(
		versions,
		settings...,
	)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("listen to %s", addr)
	err = http.ListenAndServe(host+":"+port, proxy)
	if err != nil {
		log.Fatal(err)
	}
}

var urlRegexp = regexp.MustCompile("^VERSION_(.*)_URL$")

func versionsFromEnvVars() ([]revaboxy.Version, error) {
	var versions []revaboxy.Version
	for _, e := range os.Environ() {
		pair := strings.Split(e, "=")
		envName := pair[0]
		envVal := pair[1]

		if match := urlRegexp.FindStringSubmatch(envName); match != nil {
			name := match[1]
			probabilityStr := os.Getenv(fmt.Sprintf("VERSION_%s_PROBABILITY", name))
			probability, err := strconv.ParseFloat(probabilityStr, 64)
			if err != nil {
				return nil, fmt.Errorf(`could not parse %s probability "%s"`, name, probabilityStr)
			}

			u, err := url.Parse(envVal)
			if err != nil {
				return nil, fmt.Errorf(`could not parse %s url "%s"`, name, envVal)
			}

			versions = append(versions, revaboxy.Version{
				Name:        strings.ToLower(name),
				URL:         u,
				Probability: probability,
			})
		}
	}

	return versions, nil
}

func envOrDefault(name, def string) string {
	if value, ok := syscall.Getenv(name); ok {
		return value
	}
	return def
}
