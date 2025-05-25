// models/errors.go
package models

// ErrorResponse representa un error de validación
type ErrorResponse struct {
	Campo   string `json:"campo"`
	Mensaje string `json:"mensaje"`
}

// RowError representa un error específico en una fila del Excel
type RowError struct {
	Row    int    `json:"row"`           // Número de fila (empezando desde 1)
	Column string `json:"column"`        // Letra de columna (A, B, C, D)
	Field  string `json:"field"`         // Nombre del campo
	Value  string `json:"value"`         // Valor que causó el error
	Error  string `json:"error"`         // Descripción del error
}

// ExcelValidationReport representa un reporte de validación del Excel
type ExcelValidationReport struct {
	TotalRows      int        `json:"totalRows"`       // Total de filas procesadas
	ValidRows      int        `json:"validRows"`       // Filas válidas cargadas
	InvalidRows    int        `json:"invalidRows"`     // Filas con errores
	Errors         []RowError `json:"errors"`          // Lista de errores encontrados
	LoadTimestamp  string     `json:"loadTimestamp"`   // Timestamp de cuando se cargó
}

// APIResponse representa una respuesta estándar de la API
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Errors  []ErrorResponse `json:"errors,omitempty"`
}