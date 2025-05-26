// repositories/contacto_repository.go
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

type ContactoRepositoryInterface interface {
	GetAll() ([]models.Contacto, error)
	GetByID(claveCliente int) (*models.Contacto, error)
	Create(contacto *models.Contacto) error
	Update(contacto *models.Contacto) error
	Delete(claveCliente int) error
	Search(criteria *models.ContactoDTO) ([]models.Contacto, error)
	ExistsByID(claveCliente int) (bool, error)
	GetLoadErrors() []models.RowError
	GetInvalidRowsData() []models.RowData
	ReloadExcel() ([]models.RowError, []models.RowData, error)
}

// ContactoRepository implementa el acceso a datos para contactos
type ContactoRepository struct {
	excelFile        string
	contactos        []models.Contacto
	loadErrors       []models.RowError
	invalidRowsData  []models.RowData
	
	// Optimización condicional
	useOptimization  bool       // Bandera para activar optimizaciones
	
	// Índices para búsquedas (solo si useOptimization=true)
	indiceClaveCliente map[int]*models.Contacto
	
	// Mutex simple solo para proteger operaciones concurrentes
	mu sync.RWMutex
}

// NewContactoRepository crea una nueva instancia del repositorio
func NewContactoRepository(excelFile string) *ContactoRepository {
	repo := &ContactoRepository{
		excelFile:       excelFile,
		contactos:       []models.Contacto{},
		loadErrors:      []models.RowError{},
		invalidRowsData: []models.RowData{},
		useOptimization: false, // Inicialmente desactivado, se activará automáticamente si es necesario
	}
	
	// Cargar datos al inicializar
	startTime := time.Now()
	fmt.Println("🔄 Cargando archivo Excel...")
	
	loadErrors, invalidData, err := repo.loadFromExcel()
	if err != nil {
		fmt.Printf("⚠️ Error cargando Excel: %v. Iniciando con datos vacíos.\n", err)
	}
	
	repo.loadErrors = loadErrors
	repo.invalidRowsData = invalidData
	
	// Si hay muchos contactos, activar optimizaciones
	if len(repo.contactos) > 1000 {
		repo.useOptimization = true
		fmt.Println("🚀 Activando optimizaciones para conjunto de datos grande")
		repo.buildIndices()
	} else {
		fmt.Println("✅ Usando modo estándar para conjunto de datos pequeño")
	}
	
	fmt.Printf("✅ Excel cargado en %v. %d contactos válidos, %d inválidos\n", 
		time.Since(startTime), 
		len(repo.contactos), 
		len(repo.invalidRowsData))
	
	return repo
}

// buildIndices construye índices básicos si es necesario
func (r *ContactoRepository) buildIndices() {
	if !r.useOptimization {
		return
	}
	
	startTime := time.Now()
	fmt.Println("🔍 Construyendo índices para búsquedas rápidas...")
	
	// Solo crear el índice por clave cliente (el más crítico)
	r.indiceClaveCliente = make(map[int]*models.Contacto, len(r.contactos))
	
	for i := range r.contactos {
		contacto := &r.contactos[i]
		r.indiceClaveCliente[contacto.ClaveCliente] = contacto
	}
	
	fmt.Printf("✅ Índice básico construido en %v para %d contactos\n", 
		time.Since(startTime), 
		len(r.contactos))
}

// GetAll retorna todos los contactos
func (r *ContactoRepository) GetAll() ([]models.Contacto, error) {
	// Para conjuntos pequeños no necesitamos mutex
	if !r.useOptimization {
		return r.contactos, nil
	}
	
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.contactos, nil
}

// GetByID busca un contacto por su clave cliente
func (r *ContactoRepository) GetByID(claveCliente int) (*models.Contacto, error) {
	// Usar índice si está disponible
	if r.useOptimization {
		r.mu.RLock()
		defer r.mu.RUnlock()
		
		if contacto, ok := r.indiceClaveCliente[claveCliente]; ok {
			copiado := *contacto
			return &copiado, nil
		}
	} else {
		// Búsqueda secuencial rápida para conjuntos pequeños
		for i, contacto := range r.contactos {
			if contacto.ClaveCliente == claveCliente {
				return &r.contactos[i], nil
			}
		}
	}
	
	return nil, fmt.Errorf("contacto con clave %d no encontrado", claveCliente)
}

