package controllers

import (
	"html/template"
	"log"
	"main/services"
	"net/http"
	"strconv"
)

func Survey(template *template.Template) {
	http.HandleFunc("/survey", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			surveyId, _ := strconv.ParseInt(r.URL.Query().Get("surveyId"), 10, 0)
			survey := services.GetSurvey(surveyId)

			err := template.ExecuteTemplate(w, "user-survey.html", survey)
			if err != nil {
				log.Println(err)
			}
		}
		if r.Method == "POST" {
			r.ParseForm()
			surveyId, _ := strconv.ParseInt(r.URL.Query().Get("surveyId"), 10, 0)
			response := services.Response{}
			for k, v := range r.PostForm {
				questionId, _ := strconv.ParseInt(k, 10, 0)
				question := services.GetQuestion(questionId)
				answerIndex, _ := strconv.ParseInt(v[0], 10, 0)
				response[k] = question.Options[answerIndex].Label
			}

			services.RecordResponse(surveyId, response)

			http.Redirect(w, r, "/goodbye", http.StatusSeeOther)
		}
	})
}
