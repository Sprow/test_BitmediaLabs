package handler

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/pkg/errors"
	"net/http"
	"strings"
	"test_BitmediaLabs/core/transactions"
)

type Handler struct {
	storage  *transactions.MongoStorage
}

func newHandler(storage  *transactions.MongoStorage) *Handler {
	return &Handler{storage: storage}
}

func (h *Handler) Register(r *chi.Mux) {
	r.Post("/api/v1/transactions/", h.getTXsData)
}

func Init(storage  *transactions.MongoStorage) http.Handler {
	router := chi.NewRouter()
	router.Use(middleware.Logger) // uncomment for debug api requests

	h := newHandler(storage)
	h.Register(router)
	return router
}

func (h *Handler) jsonError(w http.ResponseWriter, err error, code int) () {
	newErr := strings.Replace(err.Error(), "\"", "'", len(err.Error()))
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	w.Write([]byte(fmt.Sprintf(`{"error": "%s"}`, errors.New(newErr))))
}