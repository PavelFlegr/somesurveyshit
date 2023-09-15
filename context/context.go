package context

import (
	"database/sql"
	"github.com/gorilla/securecookie"
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
	err = Ctx.Sc.Decode("userId", cookie.Value, &value)
	if err != nil {
		log.Println("Authentication failed", err)
	}

	return value, err
}

type AppContext struct {
	Db *sql.DB
	Sc *securecookie.SecureCookie
}

var Ctx AppContext
