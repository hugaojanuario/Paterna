package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/hugaojanuario/Paterna/internal/api"
	"github.com/hugaojanuario/Paterna/internal/repository"
)

func main() {
	if err := repository.Init(); err != nil {
		fmt.Fprintln(os.Stderr, "repository init:", err)
		os.Exit(1)
	}

	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Println("Paterna API rodando em :" + port)

	if err := http.ListenAndServe(":"+port, api.Router()); err != nil {
		fmt.Fprintln(os.Stderr, "server:", err)
		os.Exit(1)
	}
}