// Create crea un nuevo contacto
func (r *ContactoRepository) Create(contacto *models.Contacto) error {
	if r.useOptimization {
		r.mu.Lock()
		defer r.mu.Unlock()
	}
	
	// Verificar si ya existe
	exists, err := r.existsByIDInternal(contacto.ClaveCliente)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("contacto con clave %d ya existe", contacto.ClaveCliente)
	}

	// Agregar al slice
	r.contactos = append(r.contactos, *contacto)
	
	// Actualizar índice si está activo
	if r.useOptimization && r.indiceClaveCliente != nil {
		nuevoContacto := &r.contactos[len(r.contactos)-1]
		r.indiceClaveCliente[contacto.ClaveCliente] = nuevoContacto
	}
	
	// Guardar en Excel
	return r.saveToExcel()
}

// existsByIDInternal verifica existencia sin adquirir mutex (para uso interno)
func (r *ContactoRepository) existsByIDInternal(claveCliente int) (bool, error) {
	if r.useOptimization && r.indiceClaveCliente != nil {
		_, ok := r.indiceClaveCliente[claveCliente]
		return ok, nil
	}
	
	// Búsqueda secuencial para conjuntos pequeños
	for _, contacto := range r.contactos {
		if contacto.ClaveCliente == claveCliente {
			return true, nil
		}
	}
	
	return false, nil
}

// Update actualiza un contacto existente
func (r *ContactoRepository) Update(contacto *models.Contacto) error {
	if r.useOptimization {
		r.mu.Lock()
		defer r.mu.Unlock()
	}
	
	// Buscar índice del contacto
	var encontrado bool
	var indice int
	
	if r.useOptimization && r.indiceClaveCliente != nil {
		if existente, ok := r.indiceClaveCliente[contacto.ClaveCliente]; ok {
			// Actualizar directamente la referencia
			*existente = *contacto
			encontrado = true
		}
	} else {
		// Búsqueda secuencial
		for i, c := range r.contactos {
			if c.ClaveCliente == contacto.ClaveCliente {
				r.contactos[i] = *contacto
				encontrado = true
				indice = i
				break
			}
		}
	}
	
	if !encontrado {
		return fmt.Errorf("contacto con clave %d no encontrado para actualizar", contacto.ClaveCliente)
	}
	
	// Actualizar índice si es necesario
	if r.useOptimization && r.indiceClaveCliente != nil && !encontrado {
		r.indiceClaveCliente[contacto.ClaveCliente] = &r.contactos[indice]
	}
	
	return r.saveToExcel()
}

// Delete elimina un contacto
func (r *ContactoRepository) Delete(claveCliente int) error {
	if r.useOptimization {
		r.mu.Lock()
		defer r.mu.Unlock()
	}
	
	// Buscar el contacto
	encontrado := false
	var indice int
	
	for i, contacto := range r.contactos {
		if contacto.ClaveCliente == claveCliente {
			indice = i
			encontrado = true
			break
		}
	}
	
	if !encontrado {
		return fmt.Errorf("contacto con clave %d no encontrado para eliminar", claveCliente)
	}
	
	// Eliminar del slice (usando la técnica rápida de reemplazo)
	r.contactos[indice] = r.contactos[len(r.contactos)-1]
	r.contactos = r.contactos[:len(r.contactos)-1]
	
	// Actualizar índice si está activo
	if r.useOptimization && r.indiceClaveCliente != nil {
		delete(r.indiceClaveCliente, claveCliente)
	}
	
	return r.saveToExcel()
}

