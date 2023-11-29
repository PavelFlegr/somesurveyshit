package question

import (
	"log"
	"main/global"
	"main/services"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
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

	err := global.Template.ExecuteTemplate(w, "manage/question", question)
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
	configuration := services.Configuration{
		QuestionType: "description",
		Options:      []services.Option{},
	}
	question := services.CreateQuestion(surveyId, userId, blockId, configuration)

	index, err := strconv.Atoi(r.PostFormValue("index"))
	if err == nil {
		services.ReorderQuestion(surveyId, question.Id, blockId, index)
	}

	err = global.Template.ExecuteTemplate(w, "manage/edit-question", question)
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

	err := global.Template.ExecuteTemplate(w, "manage/question", question)
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
	questionType := r.FormValue("questionType")
	randomize := r.FormValue("randomize") == "true"
	var options []services.Option
	for _, option := range r.PostForm["option"] {
		options = append(options, services.Option{Label: option})
	}
	question := services.Question{
		Id:          questionId,
		Title:       title,
		Description: description,
		Configuration: services.Configuration{
			QuestionType: questionType,
			Randomize:    randomize,
			Options:      options,
		},
		SurveyId: surveyId,
	}
	services.UpdateQuestion(surveyId, &question)

	err := global.Template.ExecuteTemplate(w, "manage/question", question)
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

	err := global.Template.ExecuteTemplate(w, "manage/edit-question", question)
	if err != nil {
		log.Println(err)
	}
}
