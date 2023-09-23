package global

import (
	"database/sql"
	"github.com/gorilla/securecookie"
	"html/template"
	"log"
	"net/http"
)

func CheckAuth(r *http.Request) (int64, error) {
	cookie, err := r.Cookie("userId")
	if err != nil {
		log.Println("Authentication failed", err)
		return 0, err
	}
	var value int64
	err = Sc.Decode("userId", cookie.Value, &value)
	if err != nil {
		log.Println("Authentication failed", err)
	}

	return value, err
}

var Db *sql.DB
var Template *template.Template
var Sc *securecookie.SecureCookie
