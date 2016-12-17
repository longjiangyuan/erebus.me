package model

import (
	"database/sql"
	"net/http"
)

type User struct {
	Email     string
	Nickname  string
	AvatarURL string
	URL       string
}

func (user *User) SetToken(token string) {
	ExecUpdate("UPDATE user SET token=? WHERE email=?", token, user.Email)
}

func GetLogin(r *http.Request) *User {
	var user User
	cookie, err := r.Cookie("token")
	if err != nil {
		return nil
	}
	if cookie.Value == "" {
		return nil
	}

	row := db.QueryRow("SELECT email,nickname,avatar_url,url FROM user WHERE token=?", cookie.Value)
	if err = row.Scan(&user.Email, &user.Nickname, &user.AvatarURL, &user.URL); err == sql.ErrNoRows {
		return nil
	} else if err != nil {
		Fatal(http.StatusInternalServerError, err.Error())
	}
	return &user
}

func Signin(email string, password string) *User {
	var user User
	row := db.QueryRow("SELECT email,nickname,avatar_url,url FROM user WHERE email=? AND password=SHA1(?)", email, password)

	if err := row.Scan(&user.Email, &user.Nickname, &user.AvatarURL, &user.URL); err == sql.ErrNoRows {
		Fatal(http.StatusInternalServerError, "Incorrect email or password")
	} else if err != nil {
		Fatal(http.StatusInternalServerError, err.Error())
	}

	return &user
}
