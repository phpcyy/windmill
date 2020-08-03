package models

import "testing"
import _ "github.com/go-sql-driver/mysql"

func TestInitDb(t *testing.T) {
	InitDb()
}
