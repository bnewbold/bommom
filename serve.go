package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"regexp"
)

var (
	tmplHome, tmplView *template.Template
)

func baseHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	log.Printf("serving %s\n", r.URL.Path)

	bomUrlPattern := regexp.MustCompile("^/([a-zA-Z][a-zA-Z0-9_]*)/([a-zA-Z][a-zA-Z0-9_]*)/(.*)$")
	userUrlPattern := regexp.MustCompile("^/([a-zA-Z][a-zA-Z0-9_]*)/$")

	switch {
	case r.URL.Path == "/":
		err = tmplHome.Execute(w, nil)
	case bomUrlPattern.MatchString(r.URL.Path):
		match := bomUrlPattern.FindStringSubmatch(r.URL.Path)
		bomController(w, r, match[1], match[2], match[3])
	case userUrlPattern.MatchString(r.URL.Path):
		match := userUrlPattern.FindStringSubmatch(r.URL.Path)
		fmt.Fprintf(w, "will show BOM list here for user %s", match[1])
	default:
		// 404
		log.Println("warning: 404")
		http.NotFound(w, r)
		return
	}
	if err != nil {
		// this could cause multiple responses?
		http.Error(w, "Internal error (check logs)", 500)
		log.Println("error, 500: " + err.Error())
	}
}

func bomController(w http.ResponseWriter, r *http.Request, user, name, extra string) {
	fmt.Fprintf(w, "will show BOM %s/%s here?", user, name)
}

func serveCmd() {
	var err error

	// load and parse templates
    baseTmplPath := *templatePath + "/base.html"
	tmplHome = template.Must(template.ParseFiles(*templatePath + "/home.html", baseTmplPath))
	if err != nil {
		log.Fatal(err)
	}

	// serve template static assets (images, CSS, JS)
	http.Handle("/static/", http.FileServer(http.Dir(*templatePath+"/")))
	http.Handle("/favicon.ico", http.FileServer(http.Dir(*templatePath+"/static/")))

	// fall through to default handler
	http.HandleFunc("/", baseHandler)

	listenString := fmt.Sprintf("%s:%d", *listenHost, *listenPort)
	http.ListenAndServe(listenString, nil)
	fmt.Println("Serving at " + listenString)
}
