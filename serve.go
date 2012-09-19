package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"regexp"
    "path/filepath"
    "time"
)

var (
	tmplHome, tmplView, tmplUser, tmplBomView, tmplBomUpload *template.Template
)

func baseHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	log.Printf("serving %s\n", r.URL.Path)

	bomUrlPattern := regexp.MustCompile("^/([a-zA-Z][a-zA-Z0-9_]*)/([a-zA-Z][a-zA-Z0-9_]*)/$")
	bomUploadUrlPattern := regexp.MustCompile("^/([a-zA-Z][a-zA-Z0-9_]*)/([a-zA-Z][a-zA-Z0-9_]*)/_upload/$")
	userUrlPattern := regexp.MustCompile("^/([a-zA-Z][a-zA-Z0-9_]*)/$")

	switch {
	case r.URL.Path == "/":
		err = homeController(w, r)
	case bomUploadUrlPattern.MatchString(r.URL.Path):
		match := bomUploadUrlPattern.FindStringSubmatch(r.URL.Path)
		err = bomUploadController(w, r, match[1], match[2])
	case bomUrlPattern.MatchString(r.URL.Path):
		match := bomUrlPattern.FindStringSubmatch(r.URL.Path)
		err = bomController(w, r, match[1], match[2])
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

func bomController(w http.ResponseWriter, r *http.Request, user, name string) (err error) {
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
		http.Error(w, "404 couldn't open bom: " + user + "/" + name, 404)
		return nil
	}
	err = tmplBomView.Execute(w, context)
	return
}

func bomUploadController(w http.ResponseWriter, r *http.Request, user, name string) (err error) {

	if !isShortName(user) {
		http.Error(w, "invalid username: "+user, 400)
		return
	}
	if !isShortName(name) {
		http.Error(w, "invalid bom name: "+name, 400)
		return
	}
	context := make(map[string]interface{})
    context["user"] = ShortName(user)
    context["name"] = ShortName(name)
	context["BomMeta"], context["Bom"], err = bomstore.GetHead(ShortName(user), ShortName(name))

    switch r.Method {
	case "POST":


        err := r.ParseMultipartForm(1024*1024*2)
        if err != nil {
            log.Println(err)
            http.Error(w, err.Error(), 400)
            return nil
        }
        file, fileheader, err := r.FormFile("bomfile")
        if err != nil {
            log.Println(err)
            context["error"] = "bomfile was nil!"
            err = tmplBomUpload.Execute(w, context)
            return err
        }
        if file == nil {
            log.Println("bomfile was nil")
            context["error"] = "bomfile was nil!"
            err = tmplBomUpload.Execute(w, context)
            return err
        }
        versionStr := r.FormValue("version")
        if len(versionStr) == 0 || isShortName(versionStr) == false {
            context["error"] = "Version must be specified and a ShortName!"
            context["version"] = versionStr
            err = tmplBomUpload.Execute(w, context)
            return err
        }

        //contentType := fileheader.Header["Content-Type"][0]
        var b *Bom
        var bm *BomMeta

        switch filepath.Ext(fileheader.Filename) {
            case ".json":
                bm, b, err = LoadBomFromJSON(file)
                if err != nil {
                    context["error"] = "Problem loading JSON file"
                    err = tmplBomUpload.Execute(w, context)
                    return err
                }
            case ".csv":
                b, err = LoadBomFromCSV(file)
                bm = &BomMeta{}
                if err != nil {
                    context["error"] = "Problem loading XML file"
                    err = tmplBomUpload.Execute(w, context)
                    return err
                }
            case ".xml":
                bm, b, err = LoadBomFromXML(file)
                if err != nil {
                    context["error"] = "Problem loading XML file"
                    err = tmplBomUpload.Execute(w, context)
                    return err
                }
            default:
                context["error"] = "Unknown file type: " + string(fileheader.Filename)
                err = tmplBomUpload.Execute(w, context)
                return err
        }
        bm.Owner = user 
        bm.Name = name
        b.Progeny = "File uploaded from " + fileheader.Filename
        b.Created = time.Now()
        b.Version = string(versionStr)
        if err := bomstore.Persist(bm, b, ShortName(versionStr)); err != nil {
            context["error"] = "Problem saving to datastore: " + err.Error()
            err = tmplBomUpload.Execute(w, context)
        }
        http.Redirect(w, r, "//" + user + "/" + name + "/", 302)
	case "GET":
        err = tmplBomUpload.Execute(w, context)
    default:
        http.Error(w, "bad method", 405)
        return nil
    }
	return
}


func serveCmd() {
	var err error

	// load and parse templates
	baseTmplPath := *templatePath + "/base.html"
	tmplHome = template.Must(template.ParseFiles(*templatePath+"/home.html", baseTmplPath))
	tmplUser = template.Must(template.ParseFiles(*templatePath+"/user.html", baseTmplPath))
	tmplBomView = template.Must(template.ParseFiles(*templatePath+"/bom_view.html", baseTmplPath))
	tmplBomUpload = template.Must(template.ParseFiles(*templatePath+"/bom_upload.html", baseTmplPath))
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