// Search busca contactos basado en criterios
func (r *ContactoRepository) Search(criteria *models.ContactoDTO) ([]models.Contacto, error) {
	if r.useOptimization {
		r.mu.RLock()
		defer r.mu.RUnlock()
	}
	
	// Optimización rápida para búsqueda por clave cliente
	if criteria.ClaveCliente != "" {
		clave, err := strconv.Atoi(criteria.ClaveCliente)
		if err == nil {
			if r.useOptimization && r.indiceClaveCliente != nil {
				// Búsqueda por índice
				if contacto, ok := r.indiceClaveCliente[clave]; ok {
					return []models.Contacto{*contacto}, nil
				}
				return []models.Contacto{}, nil
			} else {
				// Búsqueda secuencial rápida para claves
				for _, contacto := range r.contactos {
					if contacto.ClaveCliente == clave {
						return []models.Contacto{contacto}, nil
					}
				}
				return []models.Contacto{}, nil
			}
		}
	}
	
	// Para conjuntos pequeños, usar búsqueda simple y rápida
	// Este algoritmo es eficiente para menos de ~1000 elementos
	var resultados []models.Contacto
	
	for _, contacto := range r.contactos {
		match := true
		
		// Filtrar por clave cliente
		if criteria.ClaveCliente != "" {
			clave, err := strconv.Atoi(criteria.ClaveCliente)
			if err != nil || contacto.ClaveCliente != clave {
				match = false
			}
		}
		
		// Filtrar por nombre (case insensitive, partial match)
		if criteria.Nombre != "" && !strings.Contains(
			strings.ToLower(contacto.Nombre), 
			strings.ToLower(criteria.Nombre),
		) {
			match = false
		}
		
		// Filtrar por correo (case insensitive, partial match)
		if criteria.Correo != "" && !strings.Contains(
			strings.ToLower(contacto.Correo), 
			strings.ToLower(criteria.Correo),
		) {
			match = false
		}
		
		// Filtrar por teléfono (partial match)
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
	
	return resultados, nil
}

// ExistsByID verifica si existe un contacto con la clave dada
func (r *ContactoRepository) ExistsByID(claveCliente int) (bool, error) {
	if r.useOptimization {
		r.mu.RLock()
		defer r.mu.RUnlock()
	}
	
	return r.existsByIDInternal(claveCliente)
}

// GetLoadErrors retorna los errores de carga del Excel
func (r *ContactoRepository) GetLoadErrors() []models.RowError {
	if r.useOptimization {
		r.mu.RLock()
		defer r.mu.RUnlock()
	}
	
	return r.loadErrors
}

// GetInvalidRowsData retorna los datos completos de filas inválidas
func (r *ContactoRepository) GetInvalidRowsData() []models.RowData {
	if r.useOptimization {
		r.mu.RLock()
		defer r.mu.RUnlock()
	}
	
	return r.invalidRowsData
}

// ReloadExcel recarga el archivo Excel
func (r *ContactoRepository) ReloadExcel() ([]models.RowError, []models.RowData, error) {
	startTime := time.Now()
	fmt.Println("🔄 Recargando Excel...")
	
	loadErrors, invalidData, err := r.loadFromExcel()
	
	if r.useOptimization {
		r.mu.Lock()
		r.loadErrors = loadErrors
		r.invalidRowsData = invalidData
		
		// Reconstruir índices si es necesario
		if len(r.contactos) > 1000 {
			r.buildIndices()
		}
		r.mu.Unlock()
	} else {
		r.loadErrors = loadErrors
		r.invalidRowsData = invalidData
	}
	
	fmt.Printf("✅ Excel recargado en %v\n", time.Since(startTime))
	
	return loadErrors, invalidData, err
}

// loadFromExcel carga datos desde Excel - versión simplificada y rápida
func (r *ContactoRepository) loadFromExcel() ([]models.RowError, []models.RowData, error) {
	file, err := xlsx.OpenFile(r.excelFile)
	if err != nil {
		return nil, nil, fmt.Errorf("error abriendo archivo Excel: %w", err)
	}

	if len(file.Sheets) == 0 {
		return nil, nil, fmt.Errorf("el archivo Excel no tiene hojas")
	}

	sheet := file.Sheets[0]
	contactos := []models.Contacto{} // Lista temporal para construir
	var loadErrors []models.RowError
	var invalidRowsData []models.RowData

	// Usar ForEachRow para procesar filas
	rowIndex := 0
	err = sheet.ForEachRow(func(row *xlsx.Row) error {
		if rowIndex == 0 { // Saltar encabezados
			rowIndex++
			return nil
		}

		currentRow := rowIndex + 1

		// Verificar que la fila tenga al menos 4 celdas
		cellCount := 0
		row.ForEachCell(func(cell *xlsx.Cell) error {
			cellCount++
			return nil
		})

		if cellCount < 4 {
			// Fila incompleta
			hasContent := false
			var partialData []string
			row.ForEachCell(func(cell *xlsx.Cell) error {
				cellValue := strings.TrimSpace(cell.String())
				partialData = append(partialData, cellValue)
				if cellValue != "" {
					hasContent = true
				}
				return nil
			})
			
			if hasContent {
				// Crear RowData para fila incompleta
				rowData := models.RowData{
					HasErrors:  true,
					ErrorCount: 1,
				}
				
				// Asignar datos parciales
				if len(partialData) > 0 { rowData.ClaveCliente = partialData[0] }
				if len(partialData) > 1 { rowData.Nombre = partialData[1] }
				if len(partialData) > 2 { rowData.Correo = partialData[2] }
				if len(partialData) > 3 { rowData.TelefonoContacto = partialData[3] }
				
				invalidRowsData = append(invalidRowsData, rowData)
				
				loadErrors = append(loadErrors, models.RowError{
					Row:     currentRow,
					Column:  "general",
					Field:   "estructura",
					Value:   "",
					Error:   "La fila debe contener exactamente 4 columnas: ClaveCliente, Nombre, Correo, TelefonoContacto",
					RowData: &rowData,
				})
			}
			rowIndex++
			return nil
		}

		// Obtener valores de las celdas
		cells := make([]*xlsx.Cell, 4)
		cellIndex := 0
		row.ForEachCell(func(cell *xlsx.Cell) error {
			if cellIndex < 4 {
				cells[cellIndex] = cell
				cellIndex++
			}
			return nil
		})

		claveStr := strings.TrimSpace(cells[0].String())
		nombre := strings.TrimSpace(cells[1].String())
		correo := strings.TrimSpace(cells[2].String())
		telefono := strings.TrimSpace(cells[3].String())

		// Crear RowData
		rowData := models.RowData{
			ClaveCliente:     claveStr,
			Nombre:           nombre,
			Correo:           correo,
			TelefonoContacto: telefono,
			HasErrors:        false,
			ErrorCount:       0,
		}

		// Validar datos
		rowErrors := []models.RowError{}
		
		if claveStr == "" {
			rowData.AddError()
			rowErrors = append(rowErrors, models.RowError{
				Row:     currentRow,
				Column:  "A",
				Field:   "claveCliente",
				Value:   claveStr,
				Error:   "La clave cliente no puede estar vacía",
				RowData: &rowData,
			})
		}
		
		if nombre == "" {
			rowData.AddError()
			rowErrors = append(rowErrors, models.RowError{
				Row:     currentRow,
				Column:  "B",
				Field:   "nombre",
				Value:   nombre,
				Error:   "El nombre no puede estar vacío",
				RowData: &rowData,
			})
		}
		
		if correo == "" {
			rowData.AddError()
			rowErrors = append(rowErrors, models.RowError{
				Row:     currentRow,
				Column:  "C",
				Field:   "correo",
				Value:   correo,
				Error:   "El correo no puede estar vacío",
				RowData: &rowData,
			})
		}
		
		if telefono == "" {
			rowData.AddError()
			rowErrors = append(rowErrors, models.RowError{
				Row:     currentRow,
				Column:  "D",
				Field:   "telefonoContacto",
				Value:   telefono,
				Error:   "El teléfono no puede estar vacío",
				RowData: &rowData,
			})
		}

		// Validaciones adicionales si hay datos
		if claveStr != "" {
			// Validar formato de clave cliente
			clave, err := strconv.Atoi(claveStr)
			if err != nil {
				rowData.AddError()
				rowErrors = append(rowErrors, models.RowError{
					Row:     currentRow,
					Column:  "A",
					Field:   "claveCliente",
					Value:   claveStr,
					Error:   "La clave cliente debe ser un número entero válido",
					RowData: &rowData,
				})
			} else if clave <= 0 {
				rowData.AddError()
				rowErrors = append(rowErrors, models.RowError{
					Row:     currentRow,
					Column:  "A",
					Field:   "claveCliente",
					Value:   claveStr,
					Error:   "La clave cliente debe ser un número mayor a 0",
					RowData: &rowData,
				})
			} else {
				// Verificar duplicados de clave cliente
				for _, existingContacto := range contactos {
					if existingContacto.ClaveCliente == clave {
						rowData.AddError()
						rowErrors = append(rowErrors, models.RowError{
							Row:     currentRow,
							Column:  "A",
							Field:   "claveCliente",
							Value:   claveStr,
							Error:   fmt.Sprintf("La clave cliente %d ya existe en el archivo", clave),
							RowData: &rowData,
						})
						break
					}
				}
			}
		}

		// Validar teléfono si no está vacío
		if telefono != "" {
			if len(telefono) != 10 {
				rowData.AddError()
				rowErrors = append(rowErrors, models.RowError{
					Row:     currentRow,
					Column:  "D",
					Field:   "telefonoContacto",
					Value:   telefono,
					Error:   "El teléfono debe tener exactamente 10 dígitos",
					RowData: &rowData,
				})
			}

			// Validar que teléfono sean solo números
			for _, char := range telefono {
				if char < '0' || char > '9' {
					rowData.AddError()
					rowErrors = append(rowErrors, models.RowError{
						Row:     currentRow,
						Column:  "D",
						Field:   "telefonoContacto",
						Value:   telefono,
						Error:   "El teléfono debe contener solo números",
						RowData: &rowData,
					})
					break
				}
			}
		}

		// Validar formato básico de correo si no está vacío
		if correo != "" && !strings.Contains(correo, "@") {
			rowData.AddError()
			rowErrors = append(rowErrors, models.RowError{
				Row:     currentRow,
				Column:  "C",
				Field:   "correo",
				Value:   correo,
				Error:   "El correo debe contener @",
				RowData: &rowData,
			})
		}

		// Agregar errores a la lista principal
		loadErrors = append(loadErrors, rowErrors...)

		// Si la fila tiene errores, agregarla a invalidRowsData
		if rowData.HasErrors {
			invalidRowsData = append(invalidRowsData, rowData)
		} else {
			// Solo agregar el contacto si no hay errores
			clave, _ := strconv.Atoi(claveStr) // Ya validamos que sea un int válido
			tempContacto := models.Contacto{
				ClaveCliente:     clave,
				Nombre:           nombre,
				Correo:           correo,
				TelefonoContacto: telefono,
			}
			contactos = append(contactos, tempContacto)
		}

		rowIndex++
		return nil
	})

	if err != nil {
		return loadErrors, invalidRowsData, fmt.Errorf("error iterando filas: %w", err)
	}
	
	// Actualizar lista de contactos
	r.contactos = contactos

	fmt.Printf("✅ Procesadas %d filas del Excel\n", rowIndex-1)
	fmt.Printf("✅ Cargados %d contactos válidos\n", len(contactos))
	fmt.Printf("⚠️ Encontradas %d filas con errores\n", len(invalidRowsData))
	
	return loadErrors, invalidRowsData, nil
}

// saveToExcel guarda los contactos en el archivo Excel - versión simple
func (r *ContactoRepository) saveToExcel() error {
	startTime := time.Now()
	
	file := xlsx.NewFile()
	sheet, err := file.AddSheet("Contactos")
	if err != nil {
		return fmt.Errorf("error creando hoja Excel: %w", err)
	}

	// Agregar encabezados
	headerRow := sheet.AddRow()
	headerRow.AddCell().Value = "ClaveCliente"
	headerRow.AddCell().Value = "Nombre"
	headerRow.AddCell().Value = "Correo"
	headerRow.AddCell().Value = "TelefonoContacto"

	// Agregar datos - versión simple y directa
	for _, contacto := range r.contactos {
		row := sheet.AddRow()
		row.AddCell().Value = strconv.Itoa(contacto.ClaveCliente)
		row.AddCell().Value = contacto.Nombre
		row.AddCell().Value = contacto.Correo
		row.AddCell().Value = contacto.TelefonoContacto
	}

	if err := file.Save(r.excelFile); err != nil {
		return fmt.Errorf("error guardando archivo Excel: %w", err)
	}

	fmt.Printf("✅ Guardados %d contactos en Excel en %v\n", len(r.contactos), time.Since(startTime))
	return nil
}