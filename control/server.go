package control

import (
	"log"
	"net/http"
	"path"
	"runtime/debug"
	"strings"

	"erebus.me/model"
	"erebus.me/view"
)

type Server struct{}

func recoverPanic(w http.ResponseWriter, r *http.Request) {
	rc := recover()
	if rc == nil {
		return
	}

	//debug.PrintStack()
	if err, ok := rc.(*model.Error); ok {
		log.Printf("\"%s %s %s\" %d: panic error %s, stack:\n%s", r.Method, r.RequestURI, r.Proto, err.Code, err.Reason, debug.Stack())
		view.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		log.Printf("\"%s %s %s\" %d: panic %v, stack:\n%s", r.Method, r.RequestURI, r.Proto, http.StatusInternalServerError, rc, debug.Stack())
		view.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (server *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer recoverPanic(w, r)

	newPath := path.Join(r.URL.Path)
	if r.URL.Path != "/" && strings.HasSuffix(r.URL.Path, "/") {
		r.URL.Path = newPath + "/"
	} else {
		r.URL.Path = newPath
	}

	if r.URL.Path == "/" {
		Index(w, r)
		return
	}

	//r.URL.Path = path.Join(r.URL.Path)
	/*
		if strings.HasSuffix(r.URL.Path, "/") {
			r.URL.Path += "index.html"
		}
	*/

	//log.Println("web: path:", r.URL.Path)
	handler, pattern := http.DefaultServeMux.Handler(r)
	//log.Println("web: path:", r.URL.Path, "handler:", pattern)
	if pattern != "" {
		handler.ServeHTTP(w, r)
	} else {
		view.ServeHTTP(w, r)
	}
}

func ListenAndServe(addr string) {
	server := Server{}
	err := http.ListenAndServe(addr, &server)
	log.Fatal(err)
}
