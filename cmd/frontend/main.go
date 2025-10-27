package main

import (
	"log"
	"net/http"
)

func main() {
	fs := http.FileServer(http.Dir("./cmd/frontend"))
	http.Handle("/", fs)

	log.Println("Сервер запущен на http://localhost:5500")
	err := http.ListenAndServe(":5500", nil)
	if err != nil {
		log.Fatal(err)
	}
}
