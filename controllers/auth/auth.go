package auth

import (
	"golang.org/x/crypto/bcrypt"
	"log"
	"main/global"
	"main/services"
	"net/http"
	"net/mail"
)

func GetLogin(w http.ResponseWriter, r *http.Request) {
	_, authErr := global.CheckAuth(r)
	if r.Header.Get("Hx-Request") == "true" {
		err := global.Template.ExecuteTemplate(w, "auth/login", nil)
		if err != nil {
			log.Println(err)
		}
		return
	}
	err := global.Template.ExecuteTemplate(w, "auth/login.html", services.TemplateData{
		LoggedIn: authErr == nil,
	})
	if err != nil {
		log.Println(err)
		return
	}
}

func Login(w http.ResponseWriter, r *http.Request) {
	email := r.PostFormValue("email")
	password := r.PostFormValue("password")
	user, err := services.GetUserByEmail(email)
	if err != nil {
		err := global.Template.ExecuteTemplate(w, "error", "Account with this email does not exist")
		if err != nil {
			log.Println(err)
		}
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		err := global.Template.ExecuteTemplate(w, "error", "Password is incorrect")
		if err != nil {
			log.Println(err)
		}
		return
	}

	cookie, _ := global.Sc.Encode("userId", user.Id)

	http.SetCookie(w, &http.Cookie{
		HttpOnly: true,
		Value:    cookie,
		Name:     "userId",
		SameSite: http.SameSiteStrictMode},
	)

	w.Header().Set("HX-Redirect", "/manage/dashboard")
}

func GetRegister(w http.ResponseWriter, r *http.Request) {
	err := global.Template.ExecuteTemplate(w, "auth/register", nil)
	if err != nil {
		log.Println(err)
	}
}

func Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "userId",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func Register(w http.ResponseWriter, r *http.Request) {
	email, err := mail.ParseAddress(r.PostFormValue("email"))
	if err != nil {
		err := global.Template.ExecuteTemplate(w, "error", "Invalid email address")
		if err != nil {
			log.Println(err)
		}
		return
	}

	password := r.PostFormValue("password")
	password2 := r.PostFormValue("password2")

	if len(password) < 5 {
		err := global.Template.ExecuteTemplate(w, "error", "Password must be at least 5 characters long")
		if err != nil {
			log.Println(err)
		}
		return
	}

	if password != password2 {
		err := global.Template.ExecuteTemplate(w, "error", "Passwords don't match")
		if err != nil {
			log.Println(err)
		}
		return
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(password), 10)
	var id int64
	id, err = services.CreateUser(email.Address, string(hash))
	if err != nil {
		err := global.Template.ExecuteTemplate(w, "error", "An account with this email already exists")
		if err != nil {
			log.Println(err)
		}
	}

	cookie, _ := global.Sc.Encode("userId", id)

	http.SetCookie(w, &http.Cookie{
		HttpOnly: true,
		Value:    cookie,
		Name:     "userId",
		SameSite: http.SameSiteStrictMode},
	)

	w.Header().Set("HX-Redirect", "/manage/dashboard")
}
