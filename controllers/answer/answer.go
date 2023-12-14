package answer

import (
	"fmt"
	"log"
	"main/global"
	"main/services"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

var s = rand.NewSource(time.Now().UnixNano())

type SurveyPage struct {
	Survey     services.Survey
	Block      services.Block
	Page       int64
	ResponseId int64
}

func randomizeBlock(block *services.Block) {
	shuffled := make([]services.Question, len(block.Questions))
	for i, j := range rand.New(s).Perm(len(block.Questions)) {
		shuffled[i] = block.Questions[j]
	}
	block.Questions = shuffled
}

func randomizeQuestions(questions []services.Question) {
	for q := range questions {
		switch questions[q].Configuration.QuestionType {
		case "multiple", "single":
		default:
			continue
		}
		if !questions[q].Configuration.Randomize {
			for i := range questions[q].Configuration.Options {
				questions[q].Configuration.Options[i].Id = int64(i)
			}
			return
		}
		shuffled := make([]services.Option, len(questions[q].Configuration.Options))
		for i, j := range rand.New(s).Perm(len(questions[q].Configuration.Options)) {
			shuffled[i] = questions[q].Configuration.Options[j]
			shuffled[i].Id = int64(j)
		}
		questions[q].Configuration.Options = shuffled
	}
}

func RenderPage(w http.ResponseWriter, page SurveyPage) {
	if page.Block.Randomize {
		randomizeBlock(&page.Block)
	}
	randomizeQuestions(page.Block.Questions)
	err := global.Template.ExecuteTemplate(w, "answer/survey.html", page)
	if err != nil {
		log.Println(err)
	}
}

func GetSurvey(w http.ResponseWriter, r *http.Request) {
	surveyId, _ := strconv.ParseInt(chi.URLParam(r, "surveyId"), 10, 0)
	pageNumber, err := strconv.ParseInt(r.URL.Query().Get("page"), 10, 0)
	if err != nil {
		pageNumber = 0
	}

	var page SurveyPage
	page.Survey = services.GetSurvey(surveyId)
	page.Block, _ = services.GetBlock(page.Survey.Blocks[pageNumber].Id, surveyId)
	page.Page = pageNumber

	RenderPage(w, page)
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

	response := services.Response{
		Questions: map[string]services.QuestionResponse{},
		Blocks:    map[string]services.BlockResponse{},
	}
	for k, v := range r.PostForm {
		if k == "responseId" || k == "clickTime" || k == "submitTime" || k == "blockId" {
			continue
		}
		response.Questions[k] = v
	}

	clickTime, _ := strconv.ParseInt(r.PostFormValue("clickTime"), 10, 0)
	submitTime, _ := strconv.ParseInt(r.PostFormValue("submitTime"), 10, 0)

	response.Blocks[fmt.Sprint(r.PostFormValue("blockId"))] = services.BlockResponse{
		ClickTime:  clickTime,
		SubmitTime: submitTime,
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
		page.Block, _ = services.GetBlock(page.Survey.Blocks[page.Page].Id, surveyId)
		RenderPage(w, page)
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
