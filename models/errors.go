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

// RowError representa un error específico en una fila del Excel
type RowError struct {
	Row     int               `json:"row"`              // Número de fila (empezando desde 1)
	Column  string            `json:"column"`           // Letra de columna (A, B, C, D)
	Field   string            `json:"field"`            // Nombre del campo
	Value   string            `json:"value"`            // Valor que causó el error
	Error   string            `json:"error"`            // Descripción del error
	// 🆕 NUEVO: Datos completos de la fila para contexto y corrección
	RowData *RowData          `json:"rowData,omitempty"` // Datos completos de la fila
}

// 🆕 NUEVO: RowData contiene todos los datos de una fila para facilitar la corrección
type RowData struct {
	ClaveCliente     string `json:"claveCliente"`     // Valor original como string
	Nombre           string `json:"nombre"`           // Nombre completo
	Correo           string `json:"correo"`           // Correo electrónico
	TelefonoContacto string `json:"telefonoContacto"` // Teléfono de contacto
	// Metadatos adicionales
	HasErrors        bool   `json:"hasErrors"`        // Si la fila tiene errores
	ErrorCount       int    `json:"errorCount"`       // Cantidad de errores en la fila
}

// ExcelValidationReport representa un reporte de validación del Excel
type ExcelValidationReport struct {
	TotalRows       int        `json:"totalRows"`        // Total de filas procesadas
	ValidRows       int        `json:"validRows"`        // Filas válidas cargadas
	InvalidRows     int        `json:"invalidRows"`      // Filas con errores
	Errors          []RowError `json:"errors"`           // Lista de errores encontrados
	LoadTimestamp   string     `json:"loadTimestamp"`    // Timestamp de cuando se cargó
	// 🆕 NUEVO: Filas inválidas completas para corrección
	InvalidRowsData []RowData  `json:"invalidRowsData"`  // Datos completos de filas inválidas
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