package view

import (
	"net/http"
	"net/url"

	"erebus.me/model"
)

type ErrorPage struct {
	Login   *model.User
	Code    int
	Status  string
	Message string
}

func Login(w http.ResponseWriter, r *http.Request) {
	args := url.Values{
		"callback": {r.RequestURI},
	}
	http.Redirect(w, r, "/login?"+args.Encode(), http.StatusFound)
}

func NotFound(w http.ResponseWriter, r *http.Request) {
	page := ErrorPage{
		Code:    http.StatusNotFound,
		Status:  http.StatusText(http.StatusNotFound),
		Message: "File not found: " + r.URL.RequestURI(),
	}

	getTemplate().Execute(w, http.StatusNotFound, "/50x.html", &page)
	//executeTemplate(w, http.StatusNotFound, tpl, "/50x.html", &page)
}

func Error(w http.ResponseWriter, r *http.Request, msg string, status int) {
	page := ErrorPage{
		Code:    status,
		Status:  http.StatusText(status),
		Message: msg,
	}
	getTemplate().Execute(w, status, "/50x.html", &page)
}

func Display(w http.ResponseWriter, name string) {
	getTemplate().Execute(w, http.StatusOK, name, nil)
}

func Render(w http.ResponseWriter, name string, data interface{}) {
	getTemplate().Execute(w, http.StatusOK, name, data)
}

func Execute(w http.ResponseWriter, status int, name string, data interface{}) {
	getTemplate().Execute(w, status, name, data)
}
