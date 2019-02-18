package app

import (
	"context"
	"github.com/go-chi/chi"
	"net/http/httptest"
	"testing"
)

type MockDAO struct {
	DAO
	findByID func(id uint) (model *Model, err error)
}

func (m MockDAO) FindByID(id uint) (model *Model, err error) {
	return m.findByID(id)
}

func TestCreate(t *testing.T) {
	mockDAO := MockDAO{
		findByID: func(id uint) (model *Model, err error) {
			return &Model{ID: 1}, nil
		},
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/v1/models/1", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	controller := NewController(mockDAO, nil)
	controller.Show(w, r)

	if w.Code != 200 {
		t.Fatalf("failed with code %d", w.Code)
	}
}
