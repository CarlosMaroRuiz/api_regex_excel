// models/errors.go
package models

import (
	"fmt"
	"strconv"
)

// ErrorResponse representa un error de validación
type ErrorResponse struct {
	Campo   string `json:"campo"`
	Mensaje string `json:"mensaje"`
}



// APIResponse representa una respuesta estándar de la API
type APIResponse struct {
	Success bool              `json:"success"`
	Data    interface{}       `json:"data,omitempty"`
	Error   string            `json:"error,omitempty"`
	Errors  []ErrorResponse   `json:"errors,omitempty"`
}

// 🆕 NUEVO: Helper methods para RowData

// IsValid verifica si la fila no tiene errores
func (rd *RowData) IsValid() bool {
	return !rd.HasErrors
}

// AddError marca la fila como que tiene errores
func (rd *RowData) AddError() {
	rd.HasErrors = true
	rd.ErrorCount++
}

// ToContactoRequest convierte RowData a ContactoRequest si es válida
func (rd *RowData) ToContactoRequest() (*ContactoRequest, error) {
	if rd.HasErrors {
		return nil, fmt.Errorf("no se puede convertir fila con errores a ContactoRequest")
	}
	
	// Convertir clave cliente a int
	clave, err := strconv.Atoi(rd.ClaveCliente)
	if err != nil {
		return nil, fmt.Errorf("error convirtiendo clave cliente: %w", err)
	}
	
	return &ContactoRequest{
		ClaveCliente:     clave,
		Nombre:           rd.Nombre,
		Correo:           rd.Correo,
		TelefonoContacto: rd.TelefonoContacto,
	}, nil
}