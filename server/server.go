// server/server.go
package server

import (
	"backend-avanzada/config"
	"backend-avanzada/logger"
	"backend-avanzada/models"
	"backend-avanzada/repository"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Server struct {
	DB               *gorm.DB
	Config           *config.Config
	PeopleRepository *repository.PeopleRepository
	KillRepository   *repository.KillRepository
	logger           *logger.Logger
	taskQueue        *TaskQueue
}

func NewServer() *Server {
	s := &Server{
		logger:    logger.NewLogger(),
		taskQueue: NewTaskQueue(),
	}
	var config config.Config
	configFile, err := os.ReadFile("config/config.json")
	if err != nil {
		s.logger.Fatal(err)
	}
	if err := json.Unmarshal(configFile, &config); err != nil {
		s.logger.Fatal(err)
	}
	s.Config = &config
	return s
}

func NewTestServer(cfg *config.Config) *Server {
	s := &Server{
		Config:    cfg,
		logger:    logger.NewLogger(),
		taskQueue: NewTaskQueue(),
	}
	s.initDB()
	s.PeopleRepository = repository.NewPeopleRepository(s.DB)
	s.KillRepository = repository.NewKillRepository(s.DB)
	return s
}

func (s *Server) StartServer() {
	fmt.Println("Inicializando base de datos...")
	s.initDB()
	fmt.Println("Inicializando mux...")
	srv := &http.Server{
		Addr:    s.Config.Address,
		Handler: s.GetRouter(),
	}
	fmt.Println("Escuchando en el puerto ", s.Config.Address)
	if err := srv.ListenAndServe(); err != nil {
		s.logger.Fatal(err)
	}
}

func (s *Server) initDB() {
	switch s.Config.Database {
	case "sqlite":
		db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
		if err != nil {
			s.logger.Fatal(err)
		}
		s.DB = db
	case "postgres":
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
			os.Getenv("POSTGRES_HOST"),
			os.Getenv("POSTGRES_USER"),
			os.Getenv("POSTGRES_PASSWORD"),
			os.Getenv("POSTGRES_DB"),
		)
		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			s.logger.Fatal(err)
		}
		s.DB = db
	}
	fmt.Println("Aplicando migraciones...")
	s.DB.AutoMigrate(&models.Person{}, &models.Kill{})
	s.KillRepository = repository.NewKillRepository(s.DB)
	s.PeopleRepository = repository.NewPeopleRepository(s.DB)
}

// HandleGetConfig expone las duraciones configuradas al frontend
func (s *Server) HandleGetConfig(w http.ResponseWriter, r *http.Request) {
	cfg := map[string]int{
		"kill_duration":                  s.Config.KillDuration,
		"kill_duration_with_description": s.Config.KillDurationWithDescription,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cfg)
}

// CancelTaskForTest permite cancelar tareas desde pruebas
func (s *Server) CancelTaskForTest(id int) {
	s.taskQueue.CancelTask(id)
}
