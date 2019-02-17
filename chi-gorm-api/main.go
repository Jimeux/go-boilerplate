package main

import (
	"log"
	"net/http"

	"github.com/Jimeux/go-boilerplate/chi-gorm-api/app"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

const (
	dbURL = "dev:pass@tcp(localhost:33306)/chi_gorm_api"
)

func main() {
	db, err := gorm.Open("mysql", dbURL)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if r := recover(); r != nil {
			log.Println("recovered in main: ", r)
		}
		db.Close()
	}()

	dao := app.NewDAO(db)
	controller := app.NewController(dao)

	router := chi.NewRouter()

	router.Use(
		middleware.DefaultCompress,
		middleware.Logger,
		middleware.RealIP,
		middleware.Recoverer,
		middleware.RedirectSlashes,
		middleware.RequestID,
		render.SetContentType(render.ContentTypeJSON),
	)

	router.Route("/model", func(r chi.Router) {
		r.Get("/", controller.Index)
		r.Post("/", controller.Create)

		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", controller.Show)
			r.Put("/", controller.Edit)
			r.Delete("/", controller.Destroy)
		})
	})

	log.Fatal(http.ListenAndServe(":8080", router))
}
