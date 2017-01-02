package control

import (
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"erebus.me/model"
	"erebus.me/view"
)

func init() {
	var blog Blog

	pjax.StripPrefix("/blog/", &blog)
}

type Blog struct {
}

func (blog *Blog) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var article *model.Article

	if r.URL.Path != "" {
		id, err := strconv.Atoi(r.URL.Path)
		if err != nil {
			view.NotFound(w, r)
			return
		}

		if article = model.ArticleByID(id); article == nil {
			view.NotFound(w, r)
			return
		}
	} else {
		article = new(model.Article)
	}

	q := r.URL.Query()
	switch r.Method {
	case "GET":
		if _, ok := q["edit"]; ok {
			blog.Edit(article, w, r)
		} else if r.URL.Path == "" {
			Home(w, r)
		} else {
			blog.Get(article, w, r)
		}
	case "POST":
		if _, ok := q["comment"]; ok {
			blog.PostComment(article, w, r)
		} else {
			blog.Post(article, w, r)
		}
	}
}

func (blog *Blog) Index(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Articles model.Articles
		Page     int
		PrevPage int
		NextPage int
	}

	const PageSize = 5

	data.Page = FormInt(r, "page", 0)
	data.NextPage = data.Page + 1
	data.PrevPage = data.Page - 1

	data.Articles = model.ArticlesByTime(data.Page*PageSize, PageSize)
	view.Render(w, "/blog/index.html", &data)
}

func (blog *Blog) Edit(article *model.Article, w http.ResponseWriter, r *http.Request) {
	login := model.GetLogin(r)
	if login == nil {
		view.Login(w, r)
		return
	}

	var data struct {
		Article  *model.Article
		Content  template.HTML
		Category []string
	}

	data.Article = article
	data.Category = model.Categorys()

	if article.ID != 0 {
		data.Content = template.HTML(article.Content())
	}
	view.Render(w, "/blog/edit.html", &data)
}

func saveFormToCookie(w http.ResponseWriter, r *http.Request, name string) string {
	val := r.FormValue(name)
	if val != "" {
		if name == "url" && !strings.HasPrefix(val, "http") {
			val = "http://" + val
		}

		ck := http.Cookie{
			Name:   name,
			Value:  val,
			Path:   "/",
			MaxAge: 365 * 86400,
		}
		http.SetCookie(w, &ck)
	}
	return val
}

func (blog *Blog) Get(article *model.Article, w http.ResponseWriter, r *http.Request) {
	var data struct {
		Login    *model.User
		Article  *model.Article
		Content  template.HTML
		Comments []*model.Comment
		Related  model.Articles

		Visitor struct {
			Name  string
			Email string
			URL   string
		}
	}

	if ck, err := r.Cookie("name"); err == nil {
		data.Visitor.Name = ck.Value
	}
	if ck, err := r.Cookie("email"); err == nil {
		data.Visitor.Email = ck.Value
	}
	if ck, err := r.Cookie("url"); err == nil {
		data.Visitor.URL = ck.Value
	}

	data.Login = model.GetLogin(r)
	data.Article = article
	data.Comments = article.Comments()
	data.Article.NumComments = len(data.Comments)
	data.Content = template.HTML(article.Content())
	data.Related = model.ArticlesByCategroy(article.Category)
	view.Render(w, "/blog/view.html", data)
}

func (blog *Blog) PostComment(article *model.Article, w http.ResponseWriter, r *http.Request) {
	var comment model.Comment

	login := model.GetLogin(r)
	if login != nil {
		comment.Name = login.Nickname
		comment.Email = login.Email
		comment.URL = login.URL
	} else {
		comment.Name = saveFormToCookie(w, r, "name")
		comment.Email = saveFormToCookie(w, r, "email")
		comment.URL = saveFormToCookie(w, r, "url")
	}

	comment.ReplyID, _ = strconv.Atoi(r.FormValue("replyto"))

	if comment.Name == "" || comment.Email == "" {
		model.Fatal(http.StatusNotAcceptable, "name and email is required")
	}

	if comment.Content = r.FormValue("comments"); comment.Content == "" {
		model.Fatal(http.StatusNotAcceptable, "empty comments is not allowed")
	}
	article.PostComment(&comment)

	urlStr := "/blog/" + r.URL.Path
	view.Render(w, "/blog/comment.html", urlStr)
}

func (blog *Blog) Post(article *model.Article, w http.ResponseWriter, r *http.Request) {
	login := model.GetLogin(r)
	if login == nil {
		view.Login(w, r)
		return
	}

	article.Author = login.Nickname
	article.Title = r.FormValue("title")
	content := r.FormValue("content")
	article.Category = r.FormValue("category")

	if article.Category == "" {
		article.Category = "default"
	}
	if article.Title == "" || content == "" {
		model.Fatal(http.StatusNotAcceptable, "empty title or content")
	}

	if article.ID == 0 {
		model.PostArticle(article, content)
	} else {
		article.Update(article.Title, content, article.Category)
	}

	articleURL := "/blog/" + strconv.Itoa(article.ID)
	http.Redirect(w, r, articleURL, http.StatusFound)
}
