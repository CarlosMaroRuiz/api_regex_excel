package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"regexp"
	"fmt"

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

// ✅ FUNCIÓN HELPER: Limpiar y extraer clave numérica de texto sucio
func (h *ContactoHandler) extractNumericKey(claveInput string) (int, error) {
	if claveInput == "" {
		return 0, fmt.Errorf("clave vacía")
	}
	
	// Limpiar la clave de espacios y caracteres comunes
	claveClean := strings.TrimSpace(claveInput)
	
	// Si ya es un número válido, devolverlo
	if numero, err := strconv.Atoi(claveClean); err == nil && numero > 0 {
		return numero, nil
	}
	
	// Intentar extraer números de texto con caracteres mezclados
	// Patrones comunes en Excel:
	// - "ABC12345" -> 12345
	// - "12345XYZ" -> 12345  
	// - "CLI-12345" -> 12345
	// - "12345.0" -> 12345
	
	// Usar regex para extraer secuencias de dígitos
	re := regexp.MustCompile(`\d+`)
	matches := re.FindAllString(claveClean, -1)
	
	if len(matches) > 0 {
		// Tomar la secuencia de dígitos más larga
		numeroMasLargo := ""
		for _, match := range matches {
			if len(match) > len(numeroMasLargo) {
				numeroMasLargo = match
			}
		}
		
		if numero, err := strconv.Atoi(numeroMasLargo); err == nil && numero > 0 && numero < 999999999 {
			return numero, nil
		}
	}
	
	return 0, fmt.Errorf("no se pudo extraer clave numérica válida de: %s", claveInput)
}

// ⚡ NUEVOS HANDLERS PARA PAGINACIÓN

// GetContactosPaginated maneja GET /api/contactos/paginated
func (h *ContactoHandler) GetContactosPaginated(w http.ResponseWriter, r *http.Request) {
	// Obtener parámetros de query
	query := r.URL.Query()
	
	// Parsear page (default: 0)
	page := 0
	if pageStr := query.Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p >= 0 {
			page = p
		}
	}
	
	// Parsear size (default: 50, max: 100)
	size := 50
	if sizeStr := query.Get("size"); sizeStr != "" {
		if s, err := strconv.Atoi(sizeStr); err == nil && s > 0 {
			size = s
			if size > 100 {
				size = 100 // Límite máximo
			}
		}
	}
	
	// Obtener término de búsqueda opcional
	search := query.Get("search")
	
	// Llamar al servicio
	result, err := h.service.GetContactosPaginated(page, size, search)
	if err != nil {
		utils.InternalServerErrorResponse(w, "Error obteniendo contactos paginados: "+err.Error())
		return
	}
	
	utils.SuccessResponse(w, result)
}

// SearchContactosPaginated maneja GET /api/contactos/search
func (h *ContactoHandler) SearchContactosPaginated(w http.ResponseWriter, r *http.Request) {
	// Obtener parámetros de query
	query := r.URL.Query()
	
	// Término de búsqueda (requerido)
	searchTerm := query.Get("q")
	if searchTerm == "" {
		utils.BadRequestResponse(w, "Parámetro 'q' (término de búsqueda) es requerido")
		return
	}
	
	// Parsear page (default: 0)
	page := 0
	if pageStr := query.Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p >= 0 {
			page = p
		}
	}
	
	// Parsear size (default: 50, max: 100)
	size := 50
	if sizeStr := query.Get("size"); sizeStr != "" {
		if s, err := strconv.Atoi(sizeStr); err == nil && s > 0 {
			size = s
			if size > 100 {
				size = 100
			}
		}
	}
	
	// Llamar al servicio
	result, err := h.service.SearchContactosPaginated(searchTerm, page, size)
	if err != nil {
		utils.InternalServerErrorResponse(w, "Error buscando contactos: "+err.Error())
		return
	}
	
	utils.SuccessResponse(w, result)
}

// GetContactosCount maneja GET /api/contactos/count
func (h *ContactoHandler) GetContactosCount(w http.ResponseWriter, r *http.Request) {
	count, err := h.service.GetContactosCount()
	if err != nil {
		utils.InternalServerErrorResponse(w, "Error obteniendo conteo: "+err.Error())
		return
	}
	
	utils.SuccessResponse(w, count)
}

// 📊 HANDLERS BÁSICOS MODIFICADOS PARA CLAVES FLEXIBLES

// GetAllContactos maneja GET /api/contactos
func (h *ContactoHandler) GetAllContactos(w http.ResponseWriter, r *http.Request) {
	contactos, err := h.service.GetAllContactos()
	if err != nil {
		utils.InternalServerErrorResponse(w, "Error obteniendo contactos")
		return
	}
	utils.SuccessResponse(w, contactos)
}

