package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
)

var db *sql.DB

type Question struct {
	Id          int64
	Title       string
	Description string
	Options     []Option
}

type Option struct {
	Id         int64
	QuestionId int64
	Value      string
}

func main() {
	port := os.Args[1]
	connStr := os.Args[2]

	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	indexTmpl, err := template.New("index.html").Funcs(template.FuncMap{
		"unescape": func(val string) template.HTML {
			return template.HTML(val)
		},
	}).ParseFiles("index.html", "question.html")

	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		println(r.URL.Path)
		questions := listQuestions()
		indexTmpl.Execute(w, questions)
	})

	http.HandleFunc("/question", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			r.ParseForm()
			question := createQuestion()

			indexTmpl.ExecuteTemplate(w, "questions", []Question{question})
		}
		if r.Method == "DELETE" {
			questionId, _ := strconv.ParseInt(r.URL.Query().Get("questionId"), 10, 0)
			deleteQuestion(questionId)
		}
	})

	http.HandleFunc("/question/title", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "PUT" {
			r.ParseForm()
			query := r.URL.Query()
			questionId := query.Get("questionId")
			title := r.FormValue("title")
			updateTitle(questionId, title)
		}
	})

	http.HandleFunc("/question/description", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "PUT" {
			r.ParseForm()
			query := r.URL.Query()
			questionId, _ := strconv.ParseInt(query.Get("questionId"), 10, 0)
			description := r.FormValue("description")
			updateDescription(questionId, description)

			fmt.Fprint(w, description)
		}
	})

	http.HandleFunc("/question/option", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "DELETE" {
			query := r.URL.Query()
			questionId, _ := strconv.ParseInt(query.Get("questionId"), 10, 0)
			optionId, _ := strconv.ParseInt(query.Get("optionId"), 10, 0)

			deleteOption(questionId, optionId)
			return
		} else if r.Method == "POST" {
			r.ParseForm()
			query := r.URL.Query()
			questionId, _ := strconv.ParseInt(query.Get("questionId"), 10, 0)
			option := r.FormValue("option")
			options := addOption(questionId, option)
			indexTmpl.ExecuteTemplate(w, "option", options)
		} else if r.Method == "PUT" {
			r.ParseForm()
			query := r.URL.Query()
			questionId, _ := strconv.ParseInt(query.Get("questionId"), 10, 0)
			optionId, _ := strconv.ParseInt(query.Get("optionId"), 10, 0)
			option := r.FormValue("option")
			updateOption(questionId, optionId, option)
			indexTmpl.ExecuteTemplate(w, "option", Option{optionId, questionId, option})
		}
	})

	http.HandleFunc("/question/option/reorder", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			r.ParseForm()
			query := r.URL.Query()
			questionId, _ := strconv.ParseInt(query.Get("questionId"), 10, 0)
			oldIndex, _ := strconv.ParseInt(r.FormValue("old"), 10, 0)
			newIndex, _ := strconv.ParseInt(r.FormValue("new"), 10, 0)

			reorderOptions(questionId, oldIndex, newIndex)
		}
	})

	http.ListenAndServe(fmt.Sprintf(":%v", port), nil)
}
