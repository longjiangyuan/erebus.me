package model

import "net/http"

func NewCategory(name string) {
	_, err := db.Exec("INSERT IGNORE INTO category (name)VALUE(?)", name)
	if err != nil {
		Fatal(http.StatusInternalServerError, err.Error())
	}
}

func Categorys() []string {
	list := []string{}

	rows, err := db.Query("SELECT name FROM category ORDER by time")
	if err != nil {
		Fatal(http.StatusInternalServerError, err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		var c string

		if err := rows.Scan(&c); err != nil {
			Fatal(http.StatusInternalServerError, err.Error())
		}
		list = append(list, c)
	}
	return list
}

func ArticlesByCategory(category string) Articles {
	rows, err := db.Query("SELECT "+articleFields+" FROM article WHERE article.category=?", category)
	if err != nil {
		Fatal(http.StatusInternalServerError, err.Error())
	}
	defer rows.Close()
	return scanArticles(rows)
}
