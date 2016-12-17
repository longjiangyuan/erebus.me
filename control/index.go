package control

import (
	"net/http"

	"crypto/rand"

	"encoding/hex"

	"erebus.me/model"
	"erebus.me/view"
)

func init() {
	http.HandleFunc("/login", Login)
	http.HandleFunc("/logout", Logout)
}

func Index(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Login    *model.User
		Articles model.Articles
		Category []string
		Comments []*model.Comment
		Page     int
		PrevPage int
		NextPage int
	}

	const PageSize = 5

	data.Login = model.GetLogin(r)
	data.Page = FormInt(r, "page", 0)
	data.NextPage = data.Page + 1
	data.PrevPage = data.Page - 1

	//data.Login = model.GetLoginUser(r)
	data.Category = model.Categorys()
	data.Comments = model.RecentComments(5)
	data.Articles = model.ArticlesByTime(data.Page*PageSize, PageSize)
	view.Render(w, "/index.html", &data)
}

func Logout(w http.ResponseWriter, r *http.Request) {
	cookie := http.Cookie{
		Name:   "token",
		Value:  "",
		MaxAge: 0,
		Path:   "/",
	}
	http.SetCookie(w, &cookie)
	http.Redirect(w, r, "/", http.StatusFound)
}

func Login(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		view.Render(w, "/login.html", nil)
	case "POST":
		email := r.FormValue("email")
		password := r.FormValue("password")

		user := model.Signin(email, password)

		var tokenBin [16]byte
		n, err := rand.Read(tokenBin[:])
		if err != nil {
			model.Fatal(http.StatusInternalServerError, err.Error())
		}
		token := hex.EncodeToString(tokenBin[:n])
		user.SetToken(token)

		cookie := http.Cookie{
			Name:   "token",
			Value:  token,
			Path:   "/",
			MaxAge: 7 * 86400,
		}
		http.SetCookie(w, &cookie)
		callback := r.FormValue("callback")
		if callback == "" {
			callback = "/"
		}
		http.Redirect(w, r, callback, http.StatusFound)
	}
}

type RightPage struct {
	Category []string
}

func getRightPage() *RightPage {
	var right RightPage
	right.Category = model.Categorys()
	return &right
}
