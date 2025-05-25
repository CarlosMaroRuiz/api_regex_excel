// utils/response.go
package utils

import (
	"encoding/json"
	"net/http"

	"contactos-api/models"
)

// JSONResponse envía una respuesta JSON
func JSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// SuccessResponse envía una respuesta de éxito
func SuccessResponse(w http.ResponseWriter, data interface{}) {
	response := models.APIResponse{
		Success: true,
		Data:    data,
	}
	JSONResponse(w, http.StatusOK, response)
}

// CreatedResponse envía una respuesta de creación exitosa
func CreatedResponse(w http.ResponseWriter, data interface{}) {
	response := models.APIResponse{
		Success: true,
		Data:    data,
	}
	JSONResponse(w, http.StatusCreated, response)
}

// ErrorResponse envía una respuesta de error
func ErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	response := models.APIResponse{
		Success: false,
		Error:   message,
	}
	JSONResponse(w, statusCode, response)
}

// ValidationErrorResponse envía una respuesta de errores de validación
func ValidationErrorResponse(w http.ResponseWriter, errors []models.ErrorResponse) {
	response := models.APIResponse{
		Success: false,
		Error:   "Errores de validación",
		Errors:  errors,
	}
	JSONResponse(w, http.StatusBadRequest, response)
}

// BadRequestResponse envía una respuesta de solicitud incorrecta
func BadRequestResponse(w http.ResponseWriter, message string) {
	ErrorResponse(w, http.StatusBadRequest, message)
}

// NotFoundResponse envía una respuesta de no encontrado
func NotFoundResponse(w http.ResponseWriter, message string) {
	ErrorResponse(w, http.StatusNotFound, message)
}

// ConflictResponse envía una respuesta de conflicto
func ConflictResponse(w http.ResponseWriter, message string) {
	ErrorResponse(w, http.StatusConflict, message)
}

// InternalServerErrorResponse envía una respuesta de error interno del servidor
func InternalServerErrorResponse(w http.ResponseWriter, message string) {
	ErrorResponse(w, http.StatusInternalServerError, message)
}

// ParseJSON parsea el JSON del cuerpo de la solicitud
func ParseJSON(r *http.Request, v interface{}) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields() // No permitir campos desconocidos
	return decoder.Decode(v)
}