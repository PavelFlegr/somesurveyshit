package controllers

import (
	"html/template"
	"main/context"
	"main/services"
	"net/http"
	"strconv"
	"strings"
)

func Survey(template *template.Template) {
	http.HandleFunc("/dashboard", func(w http.ResponseWriter, r *http.Request) {
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
	http.HandleFunc("/survey", func(w http.ResponseWriter, r *http.Request) {
		userId, authErr := context.CheckAuth(r)
		if authErr != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		if r.Method == "GET" {
			surveyId, _ := strconv.ParseInt(r.URL.Query().Get("surveyId"), 10, 0)
			if !services.HasPermission(userId, "survey", surveyId, "read") {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			survey := services.GetSurvey(surveyId, userId)
			if r.Header.Get("Hx-Request") == "true" {
				template.ExecuteTemplate(w, "survey", survey)
			} else {
				template.ExecuteTemplate(w, "survey.html", services.TemplateData{
					LoggedIn: authErr == nil,
					Data:     survey,
				})
			}
		}
		if r.Method == "POST" {
			title := strings.TrimSpace(r.PostFormValue("title"))
			if len(title) < 1 {
				template.ExecuteTemplate(w, "error2", "Title can't be empty")
				return
			}
			survey := services.CreateSurvey(title, userId)
			template.ExecuteTemplate(w, "surveyItem", survey)
			template.ExecuteTemplate(w, "noerror", nil)
		}
		if r.Method == "DELETE" {
			surveyId, _ := strconv.ParseInt(r.URL.Query().Get("surveyId"), 10, 0)
			if !services.HasPermission(userId, "survey", surveyId, "edit") {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			services.DeleteSurvey(surveyId, userId)
		}
	})

	http.HandleFunc("/survey/title", func(w http.ResponseWriter, r *http.Request) {
		userId, authErr := context.CheckAuth(r)
		if authErr != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		if r.Method == "GET" {
			surveyId, _ := strconv.ParseInt(r.URL.Query().Get("surveyId"), 10, 0)
			if !services.HasPermission(userId, "survey", surveyId, "read") {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			survey := services.GetSurvey(surveyId, userId)
			template.ExecuteTemplate(w, "navigation", survey)
		}
		if r.Method == "PUT" {
			surveyId, _ := strconv.ParseInt(r.URL.Query().Get("surveyId"), 10, 0)
			if !services.HasPermission(userId, "survey", surveyId, "edit") {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			title := r.PostFormValue("title")
			services.RenameSurvey(surveyId, title)

			template.ExecuteTemplate(w, "navigation", services.Survey{Id: surveyId, Title: title})
		}
	})

	http.HandleFunc("/survey/title/edit", func(w http.ResponseWriter, r *http.Request) {
		userId, authErr := context.CheckAuth(r)
		if authErr != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		if r.Method == "GET" {
			surveyId, _ := strconv.ParseInt(r.URL.Query().Get("surveyId"), 10, 0)
			if !services.HasPermission(userId, "survey", surveyId, "read") {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			survey := services.GetSurvey(surveyId, userId)
			template.ExecuteTemplate(w, "edit-survey-title", survey)
		}
	})

	http.HandleFunc("/survey/reorder", func(w http.ResponseWriter, r *http.Request) {
		userId, authErr := context.CheckAuth(r)
		if authErr != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		if r.Method == "PUT" {
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
		}
	})

	http.HandleFunc("/question", func(w http.ResponseWriter, r *http.Request) {
		userId, authErr := context.CheckAuth(r)
		if authErr != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		if r.Method == "GET" {
			questionId, _ := strconv.ParseInt(r.URL.Query().Get("questionId"), 10, 0)
			if !services.HasPermission(userId, "question", questionId, "read") {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			question := services.GetQuestion(questionId)

			template.ExecuteTemplate(w, "question.html", question)
		}
		if r.Method == "POST" {
			surveyId, _ := strconv.ParseInt(r.URL.Query().Get("surveyId"), 10, 0)
			if !services.HasPermission(userId, "survey", surveyId, "edit") {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			question := services.CreateQuestion(surveyId, userId)

			template.ExecuteTemplate(w, "questions", []services.Question{question})
		}
		if r.Method == "PUT" {
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
		}
		if r.Method == "DELETE" {
			questionId, _ := strconv.ParseInt(r.URL.Query().Get("questionId"), 10, 0)
			if !services.HasPermission(userId, "question", questionId, "edit") {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			services.DeleteQuestion(questionId)
		}
	})

	http.HandleFunc("/question/edit", func(w http.ResponseWriter, r *http.Request) {
		userId, authErr := context.CheckAuth(r)
		if authErr != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		if r.Method == "GET" {
			questionId, _ := strconv.ParseInt(r.URL.Query().Get("questionId"), 10, 0)
			if !services.HasPermission(userId, "question", questionId, "read") {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			question := services.GetQuestion(questionId)

			template.ExecuteTemplate(w, "edit-question.html", question)
		}
	})

	http.HandleFunc("/option", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			option := services.Option{}

			template.ExecuteTemplate(w, "option", option)
		}
	})
}
