package generator

import (
	"fmt"
	"github.com/iancoleman/strcase"
	"go/format"
	"io/ioutil"
	"os"
	"strings"
)

func NewBuilder(workDir string) *Builder {
	return &Builder{
		workDir: workDir,
	}
}

type Builder struct {
	workDir string
	routes  []string
}

func (b *Builder) BuildAll() error {
	var paths []string
	dir, err := ioutil.ReadDir(b.workDir)
	if err != nil {
		return err
	}

	for _, f := range dir {
		if f.IsDir() {
			continue
		}

		if strings.HasSuffix(f.Name(), ".yml") {
			paths = append(paths, fmt.Sprintf(f.Name()))
		}
	}

	for _, path := range paths {
		err = b.Build(path)
		if err != nil {
			return err
		}
	}

	return nil
}

func (b *Builder) FlushRoutes() error {
	routeFile, err := os.OpenFile("../route.go", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}

	defer func() { _ = routeFile.Close() }()
	fileText := []byte(fmt.Sprintf(`
package main

import (
	"github.com/gorilla/mux"
	"github.com/phpcyy/windmill/controllers"
)

func InitRouter(router *mux.Router) {	
%s
}
`, strings.Join(b.routes, "\n")))
	fmt.Println(string(fileText))
	fileText, err = format.Source(fileText)
	if err != nil {
		return err
	}
	_, err = routeFile.Write(fileText)
	if err != nil {
		return err
	}
	return nil
}

func (b *Builder) Build(name string) error {
	apiFile, err := os.Open(fmt.Sprintf("%s%c%s", b.workDir, os.PathSeparator, name))
	if err != nil {
		return err
	}

	defer func() {
		_ = apiFile.Close()
	}()

	scheme, err := Decode(apiFile)
	if err != nil {
		return err
	}

	modelFile, err := os.OpenFile(fmt.Sprintf("../models/%s.go", strcase.ToLowerCamel(scheme.Name)), os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}

	defer func() { _ = modelFile.Close() }()

	modelContent, err := scheme.GenModel()
	if err != nil {
		return err
	}

	_, _ = modelFile.Write(modelContent)

	tableContent, err := scheme.GenTable()
	if err != nil {
		return err
	}

	tableFile, err := os.OpenFile(fmt.Sprintf("../models/sql/%s.sql", strcase.ToSnake(scheme.Name)), os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return err
	}

	_, _ = tableFile.WriteString(tableContent)

	controllerFile, err := os.OpenFile(fmt.Sprintf("../controllers/%s.go", strcase.ToLowerCamel(scheme.Name)), os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	fileText := []byte(scheme.GenController())
	fileText, err = format.Source(fileText)
	if err != nil {
		return err
	}

	_, err = controllerFile.Write(fileText)

	if err != nil {
		return err
	}

	b.routes = append(b.routes, scheme.GenRoutes()...)
	return nil
}
