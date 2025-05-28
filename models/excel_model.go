package models

import "time"

// RowError representa un error específico en una fila del Excel
type RowError struct {
	Row     int      `json:"row"`
	Column  string   `json:"column"`
	Field   string   `json:"field"`
	Value   string   `json:"value"`
	Error   string   `json:"error"`
	RowData *RowData `json:"rowData,omitempty"`
}

// RowData representa los datos completos de una fila (válida o inválida)
type RowData struct {
	ClaveCliente     string `json:"claveCliente"`
	Nombre           string `json:"nombre"`
	Correo           string `json:"correo"`
	TelefonoContacto string `json:"telefonoContacto"`
	HasErrors        bool   `json:"hasErrors"`
	ErrorCount       int    `json:"errorCount"`
	Errors           []string `json:"errors,omitempty"` // Lista de mensajes de error para el frontend
}



// AddErrorMessage agrega un mensaje de error específico
func (rd *RowData) AddErrorMessage(message string) {
	rd.AddError()
	if rd.Errors == nil {
		rd.Errors = make([]string, 0)
	}
	rd.Errors = append(rd.Errors, message)
}

// ExcelValidationReport representa el reporte completo de validación del Excel
type ExcelValidationReport struct {
	TotalRows       int         `json:"totalRows"`
	ValidRows       int         `json:"validRows"`
	InvalidRows     int         `json:"invalidRows"`
	Errors          []RowError  `json:"errors"`
	InvalidRowsData []RowData   `json:"invalidRowsData"`
	LoadTimestamp   string      `json:"loadTimestamp"`
	Summary         *ReportSummary `json:"summary,omitempty"`
}

// ReportSummary proporciona un resumen de los tipos de errores más comunes
type ReportSummary struct {
	ErrorsByField map[string]int `json:"errorsByField"`
	ErrorsByType  map[string]int `json:"errorsByType"`
	MostCommonErrors []CommonError `json:"mostCommonErrors"`
}

// CommonError representa un error común con su frecuencia
type CommonError struct {
	Message string `json:"message"`
	Count   int    `json:"count"`
	Field   string `json:"field"`
}

// ValidationError representa errores de validación de entrada
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   string `json:"value,omitempty"`
}

// ContactoStats representa estadísticas de contactos
type ContactoStats struct {
	Total            int                    `json:"total"`
	TotalErrores     int                    `json:"totalErrores"`
	TotalInvalidos   int                    `json:"totalInvalidos"`
	PorcentajeValidos float64               `json:"porcentajeValidos"`
	TopDominios      []DominioStats         `json:"topDominios"`
	EstadisticasPorCampo map[string]FieldStats `json:"estadisticasPorCampo"`
	Timestamp        string                 `json:"timestamp"`
}

// DominioStats representa estadísticas de dominios de correo
type DominioStats struct {
	Dominio string `json:"dominio"`
	Count   int    `json:"count"`
	Porcentaje float64 `json:"porcentaje"`
}

// FieldStats representa estadísticas de un campo específico
type FieldStats struct {
	ValoresUnicos int     `json:"valoresUnicos"`
	ValoresVacios int     `json:"valoresVacios"`
	Completitud   float64 `json:"completitud"` // Porcentaje de campos no vacíos
}

// LoadStatus representa el estado de carga del archivo Excel
type LoadStatus struct {
	IsLoaded      bool      `json:"isLoaded"`
	LastLoadTime  time.Time `json:"lastLoadTime"`
	TotalRecords  int       `json:"totalRecords"`
	ValidRecords  int       `json:"validRecords"`
	InvalidRecords int      `json:"invalidRecords"`
	ErrorCount    int       `json:"errorCount"`
}