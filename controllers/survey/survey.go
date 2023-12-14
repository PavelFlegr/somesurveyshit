package survey

import (
	"encoding/csv"
	"fmt"
	"log"
	"main/global"
	"main/services"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"golang.org/x/exp/slices"
)

func Dashboard(w http.ResponseWriter, r *http.Request) {
	userId, authErr := global.CheckAuth(r)
	if authErr != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	surveys := services.ListSurveys(userId)
	global.Template.ExecuteTemplate(w, "manage/dashboard.html", services.TemplateData{
		LoggedIn: authErr == nil,
		Data:     surveys,
	})
}

func AddSurvey(w http.ResponseWriter, r *http.Request) {
	userId, authErr := global.CheckAuth(r)
	if authErr != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	title := strings.TrimSpace(r.PostFormValue("title"))
	if len(title) < 1 {
		err := global.Template.ExecuteTemplate(w, "error2", "Title can't be empty")
		if err != nil {
			log.Println(err)
		}
		return
	}
	survey := services.CreateSurvey(title, userId)
	err := global.Template.ExecuteTemplate(w, "manage/survey-item", survey)
	if err != nil {
		log.Println(err)
	}
	err = global.Template.ExecuteTemplate(w, "noerror", nil)
	if err != nil {
		log.Println(err)
	}
}

func GetSurvey(w http.ResponseWriter, r *http.Request) {
	userId, authErr := global.CheckAuth(r)
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
		err := global.Template.ExecuteTemplate(w, "manage/survey", survey)
		if err != nil {
			log.Println(err)
		}
	} else {
		err := global.Template.ExecuteTemplate(w, "manage/survey.html", services.TemplateData{
			LoggedIn: authErr == nil,
			Data:     survey,
		})
		if err != nil {
			log.Println(err)
		}
	}
}

func DeleteSurvey(w http.ResponseWriter, r *http.Request) {
	userId, authErr := global.CheckAuth(r)
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
}

func GetSurveyTitle(w http.ResponseWriter, r *http.Request) {
	userId, authErr := global.CheckAuth(r)
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
	err := global.Template.ExecuteTemplate(w, "manage/navigation", survey)
	if err != nil {
		log.Println(err)
	}
}

func PutSurveyTitle(w http.ResponseWriter, r *http.Request) {
	userId, authErr := global.CheckAuth(r)
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

	err := global.Template.ExecuteTemplate(w, "manage/navigation", services.Survey{Id: surveyId, Title: title})
	if err != nil {
		log.Println(err)
	}
}

func GetSurveyTitleEdit(w http.ResponseWriter, r *http.Request) {
	userId, authErr := global.CheckAuth(r)
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
	err := global.Template.ExecuteTemplate(w, "manage/edit-survey-title", survey)
	if err != nil {
		log.Println(err)
	}
}

func DownloadSurvey(w http.ResponseWriter, r *http.Request) {
	surveyId, _ := strconv.ParseInt(chi.URLParam(r, "surveyId"), 10, 0)
	choice := r.FormValue("choice")

	rows, err := global.Db.Query("select response from responses where survey_id = $1", surveyId)
	if err != nil {
		log.Println("manage survey download", err)
	}

	survey := services.GetSurvey(surveyId)
	questions, _ := services.ListQuestionsBySurvey(surveyId)
	blocks, _ := services.ListBlocks(surveyId)

	csvWriter := csv.NewWriter(w)

	var record []string
	for _, block := range blocks {
		record = append(record, block.Title)
		record = append(record, "click time")
		record = append(record, "submit time")
		for _, question := range questions {
			if question.BlockId != block.Id {
				continue
			}
			record = append(record, question.Title)
		}
	}
	err = csvWriter.Write(record)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%v.csv\"", survey.Title))
	csvWriter.Flush()

	var response services.Response
	for rows.Next() {
		err := rows.Scan(&response)
		if err != nil {
			log.Println("manage survey download", err)
		}

		record = []string{}
		for _, block := range blocks {
			blockResponse := response.Blocks[fmt.Sprint(block.Id)]
			record = append(record, "")
			record = append(record, fmt.Sprint(blockResponse.ClickTime))
			record = append(record, fmt.Sprint(blockResponse.SubmitTime))
			for _, question := range questions {
				if question.BlockId != block.Id {
					continue
				}
				answers := response.Questions[fmt.Sprint(question.Id)]
				if slices.Contains([]string{"single", "multiple"}, question.Configuration.QuestionType) && choice == "label" {
					for i := range answers {
						answer, _ := strconv.Atoi(answers[i])
						answers[i] = question.Configuration.Options[answer].Label
					}
				}

				record = append(record, strings.Join(answers, ","))
			}
		}
		err = csvWriter.Write(record)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		csvWriter.Flush()
	}
}

func GetEmptyOption(w http.ResponseWriter, r *http.Request) {
	option := services.Option{}
	err := global.Template.ExecuteTemplate(w, "manage/option", option)
	if err != nil {
		log.Println(err)
	}
}
