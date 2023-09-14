package context

import "database/sql"

type AppContext struct {
	Db *sql.DB
}

var Ctx AppContext
