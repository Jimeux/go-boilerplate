package app

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
)

type RequestValidator interface {
	Struct(s interface{}) error
}

type Controller struct {
	dao      DAO
	validate RequestValidator
}

func NewController(dao DAO, validate RequestValidator) *Controller {
	return &Controller{dao: dao, validate: validate}
}

func (c *Controller) Create(w http.ResponseWriter, r *http.Request) {
	var m Model
	if err := c.decodeJSON(r.Body, &m); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	if _, err := c.dao.Create(&m); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	writeJSON(&m, w)
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

func (c *Controller) decodeJSON(body io.ReadCloser, strAdr interface{}) error {
	if err := json.NewDecoder(body).Decode(strAdr); err != nil {
		return err
	}
	if err := c.validate.Struct(strAdr); err != nil {
		return err
	}
	return nil
}

// writeJSONはdataをJSONにエンコードしwに書き込む。
func writeJSON(strAdr interface{}, w http.ResponseWriter) {
	body, err := json.Marshal(strAdr)
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
