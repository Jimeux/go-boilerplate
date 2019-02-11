package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/Jimeux/go-boilerplate/standard-api/app"
	_ "github.com/go-sql-driver/mysql"
)

const (
	dbURL = "dev:pass@tcp(localhost:33306)/standard_api"
)

func main() {
	db, err := sql.Open("mysql", dbURL)
	if err != nil {
		log.Fatal(err)
	}

	dao := app.NewDAO(db)
	controller := app.NewController(dao)

	http.HandleFunc("/model/create", controller.Create)
	http.HandleFunc("/model/destroy", controller.Destroy)
	http.HandleFunc("/model/index", controller.Index)
	http.HandleFunc("/model/show", controller.Show)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
