package view

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

type Template struct {
	t *template.Template
}

func (t *Template) Reload(basedir string) error {
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

		name := path[len(basedir):]
		src := fmt.Sprintf(`{{define "%s"}}%s{{end}}`, name, data)
		if _, err := tpl.Parse(src); err != nil {
			return err
		}
		//log.Println("view: parse template:", name)
		return nil
	}

	if err := filepath.Walk(basedir, walkFunc); err != nil {
		return err
	}
	//_template = tpl
	//log.Println("web: templates reloaded")
	t.t = tpl
	return nil
}

func (t *Template) Execute(w http.ResponseWriter, status int, name string, data interface{}) {
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	w.WriteHeader(status)
	if err := t.t.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func getTemplate() *Template {
	var tpl Template
	err := tpl.Reload("template")
	if err != nil {
		log.Fatal(err)
	}
	return &tpl
}
