package server

import (
	"testing"
	"time"

	"backend-avanzada/models"
)

func TestTaskQueueExecute(t *testing.T) {
	tq := NewTaskQueue()
	executed := false

	// Encolamos una tarea muy corta (10ms)
	tq.StartTask(1, 10*time.Millisecond, func(k *models.Kill) error {
		executed = true
		return nil
	}, &models.Kill{PersonId: 1})

	// Esperamos un poco más que 10ms
	time.Sleep(25 * time.Millisecond)
	if !executed {
		t.Error("La tarea no se ejecutó tras el delay")
	}
}

func TestTaskQueueCancel(t *testing.T) {
	tq := NewTaskQueue()
	executed := false

	// Encolamos una tarea larga (100ms)
	tq.StartTask(2, 100*time.Millisecond, func(k *models.Kill) error {
		executed = true
		return nil
	}, &models.Kill{PersonId: 2})

	// Cancelamos inmediatamente
	if !tq.CancelTask(2) {
		t.Error("CancelTask devolvió false, esperaba true")
	}

	// Dejamos pasar tiempo suficiente
	time.Sleep(150 * time.Millisecond)
	if executed {
		t.Error("La tarea se ejecutó pese a haber sido cancelada")
	}
}
