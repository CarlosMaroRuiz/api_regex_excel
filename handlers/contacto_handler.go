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

// 🆕 NUEVO: GetContactosConEstadoValidacion maneja GET /api/contactos/con-validacion
// Este endpoint retorna TODOS los contactos con información de si tienen errores o no
func (h *ContactoHandler) GetContactosConEstadoValidacion(w http.ResponseWriter, r *http.Request) {
	// Obtener todos los contactos válidos del servidor
	contactos, err := h.service.GetAllContactos()
	if err != nil {
		utils.InternalServerErrorResponse(w, "Error obteniendo contactos")
		return
	}

	// Obtener reporte de validación
	report, err := h.service.GetExcelValidationReport()
	if err != nil {
		// Si no hay reporte, retornar solo los contactos sin información de errores
		response := map[string]interface{}{
			"contactos":        contactos,
			"validationReport": nil,
			"totalContactos":   len(contactos),
			"validContactos":   len(contactos),
			"errorContactos":   0,
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

	// Preparar respuesta completa
	response := map[string]interface{}{
		"contactos":        contactos,
		"validationReport": report,
		"errorsByContact":  erroresPorContacto,
		"totalContactos":   len(contactos),
		"validContactos":   len(contactos) - contactosConErrores,
		"errorContactos":   contactosConErrores,
		"summary": map[string]interface{}{
			"hasValidationErrors": len(report.Errors) > 0,
			"totalErrors":         len(report.Errors),
			"invalidRows":         report.InvalidRows,
			"validRows":           report.ValidRows,
		},
	}

	utils.SuccessResponse(w, response)
}

// 🆕 NUEVO: Función auxiliar para determinar si un error pertenece a un contacto
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

// 🆕 NUEVO: Función auxiliar para normalizar teléfonos
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

	// Retornar respuesta con el reporte de validación
	response := map[string]interface{}{
		"message": "Archivo Excel recargado exitosamente",
		"report":  report,
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

	// Filtrar solo los errores
	response := map[string]interface{}{
		"totalErrors": len(report.Errors),
		"errors":      report.Errors,
		"errorsByField": h.groupErrorsByField(report.Errors),
		"errorsByRow":   h.groupErrorsByRow(report.Errors),
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
			"total_rows":   report.TotalRows,
			"valid_rows":   report.ValidRows,
			"invalid_rows": report.InvalidRows,
			"has_errors":   len(report.Errors) > 0,
		}
	}

	utils.SuccessResponse(w, health)
}