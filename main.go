// main.go - Versión optimizada simple y compatible
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"syscall"
	"time"

	"contactos-api/config"
	"contactos-api/repositories"
	"contactos-api/routes"
	"contactos-api/services"

	"github.com/rs/cors"
	"github.com/tealeg/xlsx/v3"
)

func main() {
	startTime := time.Now()
	
	// 🚀 CONFIGURACIÓN INICIAL
	fmt.Println("🚀 Iniciando API Optimizada para Contactos...")
	fmt.Printf("⚙️ Hardware: %d CPUs, Go %s\n", 
		runtime.NumCPU(), runtime.Version())
	
	// Cargar configuración
	cfg := config.Load()
	
	fmt.Println("🚀 Configuración Básica")
	fmt.Printf("Puerto: %s\n", cfg.Port)
	fmt.Printf("Excel: %s\n", cfg.ExcelFile)
	
	// 🧠 CONFIGURAR RUNTIME PARA RENDIMIENTO
	configureRuntime(cfg)
	
	// 📊 MÉTRICAS INICIALES
	var memBefore runtime.MemStats
	runtime.ReadMemStats(&memBefore)
	
	// 🗂️ INICIALIZAR REPOSITORIO
	fmt.Printf("📄 Cargando archivo Excel: %s\n", cfg.ExcelFile)
	
	// Crear archivo vacío si no existe
	if !fileExists(cfg.ExcelFile) {
		fmt.Printf("⚠️ Archivo no encontrado. Creando: %s\n", cfg.ExcelFile)
		createEmptyExcelFile(cfg.ExcelFile)
	}
	
	// Elegir repositorio (usar optimizado por defecto)
	var contactoRepo repositories.ContactoRepositoryInterface
	
	fmt.Println("🚀 Usando repositorio optimizado...")
	contactoRepo = repositories.NewSimpleOptimizedContactoRepository(cfg.ExcelFile)
	
	// Mostrar estadísticas si está disponible
	if optimizedRepo, ok := contactoRepo.(*repositories.SimpleOptimizedContactoRepository); ok {
		stats := optimizedRepo.GetStats()
		fmt.Printf("📊 Repo stats: %d contactos, %.1f%% cache hits\n", 
			stats["contactos_count"], stats["cache_hit_rate"])
	}
	
	// 🔧 INICIALIZAR SERVICIO
	contactoService := services.NewContactoService(contactoRepo)
	
	// 🌐 CONFIGURAR RUTAS
	router := routes.SetupRoutes(contactoService)
	
	// CORS optimizado
	c := cors.New(cors.Options{
		AllowedOrigins: []string{
			"http://localhost:3000",
			"http://localhost:5173",
			"http://localhost:4173",
		},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		MaxAge:          3600, // Cache preflight 1 hora
	})
	
	handler := c.Handler(router)
	
	// 🔧 SERVIDOR HTTP
	server := &http.Server{
		Addr:           ":" + cfg.Port,
		Handler:        handler,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 10 * 1024 * 1024, // 10MB
	}
	
	// 📈 MOSTRAR ESTADÍSTICAS DE INICIO
	loadTime := time.Since(startTime)
	contactos, _ := contactoRepo.GetAll()
	
	var memAfter runtime.MemStats
	runtime.ReadMemStats(&memAfter)
	memUsedMB := float64(memAfter.Alloc-memBefore.Alloc) / 1024 / 1024
	
	fmt.Println("\n✅ API Optimizada Lista!")
	fmt.Println("==========================================")
	fmt.Printf("⏱️  Tiempo de inicio: %v\n", loadTime)
	fmt.Printf("🧠 Memoria utilizada: %.2f MB\n", memUsedMB)
	fmt.Printf("📊 Contactos cargados: %d\n", len(contactos))
	fmt.Printf("🚀 Servidor: http://localhost:%s\n", cfg.Port)
	fmt.Printf("📡 API: http://localhost:%s/api\n", cfg.Port)
	fmt.Printf("🌐 Frontend: http://localhost:3000\n")
	fmt.Printf("❤️  Health: http://localhost:%s/api/health\n", cfg.Port)
	fmt.Printf("🎯 Perfil: OPTIMIZADO\n")
	
	// Endpoints principales
	fmt.Println("\n🔗 Endpoints Optimizados:")
	fmt.Printf("   GET  /api/contactos - Todos los contactos (%d)\n", len(contactos))
	fmt.Println("   GET  /api/contactos/buscar?nombre=X - Búsqueda optimizada")
	fmt.Println("   GET  /api/contactos/con-validacion - Con validaciones")
	fmt.Println("   GET  /api/contactos/invalid-data - Datos para corrección")
	fmt.Println("   POST /api/contactos/reload - Recargar Excel")
	fmt.Println("   GET  /api/contactos/performance-stats - Estadísticas")
	fmt.Println("==========================================")
	
	// 🔄 INICIAR SERVIDOR
	go func() {
		fmt.Printf("🟢 Servidor iniciado (PID: %d)\n", os.Getpid())
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Error servidor: %v", err)
		}
	}()
	
	// 📊 MONITOREO
	go startPerformanceMonitoring(contactoRepo)
	
	// 🛑 GRACEFUL SHUTDOWN
	setupGracefulShutdown(server)
	
	// ⏳ MANTENER SERVIDOR ACTIVO
	select {}
}

