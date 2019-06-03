package main

import (
	"database/sql"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"

	"github.com/Jimeux/go-boilerplate/standard-api/app"
	"github.com/Jimeux/go-boilerplate/standard-api/app/encrypt"
)

const (
	dbURL = "dev:pass@tcp(localhost:33306)/standard_api"
)

func main() {
	db, err := sql.Open("mysql", dbURL)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if r := recover(); r != nil {
			log.Println("recovered in main: ", r)
		}
		db.Close()
	}()

	currentVersion := encrypt.Version(1)
	keyMap := encrypt.KeyMap{
		1: []byte("itWouldBeBadIfSomebodyFoundThis!"),
		2: []byte("itWouldBeBadIfSomeoneFoundThis!!"),
	}
	encryptionManager, err := encrypt.NewEncryptionManager(currentVersion, keyMap)
	if err != nil {
		log.Fatalln(err)
	}
	dao := app.NewDAO(db, encryptionManager)
	controller := app.NewController(dao)

	http.HandleFunc("/model/create", controller.Create)
	http.HandleFunc("/model/destroy", controller.Destroy)
	http.HandleFunc("/model/edit", controller.Edit)
	http.HandleFunc("/model/index", controller.Index)
	http.HandleFunc("/model/show", controller.Show)

	log.Println("Starting server at localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
