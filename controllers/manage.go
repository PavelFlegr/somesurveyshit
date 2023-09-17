package controllers

import (
	"encoding/csv"
	"github.com/go-chi/chi/v5"
	"html/template"
	"log"
	"main/context"
	"main/services"
	"net/http"
	"strconv"
	"strings"
)

func Manage(template *template.Template, r chi.Router) {
	r.Get("/manage/dashboard", func(w http.ResponseWriter, r *http.Request) {
		userId, authErr := context.CheckAuth(r)
		if authErr != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		surveys := services.ListSurveys(userId)
		template.ExecuteTemplate(w, "dashboard.html", services.TemplateData{
			LoggedIn: authErr == nil,
			Data:     surveys,
		})
	})

	r.Get("/manage/survey", func(w http.ResponseWriter, r *http.Request) {
		userId, authErr := context.CheckAuth(r)
		if authErr != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		surveyId, _ := strconv.ParseInt(r.URL.Query().Get("surveyId"), 10, 0)
		if !services.HasPermission(userId, "survey", surveyId, "read") {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		survey := services.GetSurvey(surveyId)
		if r.Header.Get("Hx-Request") == "true" {
			template.ExecuteTemplate(w, "survey", survey)
		} else {
			err := template.ExecuteTemplate(w, "survey.html", services.TemplateData{
				LoggedIn: authErr == nil,
				Data:     survey,
			})
			if err != nil {
				log.Println("GET /survey", err)
			}
		}
	})

	r.Post("/manage/survey", func(w http.ResponseWriter, r *http.Request) {
		userId, authErr := context.CheckAuth(r)
		if authErr != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		title := strings.TrimSpace(r.PostFormValue("title"))
		if len(title) < 1 {
			template.ExecuteTemplate(w, "error2", "Title can't be empty")
			return
		}
		survey := services.CreateSurvey(title, userId)
		template.ExecuteTemplate(w, "surveyItem", survey)
		template.ExecuteTemplate(w, "noerror", nil)
	})

	r.Delete("/manage/survey", func(w http.ResponseWriter, r *http.Request) {
		userId, authErr := context.CheckAuth(r)
		if authErr != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		surveyId, _ := strconv.ParseInt(r.URL.Query().Get("surveyId"), 10, 0)
		if !services.HasPermission(userId, "survey", surveyId, "edit") {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		services.DeleteSurvey(surveyId, userId)
	})

	r.Get("/manage/survey/title", func(w http.ResponseWriter, r *http.Request) {
		userId, authErr := context.CheckAuth(r)
		if authErr != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		surveyId, _ := strconv.ParseInt(r.URL.Query().Get("surveyId"), 10, 0)
		if !services.HasPermission(userId, "survey", surveyId, "read") {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		survey := services.GetSurvey(surveyId)
		template.ExecuteTemplate(w, "navigation", survey)
	})

	r.Put("/manage/survey/title", func(w http.ResponseWriter, r *http.Request) {
		userId, authErr := context.CheckAuth(r)
		if authErr != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		surveyId, _ := strconv.ParseInt(r.URL.Query().Get("surveyId"), 10, 0)
		if !services.HasPermission(userId, "survey", surveyId, "edit") {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		title := r.PostFormValue("title")
		services.RenameSurvey(surveyId, title)

		template.ExecuteTemplate(w, "navigation", services.Survey{Id: surveyId, Title: title})
	})

	r.Get("/manage/survey/title/edit", func(w http.ResponseWriter, r *http.Request) {
		userId, authErr := context.CheckAuth(r)
		if authErr != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		surveyId, _ := strconv.ParseInt(r.URL.Query().Get("surveyId"), 10, 0)
		if !services.HasPermission(userId, "survey", surveyId, "read") {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		survey := services.GetSurvey(surveyId)
		template.ExecuteTemplate(w, "edit-survey-title", survey)
	})

	r.Put("/manage/survey/reorder", func(w http.ResponseWriter, r *http.Request) {
		userId, authErr := context.CheckAuth(r)
		if authErr != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		r.ParseForm()
		surveyId, _ := strconv.ParseInt(r.URL.Query().Get("surveyId"), 10, 0)
		if !services.HasPermission(userId, "survey", surveyId, "edit") {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		var questionsOrder []int64
		for _, question := range r.PostForm["question"] {
			id, _ := strconv.ParseInt(question, 10, 0)
			questionsOrder = append(questionsOrder, id)
		}

		services.ReorderSurvey(surveyId, questionsOrder, userId)
	})

	r.Get("/manage/question", func(w http.ResponseWriter, r *http.Request) {
		userId, authErr := context.CheckAuth(r)
		if authErr != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		questionId, _ := strconv.ParseInt(r.URL.Query().Get("questionId"), 10, 0)
		if !services.HasPermission(userId, "question", questionId, "read") {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		question := services.GetQuestion(questionId)

		template.ExecuteTemplate(w, "question.html", question)
	})

	r.Post("/manage/question", func(w http.ResponseWriter, r *http.Request) {
		userId, authErr := context.CheckAuth(r)
		if authErr != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		surveyId, _ := strconv.ParseInt(r.URL.Query().Get("surveyId"), 10, 0)
		if !services.HasPermission(userId, "survey", surveyId, "edit") {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		question := services.CreateQuestion(surveyId, userId)

		template.ExecuteTemplate(w, "questions", []services.Question{question})
	})

	r.Put("/manage/question", func(w http.ResponseWriter, r *http.Request) {
		userId, authErr := context.CheckAuth(r)
		if authErr != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		questionId, _ := strconv.ParseInt(r.URL.Query().Get("questionId"), 10, 0)
		if !services.HasPermission(userId, "question", questionId, "edit") {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		title := r.FormValue("title")
		description := r.FormValue("description")
		var options []services.Option
		for _, option := range r.PostForm["option"] {
			options = append(options, services.Option{Label: option})
		}
		question := services.Question{
			Id:          questionId,
			Title:       title,
			Description: description,
			Options:     options,
		}
		services.UpdateQuestion(question)

		template.ExecuteTemplate(w, "question.html", question)
	})

	r.Delete("/manage/question", func(w http.ResponseWriter, r *http.Request) {
		userId, authErr := context.CheckAuth(r)
		if authErr != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		questionId, _ := strconv.ParseInt(r.URL.Query().Get("questionId"), 10, 0)
		if !services.HasPermission(userId, "question", questionId, "edit") {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		services.DeleteQuestion(questionId)
	})

	r.Get("/manage/question/edit", func(w http.ResponseWriter, r *http.Request) {
		userId, authErr := context.CheckAuth(r)
		if authErr != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		questionId, _ := strconv.ParseInt(r.URL.Query().Get("questionId"), 10, 0)
		if !services.HasPermission(userId, "question", questionId, "read") {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		question := services.GetQuestion(questionId)

		template.ExecuteTemplate(w, "edit-question.html", question)
	})

	r.Get("/manage/option", func(w http.ResponseWriter, r *http.Request) {
		option := services.Option{}
		template.ExecuteTemplate(w, "option", option)
	})

	r.Get("/manage/survey/download", func(w http.ResponseWriter, r *http.Request) {
		surveyId, _ := strconv.ParseInt(r.URL.Query().Get("surveyId"), 10, 0)

		rows, err := context.Ctx.Db.Query("select response from responses where survey_id = $1", surveyId)
		if err != nil {
			log.Println("manage survey download", err)
		}

		survey := services.GetSurvey(surveyId)
		var questionIds []string

		csvWriter := csv.NewWriter(w)

		var record []string
		for _, question := range survey.Questions {
			questionIds = append(questionIds, strconv.FormatInt(question.Id, 10))
			record = append(record, question.Title)
		}
		csvWriter.Write(record)
		csvWriter.Flush()

		var response services.Response
		for rows.Next() {
			err := rows.Scan(&response)
			if err != nil {
				log.Println("manage survey download", err)
			}

			record = []string{}
			for _, questionId := range questionIds {
				record = append(record, response[questionId])
			}
			csvWriter.Write(record)
			csvWriter.Flush()
		}
	})
}
