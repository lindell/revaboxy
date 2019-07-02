package revaboxy_test

import (
	"log"
	"net/http"
	"net/url"

	"github.com/lindell/revaboxy/pkg/revaboxy"
)

func ExampleNew() {
	defaultURL, _ := url.Parse("http://defaulturl")
	greenBackgroundURL, _ := url.Parse("http://greenbackgroundurl")

	rp, err := revaboxy.New([]revaboxy.Version{
		{
			revaboxy.DefaultName,
			defaultURL,
			0.7,
		},
		{
			"green-background",
			greenBackgroundURL,
			0.3,
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Fatal(http.ListenAndServe(":8080", rp))
}
