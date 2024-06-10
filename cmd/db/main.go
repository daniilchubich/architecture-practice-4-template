package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"strings"

	"github.com/roman-mazur/architecture-practice-4-template/db/datastore"
	"github.com/roman-mazur/architecture-practice-4-template/httptools"
	"github.com/roman-mazur/architecture-practice-4-template/signal"
)

var (
	port = flag.Int("port", 8100, "server port")
	db   *datastore.Db
)

func main() {
	flag.Parse()
	var err error
	db, err = datastore.NewDb("./out")
	if err != nil {
		panic(err)
	}
	startServer()
	signal.WaitForTerminationSignal()
}

func startServer() {
	handler := http.NewServeMux()
	handler.HandleFunc("/db/", handleDb)
	server := httptools.CreateServer(*port, handler)
	server.Start()
}

func handleDb(rw http.ResponseWriter, r *http.Request) {
	key := strings.TrimPrefix(r.URL.Path, "/db/")
	switch r.Method {
	case http.MethodGet:
		data, err := get(key)
		sendResponse(rw, data, err)
	case http.MethodPost:
		value := r.FormValue("value")
		err := put(key, value)
		sendResponse(rw, nil, err)
	default:
		http.Error(rw, "This method is not allowed", http.StatusMethodNotAllowed)
	}
}

func sendResponse(rw http.ResponseWriter, data interface{}, err error) {
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
	} else if data != nil {
		if err := json.NewEncoder(rw).Encode(data); err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}
	}
}

func get(key string) (interface{}, error) {
	value, err := db.Get(key)
	if err != nil {
		return nil, err
	}
	return struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}{key, value}, nil
}

func put(key, value string) error {
	if value == "" {
		return fmt.Errorf("can't save empty value")
	}
	return db.Put(key, value)
}