// main.go
package main

import (
	"fmt"
	"log"
	"net/http"

	"contactos-api/config"
	"contactos-api/repositories"
	"contactos-api/routes"
	"contactos-api/services"

	"github.com/rs/cors"
)

func main() {
	// Cargar configuración
	cfg := config.Load()

	// Inicializar repositorio
	contactoRepo := repositories.NewContactoRepository(cfg.ExcelFile)

	// Inicializar servicio
	contactoService := services.NewContactoService(contactoRepo)

	// Configurar rutas
	router := routes.SetupRoutes(contactoService)

	// Configurar CORS
	c := cors.New(cors.Options{
		AllowedOrigins: []string{
			"http://localhost:3000",  // Create React App
			"http://localhost:5173",  // Vite
			"http://localhost:4173",  // Vite preview
		},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"*"},
		AllowCredentials: true,
	})

	handler := c.Handler(router)

	fmt.Printf("🚀 Servidor iniciado en puerto %s\n", cfg.Port)
	fmt.Printf("📊 Archivo Excel: %s\n", cfg.ExcelFile)
	fmt.Printf("🌐 Frontend URL: http://localhost:3000\n")
	fmt.Printf("🔗 API Base URL: http://localhost:%s/api\n", cfg.Port)

	log.Fatal(http.ListenAndServe(":"+cfg.Port, handler))
}