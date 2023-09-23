package question

import (
	"github.com/go-chi/chi/v5"
	"log"
	"main/global"
	"main/services"
	"net/http"
	"strconv"
)

func ReorderQuestion(w http.ResponseWriter, r *http.Request) {
	userId, authErr := global.CheckAuth(r)
	if authErr != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	surveyId, _ := strconv.ParseInt(chi.URLParam(r, "surveyId"), 10, 0)
	questionId, _ := strconv.ParseInt(chi.URLParam(r, "questionId"), 10, 0)
	blockId, _ := strconv.ParseInt(r.PostFormValue("blockId"), 10, 0)
	index, _ := strconv.Atoi(r.PostFormValue("index"))
	if !services.HasPermission(userId, "survey", surveyId, "edit") {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	services.ReorderQuestion(surveyId, questionId, blockId, index)
	question := services.GetQuestion(surveyId, questionId)

	err := global.Template.ExecuteTemplate(w, "question.html", question)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func CreateQuestion(w http.ResponseWriter, r *http.Request) {
	userId, authErr := global.CheckAuth(r)
	if authErr != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	surveyId, _ := strconv.ParseInt(chi.URLParam(r, "surveyId"), 10, 0)
	blockId, _ := strconv.ParseInt(chi.URLParam(r, "blockId"), 10, 0)
	if !services.HasPermission(userId, "survey", surveyId, "edit") {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	question := services.CreateQuestion(surveyId, userId, blockId)

	err := global.Template.ExecuteTemplate(w, "question.html", question)
	if err != nil {
		log.Println(err)
	}
}

func GetQuestion(w http.ResponseWriter, r *http.Request) {
	userId, authErr := global.CheckAuth(r)
	if authErr != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	questionId, _ := strconv.ParseInt(chi.URLParam(r, "questionId"), 10, 0)
	surveyId, _ := strconv.ParseInt(chi.URLParam(r, "surveyId"), 10, 0)
	if !services.HasPermission(userId, "survey", surveyId, "read") {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	question := services.GetQuestion(surveyId, questionId)

	err := global.Template.ExecuteTemplate(w, "question.html", question)
	if err != nil {
		log.Println(err)
	}
}

func PutQuestion(w http.ResponseWriter, r *http.Request) {
	userId, authErr := global.CheckAuth(r)
	if authErr != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	questionId, _ := strconv.ParseInt(chi.URLParam(r, "questionId"), 10, 0)
	surveyId, _ := strconv.ParseInt(chi.URLParam(r, "surveyId"), 10, 0)
	if !services.HasPermission(userId, "survey", surveyId, "edit") {
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
		SurveyId:    surveyId,
	}
	services.UpdateQuestion(surveyId, question)

	err := global.Template.ExecuteTemplate(w, "question.html", question)
	if err != nil {
		log.Println(err)
	}
}

func DeleteQuestion(w http.ResponseWriter, r *http.Request) {
	userId, authErr := global.CheckAuth(r)
	if authErr != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	questionId, _ := strconv.ParseInt(chi.URLParam(r, "questionId"), 10, 0)
	surveyId, _ := strconv.ParseInt(chi.URLParam(r, "surveyId"), 10, 0)
	if !services.HasPermission(userId, "survey", surveyId, "edit") {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	services.DeleteQuestion(surveyId, questionId)
}

func GetQuestionEdit(w http.ResponseWriter, r *http.Request) {
	userId, authErr := global.CheckAuth(r)
	if authErr != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	questionId, _ := strconv.ParseInt(chi.URLParam(r, "questionId"), 10, 0)
	surveyId, _ := strconv.ParseInt(chi.URLParam(r, "surveyId"), 10, 0)
	if !services.HasPermission(userId, "survey", surveyId, "read") {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	question := services.GetQuestion(surveyId, questionId)

	err := global.Template.ExecuteTemplate(w, "edit-question.html", question)
	if err != nil {
		log.Println(err)
	}
}
