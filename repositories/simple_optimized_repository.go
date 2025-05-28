// repositories/simple_optimized_repository.go
package repositories

import (
	"fmt"
	
	"strconv"
	"strings"
	"sync"
	"time"

	"contactos-api/models"

	"github.com/tealeg/xlsx/v3"
)

// SimpleOptimizedContactoRepository - Versión optimizada compatible con la interfaz existente
type SimpleOptimizedContactoRepository struct {
	// Campos del repositorio original
	excelFile        string
	contactos        []models.Contacto
	loadErrors       []models.RowError
	invalidRowsData  []models.RowData
	
	// 🚀 OPTIMIZACIONES BÁSICAS
	indiceClaveCliente map[int]*models.Contacto
	indiceCorreo       map[string]*models.Contacto
	searchCache        map[string][]models.Contacto
	
	// Configuración
	useOptimization  bool
	cacheMaxSize     int
	
	// Métricas
	searchCount     int64
	cacheHits       int64
	cacheMisses     int64
	loadTime        time.Duration
	
	// Sincronización
	mu      sync.RWMutex
	cacheMu sync.RWMutex
}

// NewSimpleOptimizedContactoRepository crea repositorio optimizado simple
func NewSimpleOptimizedContactoRepository(excelFile string) *SimpleOptimizedContactoRepository {
	repo := &SimpleOptimizedContactoRepository{
		excelFile:       excelFile,
		contactos:       make([]models.Contacto, 0),
		loadErrors:      make([]models.RowError, 0),
		invalidRowsData: make([]models.RowData, 0),
		useOptimization: true,
		cacheMaxSize:    500, // Cache más pequeño pero efectivo
		searchCache:     make(map[string][]models.Contacto),
	}
	
	startTime := time.Now()
	fmt.Println("🚀 Iniciando carga optimizada...")
	
	// Cargar datos
	if err := repo.loadFromExcel(); err != nil {
		fmt.Printf("⚠️ Error cargando Excel: %v\n", err)
	}
	
	repo.loadTime = time.Since(startTime)
	
	// Construir índices si hay suficientes contactos
	if len(repo.contactos) > 100 {
		repo.buildBasicIndices()
		fmt.Printf("🔍 Índices construidos para %d contactos\n", len(repo.contactos))
	}
	
	fmt.Printf("✅ Carga completada en %v - %d contactos válidos, %d inválidos\n", 
		repo.loadTime, len(repo.contactos), len(repo.invalidRowsData))
	
	return repo
}

// buildBasicIndices construye índices básicos para búsquedas rápidas
func (r *SimpleOptimizedContactoRepository) buildBasicIndices() {
	r.indiceClaveCliente = make(map[int]*models.Contacto, len(r.contactos))
	r.indiceCorreo = make(map[string]*models.Contacto, len(r.contactos))
	
	for i := range r.contactos {
		contacto := &r.contactos[i]
		r.indiceClaveCliente[contacto.ClaveCliente] = contacto
		r.indiceCorreo[strings.ToLower(contacto.Correo)] = contacto
	}
}

// 🚀 IMPLEMENTACIÓN DE LA INTERFAZ CON OPTIMIZACIONES

func (r *SimpleOptimizedContactoRepository) GetAll() ([]models.Contacto, error) {
	if r.useOptimization && len(r.contactos) > 1000 {
		r.mu.RLock()
		defer r.mu.RUnlock()
	}
	return r.contactos, nil
}

func (r *SimpleOptimizedContactoRepository) GetByID(claveCliente int) (*models.Contacto, error) {
	// Usar índice si está disponible
	if r.useOptimization && r.indiceClaveCliente != nil {
		r.mu.RLock()
		defer r.mu.RUnlock()
		
		if contacto, exists := r.indiceClaveCliente[claveCliente]; exists {
			copia := *contacto
			return &copia, nil
		}
		return nil, fmt.Errorf("contacto con clave %d no encontrado", claveCliente)
	}
	
	// Búsqueda secuencial como fallback
	for i, contacto := range r.contactos {
		if contacto.ClaveCliente == claveCliente {
			return &r.contactos[i], nil
		}
	}
	
	return nil, fmt.Errorf("contacto con clave %d no encontrado", claveCliente)
}

