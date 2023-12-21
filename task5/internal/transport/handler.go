package transport

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
	"github.com/tunx321/task5/internal/db"
)

type TxService interface {
	GetAllTransactions(string) ([]db.FrontData, error)
}

type Handler struct {
	Router  *mux.Router
	Service TxService
	Server  *http.Server
}

func NewHandler(service TxService) *Handler {
	h := &Handler{
		Service: service,
	}
	h.Router = mux.NewRouter()
	h.mapRoutes()
	h.Server = &http.Server{
		Addr:    ":8080",
		Handler: h.Router,
	}
	return h
}

func (h *Handler) mapRoutes() {
	h.Router.HandleFunc("/alive", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "I am alive")
	})
	h.Router.HandleFunc("/wallet", h.GetTransactions).Methods("POST")
	h.Router.HandleFunc("/", h.SearchForAddress).Methods("GET")

}

func (h *Handler) GetTransactions(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	address := r.FormValue("address")
	if address == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	data, err := h.Service.GetAllTransactions(address)
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	t, err := template.ParseFiles("./templates/index.html")
	if err != nil{
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err = t.Execute(w, data); err != nil{
		log.Println(err)
		return
	}
}


func (h* Handler) SearchForAddress(w http.ResponseWriter, r *http.Request){
	t, err := template.ParseFiles("./templates/find.html")
	if err != nil{
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err = t.Execute(w, nil); err != nil{
		log.Println(err)
		return
	}
}


func (h *Handler) Serve() error {
	go func() {
		if err := h.Server.ListenAndServe(); err != nil {
			log.Println(err.Error())
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	h.Server.Shutdown(ctx)
	log.Println("shut down gracefully")

	return nil
}
