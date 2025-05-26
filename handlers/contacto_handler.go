// handlers/contacto_handler.go
package handlers

import (
	"net/http"
	"strconv"

	"contactos-api/models"
	"contactos-api/services"
	"contactos-api/utils"

	"github.com/gorilla/mux"
)

// ContactoHandler maneja las peticiones HTTP para contactos
type ContactoHandler struct {
	service services.ContactoServiceInterface
}

// NewContactoHandler crea una nueva instancia del handler
func NewContactoHandler(service services.ContactoServiceInterface) *ContactoHandler {
	return &ContactoHandler{
		service: service,
	}
}

// GetAllContactos maneja GET /api/contactos
func (h *ContactoHandler) GetAllContactos(w http.ResponseWriter, r *http.Request) {
	contactos, err := h.service.GetAllContactos()
	if err != nil {
		utils.InternalServerErrorResponse(w, "Error obteniendo contactos")
		return
	}

	utils.SuccessResponse(w, contactos)
}

// 🆕 MEJORADO: GetContactosConEstadoValidacion ahora incluye datos inválidos completos
func (h *ContactoHandler) GetContactosConEstadoValidacion(w http.ResponseWriter, r *http.Request) {
	// Obtener todos los contactos válidos del servidor
	contactos, err := h.service.GetAllContactos()
	if err != nil {
		utils.InternalServerErrorResponse(w, "Error obteniendo contactos")
		return
	}

	// Obtener reporte de validación completo
	report, err := h.service.GetExcelValidationReport()
	if err != nil {
		// Si no hay reporte, retornar solo los contactos sin información de errores
		response := map[string]interface{}{
			"contactos":        contactos,
			"validationReport": nil,
			"totalContactos":   len(contactos),
			"validContactos":   len(contactos),
			"errorContactos":   0,
			"invalidRowsData":  []models.RowData{}, // 🆕 NUEVO: Lista vacía
		}
		utils.SuccessResponse(w, response)
		return
	}

	// Crear mapa de errores por contacto para fácil acceso
	erroresPorContacto := make(map[int][]models.RowError)
	
	if report != nil && len(report.Errors) > 0 {
		for _, error := range report.Errors {
			// Si el error tiene valor de clave cliente, asociarlo
			if error.Field == "claveCliente" && error.Value != "" {
				if clave, err := strconv.Atoi(error.Value); err == nil {
					erroresPorContacto[clave] = append(erroresPorContacto[clave], error)
				}
			} else {
				// Para otros campos, intentar asociar por valor
				for _, contacto := range contactos {
					if h.errorBelongsToContact(error, contacto) {
						erroresPorContacto[contacto.ClaveCliente] = append(erroresPorContacto[contacto.ClaveCliente], error)
						break
					}
				}
			}
		}
	}

	// Contar contactos con errores
	contactosConErrores := 0
	for _, contacto := range contactos {
		if _, tieneErrores := erroresPorContacto[contacto.ClaveCliente]; tieneErrores {
			contactosConErrores++
		}
	}

	// Preparar respuesta completa con datos inválidos
	response := map[string]interface{}{
		"contactos":        contactos,
		"validationReport": report,
		"errorsByContact":  erroresPorContacto,
		"totalContactos":   len(contactos),
		"validContactos":   len(contactos) - contactosConErrores,
		"errorContactos":   contactosConErrores,
		"invalidRowsData":  report.InvalidRowsData, // 🆕 NUEVO: Datos completos para corrección
		"summary": map[string]interface{}{
			"hasValidationErrors": len(report.Errors) > 0,
			"totalErrors":         len(report.Errors),
			"invalidRows":         report.InvalidRows,
			"validRows":           report.ValidRows,
			"canCorrectErrors":    len(report.InvalidRowsData) > 0, // 🆕 NUEVO: Indicar si se pueden corregir
		},
	}

	utils.SuccessResponse(w, response)
}

// 🆕 NUEVO: GetInvalidContactsForCorrection maneja GET /api/contactos/invalid-data
// Este endpoint retorna SOLO los datos inválidos para corrección
func (h *ContactoHandler) GetInvalidContactsForCorrection(w http.ResponseWriter, r *http.Request) {
	invalidData, err := h.service.GetInvalidContactsForCorrection()
	if err != nil {
		utils.InternalServerErrorResponse(w, "Error obteniendo datos inválidos")
		return
	}

	// Preparar respuesta con información útil para corrección
	response := map[string]interface{}{
		"invalidRowsData": invalidData,
		"totalInvalid":    len(invalidData),
		"message":         "Datos inválidos disponibles para corrección",
		"instructions": map[string]string{
			"claveCliente":     "Debe ser un número entero mayor a 0",
			"nombre":           "No debe contener números ni estar vacío",
			"correo":           "Debe ser de un proveedor conocido (gmail, yahoo, hotmail, outlook, live, icloud, protonmail)",
			"telefonoContacto": "Debe tener exactamente 10 dígitos numéricos",
		},
	}

	utils.SuccessResponse(w, response)
}

