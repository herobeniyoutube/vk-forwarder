package main

import (
	"log"
	"net/http"

	"github.com/herobeniyoutube/vk-forwarder/application"
	"github.com/herobeniyoutube/vk-forwarder/config"
	"github.com/herobeniyoutube/vk-forwarder/httpapi"
	"github.com/herobeniyoutube/vk-forwarder/infrastructure"
	postgresql "github.com/herobeniyoutube/vk-forwarder/storage/postresql"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	config := config.NewConfig()
	//storage
	db := postgresql.NewDb(config.DBConnectionString)

	//app
	downloader := application.NewVideoDownloader()
	tg := infrastructure.NewTgService(config)
	vk := infrastructure.NewVkService(config)

	//web
	handler := application.NewVkEventHandler(tg, vk, downloader, db)
	router := mux.NewRouter()
	httpapi.NewController(router, handler)

	port := "14888"
	addr := ":" + port
	log.Printf("listening on %s", addr)

	err := http.ListenAndServe(addr, router)
	if err != nil {
		panic("Couldn't start the server: " + err.Error())
	}
}
