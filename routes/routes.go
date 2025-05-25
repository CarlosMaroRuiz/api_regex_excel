// routes/routes.go
package routes

import (
	"net/http"

	"contactos-api/handlers"
	"contactos-api/services"

	"github.com/gorilla/mux"
)

// SetupRoutes configura todas las rutas de la API
func SetupRoutes(contactoService services.ContactoServiceInterface) *mux.Router {
	router := mux.NewRouter()

	// Crear handler
	contactoHandler := handlers.NewContactoHandler(contactoService)

	// Configurar rutas API
	api := router.PathPrefix("/api").Subrouter()

	// Rutas de contactos
	contactos := api.PathPrefix("/contactos").Subrouter()
	
	// 🆕 NUEVO: GET /api/contactos/con-validacion - Obtener contactos con estado de validación
	contactos.HandleFunc("/con-validacion", contactoHandler.GetContactosConEstadoValidacion).Methods("GET")
	
	// GET /api/contactos/buscar - Buscar contactos (debe ir antes que /{clave})
	contactos.HandleFunc("/buscar", contactoHandler.SearchContactos).Methods("GET")
	
	// GET /api/contactos/stats - Estadísticas
	contactos.HandleFunc("/stats", contactoHandler.GetContactoStats).Methods("GET")
	
	// GET /api/contactos/validation - Reporte de validación del Excel
	contactos.HandleFunc("/validation", contactoHandler.GetExcelValidationReport).Methods("GET")
	
	// GET /api/contactos/errors - Errores de validación detallados
	contactos.HandleFunc("/errors", contactoHandler.GetValidationErrors).Methods("GET")
	
	// POST /api/contactos/reload - Recargar archivo Excel
	contactos.HandleFunc("/reload", contactoHandler.ReloadExcel).Methods("POST")
	
	// GET /api/contactos - Obtener todos los contactos
	contactos.HandleFunc("", contactoHandler.GetAllContactos).Methods("GET")
	
	// POST /api/contactos - Crear nuevo contacto
	contactos.HandleFunc("", contactoHandler.CreateContacto).Methods("POST")
	
	// GET /api/contactos/{clave} - Obtener contacto específico
	contactos.HandleFunc("/{clave:[0-9]+}", contactoHandler.GetContactoByID).Methods("GET")
	
	// PUT /api/contactos/{clave} - Actualizar contacto
	contactos.HandleFunc("/{clave:[0-9]+}", contactoHandler.UpdateContacto).Methods("PUT")
	
	// DELETE /api/contactos/{clave} - Eliminar contacto
	contactos.HandleFunc("/{clave:[0-9]+}", contactoHandler.DeleteContacto).Methods("DELETE")

	// Ruta de salud
	api.HandleFunc("/health", contactoHandler.HealthCheck).Methods("GET")

	// Middleware para logging (opcional)
	router.Use(loggingMiddleware)

	return router
}

// loggingMiddleware middleware para logging de requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Aquí puedes agregar logging
		// fmt.Printf("[%s] %s %s\n", time.Now().Format("2006-01-02 15:04:05"), r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}