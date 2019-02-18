package main

import (
	"gopkg.in/go-playground/validator.v9"
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
	dbURL = "dev:pass@tcp(localhost:33306)/chi_gorm_api?parseTime=true"
)

func main() {
	db, err := gorm.Open("mysql", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	db.LogMode(true)

	defer func() {
		if r := recover(); r != nil {
			log.Println("recovered in main: ", r)
		}
		db.Close()
	}()

	dao := app.NewDAO(db)

	validate := validator.New()
	controller := app.NewController(dao, validate)

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

	router.Route("/v1", func(v1 chi.Router) {
		v1.Route("/models", func(models chi.Router) {
			models.Post("/", controller.Create)
			models.Get("/", controller.Index)

			models.Route("/{id}", func(r chi.Router) {
				r.Get("/", controller.Show)
				r.Put("/", controller.Edit)
				r.Delete("/", controller.Destroy)
			})
		})
	})

	log.Fatal(http.ListenAndServe(":8080", router))
}