// 🆕 NUEVO: GetDetailedValidationReport maneja GET /api/contactos/detailed-validation
// Este endpoint retorna un reporte detallado con datos inválidos y sugerencias de corrección
func (h *ContactoHandler) GetDetailedValidationReport(w http.ResponseWriter, r *http.Request) {
	report, err := h.service.GetExcelValidationReport()
	if err != nil {
		utils.InternalServerErrorResponse(w, "Error obteniendo reporte detallado")
		return
	}

	// Agrupar errores por fila para mejor presentación
	errorsByRow := h.groupErrorsByRow(report.Errors)
	
	// Crear sugerencias de corrección para cada fila inválida
	corrections := make([]map[string]interface{}, 0)
	for _, rowData := range report.InvalidRowsData {
		correction := map[string]interface{}{
			"originalData": rowData,
			"errors":       h.getErrorsForRowData(rowData, report.Errors),
			"suggestions":  h.generateCorrectionSuggestions(rowData),
		}
		corrections = append(corrections, correction)
	}

	response := map[string]interface{}{
		"summary": map[string]interface{}{
			"totalRows":    report.TotalRows,
			"validRows":    report.ValidRows,
			"invalidRows":  report.InvalidRows,
			"totalErrors":  len(report.Errors),
			"successRate":  h.calculateSuccessRate(report.ValidRows, report.InvalidRows),
		},
		"validationReport": report,
		"errorsByRow":      errorsByRow,
		"corrections":      corrections,
		"loadTimestamp":    report.LoadTimestamp,
	}

	utils.SuccessResponse(w, response)
}

// 🆕 NUEVO: Función auxiliar para obtener errores específicos de una fila
func (h *ContactoHandler) getErrorsForRowData(rowData models.RowData, allErrors []models.RowError) []models.RowError {
	var rowErrors []models.RowError
	for _, error := range allErrors {
		if error.RowData != nil {
			// Comparar los datos para ver si pertenecen a la misma fila
			if error.RowData.ClaveCliente == rowData.ClaveCliente &&
			   error.RowData.Nombre == rowData.Nombre &&
			   error.RowData.Correo == rowData.Correo &&
			   error.RowData.TelefonoContacto == rowData.TelefonoContacto {
				rowErrors = append(rowErrors, error)
			}
		}
	}
	return rowErrors
}

// 🆕 NUEVO: Generar sugerencias de corrección
func (h *ContactoHandler) generateCorrectionSuggestions(rowData models.RowData) map[string]string {
	suggestions := make(map[string]string)

	// Sugerencias para clave cliente
	if rowData.ClaveCliente == "" {
		suggestions["claveCliente"] = "Agregue un número entero mayor a 0"
	} else if _, err := strconv.Atoi(rowData.ClaveCliente); err != nil {
		suggestions["claveCliente"] = "Cambie a un número entero válido (ej: 123)"
	}

	// Sugerencias para nombre
	if rowData.Nombre == "" {
		suggestions["nombre"] = "Agregue un nombre válido sin números"
	}

	// Sugerencias para correo
	if rowData.Correo == "" {
		suggestions["correo"] = "Agregue un correo electrónico válido"
	} else if !h.containsAt(rowData.Correo) {
		suggestions["correo"] = "Agregue @ al correo (ej: usuario@gmail.com)"
	}

	// Sugerencias para teléfono
	if rowData.TelefonoContacto == "" {
		suggestions["telefonoContacto"] = "Agregue un teléfono de 10 dígitos"
	} else if len(rowData.TelefonoContacto) != 10 {
		suggestions["telefonoContacto"] = "El teléfono debe tener exactamente 10 dígitos"
	} else if !h.isNumeric(rowData.TelefonoContacto) {
		suggestions["telefonoContacto"] = "El teléfono debe contener solo números"
	}

	return suggestions
}

// 🆕 NUEVO: Función auxiliar para verificar si contiene @
func (h *ContactoHandler) containsAt(email string) bool {
	for _, char := range email {
		if char == '@' {
			return true
		}
	}
	return false
}

