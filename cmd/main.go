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

var urlRegexp = regexp.MustCompile("^VERSION_(.*)_URL")

func main() {
	host := envOrDefault("HOST", "")
	port := envOrDefault("PORT", "8000")
	addr := host + ":" + port

	var versions []revaboxy.Version
	for _, e := range os.Environ() {
		pair := strings.Split(e, "=")
		envName := pair[0]
		envVal := pair[1]

		if match := urlRegexp.FindStringSubmatch(envName); match != nil {
			name := match[1]
			percentageStr := os.Getenv(fmt.Sprintf("VERSION_%s_PERCENTAGE", name))
			percentage, err := strconv.ParseFloat(percentageStr, 64)
			if err != nil {
				log.Fatalln(err.Error())
			}

			u, err := url.Parse(envVal)
			if err != nil {
				log.Fatalln(err.Error())
			}

			versions = append(versions, revaboxy.Version{
				Name:       strings.ToLower(name),
				URL:        u,
				Percentage: percentage,
			})
		}
	}

	proxy, err := revaboxy.New(
		versions,
		revaboxy.UseLogger(&logger{}),
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

type logger struct{}

func (l *logger) Log(s string) {
	fmt.Println("logger: " + s)
}

func envOrDefault(name, def string) string {
	if value, ok := syscall.Getenv(name); ok {
		return value
	}
	return def
}
