package services

import (
	"log"
	"main/context"
)

func CreateUser(email string, password string) error {
	_, err := context.Ctx.Db.Exec("insert into users (email, password) values ($1, $2)", email, password)
	if err != nil {
		log.Println("Couldn't create user", err)
	}

	return err
}

func GetUserByEmail(email string) (User, error) {
	var user User
	err := context.Ctx.Db.QueryRow("select id, email, password from users where email = $1", email).Scan(&user.Id, &user.Email, &user.Password)
	if err != nil {
		log.Println("Couldn't get user by email", err)
	}

	return user, err
}
