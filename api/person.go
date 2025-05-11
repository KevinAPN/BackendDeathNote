package api

type PersonRequestDto struct {
	Nombre string `form:"name"` // vendrá en multipart form
	Edad   int32  `form:"age"`
	// esto tendre que mandarlo en multipart debido a que el json puro no acepta imagenes asi que toco multipart form data
}

type PersonResponseDto struct {
	ID            int     `json:"person_id"`
	Nombre        string  `json:"name"`
	Edad          int     `json:"age"`
	FotoURL       string  `json:"photo_url"`
	FechaCreacion string  `json:"created_at"`
	Estado        string  `json:"status"`
	Cause         *string `json:"cause,omitempty"` // este campó segun la logica sera opcional
	Details       *string `json:"details,omitempty"`
	DeathTime     *string `json:"death_time,omitempty"`
}

type ErrorResponse struct {
	Status      int    `json:"status"`
	Description string `json:"description"`
	Message     string `json:"message"`
}
