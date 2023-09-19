package services

import (
	"fmt"
	"log"
	"main/context"
)

func CreateUser(email string, password string) error {
	_, err := context.Ctx.Db.Exec("insert into users (email, password) values ($1, $2)", email, password)
	if err != nil {
		log.Println(err)
	}

	return err
}

func GetUserByEmail(email string) (User, error) {
	var user User
	err := context.Ctx.Db.QueryRow("select id, email, password from users where email = $1", email).Scan(&user.Id, &user.Email, &user.Password)
	if err != nil {
		log.Println(err)
	}

	return user, err
}

func AddPermission(userId int64, entityType string, entityId int64, action string) {
	query := fmt.Sprintf("insert into %v_permissions (user_id, action, entity_id) VALUES ($1, $2, $3)", entityType)
	_, err := context.Ctx.Db.Exec(query, userId, action, entityId)
	if err != nil {
		log.Println(err)
	}
}

func HasPermission(userId int64, entityType string, entityId int64, action string) bool {
	query := fmt.Sprintf("select exists(select 1 from %v_permissions where user_id = $1 and entity_id = $2 and action = $3)", entityType)
	var found bool
	err := context.Ctx.Db.QueryRow(query, userId, entityId, action).Scan(&found)
	if err != nil {
		log.Println(err)
	}
	return found
}

func RemovePermission(userId int64, entityType string, entityId int64, action string) {
	query := fmt.Sprintf("delete from %v_permissions where user_id = $1 and entity_id = $2 and action = $3", entityType)
	_, err := context.Ctx.Db.Exec(query, userId, entityId, action)
	if err != nil {
		log.Println(err)
	}
}
