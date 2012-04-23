package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"regexp"
)

var (
	tmplHome, tmplView, tmplUser, tmplBomView *template.Template
)

func baseHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	log.Printf("serving %s\n", r.URL.Path)

	bomUrlPattern := regexp.MustCompile("^/([a-zA-Z][a-zA-Z0-9_]*)/([a-zA-Z][a-zA-Z0-9_]*)(/.*)$")
	userUrlPattern := regexp.MustCompile("^/([a-zA-Z][a-zA-Z0-9_]*)/$")

	switch {
	case r.URL.Path == "/":
		err = homeController(w, r)
	case bomUrlPattern.MatchString(r.URL.Path):
		match := bomUrlPattern.FindStringSubmatch(r.URL.Path)
		err = bomController(w, r, match[1], match[2], match[3])
	case userUrlPattern.MatchString(r.URL.Path):
		match := userUrlPattern.FindStringSubmatch(r.URL.Path)
		err = userController(w, r, match[1], "")
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

func homeController(w http.ResponseWriter, r *http.Request) (err error) {
	context := make(map[string]interface{})
	context["BomList"], err = bomstore.ListBoms("")
	if err != nil {
		return
	}
	err = tmplHome.Execute(w, context)
	return
}

func userController(w http.ResponseWriter, r *http.Request, user, extra string) (err error) {
	if !isShortName(user) {
		http.Error(w, "invalid username: "+user, 400)
		return
	}
	var email string
	email, err = auth.GetEmail(user)
	if err != nil {
		// no such user
		http.NotFound(w, r)
		return
	}
	context := make(map[string]interface{})
	context["BomList"], err = bomstore.ListBoms(ShortName(user))
	context["Email"] = email
	context["UserName"] = user
	if err != nil {
		return
	}
	err = tmplUser.Execute(w, context)
	return
}

func bomController(w http.ResponseWriter, r *http.Request, user, name, extra string) (err error) {
	if !isShortName(user) {
		http.Error(w, "invalid username: "+user, 400)
		return
	}
	if !isShortName(name) {
		http.Error(w, "invalid bom name: "+name, 400)
		return
	}
	context := make(map[string]interface{})
	context["BomMeta"], context["Bom"], err = bomstore.GetHead(ShortName(user), ShortName(name))
	if err != nil {
		return
	}
	err = tmplBomView.Execute(w, context)
	return
}

func serveCmd() {
	var err error

	// load and parse templates
	baseTmplPath := *templatePath + "/base.html"
	tmplHome = template.Must(template.ParseFiles(*templatePath+"/home.html", baseTmplPath))
	tmplUser = template.Must(template.ParseFiles(*templatePath+"/user.html", baseTmplPath))
	tmplBomView = template.Must(template.ParseFiles(*templatePath+"/bom_view.html", baseTmplPath))
	if err != nil {
		log.Fatal(err)
	}

	openBomStore()
	openAuthStore()

	// serve template static assets (images, CSS, JS)
	http.Handle("/static/", http.FileServer(http.Dir(*templatePath+"/")))
	http.Handle("/favicon.ico", http.FileServer(http.Dir(*templatePath+"/static/")))

	// fall through to default handler
	http.HandleFunc("/", baseHandler)

	listenString := fmt.Sprintf("%s:%d", *listenHost, *listenPort)
	http.ListenAndServe(listenString, nil)
	fmt.Println("Serving at " + listenString)
}
