package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/Jimeux/go-boilerplate/standard-api/app"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/chacha20poly1305"
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

	key := []byte("itWouldBeBadIfSomebodyFoundThis!")
	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		log.Fatalln("Failed to instantiate XChaCha20-Poly1305:", err)
	}

	encryptionManager := app.NewEncryptionManager(aead)
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
