package auth

import (
	"golang.org/x/crypto/bcrypt"
	"main/global"
	"main/services"
	"net/http"
	"net/mail"
)

func GetLogin(w http.ResponseWriter, r *http.Request) {
	_, authErr := global.CheckAuth(r)
	if r.Header.Get("Hx-Request") == "true" {
		global.Template.ExecuteTemplate(w, "login", nil)
		return
	}
	global.Template.ExecuteTemplate(w, "login.html", services.TemplateData{
		LoggedIn: authErr == nil,
	})
}

func Login(w http.ResponseWriter, r *http.Request) {
	email := r.PostFormValue("email")
	password := r.PostFormValue("password")
	user, err := services.GetUserByEmail(email)
	if err != nil {
		global.Template.ExecuteTemplate(w, "error", "Account with this email does not exist")
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		global.Template.ExecuteTemplate(w, "error", "Password is incorrect")
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
	global.Template.ExecuteTemplate(w, "register", nil)
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
		global.Template.ExecuteTemplate(w, "error", "Invalid email address")
		return
	}

	password := r.PostFormValue("password")
	password2 := r.PostFormValue("password2")

	if len(password) < 5 {
		global.Template.ExecuteTemplate(w, "error", "Password must be at least 5 characters long")
		return
	}

	if password != password2 {
		global.Template.ExecuteTemplate(w, "error", "Passwords don't match")
		return
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(password), 10)
	var id int64
	id, err = services.CreateUser(email.Address, string(hash))
	if err != nil {
		global.Template.ExecuteTemplate(w, "error", "An account with this email already exists")
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
