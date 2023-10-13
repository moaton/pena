package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"server/internal/service"
	"server/pkg/util"

	"github.com/gorilla/mux"
)

type Handlers interface {
	Task(w http.ResponseWriter, r *http.Request)
	Report(w http.ResponseWriter, r *http.Request)
	GetReports(w http.ResponseWriter, r *http.Request)
}

type handler struct {
	service service.Service
}

func NewRouter(service service.Service) *mux.Router {
	router := mux.NewRouter()
	handler := &handler{
		service: service,
	}
	router.HandleFunc("/task", handler.Task).Methods("GET")
	router.HandleFunc("/report", handler.Report).Methods("POST")
	router.HandleFunc("/report", handler.GetReports).Methods("GET")
	router.HandleFunc("/failss", handler.GetFails).Methods("GET")

	return router
}

func (h *handler) Task(w http.ResponseWriter, r *http.Request) {
	h.service.Serve(w, r)
}

func (h *handler) Report(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Ids []string `json:"ids"`
	}
	type response struct {
		Message string `json:"message"`
	}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&request)
	if err != nil {
		util.ResponseError(w, http.StatusBadRequest, err.Error())
		return
	}

	log.Println("req ", request)
	h.service.AddReport(request.Ids)
	util.ResponseOk(w, response{Message: "success"})
}

func (h *handler) GetReports(w http.ResponseWriter, r *http.Request) {
	var reponse struct {
		Ids []string `json:"ids"`
	}
	reponse.Ids = h.service.GetReports()
	util.ResponseOk(w, reponse)
}

func (h *handler) GetFails(w http.ResponseWriter, r *http.Request) {
	var reponse struct {
		Ids []string `json:"ids"`
	}
	reponse.Ids = h.service.GetFails()
	util.ResponseOk(w, reponse)
}
