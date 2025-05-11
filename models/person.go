package models

import (
	"backend-avanzada/api"
	"time"

	"gorm.io/gorm"
)

type Person struct {
	gorm.Model
	Name      string
	Age       int
	PhotoPath string
	Cause     *string
	Details   *string
	DeathTime *time.Time
}

// computeStatus devuelve el estado de la persona
func computeStatus(p *Person) string {
	if p.DeathTime == nil {
		// Si no tiene timestamp de muerte, sigue “Pendiente”
		return "Pendiente"
	}
	// Ya hay DeathTime, se considera “Muerto”
	return "Muerto"
}

func (p *Person) ToPersonResponseDto() *api.PersonResponseDto {
	// Convertir *time.Time a *string
	var deathTimeStr *string
	if p.DeathTime != nil {
		ts := p.DeathTime.Format(time.RFC3339)
		deathTimeStr = &ts
	}

	return &api.PersonResponseDto{
		ID:            int(p.ID),
		Nombre:        p.Name,
		Edad:          p.Age,
		FechaCreacion: p.CreatedAt.Format(time.RFC3339),
		FotoURL:       p.PhotoPath,
		Estado:        computeStatus(p),
		Cause:         p.Cause,
		Details:       p.Details,
		DeathTime:     deathTimeStr,
	}
}