func (r *SimpleOptimizedContactoRepository) Search(criteria *models.ContactoDTO) ([]models.Contacto, error) {
	r.searchCount++
	
	// Generar clave de cache
	cacheKey := r.generateCacheKey(criteria)
	
	// Verificar cache
	if r.useOptimization {
		r.cacheMu.RLock()
		if cached, exists := r.searchCache[cacheKey]; exists {
			r.cacheMu.RUnlock()
			r.cacheHits++
			return cached, nil
		}
		r.cacheMu.RUnlock()
		r.cacheMisses++
	}
	
	var resultados []models.Contacto
	
	// Búsqueda optimizada por clave cliente
	if criteria.ClaveCliente != "" {
		if clave, err := strconv.Atoi(criteria.ClaveCliente); err == nil {
			if contacto, err := r.GetByID(clave); err == nil {
				resultados = []models.Contacto{*contacto}
			}
		}
	} else if criteria.Correo != "" && r.indiceCorreo != nil {
		// Búsqueda optimizada por correo
		r.mu.RLock()
		if contacto, exists := r.indiceCorreo[strings.ToLower(criteria.Correo)]; exists {
			resultados = []models.Contacto{*contacto}
		}
		r.mu.RUnlock()
	} else {
		// Búsqueda secuencial para otros criterios
		resultados = r.sequentialSearch(criteria)
	}
	
	// Guardar en cache
	if r.useOptimization && len(resultados) < 100 {
		r.cacheMu.Lock()
		if len(r.searchCache) >= r.cacheMaxSize {
			// Limpiar cache simple
			r.searchCache = make(map[string][]models.Contacto)
		}
		r.searchCache[cacheKey] = resultados
		r.cacheMu.Unlock()
	}
	
	return resultados, nil
}

// sequentialSearch búsqueda secuencial para criterios múltiples
func (r *SimpleOptimizedContactoRepository) sequentialSearch(criteria *models.ContactoDTO) []models.Contacto {
	var resultados []models.Contacto
	
	for _, contacto := range r.contactos {
		match := true
		
		if criteria.ClaveCliente != "" {
			if clave, err := strconv.Atoi(criteria.ClaveCliente); err != nil || contacto.ClaveCliente != clave {
				match = false
			}
		}
		
		if criteria.Nombre != "" && !strings.Contains(
			strings.ToLower(contacto.Nombre), 
			strings.ToLower(criteria.Nombre),
		) {
			match = false
		}
		
		if criteria.Correo != "" && !strings.Contains(
			strings.ToLower(contacto.Correo), 
			strings.ToLower(criteria.Correo),
		) {
			match = false
		}
		
		if criteria.Telefono != "" && !strings.Contains(
			contacto.TelefonoContacto, 
			criteria.Telefono,
		) {
			match = false
		}
		
		if match {
			resultados = append(resultados, contacto)
		}
	}
	
	return resultados
}

func (r *SimpleOptimizedContactoRepository) Create(contacto *models.Contacto) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	// Verificar duplicado usando índice si está disponible
	if r.indiceClaveCliente != nil {
		if _, exists := r.indiceClaveCliente[contacto.ClaveCliente]; exists {
			return fmt.Errorf("contacto con clave %d ya existe", contacto.ClaveCliente)
		}
	}
	
	// Agregar contacto
	r.contactos = append(r.contactos, *contacto)
	nuevoContacto := &r.contactos[len(r.contactos)-1]
	
	// Actualizar índices
	if r.indiceClaveCliente != nil {
		r.indiceClaveCliente[contacto.ClaveCliente] = nuevoContacto
	}
	if r.indiceCorreo != nil {
		r.indiceCorreo[strings.ToLower(contacto.Correo)] = nuevoContacto
	}
	
	// Limpiar cache
	r.clearCache()
	
	return r.saveToExcel()
}

func (r *SimpleOptimizedContactoRepository) Update(contacto *models.Contacto) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	// Encontrar contacto usando índice
	if r.indiceClaveCliente != nil {
		if existente, exists := r.indiceClaveCliente[contacto.ClaveCliente]; exists {
			*existente = *contacto
			r.clearCache()
			return r.saveToExcel()
		}
	} else {
		// Búsqueda secuencial
		for i, c := range r.contactos {
			if c.ClaveCliente == contacto.ClaveCliente {
				r.contactos[i] = *contacto
				r.clearCache()
				return r.saveToExcel()
			}
		}
	}
	
	return fmt.Errorf("contacto con clave %d no encontrado", contacto.ClaveCliente)
}