// ✅ GetContactoByID maneja GET /api/contactos/{clave} - MODIFICADO para claves flexibles
func (h *ContactoHandler) GetContactoByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	claveStr := vars["clave"]

	// Intentar extraer clave numérica del input (que puede tener caracteres)
	clave, err := h.extractNumericKey(claveStr)
	if err != nil {
		utils.BadRequestResponse(w, fmt.Sprintf("No se pudo extraer clave numérica válida de '%s': %v", claveStr, err))
		return
	}

	contacto, err := h.service.GetContactoByID(clave)
	if err != nil {
		utils.NotFoundResponse(w, fmt.Sprintf("Contacto con clave %d (extraída de '%s') no encontrado", clave, claveStr))
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

// ✅ UpdateContacto maneja PUT /api/contactos/{clave} - MODIFICADO para claves flexibles
func (h *ContactoHandler) UpdateContacto(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	claveStr := vars["clave"]

	// Intentar extraer clave numérica del input
	clave, err := h.extractNumericKey(claveStr)
	if err != nil {
		utils.BadRequestResponse(w, fmt.Sprintf("No se pudo extraer clave numérica válida de '%s': %v", claveStr, err))
		return
	}

	var request models.ContactoRequest

	if err := utils.ParseJSON(r, &request); err != nil {
		utils.BadRequestResponse(w, "JSON inválido")
		return
	}

	contacto, errores, err := h.service.UpdateContacto(clave, &request)
	if err != nil {
		utils.NotFoundResponse(w, fmt.Sprintf("Contacto con clave %d (extraída de '%s') no encontrado para actualizar", clave, claveStr))
		return
	}

	if len(errores) > 0 {
		utils.ValidationErrorResponse(w, errores)
		return
	}

	utils.SuccessResponse(w, contacto)
}

// ✅ DeleteContacto maneja DELETE /api/contactos/{clave} - MODIFICADO para claves flexibles
func (h *ContactoHandler) DeleteContacto(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	claveStr := vars["clave"]

	// Intentar extraer clave numérica del input
	clave, err := h.extractNumericKey(claveStr)
	if err != nil {
		utils.BadRequestResponse(w, fmt.Sprintf("No se pudo extraer clave numérica válida de '%s': %v", claveStr, err))
		return
	}

	if err := h.service.DeleteContacto(clave); err != nil {
		utils.NotFoundResponse(w, fmt.Sprintf("Contacto con clave %d (extraída de '%s') no encontrado para eliminar", clave, claveStr))
		return
	}

	utils.SuccessResponse(w, map[string]interface{}{
		"message": fmt.Sprintf("Contacto con clave %d eliminado exitosamente", clave),
		"claveOriginal": claveStr,
		"claveExtraida": clave,
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

// ✅ GetContactoStats maneja GET /api/contactos/stats (CORREGIDO)
func (h *ContactoHandler) GetContactoStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.GetContactoStats()
	if err != nil {
		utils.InternalServerErrorResponse(w, "Error obteniendo estadísticas: "+err.Error())
		return
	}
	
	utils.SuccessResponse(w, stats)
}

// 🔧 HANDLERS DE VALIDACIÓN Y SISTEMA

// GetExcelValidationReport maneja GET /api/contactos/validation
func (h *ContactoHandler) GetExcelValidationReport(w http.ResponseWriter, r *http.Request) {
	report, err := h.service.GetExcelValidationReport()
	if err != nil {
		utils.InternalServerErrorResponse(w, "Error obteniendo reporte: "+err.Error())
		return
	}
	utils.SuccessResponse(w, report)
}

// ReloadExcel maneja POST /api/contactos/reload
func (h *ContactoHandler) ReloadExcel(w http.ResponseWriter, r *http.Request) {
	report, err := h.service.ReloadExcel()
	if err != nil {
		utils.InternalServerErrorResponse(w, "Error recargando Excel: "+err.Error())
		return
	}
	utils.SuccessResponse(w, report)
}

// GetValidationErrors maneja GET /api/contactos/errors
func (h *ContactoHandler) GetValidationErrors(w http.ResponseWriter, r *http.Request) {
	report, err := h.service.GetExcelValidationReport()
	if err != nil {
		utils.InternalServerErrorResponse(w, "Error obteniendo errores: "+err.Error())
		return
	}
	utils.SuccessResponse(w, report.Errors)
}

// GetContactosConEstadoValidacion maneja GET /api/contactos/con-validacion
func (h *ContactoHandler) GetContactosConEstadoValidacion(w http.ResponseWriter, r *http.Request) {
	contactos, err := h.service.GetAllContactos()
	if err != nil {
		utils.InternalServerErrorResponse(w, "Error obteniendo contactos: "+err.Error())
		return
	}
	utils.SuccessResponse(w, contactos)
}

// ✅ GetInvalidContactsForCorrection maneja GET /api/contactos/invalid-data (CORREGIDO)
func (h *ContactoHandler) GetInvalidContactsForCorrection(w http.ResponseWriter, r *http.Request) {
	data, err := h.service.GetInvalidContactsForCorrection()
	if err != nil {
		utils.InternalServerErrorResponse(w, "Error obteniendo datos inválidos: "+err.Error())
		return
	}
	
	utils.SuccessResponse(w, data)
}

// HealthCheck maneja GET /api/health
func (h *ContactoHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	utils.SuccessResponse(w, map[string]interface{}{
		"status":    "ok",
		"service":   "contactos-api",
		"timestamp": r.Header.Get("Date"),
	})
}