// 🆕 NUEVO: Función auxiliar para verificar si es numérico
func (h *ContactoHandler) isNumeric(s string) bool {
	for _, char := range s {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}

// 🆕 NUEVO: Calcular tasa de éxito
func (h *ContactoHandler) calculateSuccessRate(valid, invalid int) float64 {
	total := valid + invalid
	if total == 0 {
		return 0.0
	}
	return (float64(valid) / float64(total)) * 100
}

// Función auxiliar para determinar si un error pertenece a un contacto
func (h *ContactoHandler) errorBelongsToContact(error models.RowError, contacto models.Contacto) bool {
	switch error.Field {
	case "nombre":
		return error.Value == contacto.Nombre
	case "correo":
		return error.Value == contacto.Correo
	case "telefonoContacto":
		// Normalizar teléfonos para comparar
		errorPhone := h.normalizePhone(error.Value)
		contactPhone := h.normalizePhone(contacto.TelefonoContacto)
		return errorPhone == contactPhone
	}
	return false
}

// Función auxiliar para normalizar teléfonos
func (h *ContactoHandler) normalizePhone(phone string) string {
	result := ""
	for _, char := range phone {
		if char >= '0' && char <= '9' {
			result += string(char)
		}
	}
	return result
}

// GetContactoByID maneja GET /api/contactos/{clave}
func (h *ContactoHandler) GetContactoByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	claveStr := vars["clave"]

	clave, err := strconv.Atoi(claveStr)
	if err != nil {
		utils.BadRequestResponse(w, "Clave cliente inválida")
		return
	}

	contacto, err := h.service.GetContactoByID(clave)
	if err != nil {
		utils.NotFoundResponse(w, "Contacto no encontrado")
		return
	}

	utils.SuccessResponse(w, contacto)
}

// CreateContacto maneja POST /api/contactos
func (h *ContactoHandler) CreateContacto(w http.ResponseWriter, r *http.Request) {
	var request models.ContactoRequest

	if err := utils.ParseJSON(r, &request); err != nil {
		utils.BadRequestResponse(w, "JSON inválido")
		return
	}

	contacto, errores, err := h.service.CreateContacto(&request)
	if err != nil {
		utils.InternalServerErrorResponse(w, "Error creando contacto")
		return
	}

	if len(errores) > 0 {
		utils.ValidationErrorResponse(w, errores)
		return
	}

	utils.CreatedResponse(w, contacto)
}

// UpdateContacto maneja PUT /api/contactos/{clave}
func (h *ContactoHandler) UpdateContacto(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	claveStr := vars["clave"]

	clave, err := strconv.Atoi(claveStr)
	if err != nil {
		utils.BadRequestResponse(w, "Clave cliente inválida")
		return
	}

	var request models.ContactoRequest

	if err := utils.ParseJSON(r, &request); err != nil {
		utils.BadRequestResponse(w, "JSON inválido")
		return
	}

	contacto, errores, err := h.service.UpdateContacto(clave, &request)
	if err != nil {
		if err.Error() == "contacto no encontrado" || 
		   err.Error() == "contacto no encontrado: contacto con clave "+strconv.Itoa(clave)+" no encontrado" {
			utils.NotFoundResponse(w, "Contacto no encontrado")
			return
		}
		utils.InternalServerErrorResponse(w, "Error actualizando contacto")
		return
	}

	if len(errores) > 0 {
		utils.ValidationErrorResponse(w, errores)
		return
	}

	utils.SuccessResponse(w, contacto)
}

// DeleteContacto maneja DELETE /api/contactos/{clave}
func (h *ContactoHandler) DeleteContacto(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	claveStr := vars["clave"]

	clave, err := strconv.Atoi(claveStr)
	if err != nil {
		utils.BadRequestResponse(w, "Clave cliente inválida")
		return
	}

	if err := h.service.DeleteContacto(clave); err != nil {
		if err.Error() == "contacto no encontrado" {
			utils.NotFoundResponse(w, "Contacto no encontrado")
			return
		}
		utils.InternalServerErrorResponse(w, "Error eliminando contacto")
		return
	}

	utils.SuccessResponse(w, map[string]string{
		"message": "Contacto eliminado exitosamente",
	})
}

