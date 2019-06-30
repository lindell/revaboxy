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

	"github.com/lindell/revaboxy/pkg/revaboxy"
)

func main() {
	host := envOrDefault("HOST", "")
	port := envOrDefault("PORT", "8000")
	addr := host + ":" + port

	versions, err := versionsFromEnvVars()
	if err != nil {
		log.Fatal(err)
	}

	proxy, err := revaboxy.New(
		versions,
		revaboxy.WithLogger(log.New(os.Stdout, "", log.Ldate|log.Ltime|log.LUTC)),
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
				return nil, err
			}

			u, err := url.Parse(envVal)
			if err != nil {
				return nil, err
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
