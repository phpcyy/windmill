package models

import "database/sql"

var Db *sql.DB

func InitDb() {
	var err error
	Db, err = sql.Open("mysql", "root:123456@tcp(127.0.0.1:3306)/windmill?charset=utf8mb4&parseTime=True&loc=Local&interpolateParams=true")
	if err != nil {
		panic(err)
	}
}