// 🧠 configureRuntime optimiza configuración de Go runtime
func configureRuntime(cfg *config.Config) {
	fmt.Println("🔧 Optimizando runtime básico...")
	
	// Configuración básica de GC
	debug.SetGCPercent(100)
	
	fmt.Println("✅ Runtime configurado")
}

// 📊 startPerformanceMonitoring inicia monitoreo de rendimiento
func startPerformanceMonitoring(repo repositories.ContactoRepositoryInterface) {
	fmt.Println("📊 Iniciando monitoreo de rendimiento...")
	
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		
		// Estadísticas básicas
		contactos, _ := repo.GetAll()
		memMB := float64(m.Alloc) / 1024 / 1024
		
		// Estadísticas avanzadas si disponibles
		if optimizedRepo, ok := repo.(*repositories.SimpleOptimizedContactoRepository); ok {
			stats := optimizedRepo.GetStats()
			fmt.Printf("📈 [%s] Mem: %.1fMB | Contactos: %d | Búsquedas: %d | Cache: %.1f%% hits\n",
				time.Now().Format("15:04:05"),
				memMB,
				len(contactos),
				stats["search_count"],
				stats["cache_hit_rate"])
		} else {
			fmt.Printf("📈 [%s] Mem: %.1fMB | Contactos: %d\n",
				time.Now().Format("15:04:05"),
				memMB,
				len(contactos))
		}
		
		// Alertas de memoria
		if memMB > 800 { // 800MB límite básico
			fmt.Printf("⚠️ Memoria alta: %.1fMB\n", memMB)
		}
		
		// Limpieza de memoria si es necesario
		if memMB > 1000 { // 1GB límite
			fmt.Println("🧹 Ejecutando limpieza de memoria...")
			runtime.GC()
			debug.FreeOSMemory()
		}
	}
}

// 🛑 setupGracefulShutdown configura cierre elegante
func setupGracefulShutdown(server *http.Server) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	
	go func() {
		<-c
		fmt.Println("\n🛑 Señal de cierre recibida...")
		
		// Timeout para cierre
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		
		// Cerrar servidor
		fmt.Println("🔄 Cerrando servidor HTTP...")
		if err := server.Shutdown(ctx); err != nil {
			fmt.Printf("❌ Error en cierre: %v\n", err)
		}
		
		// Estadísticas finales
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		fmt.Printf("📊 Memoria final: %.2f MB\n", float64(m.Alloc)/1024/1024)
		fmt.Printf("📊 GC ejecutado: %d veces\n", m.NumGC)
		
		fmt.Println("✅ Cierre completado")
		os.Exit(0)
	}()
}

// 🛠️ FUNCIONES AUXILIARES

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

func createEmptyExcelFile(filename string) {
	file := xlsx.NewFile()
	sheet, err := file.AddSheet("Contactos")
	if err != nil {
		fmt.Printf("❌ Error creando hoja: %v\n", err)
		return
	}
	
	// Headers
	headerRow := sheet.AddRow()
	headerRow.AddCell().Value = "ClaveCliente"
	headerRow.AddCell().Value = "Nombre"
	headerRow.AddCell().Value = "Correo"
	headerRow.AddCell().Value = "TelefonoContacto"
	
	// Datos de ejemplo
	exampleRow := sheet.AddRow()
	exampleRow.AddCell().Value = "1"
	exampleRow.AddCell().Value = "Juan Pérez"
	exampleRow.AddCell().Value = "juan@gmail.com"
	exampleRow.AddCell().Value = "5551234567"
	
	if err := file.Save(filename); err != nil {
		fmt.Printf("❌ Error guardando: %v\n", err)
	} else {
		fmt.Printf("✅ Archivo Excel creado: %s\n", filename)
	}
}