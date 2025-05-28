// routes/routes.go
package routes

import (


	"contactos-api/handlers"
	"contactos-api/services"

	"github.com/gorilla/mux"
)

// SetupRoutes configura todas las rutas de la API
func SetupRoutes(contactoService services.ContactoServiceInterface) *mux.Router {
	router := mux.NewRouter()

	// Crear handler - usar el constructor que ya existe en tu proyecto
	contactoHandler := handlers.NewContactoHandler(contactoService)

	// Configurar rutas API
	api := router.PathPrefix("/api").Subrouter()

	// Rutas de contactos
	contactos := api.PathPrefix("/contactos").Subrouter()
	
	// Rutas básicas
	contactos.HandleFunc("", contactoHandler.GetAllContactos).Methods("GET")
	contactos.HandleFunc("", contactoHandler.CreateContacto).Methods("POST")
	contactos.HandleFunc("/{clave:[0-9]+}", contactoHandler.GetContactoByID).Methods("GET")
	contactos.HandleFunc("/{clave:[0-9]+}", contactoHandler.UpdateContacto).Methods("PUT")
	contactos.HandleFunc("/{clave:[0-9]+}", contactoHandler.DeleteContacto).Methods("DELETE")

	// Rutas adicionales
	contactos.HandleFunc("/buscar", contactoHandler.SearchContactos).Methods("GET")
	contactos.HandleFunc("/stats", contactoHandler.GetContactoStats).Methods("GET")
	contactos.HandleFunc("/validation", contactoHandler.GetExcelValidationReport).Methods("GET")
	contactos.HandleFunc("/errors", contactoHandler.GetValidationErrors).Methods("GET")
	contactos.HandleFunc("/reload", contactoHandler.ReloadExcel).Methods("POST")
	contactos.HandleFunc("/con-validacion", contactoHandler.GetContactosConEstadoValidacion).Methods("GET")
	contactos.HandleFunc("/invalid-data", contactoHandler.GetInvalidContactsForCorrection).Methods("GET")

	// Health check
	api.HandleFunc("/health", contactoHandler.HealthCheck).Methods("GET")

	return router
}