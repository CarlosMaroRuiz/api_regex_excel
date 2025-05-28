package utils

import (
	"encoding/json"
	"net/http"
	"contactos-api/models"
)

// APIResponse representa la estructura estándar de respuesta de la API
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// ErrorDetail representa detalles de errores de validación
type ErrorDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   string `json:"value,omitempty"`
}

// SuccessResponse envía una respuesta exitosa
func SuccessResponse(w http.ResponseWriter, data interface{}) {
	response := APIResponse{
		Success: true,
		Data:    data,
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// CreatedResponse envía una respuesta de recurso creado
func CreatedResponse(w http.ResponseWriter, data interface{}) {
	response := APIResponse{
		Success: true,
		Data:    data,
		Message: "Recurso creado exitosamente",
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// BadRequestResponse envía una respuesta de solicitud inválida
func BadRequestResponse(w http.ResponseWriter, message string) {
	response := APIResponse{
		Success: false,
		Error:   message,
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(response)
}

// NotFoundResponse envía una respuesta de recurso no encontrado
func NotFoundResponse(w http.ResponseWriter, message string) {
	response := APIResponse{
		Success: false,
		Error:   message,
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(response)
}

// InternalServerErrorResponse envía una respuesta de error interno del servidor
func InternalServerErrorResponse(w http.ResponseWriter, message string) {
	response := APIResponse{
		Success: false,
		Error:   message,
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(response)
}

// ValidationErrorResponse envía una respuesta con errores de validación
func ValidationErrorResponse(w http.ResponseWriter, errors []models.ErrorResponse) {
	// Convertir errores del modelo a detalles de error
	var errorDetails []ErrorDetail
	for _, err := range errors {
		errorDetails = append(errorDetails, ErrorDetail{
			Field:   err.Campo,
			Message: err.Mensaje,
		})
	}
	
	response := APIResponse{
		Success: false,
		Error:   "Errores de validación",
		Data:    errorDetails,
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnprocessableEntity)
	json.NewEncoder(w).Encode(response)
}

// ParseJSON parsea el JSON de la request
func ParseJSON(r *http.Request, v interface{}) error {
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	return decoder.Decode(v)
}