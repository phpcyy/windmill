package generator

import (
	"bytes"
	"fmt"
	"github.com/iancoleman/strcase"
	"go/format"
	"gopkg.in/yaml.v2"
	"io"
	"strings"
)

type Property struct {
	Name string
	Type string
	Perm int
	Desc string
}

type Scheme struct {
	Name        string
	Description string
	Path        string
	Properties  []Property
}

func Decode(r io.Reader) (*Scheme, error) {
	scheme := new(Scheme)
	decoder := yaml.NewDecoder(r)
	err := decoder.Decode(&scheme)
	for i := range scheme.Properties {
		scheme.Properties[i].Name = strcase.ToCamel(scheme.Properties[i].Name)
	}
	scheme.Properties = append(scheme.Properties, Property{
		Name: "Id",
		Type: "int",
		Perm: 0b0000,
	}, Property{
		Name: "CreateTime",
		Type: "time.Time",
		Perm: 0b0000,
	}, Property{
		Name: "UpdateTime",
		Type: "time.Time",
		Perm: 0b0000,
	})
	return scheme, err
}

func (s *Scheme) GenModel() ([]byte, error) {
	lines := [][]byte{[]byte("package models"), []byte(`import "time"`), []byte(fmt.Sprintf("type %s struct {", s.Name))}
	for _, property := range s.Properties {
		lines = append(lines, []byte(fmt.Sprintf("%s %s", property.Name, property.Type)))
	}
	lines = append(lines, []byte{'}'})
	addStr, err := s.GenAdd()
	if err != nil {
		return nil, err
	}
	lines = append(lines, addStr)

	return format.Source(bytes.Join(lines, []byte{'\n'}))
}

func (s *Scheme) GenAdd() ([]byte, error) {
	param := strcase.ToLowerCamel(s.Name)
	str := fmt.Sprintf("func (%s *%s) Add() (int64, error) {\n", param, s.Name)

	var fields []string
	var values []string
	for _, property := range s.Properties {
		fields = append(fields, fmt.Sprintf("%s=?", strcase.ToSnake(property.Name)))
		values = append(values, fmt.Sprintf("%s.%s", param, property.Name))
	}

	str += fmt.Sprintf("stmt, err := Db.Prepare(\"insert into `%s` set %s\")\n", strcase.ToSnake(s.Name), strings.Join(fields, ","))

	str += fmt.Sprintf(`if err != nil {
		return 0, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(%s)
	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}`, strings.Join(values, ","))

	return []byte(str), nil
}

var MySQLTypeMap = map[string]string{
	"int":       "int",
	"string":    "varchar(128)",
	"time.Time": "datetime",
	"bool":      "tinyint",
	"float":     "double",
}

func (s *Scheme) GenTable() (string, error) {
	fields := make([]string, 0)
	for _, property := range s.Properties {
		switch property.Name {
		case "Id":
			fields = append(fields, fmt.Sprintf("`%s` %s primary key auto_increment comment '%s'", strcase.ToSnake(property.Name), MySQLTypeMap[property.Type], property.Desc))
		case "CreateTime":
			fields = append(fields, fmt.Sprintf("`%s` %s not null default CURRENT_TIMESTAMP comment '%s'", strcase.ToSnake(property.Name), MySQLTypeMap[property.Type], property.Desc))
		case "UpdateTime":
			fields = append(fields, fmt.Sprintf("`%s` %s not null default CURRENT_TIMESTAMP on update CURRENT_TIMESTAMP comment '%s'", strcase.ToSnake(property.Name), MySQLTypeMap[property.Type], property.Desc))
		default:
			fields = append(fields, fmt.Sprintf("`%s` %s not null comment '%s'", strcase.ToSnake(property.Name), MySQLTypeMap[property.Type], property.Desc))
		}

	}
	return fmt.Sprintf("create table if not exists `%s`(%s);", strcase.ToSnake(s.Name), strings.Join(fields, ",\n")), nil
}

func (s *Scheme) GenApi() string {
	param := strcase.ToLowerCamel(s.Name)
	str := fmt.Sprintf(`
		package main

		import (
			"encoding/json"
			"fmt"
			"io/ioutil"
			"net/http"
			"github.com/phpcyy/windmill/models"
		)
		func InitRouter() *http.ServeMux {
			mux := http.ServeMux{}

			mux.HandleFunc("%s", func(writer http.ResponseWriter, request *http.Request) {
			bodyBytes, err := ioutil.ReadAll(request.Body)
			if err != nil {
				writer.WriteHeader(http.StatusBadRequest)
				return
			}

			%s := models.%s{}
			err = json.Unmarshal(bodyBytes, &%s)
			if err != nil {
				writer.WriteHeader(http.StatusBadRequest)
				return
			}
	
	
			id, err := %s.Add()
			if err != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				return
			}`, s.Path, param, s.Name, param, param)
	str += "\nwriter.Write([]byte(fmt.Sprintf(`{\"id\": %d}`, id)))\nreturn\n})"
	str += `
	return &mux
}`
	return str
}
