package main

import (
	"github.com/gorilla/sessions"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"regexp"
	"time"
)

var (
	tmplHome, tmplView, tmplAccount, tmplUser, tmplBomView, tmplBomUpload *template.Template
)

var store = sessions.NewCookieStore([]byte(*sessionSecret))

func baseHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	log.Printf("serving %s\n", r.URL.Path)

	bomUrlPattern := regexp.MustCompile("^/([a-zA-Z][a-zA-Z0-9_]*)/([a-zA-Z][a-zA-Z0-9_]*)/$")
	bomUploadUrlPattern := regexp.MustCompile("^/([a-zA-Z][a-zA-Z0-9_]*)/([a-zA-Z][a-zA-Z0-9_]*)/_upload/$")
	userUrlPattern := regexp.MustCompile("^/([a-zA-Z][a-zA-Z0-9_]*)/$")

	switch {
	case r.URL.Path == "/":
		err = homeController(w, r)
	case r.URL.Path == "/account/login/":
		err = loginController(w, r)
	case r.URL.Path == "/account/logout/":
		err = logoutController(w, r)
	//case r.URL.Path == "/account/newuser/":
	//	err = newUserController(w, r)
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
		log.Println("error, 500: " + err.Error())
		http.Error(w, "Internal error (check logs)", 500)
	}
}

func homeController(w http.ResponseWriter, r *http.Request) (err error) {
	session, _ := store.Get(r, "bommom")
	context := make(map[string]interface{})
	context["Session"] = session.Values
	log.Printf("%s\n", session.Values["UserName"])
	context["BomList"], err = bomstore.ListBoms("")
	if err != nil {
		return
	}
	err = tmplHome.Execute(w, context)
	return
}

func loginController(w http.ResponseWriter, r *http.Request) (err error) {
	session, _ := store.Get(r, "bommom")
	context := make(map[string]interface{})
	context["ActionLogin"] = true
	context["Session"] = session.Values
	if r.Method == "POST" {
		if isShortName(r.FormValue("UserName")) != true {
			context["Problem"] = "Ugh, need to use a SHORTNAME!"
			err = tmplAccount.Execute(w, context)
			return
		}
		audience := "http://bommom.com/"
		vResponse := VerifyPersonaAssertion(r.FormValue("assertion"), audience)
		if vResponse.Okay() {
			session.Values["UserName"] = r.FormValue("UserName")
			session.Values["Email"] = vResponse.Email
			session.Save(r, w)
			context["Session"] = session.Values
			http.Redirect(w, r, "/", 302)
			return
		} else {
			context["Problem"] = vResponse.Reason
			err = tmplAccount.Execute(w, context)
			return
		}
	}
	err = tmplAccount.Execute(w, context)
	return
}

func logoutController(w http.ResponseWriter, r *http.Request) (err error) {
	session, _ := store.Get(r, "bommom")
	context := make(map[string]interface{})
	delete(session.Values, "UserName")
	delete(session.Values, "Email")
	session.Save(r, w)
	context["Session"] = session.Values
	context["ActionLogout"] = true
	err = tmplAccount.Execute(w, context)
	return
}

func userController(w http.ResponseWriter, r *http.Request, user, extra string) (err error) {
	session, _ := store.Get(r, "bommom")
	if !isShortName(user) {
		http.Error(w, "invalid username: "+user, 400)
		return
	}
	if err != nil {
		// no such user
		http.NotFound(w, r)
		return
	}
	context := make(map[string]interface{})
	context["BomList"], err = bomstore.ListBoms(ShortName(user))
	if user == "common" {
		context["IsCommon"] = true
	}
	context["UserName"] = user
	context["Session"] = session.Values
	if err != nil {
		return
	}
	err = tmplUser.Execute(w, context)
	return
}

func bomController(w http.ResponseWriter, r *http.Request, user, name string) (err error) {
	session, _ := store.Get(r, "bommom")
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
	context["Session"] = session.Values
	if err != nil {
		http.Error(w, "404 couldn't open bom: "+user+"/"+name, 404)
		return nil
	}
    err = pricingSource.AttachMarketInfoBom(context["Bom"].(*Bom))
    if err != nil {
        log.Println("error attaching market info: " + err.Error())
    }
	err = tmplBomView.Execute(w, context)
	return
}

func bomUploadController(w http.ResponseWriter, r *http.Request, user, name string) (err error) {
	session, _ := store.Get(r, "bommom")

	if !isShortName(user) {
		http.Error(w, "invalid username: "+user, 400)
		return
	}
	if !isShortName(name) {
		http.Error(w, "invalid bom name: "+name, 400)
		return
	}
	context := make(map[string]interface{})
	context["Session"] = session.Values
	context["user"] = ShortName(user)
	context["name"] = ShortName(name)
	context["BomMeta"], context["Bom"], err = bomstore.GetHead(ShortName(user), ShortName(name))

	switch r.Method {
	case "POST":
		err := r.ParseMultipartForm(1024 * 1024 * 2)
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
				context["error"] = "Problem loading CSV file: " + err.Error()
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
			log.Fatal(context["error"])
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
		http.Redirect(w, r, "//"+user+"/"+name+"/", 302)
		return err
	case "GET":
		err = tmplBomUpload.Execute(w, context)
		return err
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
	tmplAccount = template.Must(template.ParseFiles(*templatePath+"/account.html", baseTmplPath))
	//tmplLogout = template.Must(template.ParseFiles(*templatePath+"/logout.html", baseTmplPath))
	//tmplNewUser = template.Must(template.ParseFiles(*templatePath+"/newuser.html", baseTmplPath))
	tmplUser = template.Must(template.ParseFiles(*templatePath+"/user.html", baseTmplPath))
	tmplBomView = template.Must(template.ParseFiles(*templatePath+"/bom_view.html", baseTmplPath))
	tmplBomUpload = template.Must(template.ParseFiles(*templatePath+"/bom_upload.html", baseTmplPath))
	if err != nil {
		log.Fatal(err)
	}

	openBomStore()
	openAuthStore()
    openPricingSource()

	// serve template static assets (images, CSS, JS)
	http.Handle("/static/", http.FileServer(http.Dir(*templatePath+"/")))
	http.Handle("/favicon.ico", http.FileServer(http.Dir(*templatePath+"/static/")))

	// fall through to default handler
	http.HandleFunc("/", baseHandler)

	listenString := fmt.Sprintf("%s:%d", *listenHost, *listenPort)
	http.ListenAndServe(listenString, nil)
	fmt.Println("Serving at " + listenString)
}
