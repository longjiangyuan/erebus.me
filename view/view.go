package view

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"erebus.me/model"
)

var templateDir string
var fileserver http.Handler

type ErrorPage struct {
	Login   *model.User
	Code    int
	Status  string
	Message string
}

func SetTemplateDir(dir string) {
	templateDir = dir
}

func SetDocumentRoot(root string) {
	fileserver = http.FileServer(http.Dir(root))
}

func reloadTemplate() (*template.Template, error) {
	tpl := template.New("")
	walkFunc := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		name := path[len(templateDir):]
		src := fmt.Sprintf(`{{define "%s"}}%s{{end}}`, name, data)
		if _, err := tpl.Parse(src); err != nil {
			return err
		}
		//log.Println("web: parse template:", name)
		return nil
	}

	if err := filepath.Walk(templateDir, walkFunc); err != nil {
		return nil, err
	}
	//_template = tpl
	//log.Println("web: templates reloaded")
	return tpl, nil
}

func Login(w http.ResponseWriter, r *http.Request) {
	args := url.Values{
		"callback": {r.RequestURI},
	}
	http.Redirect(w, r, "/signin?"+args.Encode(), http.StatusFound)
}

func NotFound(w http.ResponseWriter) {
	tpl, err := reloadTemplate()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	page := ErrorPage{
		Code:   http.StatusNotFound,
		Status: http.StatusText(http.StatusNotFound),
	}

	w.WriteHeader(http.StatusNotFound)
	tpl.ExecuteTemplate(w, "/50x.html", &page)
}

func Error(w http.ResponseWriter, msg string, status int) {
	tpl, err := reloadTemplate()
	if err != nil {
		log.Print(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	page := ErrorPage{
		Code:    status,
		Status:  http.StatusText(status),
		Message: msg,
	}

	w.WriteHeader(status)
	if err := tpl.ExecuteTemplate(w, "/50x.html", &page); err != nil {
		log.Print(err)
	}
}

func Render(w http.ResponseWriter, name string, data interface{}) {
	tpl, err := reloadTemplate()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h := w.Header()
	h.Set("Content-Type", "text/html; charset=utf-8")
	if err := tpl.ExecuteTemplate(w, name, data); err != nil {
		log.Printf("view: render template %s error: %s", name, err)
	}
}

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fileserver.ServeHTTP(w, r)
}
