package routes

import (
	"contactos-api/handlers"
	"contactos-api/services"

	"github.com/gorilla/mux"
)

// SetupRoutes configura todas las rutas de la API incluyendo paginación
func SetupRoutes(contactoService services.ContactoServiceInterface) *mux.Router {
	router := mux.NewRouter()

	// Crear handler
	contactoHandler := handlers.NewContactoHandler(contactoService)

	// Configurar rutas API
	api := router.PathPrefix("/api").Subrouter()

	// Rutas de contactos
	contactos := api.PathPrefix("/contactos").Subrouter()
	
	// ⚡ RUTAS OPTIMIZADAS PARA GRANDES DATASETS (agregar primero)
	contactos.HandleFunc("/paginated", contactoHandler.GetContactosPaginated).Methods("GET")
	contactos.HandleFunc("/search", contactoHandler.SearchContactosPaginated).Methods("GET")
	contactos.HandleFunc("/count", contactoHandler.GetContactosCount).Methods("GET")
	
	// ✅ RUTAS DE VALIDACIÓN Y SISTEMA (corregidas)
	contactos.HandleFunc("/stats", contactoHandler.GetContactoStats).Methods("GET")
	contactos.HandleFunc("/validation", contactoHandler.GetExcelValidationReport).Methods("GET")
	contactos.HandleFunc("/errors", contactoHandler.GetValidationErrors).Methods("GET")
	contactos.HandleFunc("/invalid-data", contactoHandler.GetInvalidContactsForCorrection).Methods("GET")
	contactos.HandleFunc("/con-validacion", contactoHandler.GetContactosConEstadoValidacion).Methods("GET")
	contactos.HandleFunc("/reload", contactoHandler.ReloadExcel).Methods("POST")
	
	// 📊 RUTAS BÁSICAS - MODIFICADAS para aceptar claves alfanuméricas
	contactos.HandleFunc("", contactoHandler.GetAllContactos).Methods("GET")
	contactos.HandleFunc("", contactoHandler.CreateContacto).Methods("POST")
	// ✅ Permitir claves alfanuméricas (letras, números, guiones, etc.)
	contactos.HandleFunc("/{clave:[A-Za-z0-9._-]+}", contactoHandler.GetContactoByID).Methods("GET")
	contactos.HandleFunc("/{clave:[A-Za-z0-9._-]+}", contactoHandler.UpdateContacto).Methods("PUT")
	contactos.HandleFunc("/{clave:[A-Za-z0-9._-]+}", contactoHandler.DeleteContacto).Methods("DELETE")

	// Rutas adicionales existentes
	contactos.HandleFunc("/buscar", contactoHandler.SearchContactos).Methods("GET")

	// Health check
	api.HandleFunc("/health", contactoHandler.HealthCheck).Methods("GET")

	return router
}