func (r *SimpleOptimizedContactoRepository) Delete(claveCliente int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	// Encontrar índice
	indice := -1
	for i, contacto := range r.contactos {
		if contacto.ClaveCliente == claveCliente {
			indice = i
			break
		}
	}
	
	if indice == -1 {
		return fmt.Errorf("contacto con clave %d no encontrado", claveCliente)
	}
	
	// Eliminar del slice
	r.contactos = append(r.contactos[:indice], r.contactos[indice+1:]...)
	
	// Actualizar índices
	if r.indiceClaveCliente != nil {
		delete(r.indiceClaveCliente, claveCliente)
	}
	
	// Reconstruir índice de correo (simple)
	if r.indiceCorreo != nil {
		r.buildBasicIndices()
	}
	
	r.clearCache()
	return r.saveToExcel()
}

func (r *SimpleOptimizedContactoRepository) ExistsByID(claveCliente int) (bool, error) {
	if r.indiceClaveCliente != nil {
		r.mu.RLock()
		defer r.mu.RUnlock()
		_, exists := r.indiceClaveCliente[claveCliente]
		return exists, nil
	}
	
	// Búsqueda secuencial
	for _, contacto := range r.contactos {
		if contacto.ClaveCliente == claveCliente {
			return true, nil
		}
	}
	
	return false, nil
}

func (r *SimpleOptimizedContactoRepository) GetLoadErrors() []models.RowError {
	return r.loadErrors
}

func (r *SimpleOptimizedContactoRepository) GetInvalidRowsData() []models.RowData {
	return r.invalidRowsData
}

func (r *SimpleOptimizedContactoRepository) ReloadExcel() ([]models.RowError, []models.RowData, error) {
	startTime := time.Now()
	fmt.Println("🔄 Recargando Excel...")
	
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if err := r.loadFromExcel(); err != nil {
		return r.loadErrors, r.invalidRowsData, err
	}
	
	// Reconstruir índices
	if len(r.contactos) > 100 {
		r.buildBasicIndices()
	}
	
	r.clearCache()
	r.loadTime = time.Since(startTime)
	
	fmt.Printf("✅ Recarga completada en %v\n", r.loadTime)
	return r.loadErrors, r.invalidRowsData, nil
}

// 🔧 FUNCIONES AUXILIARES

func (r *SimpleOptimizedContactoRepository) generateCacheKey(criteria *models.ContactoDTO) string {
	return fmt.Sprintf("c:%s|n:%s|e:%s|t:%s", 
		criteria.ClaveCliente, criteria.Nombre, criteria.Correo, criteria.Telefono)
}

func (r *SimpleOptimizedContactoRepository) clearCache() {
	if r.useOptimization {
		r.cacheMu.Lock()
		r.searchCache = make(map[string][]models.Contacto)
		r.cacheMu.Unlock()
	}
}

// GetStats retorna estadísticas básicas
func (r *SimpleOptimizedContactoRepository) GetStats() map[string]interface{} {
	cacheHitRate := 0.0
	if r.cacheHits+r.cacheMisses > 0 {
		cacheHitRate = (float64(r.cacheHits) / float64(r.cacheHits+r.cacheMisses)) * 100
	}
	
	return map[string]interface{}{
		"contactos_count":    len(r.contactos),
		"load_time_ms":       r.loadTime.Milliseconds(),
		"search_count":       r.searchCount,
		"cache_hit_rate":     cacheHitRate,
		"cache_hits":         r.cacheHits,
		"cache_misses":       r.cacheMisses,
		"use_optimization":   r.useOptimization,
		"index_sizes": map[string]int{
			"clave_cliente": len(r.indiceClaveCliente),
			"correo":        len(r.indiceCorreo),
		},
	}
}

// 📄 CARGA Y GUARDADO OPTIMIZADOS

