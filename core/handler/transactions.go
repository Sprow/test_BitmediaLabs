package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"test_BitmediaLabs/core/transactions"
)



func (h *Handler) getTXsData(w http.ResponseWriter, r *http.Request) {
	var filter transactions.TXFilter
	err := json.NewDecoder(r.Body).Decode(&filter)
	if err != nil {
		h.jsonError(w, fmt.Errorf("invalid data"), http.StatusBadRequest)
		log.Println(err)
		return
	}

	err = filter.Validate()
	if err != nil {
		h.jsonError(w, err, http.StatusBadRequest)
		log.Println(err)
		return
	}

	date, err := h.storage.FindData(r.Context(), filter)
	if err != nil {
		h.jsonError(w, fmt.Errorf("internal server error"), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	if len(date) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	err = json.NewEncoder(w).Encode(date)
	if err != nil {
		h.jsonError(w, fmt.Errorf("internal server error"), http.StatusInternalServerError)
		log.Println(err)
		return
	}
}

