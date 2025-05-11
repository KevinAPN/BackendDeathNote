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
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func createTestServer(t *testing.T) *server.Server {
	cfg := &config.Config{
		Database:                    "postgres",
		KillDuration:                2,
		KillDurationWithDescription: 4,
	}
	s := server.NewTestServer(cfg)

	// Limpiar tablas
	s.DB.Exec("DELETE FROM kills")
	s.DB.Exec("DELETE FROM people")
	return s
}

func TestCreateAndGetPerson(t *testing.T) {
	s := createTestServer(t)

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	writer.WriteField("name", "L")
	writer.WriteField("age", "25")

	file, err := os.Open("./testdata/light.jpg")
	if err != nil {
		t.Fatalf("falta ./testdata/light.jpg: %v", err)
	}
	defer file.Close()

	part, _ := writer.CreateFormFile("photo", filepath.Base("light.jpg"))
	io.Copy(part, file)
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/people", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	s.GetRouter().ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("esperado 201, obtuve %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAddCauseAndDetailsAndStatus(t *testing.T) {
	s := createTestServer(t)

	// Crear persona
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	writer.WriteField("name", "Misa")
	writer.WriteField("age", "22")
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
		t.Fatalf("creación falló: %s", rec.Body.String())
	}

	var created map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &created)
	id := int(created["person_id"].(float64))

	// Añadir causa
	causePayload := `{ "cause": "accidente" }`
	req = httptest.NewRequest(http.MethodPost, "/people/"+strconv.Itoa(id)+"/cause", strings.NewReader(causePayload))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	s.GetRouter().ServeHTTP(rec, req)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("add cause falló: %s", rec.Body.String())
	}

	// Añadir detalles
	detailsPayload := `{ "details": "murió en Tokio" }`
	req = httptest.NewRequest(http.MethodPost, "/people/"+strconv.Itoa(id)+"/details", strings.NewReader(detailsPayload))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	s.GetRouter().ServeHTTP(rec, req)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("add details falló: %s", rec.Body.String())
	}

	// Esperar muerte (killDuration=2)
	time.Sleep(3 * time.Second)

	// Consultar estado
	req = httptest.NewRequest(http.MethodGet, "/people/"+strconv.Itoa(id)+"/status", nil)
	rec = httptest.NewRecorder()
	s.GetRouter().ServeHTTP(rec, req)

	if !strings.Contains(rec.Body.String(), "Muerto") {
		t.Errorf("esperado estado Muerto, got: %s", rec.Body.String())
	}
}

func TestGetAllKills(t *testing.T) {
	s := createTestServer(t)

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
