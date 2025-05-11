package server

import (
	"net/http"

	"github.com/gorilla/mux"
)

// GetRouter expone el router con CORS, logging y todas las rutas
func (s *Server) GetRouter() http.Handler {
	router := mux.NewRouter()

	// Middleware CORS
	router.Use(middlewareCORS)
	// Middleware logging
	router.Use(s.logger.RequestLogger)

	// Servir archivos estáticos desde uploads/ en /static/
	router.PathPrefix("/static/").
		Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("uploads/"))))

	// Rutas de personas
	router.HandleFunc("/people", s.HandlePeople).
		Methods(http.MethodGet, http.MethodPost, http.MethodOptions)
	router.HandleFunc("/people/{id}", s.HandlePeopleWithId).
		Methods(http.MethodGet, http.MethodPut, http.MethodDelete, http.MethodOptions)
	router.HandleFunc("/people/{id}/cause", s.HandleAddCause).
		Methods(http.MethodPost, http.MethodOptions)
	router.HandleFunc("/people/{id}/details", s.HandleAddDetails).
		Methods(http.MethodPost, http.MethodOptions)
	router.HandleFunc("/people/{id}/status", s.HandleGetStatus).
		Methods(http.MethodGet, http.MethodOptions)

	// Rutas de kills
	router.HandleFunc("/kills", s.HandleKills).
		Methods(http.MethodGet, http.MethodOptions)
	router.HandleFunc("/kills/{id}", s.HandleKillsWithId).
		Methods(http.MethodPost, http.MethodDelete, http.MethodOptions)

	// Ruta de configuración
	router.HandleFunc("/config", s.HandleGetConfig).
		Methods(http.MethodGet, http.MethodOptions)

	return router
}
