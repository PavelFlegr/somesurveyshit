package answer

import (
	"github.com/go-chi/chi/v5"
	"log"
	"main/global"
	"main/services"
	"net/http"
	"strconv"
)

type SurveyPage struct {
	Survey     services.Survey
	Block      services.Block
	Page       int64
	ResponseId int64
}

func GetSurvey(w http.ResponseWriter, r *http.Request) {
	surveyId, _ := strconv.ParseInt(chi.URLParam(r, "surveyId"), 10, 0)
	pageNumber, err := strconv.ParseInt(r.URL.Query().Get("page"), 10, 0)
	if err != nil {
		pageNumber = 0
	}

	var page SurveyPage
	page.Survey = services.GetSurvey(surveyId)
	page.Block, err = services.GetBlock(page.Survey.Blocks[pageNumber].Id, surveyId)
	page.Page = pageNumber
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = global.Template.ExecuteTemplate(w, "answer/survey.html", page)
	if err != nil {
		log.Println(err)
	}
}

func SubmitAnswers(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
	}
	surveyId, _ := strconv.ParseInt(chi.URLParam(r, "surveyId"), 10, 0)
	var pageNumber int64
	pageNumber, err = strconv.ParseInt(r.URL.Query().Get("page"), 10, 0)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	response := services.Response{}
	for k, v := range r.PostForm {
		if k == "responseId" {
			continue
		}
		questionId, _ := strconv.ParseInt(k, 10, 0)
		question := services.GetQuestion(surveyId, questionId)
		answerIndex, _ := strconv.ParseInt(v[0], 10, 0)
		response[k] = question.Options[answerIndex].Label
	}

	var responseId int64
	if pageNumber == 0 {
		responseId, err = services.RecordResponse(surveyId, response)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		responseId, err = strconv.ParseInt(r.PostFormValue("responseId"), 10, 0)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err := services.MergeResponse(responseId, response)
		if err != nil {
			log.Println(err)
		}
	}

	var page SurveyPage
	page.Survey = services.GetSurvey(surveyId)
	page.Page = pageNumber + 1
	page.ResponseId = responseId

	if page.Page < int64(len(page.Survey.Blocks)) {
		page.Block, err = services.GetBlock(page.Survey.Blocks[page.Page].Id, surveyId)
		err := global.Template.ExecuteTemplate(w, "answer/survey.html", page)
		if err != nil {
			log.Println(err)
		}
		return
	}

	http.Redirect(w, r, "/goodbye", http.StatusSeeOther)
}

func ShowGoodbye(w http.ResponseWriter, r *http.Request) {
	err := global.Template.ExecuteTemplate(w, "answer/goodbye.html", nil)
	if err != nil {
		log.Println(err)
	}
}
