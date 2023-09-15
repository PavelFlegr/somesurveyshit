package main

import (
	"database/sql"
	"fmt"
	"github.com/gorilla/securecookie"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"html/template"
	"log"
	"main/context"
	_ "main/context"
	"main/controllers"
	"main/services"
	"net/http"
	"net/mail"
	"os"
)

func main() {
	port := os.Args[1]
	connStr := os.Args[2]
	hashKey := []byte(os.Args[3])

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	context.Ctx = context.AppContext{Db: db, Sc: securecookie.New(hashKey, nil)}
	indexTmpl, err := template.New("index.html").Funcs(template.FuncMap{
		"unescape": func(val string) template.HTML {
			return template.HTML(val)
		},
	}).ParseGlob("templates/*")
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, authErr := context.CheckAuth(r)
		indexTmpl.Execute(w, services.TemplateData{
			LoggedIn: authErr == nil,
		})
	})

	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		_, authErr := context.CheckAuth(r)
		if r.Method == "GET" {
			if r.Header.Get("Hx-Request") == "true" {
				indexTmpl.ExecuteTemplate(w, "login", nil)
				return
			}
			indexTmpl.ExecuteTemplate(w, "login.html", services.TemplateData{
				LoggedIn: authErr == nil,
			})
		}
		if r.Method == "POST" {
			email := r.PostFormValue("email")
			password := r.PostFormValue("password")
			user, err := services.GetUserByEmail(email)
			if err != nil {
				indexTmpl.ExecuteTemplate(w, "error", "Account with this email does not exist")
				return
			}

			err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
			if err != nil {
				indexTmpl.ExecuteTemplate(w, "error", "Password is incorrect")
				return
			}

			cookie, _ := context.Ctx.Sc.Encode("userId", user.Id)

			http.SetCookie(w, &http.Cookie{
				HttpOnly: true,
				Value:    cookie,
				Name:     "userId",
				SameSite: http.SameSiteStrictMode},
			)

			w.Header().Set("HX-Redirect", "/dashboard")
		}
	})

	http.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			http.SetCookie(w, &http.Cookie{
				Name:     "userId",
				Value:    "",
				Path:     "/",
				MaxAge:   -1,
				HttpOnly: true,
			})
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			indexTmpl.ExecuteTemplate(w, "register", nil)
		}

		if r.Method == "POST" {
			email, err := mail.ParseAddress(r.PostFormValue("email"))
			if err != nil {
				indexTmpl.ExecuteTemplate(w, "error", "Invalid email address")
				return
			}

			password := r.PostFormValue("password")
			password2 := r.PostFormValue("password2")

			if len(password) < 5 {
				indexTmpl.ExecuteTemplate(w, "error", "Password must be at least 5 characters long")
				return
			}

			if password != password2 {
				indexTmpl.ExecuteTemplate(w, "error", "Passwords don't match")
				return
			}

			hash, _ := bcrypt.GenerateFromPassword([]byte(password), 10)
			err = services.CreateUser(email.Address, string(hash))
			if err != nil {
				indexTmpl.ExecuteTemplate(w, "error", "An account with this email already exists")
			}

			w.Header().Set("HX-Redirect", "/login")
		}
	})

	controllers.Survey(indexTmpl)

	http.ListenAndServe(fmt.Sprintf(":%v", port), nil)
}
