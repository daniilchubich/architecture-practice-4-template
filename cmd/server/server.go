package main

import (
	"context"
	"flag"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/roman-mazur/architecture-practice-4-template/httptools"
	"github.com/roman-mazur/architecture-practice-4-template/signal"
)

var port = flag.Int("port", 8080, "server port")
var dbUrl = flag.String("db-url", "db:8100", "db url")
var delay = flag.Duration("delay", 0, "response delay")

const confHealthFailure = "CONF_HEALTH_FAILURE"

func main() {
	h := new(http.ServeMux)
	createTeam()

	h.HandleFunc("/health", func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("content-type", "text/plain")
		if failConfig := os.Getenv(confHealthFailure); failConfig == "true" {
			rw.WriteHeader(http.StatusInternalServerError)
			_, _ = rw.Write([]byte("FAILURE"))
		} else {
			rw.WriteHeader(http.StatusOK)
			_, _ = rw.Write([]byte("OK"))
		}
	})

	report := make(Report)

	h.HandleFunc("/api/v1/some-data", func(rw http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		ctx, cancel := context.WithTimeout(r.Context(), time.Duration(10)*time.Second)
		defer cancel()
		
		fwdRequest := r.Clone(ctx)
		fwdRequest.RequestURI = ""
		fwdRequest.URL.Host = *dbUrl
		fwdRequest.Host = *dbUrl
		fwdRequest.URL.Scheme = "http"
		fwdRequest.URL.Path = "/db/" + key

		resp, err := http.DefaultClient.Do(fwdRequest)
		if *delay > 0 && *delay < 300 {
			time.Sleep(time.Duration(*delay) * time.Millisecond)
		}
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		if resp.StatusCode == http.StatusBadRequest {
			body, err := ioutil.ReadAll(resp.Body)
			if err == nil && string(body) == "record does not exist\n" {
				rw.WriteHeader(http.StatusNotFound)
				return
			}
		}
	
		report.Process(r)
	
		rw.WriteHeader(resp.StatusCode)
		rw.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
		rw.Header().Set("Content-Length", resp.Header.Get("Content-Length"))
		io.Copy(rw, resp.Body)
		resp.Body.Close()
	})

	h.Handle("/report", report)

	server := httptools.CreateServer(*port, h)
	server.Start()
	signal.WaitForTerminationSignal()
}

func createTeam() {
	formData := url.Values{}
	formData.Set("value", time.Now().Format("2006-01-02"))

	resp, err := http.PostForm("http" + "://" + *dbUrl + "/db/" + "gods", formData)
	if err != nil || resp.StatusCode != http.StatusOK {
		panic("Error occured when initializing DB")
	}
}
