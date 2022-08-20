package server

import (
	"dispatcher/registrant"
	"log"
	"net/http"
)

func InitServer(register registrant.RegisterDispatcher) {
	http.HandleFunc("/", register.MainFunc)
	log.Fatal(http.ListenAndServe(":"+register.Port, nil))
}
