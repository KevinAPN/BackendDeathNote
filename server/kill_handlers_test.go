package server_test

import (
	"backend-avanzada/config"
	"backend-avanzada/server"
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
)

func setupKillTestServer(t *testing.T) (*server.Server, int) {
	cfg := &config.Config{
		Database:                    "postgres",
		KillDuration:                2,
		KillDurationWithDescription: 4,
	}
	s := server.NewTestServer(cfg)
	s.DB.Exec("DELETE FROM kills")
	s.DB.Exec("DELETE FROM people")

	// Crear persona con foto
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	writer.WriteField("name", "Near")
	writer.WriteField("age", "20")
	file, _ := os.Open("./testdata/light.jpg")
	defer file.Close()
	part, _ := writer.CreateFormFile("photo", "light.jpg")
	io.Copy(part, file)
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/people", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	s.GetRouter().ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("falló creación de persona: %s", rec.Body.String())
	}

	var created map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &created)
	id := int(created["person_id"].(float64))

	s.CancelTaskForTest(id) // Evita conflicto por tarea de muerte automática

	return s, id
}

func TestCreateKillWithoutDescription(t *testing.T) {
	s, id := setupKillTestServer(t)

	payload := `{ "description": "" }`
	req := httptest.NewRequest(http.MethodPost, "/kills/"+strconv.Itoa(id), strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	s.GetRouter().ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("esperado 201, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "In progress") {
		t.Errorf("esperaba respuesta con 'In progress': %s", rec.Body.String())
	}
}

func TestCreateKillWithDescription(t *testing.T) {
	s, id := setupKillTestServer(t)

	payload := `{ "description": "asesinado por Kira" }`
	req := httptest.NewRequest(http.MethodPost, "/kills/"+strconv.Itoa(id), strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	s.GetRouter().ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("esperado 201, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "In progress") {
		t.Errorf("esperaba respuesta con 'In progress': %s", rec.Body.String())
	}
}

func TestGetAllKillsEmptyInitially(t *testing.T) {
	s, _ := setupKillTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/kills", nil)
	rec := httptest.NewRecorder()
	s.GetRouter().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("esperado 200, got %d", rec.Code)
	}
	if !strings.HasPrefix(rec.Body.String(), "[") {
		t.Errorf("respuesta no válida: %s", rec.Body.String())
	}
}
