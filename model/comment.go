package model

import (
	"database/sql"
	"net/http"
	"time"
)

type Comment struct {
	ID        int
	ArticleID int
	Name      string
	Email     string
	URL       string
	Time      time.Time
	Content   string
	ReplyID   int
	ReplyTo   *Comment
}

const commentFields = "id,article_id,name,email,url,content,UNIX_TIMESTAMP(time) as time,reply_to"

func (c *Comment) scanRows(rows *sql.Rows) {
	var unixtime int64
	if err := rows.Scan(&c.ID, &c.ArticleID, &c.Name, &c.Email, &c.URL, &c.Content, &unixtime, &c.ReplyID); err != nil {
		Fatal(http.StatusInternalServerError, err.Error())
	}
	c.Time = time.Unix(unixtime, 0)
}

func (article *Article) PostComment(c *Comment) {
	ExecUpdate("INSERT INTO article_comment SET article_id=?,name=?,email=?,url=?,time=NOW(),content=?,reply_to=?",
		article.ID, c.Name, c.Email, c.URL, c.Content, c.ReplyID)
	ExecUpdate("UPDATE article SET comments=comments+1 WHERE id=?", article.ID)
}

func (article *Article) Comments() []*Comment {
	list := []*Comment{}
	rows := Query("SELECT "+commentFields+" FROM article_comment WHERE article_id=? ORDER BY time DESC", article.ID)
	defer rows.Close()

	for rows.Next() {
		var c Comment
		c.scanRows(rows)
		list = append(list, &c)
	}

	for _, c := range list {
		if c.ReplyID != 0 {
			c.ReplyTo = findCommentByID(list, c.ReplyID)
		}
	}
	return list
}

func RecentComments(limit int) []*Comment {
	list := []*Comment{}
	rows := Query("SELECT "+commentFields+" FROM article_comment ORDER BY time DESC LIMIT ?", limit)
	defer rows.Close()

	for rows.Next() {
		var c Comment
		c.scanRows(rows)
		list = append(list, &c)
	}

	return list
}

func findCommentByID(comments []*Comment, id int) *Comment {
	for _, c := range comments {
		if c.ID == id {
			return c
		}
	}
	return nil
}
