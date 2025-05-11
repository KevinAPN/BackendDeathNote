package server

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (s *Server) router() http.Handler {
	router := mux.NewRouter()

	// Middleware de logging
	router.Use(s.logger.RequestLogger)

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Death Note API is running"))
	}).Methods(http.MethodGet)

	// Servir archivos estáticos desde uploads/ en /static/
	router.
		PathPrefix("/static/").
		Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("uploads/"))))

	// Endpoints para personas
	router.HandleFunc("/people", s.HandlePeople).Methods(http.MethodGet, http.MethodPost)
	router.HandleFunc("/people/{id}", s.HandlePeopleWithId).
		Methods(http.MethodGet, http.MethodPut, http.MethodDelete)

	// Registrar causa y detalles
	router.HandleFunc("/people/{id}/cause", s.HandleAddCause).
		Methods(http.MethodPost)
	router.HandleFunc("/people/{id}/details", s.HandleAddDetails).
		Methods(http.MethodPost)

	// Endpoint para obtener la duración de una persona viva
	router.HandleFunc("/config", s.HandleGetConfig).Methods(http.MethodGet)

	// Endpoint para obtener estado de una persona
	router.HandleFunc("/people/{id}/status", s.HandleGetStatus).
		Methods(http.MethodGet)

	// Endpoints para kills (muerte)
	router.HandleFunc("/kills", s.HandleKills).Methods(http.MethodGet)
	router.HandleFunc("/kills/{id}", s.HandleKillsWithId).
		Methods(http.MethodPost, http.MethodDelete)

	return router
}
