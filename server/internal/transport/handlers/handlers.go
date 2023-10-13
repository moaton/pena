package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"server/internal/models"
	"server/internal/service"
	"server/pkg/util"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/xid"
)

type Handlers interface {
	NewRouter() *mux.Router
}

type handlers struct {
	service service.Service
}

func NewHandlers(service service.Service) Handlers {
	return &handlers{
		service: service,
	}
}

func (h *handlers) NewRouter() *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/task", h.Tasks).Methods("GET")
	router.HandleFunc("/report", h.Report).Methods("POST")

	router.HandleFunc("/reports", h.GetReports).Methods("GET")
	router.HandleFunc("/fails", h.GetFails).Methods("GET")

	return router
}

func (h *handlers) Tasks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	uuid := xid.NewWithTime(time.Now().UTC())
	cookie := http.Cookie{
		Name:  "uuid",
		Value: uuid.String(),
	}
	log.Println("uuid ", uuid)
	http.SetCookie(w, &cookie)

	batchsizeStr := r.FormValue("batchsize")
	batchsize, err := strconv.Atoi(batchsizeStr)
	if err != nil {
		util.ResponseError(w, http.StatusBadRequest, err.Error())
		return
	}
	if batchsize == 0 {
		util.ResponseError(w, http.StatusBadRequest, "batchsize equal 0")
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		util.ResponseError(w, http.StatusBadRequest, "streaming unsupported")
	}

	h.service.GetTasks(&models.Client{
		ID:        uuid.String(),
		Batchsize: batchsize,
		Writer:    w,
		Flusher:   flusher,
		ReqCtx:    ctx,
	})
}

func (h *handlers) Report(w http.ResponseWriter, r *http.Request) {
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
	id, err := r.Cookie("uuid")
	if err != nil {
		log.Println("cookie not found ", err)
	}
	h.service.Report(id.Value, request.Ids)
	util.ResponseOk(w, response{Message: "success"})
}

func (h *handlers) GetReports(w http.ResponseWriter, r *http.Request) {
	var reponse struct {
		Ids []string `json:"ids"`
	}
	reponse.Ids = h.service.GetReports()
	util.ResponseOk(w, reponse)
}

func (h *handlers) GetFails(w http.ResponseWriter, r *http.Request) {
	var reponse struct {
		Ids []string `json:"ids"`
	}
	reponse.Ids = h.service.GetFailTasks()
	util.ResponseOk(w, reponse)
}
