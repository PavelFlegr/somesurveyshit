package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"main/controllers/answer"
	"main/controllers/auth"
	"main/controllers/block"
	"main/controllers/question"
	"main/controllers/survey"
	"main/global"
	_ "main/global"
	"main/services"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/securecookie"
	_ "github.com/lib/pq"
)

func main() {
	port := os.Args[1]
	connStr := os.Args[2]
	hashKey := []byte(os.Args[3])

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	global.Db = db
	global.Sc = securecookie.New(hashKey, nil)
	var tmpl *template.Template
	tmpl, err = template.New("index.html").Funcs(template.FuncMap{
		"unescape": func(val string) template.HTML {
			return template.HTML(val)
		},
	}).ParseGlob("templates/*.html")
	tmpl, err = tmpl.ParseGlob("templates/*/*.html")
	if err != nil {
		log.Fatal(err)
	}

	global.Template = tmpl

	log.SetFlags(log.Llongfile | log.Ltime | log.Ldate | log.Lmsgprefix)

	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		_, authErr := global.CheckAuth(r)
		err := tmpl.Execute(w, services.TemplateData{
			LoggedIn: authErr == nil,
		})
		if err != nil {
			log.Println(err)
		}
	})

	RegisterRoutes(r)

	err = http.ListenAndServe(fmt.Sprintf("127.0.0.1:%v", port), r)
	if err != nil {
		log.Println(err)
	}
}

func RegisterRoutes(r chi.Router) {
	r.Get("/manage/dashboard", survey.Dashboard)
	r.Post("/manage/survey", survey.AddSurvey)
	r.Get("/manage/survey/{surveyId}", survey.GetSurvey)
	r.Delete("/manage/survey/{surveyId}", survey.DeleteSurvey)
	r.Get("/manage/survey/{surveyId}/title", survey.GetSurveyTitle)
	r.Get("/manage/survey/{surveyId}/title/edit", survey.GetSurveyTitleEdit)
	r.Put("/manage/survey/{surveyId}/title", survey.PutSurveyTitle)
	r.Get("/manage/option", survey.GetEmptyOption)
	r.Get("/manage/survey/{surveyId}/download", survey.DownloadSurvey)

	r.Put("/manage/survey/{surveyId}/block/{blockId}/title", block.PutBlockTitle)
	r.Get("/manage/survey/{surveyId}/block/{blockId}/title", block.GetBlockTitle)
	r.Put("/manage/survey/{surveyId}/block/{blockId}/randomize", block.PutBlockRandomize)
	r.Get("/manage/survey/{surveyId}/block/{blockId}/title/edit", block.GetBlockTitleEdit)
	r.Post("/manage/survey/{surveyId}/block/{blockId}/reorder", block.ReorderBlock)
	r.Post("/manage/survey/{surveyId}/block", block.CreateBlock)
	r.Get("/manage/survey/{surveyId}/block/{blockId}/question", block.ListQuestions)
	r.Delete("/manage/survey/{surveyId}/block/{blockId}", block.DeleteBlock)

	r.Post("/manage/survey/{surveyId}/question/{questionId}/reorder", question.ReorderQuestion)
	r.Post("/manage/survey/{surveyId}/block/{blockId}/question", question.CreateQuestion)
	r.Get("/manage/survey/{surveyId}/question/{questionId}", question.GetQuestion)
	r.Put("/manage/survey/{surveyId}/question/{questionId}", question.PutQuestion)
	r.Delete("/manage/survey/{surveyId}/question/{questionId}", question.DeleteQuestion)
	r.Get("/manage/survey/{surveyId}/question/{questionId}/edit", question.GetQuestionEdit)

	r.Get("/survey/{surveyId}", answer.GetSurvey)
	r.Post("/survey/{surveyId}", answer.SubmitAnswers)
	r.Get("/goodbye", answer.ShowGoodbye)

	r.Get("/login", auth.GetLogin)
	r.Post("/login", auth.Login)
	r.Get("/logout", auth.Logout)
	r.Get("/register", auth.GetRegister)
	r.Post("/register", auth.Register)
}
