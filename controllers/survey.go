package controllers

import (
	"github.com/go-chi/chi/v5"
	"html/template"
	"log"
	"main/services"
	"net/http"
	"strconv"
)

func Survey(template *template.Template, r chi.Router) {
	r.Route("/survey/{surveyId}", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			surveyId, _ := strconv.ParseInt(chi.URLParam(r, "surveyId"), 10, 0)
			survey := services.GetSurvey(surveyId)

			err := template.ExecuteTemplate(w, "user-survey.html", survey)
			if err != nil {
				log.Println(err)
			}
		})

		r.Post("/", func(w http.ResponseWriter, r *http.Request) {
			r.ParseForm()
			surveyId, _ := strconv.ParseInt(chi.URLParam(r, "surveyId"), 10, 0)
			response := services.Response{}
			for k, v := range r.PostForm {
				questionId, _ := strconv.ParseInt(k, 10, 0)
				question := services.GetQuestion(questionId)
				answerIndex, _ := strconv.ParseInt(v[0], 10, 0)
				response[k] = question.Options[answerIndex].Label
			}

			services.RecordResponse(surveyId, response)

			http.Redirect(w, r, "/goodbye", http.StatusSeeOther)
		})
	})
}
