package control

import (
	"net/http"

	"erebus.me/model"
	"erebus.me/view"
)

type Category struct {
}

func init() {
	var category Category
	pjax.StripPrefix("/category/", &category)
}

func (category *Category) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Name     string
		Articles model.Articles
		Category []string
	}

	if r.URL.Path == "" {
		view.NotFound(w, r)
		return
	}

	data.Name = r.URL.Path
	data.Articles = model.ArticlesByCategory(data.Name)
	data.Category = model.Categorys()
	view.Render(w, "/category/view.html", &data)
}
