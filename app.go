package main

import (
	"github.com/fvbock/endless"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/phpcyy/windmill/generator"
	"github.com/phpcyy/windmill/models"
	"net/http"
	"os/exec"
	"syscall"
)

func init() {
	models.InitDb()
}

func main() {
	router := mux.NewRouter()
	InitRouter(router)
	router.HandleFunc("/_/build", build)
	endless.NewServer(":9111", router).ListenAndServe()
}

func build(writer http.ResponseWriter, request *http.Request) {
	name := request.URL.Query().Get("scheme")
	var err error
	if name == "" {
		err = generator.NewBuilder("schemes/entity").BuildAll()
	} else {
		err = generator.NewBuilder("schemes/entity").Build(name + ".yml")
	}

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