func (r *SimpleOptimizedContactoRepository) loadFromExcel() error {
	file, err := xlsx.OpenFile(r.excelFile)
	if err != nil {
		return fmt.Errorf("error abriendo Excel: %w", err)
	}

	if len(file.Sheets) == 0 {
		return fmt.Errorf("archivo sin hojas")
	}

	sheet := file.Sheets[0]
	
	// Limpiar datos anteriores
	r.contactos = r.contactos[:0]
	r.loadErrors = r.loadErrors[:0]
	r.invalidRowsData = r.invalidRowsData[:0]
	
	// Procesar filas
	rowIndex := 0
	err = sheet.ForEachRow(func(row *xlsx.Row) error {
		if rowIndex == 0 { // Saltar header
			rowIndex++
			return nil
		}

		currentRow := rowIndex + 1

		// Obtener celdas
		var cells [4]string
		cellIndex := 0
		row.ForEachCell(func(cell *xlsx.Cell) error {
			if cellIndex < 4 {
				cells[cellIndex] = strings.TrimSpace(cell.String())
				cellIndex++
			}
			return nil
		})

		if cellIndex < 4 {
			// Fila incompleta, agregar error
			rowData := models.RowData{
				ClaveCliente:     cells[0],
				Nombre:           cells[1],
				Correo:           cells[2],
				TelefonoContacto: cells[3],
				HasErrors:        true,
				ErrorCount:       1,
			}
			
			r.invalidRowsData = append(r.invalidRowsData, rowData)
			r.loadErrors = append(r.loadErrors, models.RowError{
				Row:     currentRow,
				Column:  "general",
				Field:   "estructura",
				Error:   "Fila incompleta",
				RowData: &rowData,
			})
			
			rowIndex++
			return nil
		}

		// Validar y procesar fila completa
		claveStr, nombre, correo, telefono := cells[0], cells[1], cells[2], cells[3]

		rowData := models.RowData{
			ClaveCliente:     claveStr,
			Nombre:           nombre,
			Correo:           correo,
			TelefonoContacto: telefono,
			HasErrors:        false,
			ErrorCount:       0,
		}

		var rowErrors []models.RowError

		// Validaciones básicas
		if claveStr == "" || nombre == "" || correo == "" || telefono == "" {
			rowData.HasErrors = true
			rowData.ErrorCount++
			rowErrors = append(rowErrors, models.RowError{
				Row: currentRow, Field: "general", Error: "Campos vacíos", RowData: &rowData,
			})
		}

		// Validar clave cliente
		clave := 0
		if claveStr != "" {
			if c, err := strconv.Atoi(claveStr); err != nil || c <= 0 {
				rowData.HasErrors = true
				rowData.ErrorCount++
				rowErrors = append(rowErrors, models.RowError{
					Row: currentRow, Field: "claveCliente", Error: "Clave inválida", RowData: &rowData,
				})
			} else {
				clave = c
			}
		}

		// Validar teléfono
		if telefono != "" && len(telefono) != 10 {
			rowData.HasErrors = true
			rowData.ErrorCount++
			rowErrors = append(rowErrors, models.RowError{
				Row: currentRow, Field: "telefonoContacto", Error: "Teléfono debe tener 10 dígitos", RowData: &rowData,
			})
		}

		// Validar correo básico
		if correo != "" && !strings.Contains(correo, "@") {
			rowData.HasErrors = true
			rowData.ErrorCount++
			rowErrors = append(rowErrors, models.RowError{
				Row: currentRow, Field: "correo", Error: "Correo sin @", RowData: &rowData,
			})
		}

		r.loadErrors = append(r.loadErrors, rowErrors...)

		if rowData.HasErrors {
			r.invalidRowsData = append(r.invalidRowsData, rowData)
		} else {
			// Crear contacto válido
			contacto := models.Contacto{
				ClaveCliente:     clave,
				Nombre:           nombre,
				Correo:           correo,
				TelefonoContacto: telefono,
			}
			r.contactos = append(r.contactos, contacto)
		}

		rowIndex++
		return nil
	})

	return err
}

func (r *SimpleOptimizedContactoRepository) saveToExcel() error {
	file := xlsx.NewFile()
	sheet, err := file.AddSheet("Contactos")
	if err != nil {
		return fmt.Errorf("error creando hoja: %w", err)
	}

	// Headers
	headerRow := sheet.AddRow()
	headerRow.AddCell().Value = "ClaveCliente"
	headerRow.AddCell().Value = "Nombre"
	headerRow.AddCell().Value = "Correo"
	headerRow.AddCell().Value = "TelefonoContacto"

	// Datos
	for _, contacto := range r.contactos {
		row := sheet.AddRow()
		row.AddCell().Value = strconv.Itoa(contacto.ClaveCliente)
		row.AddCell().Value = contacto.Nombre
		row.AddCell().Value = contacto.Correo
		row.AddCell().Value = contacto.TelefonoContacto
	}

	return file.Save(r.excelFile)
}