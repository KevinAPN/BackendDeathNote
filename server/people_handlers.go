package server

import (
	"backend-avanzada/api"
	"backend-avanzada/models"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

func (s *Server) HandlePeople(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleGetAllPeople(w, r)
		return
	case http.MethodPost:
		s.handleCreatePerson(w, r)
		return
	}
}

func (s *Server) HandlePeopleWithId(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleGetPersonById(w, r)
		return
	case http.MethodPut:
		s.handleEditPerson(w, r)
		return
	case http.MethodDelete:
		s.handleDeletePerson(w, r)
		return
	}
}

func (s *Server) handleGetAllPeople(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	people, err := s.PeopleRepository.FindAll()
	if err != nil {
		s.HandleError(w, http.StatusInternalServerError, r.URL.Path, err)
		return
	}

	// Convertimos cada modelo a su DTO (incluye foto, estado, causa, detalles, muerte)
	var result []*api.PersonResponseDto
	for _, p := range people {
		result = append(result, p.ToPersonResponseDto())
	}

	response, err := json.Marshal(result)
	if err != nil {
		s.HandleError(w, http.StatusInternalServerError, r.URL.Path, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
	s.logger.Info(http.StatusOK, r.URL.Path, start)
}

func (s *Server) handleGetPersonById(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 32)
	if err != nil {
		s.HandleError(w, http.StatusBadRequest, r.URL.Path, err)
		return
	}
	p, err := s.PeopleRepository.FindById(int(id))
	if p == nil && err == nil {
		s.HandleError(w, http.StatusNotFound, r.URL.Path, fmt.Errorf("person with id %d not found", id))
		return
	}
	if err != nil {
		s.HandleError(w, http.StatusInternalServerError, r.URL.Path, err)
		return
	}
	resp := &api.PersonResponseDto{
		ID:            int(p.ID),
		Nombre:        p.Name,
		Edad:          p.Age,
		FechaCreacion: p.CreatedAt.String(),
	}
	response, err := json.Marshal(resp)
	if err != nil {
		s.HandleError(w, http.StatusInternalServerError, r.URL.Path, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
	s.logger.Info(http.StatusOK, r.URL.Path, start)
}

func (s *Server) handleCreatePerson(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// 1) Límite de tamaño (por ejemplo 10 MB)
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)

	// 2) Parsear multipart
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		s.HandleError(w, http.StatusBadRequest, r.URL.Path, fmt.Errorf("error parsing form: %w", err))
		return
	}
	name := r.FormValue("name")
	ageStr := r.FormValue("age")
	if name == "" || ageStr == "" {
		s.HandleError(w, http.StatusBadRequest, r.URL.Path, fmt.Errorf("name and age are required"))
		return
	}
	age, err := strconv.Atoi(ageStr)
	if err != nil || age <= 0 {
		s.HandleError(w, http.StatusBadRequest, r.URL.Path, fmt.Errorf("invalid age"))
		return
	}

	// 3) Obtener foto
	file, header, err := r.FormFile("photo")
	if err != nil {
		s.HandleError(w, http.StatusBadRequest, r.URL.Path, fmt.Errorf("photo is required"))
		return
	}
	defer file.Close()

	// 4) Guardar archivo en disco
	uploadsDir := "uploads/"
	os.MkdirAll(uploadsDir, os.ModePerm)
	filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), header.Filename)
	outPath := filepath.Join(uploadsDir, filename)
	outFile, err := os.Create(outPath)
	if err != nil {
		s.HandleError(w, http.StatusInternalServerError, r.URL.Path, err)
		return
	}
	defer outFile.Close()
	if _, err := io.Copy(outFile, file); err != nil {
		s.HandleError(w, http.StatusInternalServerError, r.URL.Path, err)
		return
	}

	// 5) Crear persona en BD
	person := &models.Person{
		Name:      name,
		Age:       age,
		PhotoPath: "/static/" + filename, // asumiendo servir uploads como /static/
	}
	person, err = s.PeopleRepository.Save(person)
	if err != nil {
		s.HandleError(w, http.StatusInternalServerError, r.URL.Path, err)
		return
	}

	// 6) Encolar muerte inicial (40s)
	kill := &models.Kill{PersonId: person.ID}
	duration := time.Duration(s.Config.KillDuration) * time.Second
	s.taskQueue.StartTask(int(person.ID), duration, func(k *models.Kill) error {
		return s.PeopleRepository.MarkHeartAttack(k.PersonId) // método a implementar
	}, kill)

	// 7) Responder
	resp := person.ToPersonResponseDto()
	data, _ := json.Marshal(resp)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(data)
	s.logger.Info(http.StatusCreated, r.URL.Path, start)
}

