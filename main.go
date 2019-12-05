package main

import (
	"context"
	"context_demo/newsapi"
	"flag"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"
)

var tpl = template.Must(template.ParseFiles("templates/index.html"))

var apiKey *string

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	tpl.Execute(w, nil)
}

func SearchHandler(w http.ResponseWriter, r *http.Request) {

	var (
		ctx    context.Context
		cancel context.CancelFunc
	)
	timeout, err := time.ParseDuration("1s")
	if err == nil {
		// The request has a timeout, so create a context that is
		// canceled automatically when the timeout expires.
		ctx, cancel = context.WithTimeout(context.Background(), timeout)
	} else {
		ctx, cancel = context.WithCancel(context.Background())
	}
	defer cancel() // Cancel ctx as soon as handleSearch returns.
	query := r.FormValue("q")
	if query == "" {
		http.Error(w, "no query", http.StatusBadRequest)
		return
	}

	page := r.FormValue("page")
	if page == "" {
		page = "1"
	}

	next, err := strconv.Atoi(page)
	if err != nil {
		http.Error(w, "Unexpected server error", http.StatusInternalServerError)
		return
	}

	ctx = context.WithValue(ctx, newsapi.ApiKeyKey, *apiKey)

	start := time.Now()
	search, err := newsapi.SearchNews(ctx, query, next)
	elapsed := time.Since(start)

	search.Timeout = timeout
	search.Elapsed = elapsed

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tpl.Execute(w, search)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

}

func main() {
	apiKey = flag.String("apikey", "", "Newsapi.org access key")
	flag.Parse()
	if *apiKey == "" {
		log.Fatal("apiKey must be set")
	}
	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir("assets"))
	mux.Handle("/assets/", http.StripPrefix("/assets/", fs))
	mux.HandleFunc("/", IndexHandler)
	mux.HandleFunc("/search", SearchHandler)
	http.ListenAndServe(":80", mux)
}