// SearchContactos maneja GET /api/contactos/buscar
func (h *ContactoHandler) SearchContactos(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	criteria := &models.ContactoDTO{
		ClaveCliente: query.Get("claveCliente"),
		Nombre:       query.Get("nombre"),
		Correo:       query.Get("correo"),
		Telefono:     query.Get("telefono"),
	}

	contactos, errores, err := h.service.SearchContactos(criteria)
	if err != nil {
		utils.InternalServerErrorResponse(w, "Error buscando contactos")
		return
	}

	if len(errores) > 0 {
		utils.ValidationErrorResponse(w, errores)
		return
	}

	utils.SuccessResponse(w, contactos)
}

// GetContactoStats maneja GET /api/contactos/stats
func (h *ContactoHandler) GetContactoStats(w http.ResponseWriter, r *http.Request) {
	// Verificar si el servicio tiene el método GetContactoStats
	if service, ok := h.service.(*services.ContactoService); ok {
		stats, err := service.GetContactoStats()
		if err != nil {
			utils.InternalServerErrorResponse(w, "Error obteniendo estadísticas")
			return
		}
		utils.SuccessResponse(w, stats)
	} else {
		utils.InternalServerErrorResponse(w, "Estadísticas no disponibles")
	}
}

// GetExcelValidationReport maneja GET /api/contactos/validation
func (h *ContactoHandler) GetExcelValidationReport(w http.ResponseWriter, r *http.Request) {
	report, err := h.service.GetExcelValidationReport()
	if err != nil {
		utils.InternalServerErrorResponse(w, "Error obteniendo reporte de validación")
		return
	}

	utils.SuccessResponse(w, report)
}

// ReloadExcel maneja POST /api/contactos/reload
func (h *ContactoHandler) ReloadExcel(w http.ResponseWriter, r *http.Request) {
	report, err := h.service.ReloadExcel()
	if err != nil {
		utils.InternalServerErrorResponse(w, "Error recargando archivo Excel")
		return
	}

	// Retornar respuesta con el reporte de validación mejorado
	response := map[string]interface{}{
		"message":          "Archivo Excel recargado exitosamente",
		"report":           report,
		"invalidRowsData":  report.InvalidRowsData, // 🆕 NUEVO: Datos para corrección
		"summary": map[string]interface{}{
			"totalProcessed":   report.TotalRows,
			"validContactos":   report.ValidRows,
			"invalidContactos": report.InvalidRows,
			"canCorrect":      len(report.InvalidRowsData) > 0,
		},
	}

	utils.SuccessResponse(w, response)
}

// GetValidationErrors maneja GET /api/contactos/errors
func (h *ContactoHandler) GetValidationErrors(w http.ResponseWriter, r *http.Request) {
	report, err := h.service.GetExcelValidationReport()
	if err != nil {
		utils.InternalServerErrorResponse(w, "Error obteniendo errores de validación")
		return
	}

	// Filtrar solo los errores con información mejorada
	response := map[string]interface{}{
		"totalErrors":      len(report.Errors),
		"errors":           report.Errors,
		"errorsByField":    h.groupErrorsByField(report.Errors),
		"errorsByRow":      h.groupErrorsByRow(report.Errors),
		"invalidRowsData":  report.InvalidRowsData, // 🆕 NUEVO: Datos completos
	}

	utils.SuccessResponse(w, response)
}

// groupErrorsByField agrupa los errores por campo
func (h *ContactoHandler) groupErrorsByField(errors []models.RowError) map[string][]models.RowError {
	grouped := make(map[string][]models.RowError)
	
	for _, error := range errors {
		grouped[error.Field] = append(grouped[error.Field], error)
	}
	
	return grouped
}

// groupErrorsByRow agrupa los errores por fila
func (h *ContactoHandler) groupErrorsByRow(errors []models.RowError) map[int][]models.RowError {
	grouped := make(map[int][]models.RowError)
	
	for _, error := range errors {
		grouped[error.Row] = append(grouped[error.Row], error)
	}
	
	return grouped
}

// HealthCheck maneja GET /api/health
func (h *ContactoHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	// Obtener información de validación para el health check
	report, err := h.service.GetExcelValidationReport()
	
	health := map[string]interface{}{
		"status":    "ok",
		"service":   "contactos-api",
		"version":   "1.0.0",
		"timestamp": r.Header.Get("Date"),
	}

	// Agregar información de validación si está disponible
	if err == nil {
		health["excel_status"] = map[string]interface{}{
			"total_rows":     report.TotalRows,
			"valid_rows":     report.ValidRows,
			"invalid_rows":   report.InvalidRows,
			"has_errors":     len(report.Errors) > 0,
			"can_correct":    len(report.InvalidRowsData) > 0, // 🆕 NUEVO
		}
	}

	utils.SuccessResponse(w, health)
}