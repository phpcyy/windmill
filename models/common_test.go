package models

import (
	"fmt"
	"log"
	"testing"
)
import _ "github.com/go-sql-driver/mysql"

func TestInitDb(t *testing.T) {
	InitDb()
}

func Query() ([]*User, error) {
	stmt, err := Db.Prepare("select `name` from `user`")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		return nil, err
	}

	var userList []*User
	for rows.Next() {
		var user User
		err = rows.Scan(&user.Name)
		if err != nil {
			log.Println(err)
			continue
		}
		userList = append(userList, &user)
	}

	return userList, nil
}

func TestQuery(t *testing.T) {
	InitDb()
	rows, err := Query()
	if err != nil {
		t.Error(err)
		return
	}

	for _, row := range rows {
		fmt.Println(row)
	}
}
