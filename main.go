package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"html/template"
	"log"
	"main/context"
	_ "main/context"
	"main/services"
	"net/http"
	"os"
	"strconv"
)

func main() {
	port := os.Args[1]
	connStr := os.Args[2]

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	context.Ctx = context.AppContext{Db: db}
	indexTmpl, err := template.New("index.html").Funcs(template.FuncMap{
		"unescape": func(val string) template.HTML {
			return template.HTML(val)
		},
	}).ParseGlob("templates/*")

	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		surveys := services.ListSurveys()
		indexTmpl.Execute(w, surveys)
	})

	http.HandleFunc("/survey", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			surveyId, _ := strconv.ParseInt(r.URL.Query().Get("surveyId"), 10, 0)
			survey := services.GetSurvey(surveyId)
			if r.Header.Get("Hx-Request") == "true" {
				indexTmpl.ExecuteTemplate(w, "survey", survey)
			} else {
				indexTmpl.ExecuteTemplate(w, "survey.html", survey)
			}
		}
		if r.Method == "POST" {
			r.ParseForm()
			title := r.FormValue("title")
			survey := services.CreateSurvey(title)

			indexTmpl.ExecuteTemplate(w, "surveyItem", survey)
		}
		if r.Method == "DELETE" {
			surveyId, _ := strconv.ParseInt(r.URL.Query().Get("surveyId"), 10, 0)
			services.DeleteSurvey(surveyId)
		}
	})

	http.HandleFunc("/survey/title", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			surveyId, _ := strconv.ParseInt(r.URL.Query().Get("surveyId"), 10, 0)
			survey := services.GetSurvey(surveyId)
			indexTmpl.ExecuteTemplate(w, "navigation", survey)
		}
		if r.Method == "PUT" {
			surveyId, _ := strconv.ParseInt(r.URL.Query().Get("surveyId"), 10, 0)
			title := r.PostFormValue("title")
			services.RenameSurvey(surveyId, title)

			indexTmpl.ExecuteTemplate(w, "navigation", services.Survey{Id: surveyId, Title: title})
		}
	})

	http.HandleFunc("/survey/title/edit", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			surveyId, _ := strconv.ParseInt(r.URL.Query().Get("surveyId"), 10, 0)
			survey := services.GetSurvey(surveyId)
			indexTmpl.ExecuteTemplate(w, "edit-survey-title", survey)
		}
	})

	http.HandleFunc("/survey/reorder", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "PUT" {
			r.ParseForm()
			surveyId, _ := strconv.ParseInt(r.URL.Query().Get("surveyId"), 10, 0)
			var questionsOrder []int64
			for _, question := range r.PostForm["question"] {
				id, _ := strconv.ParseInt(question, 10, 0)
				questionsOrder = append(questionsOrder, id)
			}

			services.ReorderSurvey(surveyId, questionsOrder)
		}
	})

	http.HandleFunc("/question", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			questionId, _ := strconv.ParseInt(r.URL.Query().Get("questionId"), 10, 0)
			question := services.GetQuestion(questionId)

			indexTmpl.ExecuteTemplate(w, "question.html", question)
		}
		if r.Method == "POST" {
			surveyId, _ := strconv.ParseInt(r.URL.Query().Get("surveyId"), 10, 0)
			question := services.CreateQuestion(surveyId)

			indexTmpl.ExecuteTemplate(w, "questions", []services.Question{question})
		}
		if r.Method == "PUT" {
			questionId, _ := strconv.ParseInt(r.URL.Query().Get("questionId"), 10, 0)
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

			indexTmpl.ExecuteTemplate(w, "question.html", question)
		}
		if r.Method == "DELETE" {
			questionId, _ := strconv.ParseInt(r.URL.Query().Get("questionId"), 10, 0)
			services.DeleteQuestion(questionId)
		}
	})

	http.HandleFunc("/question/edit", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			questionId, _ := strconv.ParseInt(r.URL.Query().Get("questionId"), 10, 0)
			question := services.GetQuestion(questionId)

			indexTmpl.ExecuteTemplate(w, "edit-question.html", question)
		}
	})

	http.HandleFunc("/option", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			option := services.Option{}

			indexTmpl.ExecuteTemplate(w, "option", option)
		}
	})

	http.ListenAndServe(fmt.Sprintf(":%v", port), nil)
}
