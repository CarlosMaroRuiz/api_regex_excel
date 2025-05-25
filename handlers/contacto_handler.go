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

// HealthCheck maneja GET /api/health
func (h *ContactoHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":    "ok",
		"service":   "contactos-api",
		"version":   "1.0.0",
		"timestamp": r.Header.Get("Date"),
	}

	utils.SuccessResponse(w, health)
}