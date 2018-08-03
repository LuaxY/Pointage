package adp

import (
	"Pointage/config"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/http/httputil"
)

var client = httpClient()

func httpClient() *http.Client {
	jar, err := cookiejar.New(&cookiejar.Options{})

	if err != nil {
		log.Fatal(err)
	}

	return &http.Client{
		Jar: jar,
	}
}

func dumpRequest(req *http.Request) {
	if config.DebugMode {
		dump, _ := httputil.DumpRequest(req, true)
		log.Println(string(dump))
	}

}

func dumpResponse(res *http.Response) {
	if config.DebugMode {
		dump, _ := httputil.DumpResponse(res, true)
		log.Println(string(dump))
	}
}
