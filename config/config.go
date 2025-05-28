// config/config.go - Versión compatible con optimizaciones
package config

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
)

// Config estructura básica (mantener compatibilidad)
type Config struct {
	Port      string
	ExcelFile string
	APIURL    string
}

// OptimizedConfig configuración extendida para optimizaciones
type OptimizedConfig struct {
	// Configuración básica (heredada)
	Config
	
	// 🚀 CONFIGURACIONES DE RENDIMIENTO
	MaxWorkers        int    // Número máximo de workers
	BatchSize         int    // Tamaño de lote para procesamiento
	CacheSize         int    // Tamaño del cache de búsquedas
	UseOptimizations  bool   // Activar optimizaciones avanzadas
	
	// 📊 CONFIGURACIONES DE MONITOREO
	EnableMetrics     bool   // Habilitar métricas de rendimiento
	LogLevel          string // Nivel de logging
	
	// 🛡️ CONFIGURACIONES DE MEMORIA
	MaxMemoryMB       int    // Límite de memoria en MB
	GCPercent         int    // Porcentaje para garbage collection
	
	// ⚡ CONFIGURACIONES DE RED
	ReadTimeout       int    // Timeout de lectura en segundos
	WriteTimeout      int    // Timeout de escritura en segundos
	MaxRequestSize    int64  // Tamaño máximo de request en bytes
}

// Load carga configuración básica (mantener compatibilidad)
func Load() *Config {
	return &Config{
		Port:      getEnv("PORT", "8080"),
		ExcelFile: getEnv("EXCEL_FILE", "contactos.xlsx"),
		APIURL:    getEnv("API_URL", "http://localhost:8080"),
	}
}

// LoadOptimizedConfig carga configuración optimizada
func LoadOptimizedConfig() *OptimizedConfig {
	numCPU := runtime.NumCPU()
	
	// Cargar configuración básica primero
	basicConfig := Load()
	
	config := &OptimizedConfig{
		Config: *basicConfig,
		
		// Configuraciones inteligentes basadas en hardware
		MaxWorkers:       getEnvInt("MAX_WORKERS", numCPU*2),
		BatchSize:        getEnvInt("BATCH_SIZE", 1000),
		CacheSize:        getEnvInt("CACHE_SIZE", 500),
		UseOptimizations: getEnvBool("USE_OPTIMIZATIONS", true),
		
		// Monitoreo
		EnableMetrics:    getEnvBool("ENABLE_METRICS", true),
		LogLevel:         getEnv("LOG_LEVEL", "INFO"),
		
		// Memoria
		MaxMemoryMB:      getEnvInt("MAX_MEMORY_MB", 1024),
		GCPercent:        getEnvInt("GC_PERCENT", 100),
		
		// Red
		ReadTimeout:      getEnvInt("READ_TIMEOUT", 30),
		WriteTimeout:     getEnvInt("WRITE_TIMEOUT", 30),
		MaxRequestSize:   getEnvInt64("MAX_REQUEST_SIZE", 10*1024*1024), // 10MB
	}
	
	// Ajustes automáticos
	config.autoTune()
	
	return config
}

// autoTune ajusta configuración automáticamente
func (c *OptimizedConfig) autoTune() {
	// Si esperamos datasets grandes, aumentar configuraciones
	if c.isLargeDatasetExpected() {
		c.MaxWorkers = c.MaxWorkers * 2
		c.BatchSize = 2000
		c.CacheSize = 1000
		c.MaxMemoryMB = c.MaxMemoryMB * 2
		c.GCPercent = 200
		
		fmt.Println("🔧 Auto-tuning para dataset grande activado")
	}
	
	// Límites de seguridad
	if c.MaxWorkers > runtime.NumCPU()*4 {
		c.MaxWorkers = runtime.NumCPU() * 4
	}
	
	if c.BatchSize > 5000 {
		c.BatchSize = 5000
	}
}

// isLargeDatasetExpected detecta si esperamos un dataset grande
func (c *OptimizedConfig) isLargeDatasetExpected() bool {
	// Verificar variable de entorno
	if getEnv("LARGE_DATASET", "false") == "true" {
		return true
	}
	
	// Verificar por nombre de archivo
	fileName := strings.ToLower(c.ExcelFile)
	largeIndicators := []string{"large", "big", "100k", "massive", "huge", "grande"}
	
	for _, indicator := range largeIndicators {
		if strings.Contains(fileName, indicator) {
			return true
		}
	}
	
	return false
}

// GetPerformanceProfile retorna perfil de rendimiento
func (c *OptimizedConfig) GetPerformanceProfile() string {
	if c.MaxWorkers >= runtime.NumCPU()*3 && c.BatchSize >= 2000 {
		return "HIGH_PERFORMANCE"
	} else if c.MaxWorkers >= runtime.NumCPU()*2 && c.BatchSize >= 1000 {
		return "BALANCED"
	}
	return "CONSERVATIVE"
}

// PrintConfig imprime la configuración actual
func (c *OptimizedConfig) PrintConfig() {
	profile := c.GetPerformanceProfile()
	
	fmt.Println("🚀 Configuración de Rendimiento")
	fmt.Println("================================")
	fmt.Printf("Perfil: %s\n", profile)
	fmt.Printf("Puerto: %s\n", c.Port)
	fmt.Printf("Excel: %s\n", c.ExcelFile)
	fmt.Printf("Workers: %d (CPUs: %d)\n", c.MaxWorkers, runtime.NumCPU())
	fmt.Printf("Batch Size: %d\n", c.BatchSize)
	fmt.Printf("Cache Size: %d\n", c.CacheSize)
	fmt.Printf("Memoria Máx: %d MB\n", c.MaxMemoryMB)
	fmt.Printf("GC Percent: %d%%\n", c.GCPercent)
	fmt.Printf("Optimizaciones: %t\n", c.UseOptimizations)
	fmt.Printf("Métricas: %t\n", c.EnableMetrics)
	fmt.Println("================================")
}

// 🛠️ FUNCIONES AUXILIARES

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}