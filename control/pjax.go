package control

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
	"path"
	"runtime/debug"
	"strconv"
	"strings"

	"erebus.me/model"
	"erebus.me/view"
)

type MainPage struct {
	Login   *model.User
	Content template.HTML
}

var pjax = &PjaxHandler{
	mux: map[string]http.Handler{},
}

func ListenAndServe(addr string) {
	err := http.ListenAndServe(addr, pjax)
	log.Fatal(err)
}

type PjaxHandler struct {
	mux map[string]http.Handler
}

func (pjax *PjaxHandler) Handle(path string, handler http.Handler) {
	pjax.mux[path] = handler
}

func (pjax *PjaxHandler) HandleFunc(path string, handler func(http.ResponseWriter, *http.Request)) {
	pjax.mux[path] = http.HandlerFunc(handler)
}

func (pjax *PjaxHandler) StripPrefix(path string, handler http.Handler) {
	fn := func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = r.URL.Path[len(path):]
		handler.ServeHTTP(w, r)
	}
	pjax.mux[path] = http.HandlerFunc(fn)
}

func recoverPanic(w http.ResponseWriter, r *http.Request) {
	rc := recover()
	if rc == nil {
		return
	}

	//debug.PrintStack()
	if err, ok := rc.(*model.Error); ok {
		log.Printf("\"%s %s %s\" %d: panic error %s, stack:\n%s", r.Method, r.RequestURI, r.Proto, err.Code, err.Reason, debug.Stack())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		log.Printf("\"%s %s %s\" %d: panic %v, stack:\n%s", r.Method, r.RequestURI, r.Proto, http.StatusInternalServerError, rc, debug.Stack())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (pjax *PjaxHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer recoverPanic(w, r)

	newPath := path.Join(r.URL.Path)
	if r.URL.Path != "/" && strings.HasSuffix(r.URL.Path, "/") {
		r.URL.Path = newPath + "/"
	} else {
		r.URL.Path = newPath
	}

	for path := r.URL.Path; len(path) > 0; {
		if path == "/" && path != r.URL.Path {
			break
		}
		if handler, ok := pjax.mux[path]; ok {
			servePjax(handler, w, r)
			return
		}

		if n := len(path); path[n-1] == '/' {
			path = path[:n-1]
		}
		if idx := strings.LastIndexByte(path, '/'); idx >= 0 {
			//log.Printf("path %s not found, try %s", path, path[:idx+1])
			path = path[:idx+1]
		}
	}
	view.NotFound(w, r)
}

func servePjax(h http.Handler, w http.ResponseWriter, r *http.Request) {
	writer := BufferResponseWriter{
		w:      w,
		status: http.StatusOK,
		header: w.Header(),
	}
	h.ServeHTTP(&writer, r)

	body := writer.body.Bytes()
	if r.Header.Get("X-Pjax") == "true" {
		writer.header.Del("Content-Encoding")
		writer.header.Set("Content-Length", strconv.Itoa(len(body)))
		w.WriteHeader(writer.status)
		w.Write(body)
	} else {
		var page MainPage
		page.Login = model.GetLogin(r)
		page.Content = template.HTML(body)
		view.Execute(w, writer.status, "/main.html", &page)
	}
}

type BufferResponseWriter struct {
	w      http.ResponseWriter
	status int
	header http.Header
	body   bytes.Buffer
}

func (writer *BufferResponseWriter) Header() http.Header {
	return writer.header
}

func (writer *BufferResponseWriter) WriteHeader(status int) {
	writer.status = status
}

func (writer *BufferResponseWriter) Write(b []byte) (int, error) {
	return writer.body.Write(b)
}
