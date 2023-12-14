package services

import (
	"log"
	"main/global"
	"time"

	"github.com/lib/pq"
)

func ListSurveys(userId int64) []Survey {
	rows, err := global.Db.Query("select id, title, created, updated from surveys inner join public.survey_permissions sp on surveys.id = sp.entity_id where sp.user_id = $1 and sp.action = 'read'", userId)
	if err != nil {
		log.Panic("Couldn't list surveys", err)
	}
	var surveys []Survey
	for rows.Next() {
		var survey Survey
		err = rows.Scan(&survey.Id, &survey.Title, &survey.Created, &survey.Updated)
		if err != nil {
			log.Panic("Couldn't scan survey", err)
		}
		surveys = append(surveys, survey)
	}

	return surveys
}

func GetSurvey(surveyId int64) Survey {
	var survey Survey
	survey.Id = surveyId
	var created string
	var updated string
	var blocksOrder []int64
	err := global.Db.QueryRow("select title, created, updated, blocks_order from surveys where id = $1", surveyId).Scan(&survey.Title, &created, &updated, (*pq.Int64Array)(&blocksOrder))
	if err != nil {
		log.Print(err)
	}

	createdTime, _ := time.Parse(time.RFC3339, created)
	updatedTime, _ := time.Parse(time.RFC3339, updated)
	survey.Created = createdTime
	survey.Updated = updatedTime

	tmp := make(map[int64]Block)
	survey.Blocks, err = ListBlocks(surveyId)
	if err != nil {
		println(err)
		return survey
	}
	for _, b := range survey.Blocks {
		tmp[b.Id] = b
	}
	for i, b := range blocksOrder {
		survey.Blocks[i] = tmp[b]
	}

	return survey
}

func CreateSurvey(title string, userId int64) Survey {
	var survey Survey
	survey.Title = title
	err := global.Db.QueryRow("insert into surveys (title, created, updated, user_id) values ($1, now(), now(), $2) returning id, created, updated", title, userId).Scan(&survey.Id, &survey.Created, &survey.Updated)
	if err != nil {
		log.Print("Couldn't create survey", err)
	}

	AddPermission(userId, "survey", survey.Id, "manage")
	AddPermission(userId, "survey", survey.Id, "edit")
	AddPermission(userId, "survey", survey.Id, "read")

	return survey
}

func RenameSurvey(surveyId int64, title string) {
	_, err := global.Db.Exec("update surveys set updated = now(), title = $1 where id = $2", title, surveyId)
	if err != nil {
		log.Print("Couldn't update survey title", err)
	}
}

func DeleteSurvey(surveyId int64, userId int64) {
	_, err := global.Db.Exec("delete from surveys where id = $1 and user_id = $2", surveyId, userId)
	if err != nil {
		log.Print("Couldn't delete survey", err)
	}
}

func ReorderBlock(surveyId int64, blockId int64, index int) {
	_, err := global.Db.Exec("update surveys set blocks_order = array_remove(blocks_order, $1) where id = $2", blockId, surveyId)
	if err != nil {
		log.Println(err)
		return
	}

	var blocksOrder []int64
	err = global.Db.QueryRow("select blocks_order from surveys where id = $1", surveyId).Scan((*pq.Int64Array)(&blocksOrder))
	if err != nil {
		log.Println(err)
		return
	}

	var newOrder []int64
	if index == 0 {
		newOrder = append(newOrder, blockId)
	}
	for i := 0; i < len(blocksOrder); i++ {
		newOrder = append(newOrder, blocksOrder[i])
		if i+1 == index {
			newOrder = append(newOrder, blockId)
		}
	}

	_, err = global.Db.Exec("update surveys set blocks_order = $1, updated = now() where id = $2", pq.Array(newOrder), surveyId)
	if err != nil {
		log.Println(err)
		return
	}
}

func RecordResponse(surveyId int64, response Response) (int64, error) {
	var responseId int64
	err := global.Db.QueryRow("insert into responses (survey_id, response) VALUES ($1, $2) returning id", surveyId, response).Scan(&responseId)
	if err != nil {
		log.Println(err)
	}

	return responseId, err
}

func MergeResponse(responseId int64, response Response) error {
	var old Response
	global.Db.QueryRow("select response from responses where id = $1", responseId).Scan(&old)
	for k, v := range response.Blocks {
		old.Blocks[k] = v
	}
	for k, v := range response.Questions {
		old.Questions[k] = v
	}
	_, err := global.Db.Exec("update responses set response = $1 where id = $2", old, responseId)
	if err != nil {
		log.Println(err)
	}

	return err
}
