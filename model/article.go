package model

import (
	"database/sql"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/net/html"
)

type Article struct {
	//Uri   string
	ID          int
	Title       string
	Author      string
	PostAt      time.Time
	Abstract    string
	Thumbnail   string
	NumComments int
	Category    string
}

type Articles []*Article

const articleFields = "article.id,article.title,article.author,UNIX_TIMESTAMP(article.post_at) AS post_at,article.abstract,article.thumbnail,article.comments,article.category"
const maxAbstractLength = 140

var ellipsis = []byte("...")

func extractAbstract(htmlContent string) (string, string) {
	tokenizer := html.NewTokenizer(strings.NewReader(htmlContent))
	abstract := []byte{}
	thumbnail := ""
	num := 0

loop:
	for {
		switch token := tokenizer.Next(); token {
		case html.ErrorToken:
			//log.Printf("error token: %s", tokenizer.Err())
			break loop
		case html.TextToken:
			text := tokenizer.Text()
			//log.Printf("text token: %s", text)
			for len(text) > 0 && num < maxAbstractLength {
				r, n := utf8.DecodeRune(text)
				if r == utf8.RuneError {
					break
				}

				abstract = append(abstract, text[:n]...)
				text = text[n:]
				num++
			}
		case html.SelfClosingTagToken:
			fallthrough
		case html.StartTagToken:
			node := tokenizer.Token()
			//log.Printf("start tag token: %s", node.Data)
			if thumbnail == "" && node.Data == "img" {
				for _, attr := range node.Attr {
					if attr.Key == "src" {
						thumbnail = attr.Val
						//log.Printf("extract thumbnail: %s", thumbnail)
						break
					}
				}
			}
		}
	}

	if num >= maxAbstractLength {
		abstract = append(abstract, ellipsis...)
	}

	return string(abstract), thumbnail
}

func (article *Article) Content() (content string) {
	err := db.QueryRow("SELECT content FROM article_content WHERE id=?", article.ID).Scan(&content)
	if err != nil {
		Fatal(http.StatusInternalServerError, err.Error())
	}
	return
}

func (article *Article) Update(title, content, category string) {
	article.Abstract, article.Thumbnail = extractAbstract(content)
	article.Category = category
	_, err := db.Exec("UPDATE article SET title=?,abstract=?,thumbnail=?,category=? WHERE id=?", title, article.Abstract, article.Thumbnail, article.Category, article.ID)
	if err != nil {
		Fatal(http.StatusInternalServerError, err.Error())
	}

	_, err = db.Exec("UPDATE article_content SET content=? WHERE id=?", content, article.ID)
	if err != nil {
		Fatal(http.StatusInternalServerError, err.Error())
	}
	//log.Printf("model: update article %d, title:%s abstract:%s thumbnail:%s", article.ID, article.Title, article.Abstract, article.Thumbnail)
}

func (article *Article) scanRows(rows *sql.Rows) {
	var t int64
	err := rows.Scan(&article.ID, &article.Title, &article.Author, &t, &article.Abstract, &article.Thumbnail, &article.NumComments, &article.Category)
	if err != nil {
		Fatal(http.StatusInternalServerError, err.Error())
	}
	article.PostAt = time.Unix(t, 0)
}

func scanArticles(rows *sql.Rows) Articles {
	articles := Articles{}
	for rows.Next() {
		var article Article
		article.scanRows(rows)
		articles = append(articles, &article)
	}
	return articles
}

func PostArticle(article *Article, content string) {
	article.Abstract, article.Thumbnail = extractAbstract(content)

	rs, err := db.Exec("INSERT INTO article SET title=?,author=?,post_at=NOW(),abstract=?,thumbnail=?,category=?",
		article.Title, article.Author, article.Abstract, article.Thumbnail, article.Category)
	if err != nil {
		Fatal(http.StatusInternalServerError, err.Error())
	}

	if id, err := rs.LastInsertId(); err != nil {
		Fatal(http.StatusInternalServerError, err.Error())
	} else {
		article.ID = int(id)
	}

	_, err = db.Exec("INSERT INTO article_content (id,content) VALUES (?,?)", article.ID, content)
	if err != nil {
		Fatal(http.StatusInternalServerError, err.Error())
	}
	//log.Printf("model: create article %d, title:%s abstract:%s thumbnail:%s", article.ID, article.Title, article.Abstract, article.Thumbnail)
}

func ArticleByID(id int) *Article {
	rows, err := db.Query("SELECT "+articleFields+" FROM article WHERE id=?", id)
	if err != nil {
		Fatal(http.StatusInternalServerError, err.Error())
	}
	defer rows.Close()

	if !rows.Next() {
		return nil
	}

	var article Article
	article.scanRows(rows)
	return &article
}

func ArticlesByCategroy(category string) Articles {
	rows, err := db.Query("SELECT "+articleFields+" FROM article WHERE article.category=? ORDER BY article.id DESC", category)
	if err != nil {
		Fatal(http.StatusInternalServerError, err.Error())
	}

	defer rows.Close()
	return scanArticles(rows)
}

func ArticlesByTime(offset int, limit int) Articles {
	rows, err := db.Query("SELECT "+articleFields+" FROM article ORDER BY article.id DESC LIMIT ?,?", offset, limit)
	if err != nil {
		Fatal(http.StatusInternalServerError, err.Error())
	}
	defer rows.Close()
	return scanArticles(rows)
}

func ArticlesByCategory(category int) Articles {
	rows, err := db.Query("SELECT "+articleFields+" FROM article,category WHERE article.category=?", category)
	if err != nil {
		Fatal(http.StatusInternalServerError, err.Error())
	}
	defer rows.Close()
	return scanArticles(rows)
}
