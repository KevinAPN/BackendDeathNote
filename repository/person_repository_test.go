package repository_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"backend-avanzada/models"
	"backend-avanzada/repository"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupRepo(t *testing.T) *repository.PeopleRepository {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=5432 sslmode=disable",
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to PostgreSQL: %v", err)
	}

	// Limpiar tabla antes del test
	_ = db.Migrator().DropTable(&models.Person{})
	if err := db.AutoMigrate(&models.Person{}); err != nil {
		t.Fatalf("migration error: %v", err)
	}

	return repository.NewPeopleRepository(db)
}

func TestAddCauseAndMarkDeath(t *testing.T) {
	repo := setupRepo(t)

	// 1) Creamos una persona
	orig := &models.Person{
		Name:      "Light Yagami",
		Age:       18,
		PhotoPath: "/static/light.jpg",
	}
	saved, err := repo.Save(orig)
	if err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// 2) Probamos AddCause
	cause := "Accidente"
	if err := repo.AddCause(saved.ID, cause); err != nil {
		t.Fatalf("AddCause() error: %v", err)
	}
	p1, _ := repo.FindById(int(saved.ID))
	if p1.Cause == nil || *p1.Cause != cause {
		t.Errorf("AddCause: esperé %q, obtuve %v", cause, p1.Cause)
	}

	// 3) Probamos MarkDeath
	if err := repo.MarkDeath(saved.ID); err != nil {
		t.Fatalf("MarkDeath() error: %v", err)
	}
	p2, _ := repo.FindById(int(saved.ID))
	if p2.DeathTime == nil || time.Since(*p2.DeathTime) > time.Second {
		t.Errorf("MarkDeath: DeathTime no establecido correctamente, got %v", p2.DeathTime)
	}
}
