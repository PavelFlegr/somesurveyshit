package block

import (
	"fmt"
	"log"
	"main/global"
	"main/services"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func PutBlockRandomize(w http.ResponseWriter, r *http.Request) {
	userId, authErr := global.CheckAuth(r)
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
	randomize := r.PostFormValue("randomize") == "true"
	err := services.SetRandomize(blockId, surveyId, randomize)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func PutBlockTitle(w http.ResponseWriter, r *http.Request) {
	userId, authErr := global.CheckAuth(r)
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
	err := services.RenameBlock(blockId, surveyId, title)
	if err != nil {
		log.Println(err)
	}
	block, _ := services.GetBlock(blockId, surveyId)
	err = global.Template.ExecuteTemplate(w, "manage/block-title", block)
	if err != nil {
		log.Println(err)
	}
}

func GetBlockTitle(w http.ResponseWriter, r *http.Request) {
	userId, authErr := global.CheckAuth(r)
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
	err := global.Template.ExecuteTemplate(w, "manage/block-title", block)
	if err != nil {
		log.Println(err)
	}
}

func GetBlockTitleEdit(w http.ResponseWriter, r *http.Request) {
	userId, authErr := global.CheckAuth(r)
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
	err := global.Template.ExecuteTemplate(w, "manage/edit-block-title", block)
	if err != nil {
		log.Println(err)
	}
}

func ReorderBlock(w http.ResponseWriter, r *http.Request) {
	userId, authErr := global.CheckAuth(r)
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
}

func CreateBlock(w http.ResponseWriter, r *http.Request) {
	userId, authErr := global.CheckAuth(r)
	if authErr != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	surveyId, _ := strconv.ParseInt(chi.URLParam(r, "surveyId"), 10, 0)
	if !services.HasPermission(userId, "survey", surveyId, "edit") {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	blockCount := services.CountBlocks(surveyId)
	block := services.Block{
		Title:    fmt.Sprintf("Block %v", blockCount+1),
		SurveyId: surveyId,
	}
	err := services.CreateBlock(&block, userId)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var index int
	index, err = strconv.Atoi(r.PostFormValue("index"))
	if err == nil {
		services.ReorderBlock(surveyId, block.Id, index)
	}

	err = global.Template.ExecuteTemplate(w, "manage/block", block)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func ListQuestions(w http.ResponseWriter, r *http.Request) {
	userId, authErr := global.CheckAuth(r)
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

	err = global.Template.ExecuteTemplate(w, "manage/questions", survey.Questions)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func DeleteBlock(w http.ResponseWriter, r *http.Request) {
	userId, authErr := global.CheckAuth(r)
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
	err := services.RemoveBlock(surveyId, blockId)
	if err != nil {
		log.Println(err)
	}
}
