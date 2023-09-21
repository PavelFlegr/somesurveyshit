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

	r.Route("/manage/survey/{surveyId}", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			userId, authErr := context.CheckAuth(r)
			if authErr != nil {
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}
			surveyId, _ := strconv.ParseInt(chi.URLParam(r, "surveyId"), 10, 0)
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
					log.Println(err)
				}
			}
		})
		r.Delete("/", func(w http.ResponseWriter, r *http.Request) {
			userId, authErr := context.CheckAuth(r)
			if authErr != nil {
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}
			surveyId, _ := strconv.ParseInt(chi.URLParam(r, "surveyId"), 10, 0)
			if !services.HasPermission(userId, "survey", surveyId, "edit") {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			services.DeleteSurvey(surveyId, userId)
		})
	})

	r.Route("/manage/survey/{surveyId}/title", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			userId, authErr := context.CheckAuth(r)
			if authErr != nil {
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}

			surveyId, _ := strconv.ParseInt(chi.URLParam(r, "surveyId"), 10, 0)
			if !services.HasPermission(userId, "survey", surveyId, "read") {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			survey := services.GetSurvey(surveyId)
			template.ExecuteTemplate(w, "navigation", survey)
		})

		r.Put("/", func(w http.ResponseWriter, r *http.Request) {
			userId, authErr := context.CheckAuth(r)
			if authErr != nil {
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}

			surveyId, _ := strconv.ParseInt(chi.URLParam(r, "surveyId"), 10, 0)
			if !services.HasPermission(userId, "survey", surveyId, "edit") {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			title := r.PostFormValue("title")
			services.RenameSurvey(surveyId, title)

			template.ExecuteTemplate(w, "navigation", services.Survey{Id: surveyId, Title: title})
		})
	})

	r.Put("/manage/survey/{surveyId}/block/{blockId}/title", func(w http.ResponseWriter, r *http.Request) {
		userId, authErr := context.CheckAuth(r)
		if authErr != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		surveyId, _ := strconv.ParseInt(chi.URLParam(r, "surveyId"), 10, 0)
		blockId, _ := strconv.ParseInt(chi.URLParam(r, "blockId"), 10, 0)
		if !services.HasPermission(userId, "survey", surveyId, "read") {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		title := r.PostFormValue("title")
		services.RenameBlock(blockId, surveyId, title)
		block, _ := services.GetBlock(blockId, surveyId)
		template.ExecuteTemplate(w, "block-title", block)
	})

	r.Get("/manage/survey/{surveyId}/block/{blockId}/title", func(w http.ResponseWriter, r *http.Request) {
		userId, authErr := context.CheckAuth(r)
		if authErr != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		surveyId, _ := strconv.ParseInt(chi.URLParam(r, "surveyId"), 10, 0)
		blockId, _ := strconv.ParseInt(chi.URLParam(r, "blockId"), 10, 0)
		if !services.HasPermission(userId, "survey", surveyId, "read") {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		block, _ := services.GetBlock(blockId, surveyId)
		template.ExecuteTemplate(w, "block-title", block)
	})

	r.Get("/manage/survey/{surveyId}/title/edit", func(w http.ResponseWriter, r *http.Request) {
		userId, authErr := context.CheckAuth(r)
		if authErr != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		surveyId, _ := strconv.ParseInt(chi.URLParam(r, "surveyId"), 10, 0)
		if !services.HasPermission(userId, "survey", surveyId, "read") {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		survey := services.GetSurvey(surveyId)
		template.ExecuteTemplate(w, "edit-survey-title", survey)
	})

	r.Get("/manage/survey/{surveyId}/block/{blockId}/title/edit", func(w http.ResponseWriter, r *http.Request) {
		userId, authErr := context.CheckAuth(r)
		if authErr != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		surveyId, _ := strconv.ParseInt(chi.URLParam(r, "surveyId"), 10, 0)
		blockId, _ := strconv.ParseInt(chi.URLParam(r, "blockId"), 10, 0)
		if !services.HasPermission(userId, "survey", surveyId, "read") {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		block, _ := services.GetBlock(blockId, surveyId)
		template.ExecuteTemplate(w, "edit-block-title", block)
	})

	r.Post("/manage/survey/{surveyId}/question/{questionId}/reorder", func(w http.ResponseWriter, r *http.Request) {
		userId, authErr := context.CheckAuth(r)
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

		err := template.ExecuteTemplate(w, "question.html", question)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})

	r.Post("/manage/survey/{surveyId}/block/{blockId}/reorder", func(w http.ResponseWriter, r *http.Request) {
		userId, authErr := context.CheckAuth(r)
		if authErr != nil {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		surveyId, _ := strconv.ParseInt(chi.URLParam(r, "surveyId"), 10, 0)
		blockId, _ := strconv.ParseInt(chi.URLParam(r, "blockId"), 10, 0)
		index, _ := strconv.Atoi(r.PostFormValue("index"))
		if !services.HasPermission(userId, "survey", surveyId, "edit") {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		services.ReorderBlock(surveyId, blockId, index)
	})

	r.Get("/manage/survey/{surveyId}/block/{blockId}/question", func(w http.ResponseWriter, r *http.Request) {
		userId, authErr := context.CheckAuth(r)
		if authErr != nil {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		surveyId, _ := strconv.ParseInt(chi.URLParam(r, "surveyId"), 10, 0)
		blockId, _ := strconv.ParseInt(chi.URLParam(r, "blockId"), 10, 0)
		if !services.HasPermission(userId, "survey", surveyId, "edit") {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		survey, err := services.GetBlock(blockId, surveyId)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = template.ExecuteTemplate(w, "questions", survey.Questions)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})

	r.Post("/manage/survey/{surveyId}/block", func(w http.ResponseWriter, r *http.Request) {
		userId, authErr := context.CheckAuth(r)
		if authErr != nil {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		surveyId, _ := strconv.ParseInt(chi.URLParam(r, "surveyId"), 10, 0)
		if !services.HasPermission(userId, "survey", surveyId, "edit") {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		block := services.Block{
			Title:    "New Block",
			SurveyId: surveyId,
		}
		err := services.CreateBlock(&block, userId)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err = template.ExecuteTemplate(w, "block", block)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})

	r.Post("/manage/survey/{surveyId}/block/{blockId}/question", func(w http.ResponseWriter, r *http.Request) {
		userId, authErr := context.CheckAuth(r)
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

		err := template.ExecuteTemplate(w, "question.html", question)
		if err != nil {
			log.Println(err)
		}
	})

	r.Route("/manage/survey/{surveyId}/question/{questionId}", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			userId, authErr := context.CheckAuth(r)
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

			template.ExecuteTemplate(w, "question.html", question)
		})

		r.Put("/", func(w http.ResponseWriter, r *http.Request) {
			userId, authErr := context.CheckAuth(r)
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

			template.ExecuteTemplate(w, "question.html", question)
		})

		r.Delete("/", func(w http.ResponseWriter, r *http.Request) {
			userId, authErr := context.CheckAuth(r)
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
		})
	})

	r.Delete("/manage/survey/{surveyId}/block/{blockId}", func(w http.ResponseWriter, r *http.Request) {
		userId, authErr := context.CheckAuth(r)
		if authErr != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		blockId, _ := strconv.ParseInt(chi.URLParam(r, "blockId"), 10, 0)
		surveyId, _ := strconv.ParseInt(chi.URLParam(r, "surveyId"), 10, 0)
		if !services.HasPermission(userId, "survey", surveyId, "edit") {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		services.RemoveBlock(surveyId, blockId)
	})

	r.Get("/manage/survey/{surveyId}/question/{questionId}/edit", func(w http.ResponseWriter, r *http.Request) {
		userId, authErr := context.CheckAuth(r)
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

		template.ExecuteTemplate(w, "edit-question.html", question)
	})

	r.Get("/manage/option", func(w http.ResponseWriter, r *http.Request) {
		option := services.Option{}
		template.ExecuteTemplate(w, "option", option)
	})

	r.Get("/manage/survey/{surveyId}/download", func(w http.ResponseWriter, r *http.Request) {
		surveyId, _ := strconv.ParseInt(chi.URLParam(r, "surveyId"), 10, 0)

		rows, err := context.Ctx.Db.Query("select response from responses where survey_id = $1", surveyId)
		if err != nil {
			log.Println("manage survey download", err)
		}

		var questions []services.Question
		questions, err = services.ListQuestionsBySurvey(surveyId)
		var questionIds []string

		csvWriter := csv.NewWriter(w)

		var record []string
		for _, question := range questions {
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
