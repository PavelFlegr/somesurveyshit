package controllers

import (
	"github.com/go-chi/chi/v5"
	"html/template"
	"log"
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

func Survey(template *template.Template, r chi.Router) {
	r.Route("/survey/{surveyId}", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
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
			err = template.ExecuteTemplate(w, "user-survey.html", page)
			if err != nil {
				log.Println(err)
			}
		})

		r.Post("/", func(w http.ResponseWriter, r *http.Request) {
			r.ParseForm()
			surveyId, _ := strconv.ParseInt(chi.URLParam(r, "surveyId"), 10, 0)
			pageNumber, err := strconv.ParseInt(r.URL.Query().Get("page"), 10, 0)
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

				services.MergeResponse(responseId, response)
			}

			var page SurveyPage
			page.Survey = services.GetSurvey(surveyId)
			page.Page = pageNumber + 1
			page.ResponseId = responseId

			if page.Page < int64(len(page.Survey.Blocks)) {
				page.Block, err = services.GetBlock(page.Survey.Blocks[page.Page].Id, surveyId)
				template.ExecuteTemplate(w, "user-survey.html", page)
				return
			}

			http.Redirect(w, r, "/goodbye", http.StatusSeeOther)
		})
	})
}
