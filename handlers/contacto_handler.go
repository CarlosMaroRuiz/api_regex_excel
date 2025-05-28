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
		utils.NotFoundResponse(w, "Contacto no encontrado")
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
		utils.NotFoundResponse(w, "Contacto no encontrado")
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
	utils.SuccessResponse(w, map[string]string{"message": "Stats no disponible"})
}

// GetExcelValidationReport maneja GET /api/contactos/validation
func (h *ContactoHandler) GetExcelValidationReport(w http.ResponseWriter, r *http.Request) {
	report, err := h.service.GetExcelValidationReport()
	if err != nil {
		utils.InternalServerErrorResponse(w, "Error obteniendo reporte")
		return
	}
	utils.SuccessResponse(w, report)
}

// ReloadExcel maneja POST /api/contactos/reload
func (h *ContactoHandler) ReloadExcel(w http.ResponseWriter, r *http.Request) {
	report, err := h.service.ReloadExcel()
	if err != nil {
		utils.InternalServerErrorResponse(w, "Error recargando Excel")
		return
	}
	utils.SuccessResponse(w, report)
}

// GetValidationErrors maneja GET /api/contactos/errors
func (h *ContactoHandler) GetValidationErrors(w http.ResponseWriter, r *http.Request) {
	report, err := h.service.GetExcelValidationReport()
	if err != nil {
		utils.InternalServerErrorResponse(w, "Error obteniendo errores")
		return
	}
	utils.SuccessResponse(w, report.Errors)
}

// GetContactosConEstadoValidacion maneja GET /api/contactos/con-validacion
func (h *ContactoHandler) GetContactosConEstadoValidacion(w http.ResponseWriter, r *http.Request) {
	contactos, err := h.service.GetAllContactos()
	if err != nil {
		utils.InternalServerErrorResponse(w, "Error obteniendo contactos")
		return
	}
	utils.SuccessResponse(w, contactos)
}

// GetInvalidContactsForCorrection maneja GET /api/contactos/invalid-data
func (h *ContactoHandler) GetInvalidContactsForCorrection(w http.ResponseWriter, r *http.Request) {
	if service, ok := h.service.(interface{ GetInvalidContactsForCorrection() ([]models.RowData, error) }); ok {
		data, err := service.GetInvalidContactsForCorrection()
		if err != nil {
			utils.InternalServerErrorResponse(w, "Error obteniendo datos inválidos")
			return
		}
		utils.SuccessResponse(w, data)
	} else {
		utils.SuccessResponse(w, []models.RowData{})
	}
}

// HealthCheck maneja GET /api/health
func (h *ContactoHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	utils.SuccessResponse(w, map[string]interface{}{
		"status":    "ok",
		"service":   "contactos-api",
		"timestamp": r.Header.Get("Date"),
	})
}