package main

import (
	"fmt"
	"github.com/fvbock/endless"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iancoleman/strcase"
	"github.com/phpcyy/windmill/generator"
	"github.com/phpcyy/windmill/models"
	"go/format"
	"net/http"
	"os"
	"os/exec"
	"syscall"
)

func init() {
	models.InitDb()
}

func main() {
	mux := InitRouter()
	mux.HandleFunc("/_/build", build)
	endless.NewServer(":9111", mux).ListenAndServe()
}

func build(writer http.ResponseWriter, request *http.Request) {
	name := request.URL.Query().Get("scheme")
	if name == "" {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}
	apiFile, err := os.Open(fmt.Sprintf("models/entity/%s.yml", name))
	if err != nil {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	defer func() {
		_ = apiFile.Close()
	}()

	scheme, err := generator.Decode(apiFile)
	if err != nil {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	routeFile, err := os.OpenFile("router.go", os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	defer func() {
		_ = routeFile.Close()
	}()

	modelFile, err := os.OpenFile(fmt.Sprintf("models/%s.go", strcase.ToLowerCamel(scheme.Name)), os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	defer modelFile.Close()

	modelContent, err := scheme.GenModel()
	if err != nil {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	modelFile.Write(modelContent)

	tableContent, err := scheme.GenTable()
	if err != nil {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	_, err = models.Db.Exec(tableContent)
	if err != nil {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	fileText := []byte(scheme.GenApi())
	fileText, err = format.Source(fileText)
	if err != nil {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	_, err = routeFile.Write(fileText)

	if err != nil {
		_, _ = writer.Write([]byte(err.Error()))
		return
	}

	err = exec.Command("go", "build", ".").Run()
	if err != nil {
		_, _ = writer.Write([]byte(err.Error()))
		return
	}

	parent := syscall.Getpid()
	err = syscall.Kill(parent, syscall.SIGHUP)
	if err != nil {
		_, _ = writer.Write([]byte(err.Error()))
		return
	}
}