func (s *Server) handleEditPerson(w http.ResponseWriter, r *http.Request) {
	var p api.PersonRequestDto
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 32)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
	person, err := s.PeopleRepository.FindById(int(id))
	if person == nil && err == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	person.Name = p.Nombre
	person.Age = int(p.Edad)
	person, err = s.PeopleRepository.Save(person)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	pResponse := &api.PersonResponseDto{
		ID:            int(person.ID),
		Nombre:        person.Name,
		Edad:          person.Age,
		FechaCreacion: person.CreatedAt.String(),
	}
	result, err := json.Marshal(pResponse)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	w.Write(result)
}

func (s *Server) handleDeletePerson(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 32)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
	person, err := s.PeopleRepository.FindById(int(id))
	if person == nil && err == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = s.PeopleRepository.Delete(person)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) HandleAddCause(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		s.HandleError(w, http.StatusBadRequest, r.URL.Path, err)
		return
	}

	// Leer JSON { "cause": "texto corto" }
	var payload struct {
		Cause string `json:"cause"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		s.HandleError(w, http.StatusBadRequest, r.URL.Path, err)
		return
	}

	// Cancelar task de 40s inicial
	s.taskQueue.CancelTask(id)

	// Actualizar causa en BD
	if err := s.PeopleRepository.AddCause(uint(id), payload.Cause); err != nil {
		s.HandleError(w, http.StatusInternalServerError, r.URL.Path, err)
		return
	}

	// Encolar la muerte 6m40s después
	duration := time.Duration(s.Config.KillDurationWithDescription) * time.Second
	s.taskQueue.StartTask(id, duration, func(_ *models.Kill) error {
		return s.PeopleRepository.MarkDeath(uint(id))
	}, nil)

	w.WriteHeader(http.StatusAccepted)
	s.logger.Info(http.StatusAccepted, r.URL.Path, start)
}

func (s *Server) HandleAddDetails(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		s.HandleError(w, http.StatusBadRequest, r.URL.Path, err)
		return
	}

	var payload struct {
		Details string `json:"details"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		s.HandleError(w, http.StatusBadRequest, r.URL.Path, err)
		return
	}

	// Cancelar task de 6m40s
	s.taskQueue.CancelTask(id)

	// Actualizar detalles en BD
	if err := s.PeopleRepository.AddDetails(uint(id), payload.Details); err != nil {
		s.HandleError(w, http.StatusInternalServerError, r.URL.Path, err)
		return
	}

	// Encolar muerte final 40s después
	duration := time.Duration(s.Config.KillDuration) * time.Second
	s.taskQueue.StartTask(id, duration, func(_ *models.Kill) error {
		return s.PeopleRepository.MarkDeath(uint(id))
	}, nil)

	w.WriteHeader(http.StatusAccepted)
	s.logger.Info(http.StatusAccepted, r.URL.Path, start)
}

func (s *Server) HandleGetStatus(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		s.HandleError(w, http.StatusBadRequest, r.URL.Path, err)
		return
	}
	person, err := s.PeopleRepository.FindById(id)
	if person == nil && err == nil {
		s.HandleError(w, http.StatusNotFound, r.URL.Path,
			fmt.Errorf("person %d not found", id))
		return
	}
	if err != nil {
		s.HandleError(w, http.StatusInternalServerError, r.URL.Path, err)
		return
	}

	// Devolver solo el DTO con estado
	resp := person.ToPersonResponseDto()
	data, _ := json.Marshal(resp)
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
	s.logger.Info(http.StatusOK, r.URL.Path, start)
}
