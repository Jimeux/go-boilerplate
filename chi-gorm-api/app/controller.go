package app

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
)

type Controller struct {
	dao DAO
}

func NewController(dao DAO) *Controller {
	return &Controller{dao: dao}
}

func (c *Controller) Create(w http.ResponseWriter, r *http.Request) {
	var m Model
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	model, err := c.dao.Create(&m)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	writeJSON(&model, w)
}

func (c *Controller) Destroy(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	exists, err := c.dao.Delete(uint(id))
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !exists {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (c *Controller) Edit(w http.ResponseWriter, r *http.Request) {
	var m Model
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	updated, err := c.dao.Update(&m)
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if updated == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	writeJSON(&updated, w)
}

func (c *Controller) Index(w http.ResponseWriter, r *http.Request) {
	pageParam := chi.URLParam(r, "page")
	page, err := strconv.Atoi(pageParam)
	if err != nil {
		page = 1
	}
	perPageParam := chi.URLParam(r, "perPage")
	perPage, err := strconv.Atoi(perPageParam)
	if err != nil {
		perPage = 10
	}

	models, err := c.dao.FindAll(page*perPage-perPage, perPage)
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	writeJSON(models, w)
}

func (c *Controller) Show(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	model, err := c.dao.FindByID(uint(id))
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if model == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	writeJSON(&model, w)
}

// writeJSONはdataをJSONにエンコードしwに書き込む。
func writeJSON(data interface{}, w http.ResponseWriter) {
	body, err := json.Marshal(data)
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(body); err != nil {
		log.Print(err)
	}
}
