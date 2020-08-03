package main

import (
	"net/http"
)

func InitRouter() *http.ServeMux {
	mux := http.ServeMux{}

	return &mux
}
