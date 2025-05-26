// repositories/contacto_repository.go
package repositories

import (
	"fmt"
	"strconv"
	"strings"

	"contactos-api/models"

	"github.com/tealeg/xlsx/v3"
)

// ContactoRepositoryInterface define la interfaz para el repositorio de contactos
type ContactoRepositoryInterface interface {
	GetAll() ([]models.Contacto, error)
	GetByID(claveCliente int) (*models.Contacto, error)
	Create(contacto *models.Contacto) error
	Update(contacto *models.Contacto) error
	Delete(claveCliente int) error
	Search(criteria *models.ContactoDTO) ([]models.Contacto, error)
	ExistsByID(claveCliente int) (bool, error)
	GetLoadErrors() []models.RowError
	// 🆕 NUEVO: Obtener datos de filas inválidas para corrección
	GetInvalidRowsData() []models.RowData
	ReloadExcel() ([]models.RowError, []models.RowData, error)
}

// ContactoRepository implementa el acceso a datos para contactos
type ContactoRepository struct {
	excelFile        string
	contactos        []models.Contacto
	loadErrors       []models.RowError
	// 🆕 NUEVO: Almacenar datos completos de filas inválidas
	invalidRowsData  []models.RowData
}

// NewContactoRepository crea una nueva instancia del repositorio
func NewContactoRepository(excelFile string) *ContactoRepository {
	repo := &ContactoRepository{
		excelFile:       excelFile,
		contactos:       []models.Contacto{},
		loadErrors:      []models.RowError{},
		invalidRowsData: []models.RowData{},
	}
	
	// Cargar datos al inicializar
	loadErrors, invalidData, err := repo.loadFromExcel()
	if err != nil {
		fmt.Printf("⚠️  Error cargando Excel: %v. Iniciando con datos vacíos.\n", err)
	}
	
	repo.loadErrors = loadErrors
	repo.invalidRowsData = invalidData
	
	return repo
}

// GetAll retorna todos los contactos
func (r *ContactoRepository) GetAll() ([]models.Contacto, error) {
	return r.contactos, nil
}

// GetByID busca un contacto por su clave cliente
func (r *ContactoRepository) GetByID(claveCliente int) (*models.Contacto, error) {
	for i, contacto := range r.contactos {
		if contacto.ClaveCliente == claveCliente {
			return &r.contactos[i], nil
		}
	}
	return nil, fmt.Errorf("contacto con clave %d no encontrado", claveCliente)
}

// Create crea un nuevo contacto
func (r *ContactoRepository) Create(contacto *models.Contacto) error {
	// Verificar si ya existe
	exists, err := r.ExistsByID(contacto.ClaveCliente)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("contacto con clave %d ya existe", contacto.ClaveCliente)
	}

	// Agregar al slice
	r.contactos = append(r.contactos, *contacto)
	
	// Guardar en Excel
	return r.saveToExcel()
}

// Update actualiza un contacto existente
func (r *ContactoRepository) Update(contacto *models.Contacto) error {
	for i, c := range r.contactos {
		if c.ClaveCliente == contacto.ClaveCliente {
			r.contactos[i] = *contacto
			return r.saveToExcel()
		}
	}
	return fmt.Errorf("contacto con clave %d no encontrado para actualizar", contacto.ClaveCliente)
}

// Delete elimina un contacto
func (r *ContactoRepository) Delete(claveCliente int) error {
	for i, contacto := range r.contactos {
		if contacto.ClaveCliente == claveCliente {
			// Eliminar del slice
			r.contactos = append(r.contactos[:i], r.contactos[i+1:]...)
			return r.saveToExcel()
		}
	}
	return fmt.Errorf("contacto con clave %d no encontrado para eliminar", claveCliente)
}

// Search busca contactos basado en criterios
func (r *ContactoRepository) Search(criteria *models.ContactoDTO) ([]models.Contacto, error) {
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
	_, err := r.GetByID(claveCliente)
	if err != nil {
		return false, nil
	}
	return true, nil
}

// GetLoadErrors retorna los errores de carga del Excel
func (r *ContactoRepository) GetLoadErrors() []models.RowError {
	return r.loadErrors
}

// 🆕 NUEVO: GetInvalidRowsData retorna los datos completos de filas inválidas
func (r *ContactoRepository) GetInvalidRowsData() []models.RowData {
	return r.invalidRowsData
}

// 🆕 MEJORADO: ReloadExcel ahora retorna también los datos inválidos
func (r *ContactoRepository) ReloadExcel() ([]models.RowError, []models.RowData, error) {
	loadErrors, invalidData, err := r.loadFromExcel()
	r.loadErrors = loadErrors
	r.invalidRowsData = invalidData
	return loadErrors, invalidData, err
}

// 🆕 MEJORADO: loadFromExcel ahora captura datos completos de filas inválidas
func (r *ContactoRepository) loadFromExcel() ([]models.RowError, []models.RowData, error) {
	file, err := xlsx.OpenFile(r.excelFile)
	if err != nil {
		return nil, nil, fmt.Errorf("error abriendo archivo Excel: %w", err)
	}

	if len(file.Sheets) == 0 {
		return nil, nil, fmt.Errorf("el archivo Excel no tiene hojas")
	}

	sheet := file.Sheets[0]
	r.contactos = []models.Contacto{}
	var loadErrors []models.RowError
	var invalidRowsData []models.RowData

	// Usar ForEachRow para xlsx/v3
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
			// Si la fila tiene contenido pero menos de 4 columnas
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
				// 🆕 NUEVO: Crear RowData para fila incompleta
				rowData := models.RowData{
					HasErrors:  true,
					ErrorCount: 1,
				}
				
				// Asignar los datos parciales que tenemos
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

		// 🆕 NUEVO: Crear RowData con todos los datos de la fila
		rowData := models.RowData{
			ClaveCliente:     claveStr,
			Nombre:           nombre,
			Correo:           correo,
			TelefonoContacto: telefono,
			HasErrors:        false,
			ErrorCount:       0,
		}

		// Validar que no estén vacíos
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
				for _, existingContacto := range r.contactos {
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

		// 🆕 NUEVO: Si la fila tiene errores, agregarla a invalidRowsData
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
			r.contactos = append(r.contactos, tempContacto)
		}

		rowIndex++
		return nil
	})

	if err != nil {
		return loadErrors, invalidRowsData, fmt.Errorf("error iterando filas: %w", err)
	}

	fmt.Printf("✅ Procesadas %d filas del Excel\n", rowIndex-1)
	fmt.Printf("✅ Cargados %d contactos válidos\n", len(r.contactos))
	fmt.Printf("⚠️  Encontradas %d filas con errores\n", len(invalidRowsData))
	if len(loadErrors) > 0 {
		fmt.Printf("⚠️  Se encontraron %d errores de validación\n", len(loadErrors))
	}

	return loadErrors, invalidRowsData, nil
}

// saveToExcel guarda los contactos en el archivo Excel
func (r *ContactoRepository) saveToExcel() error {
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

	// Agregar datos
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

	fmt.Printf("✅ Guardados %d contactos en Excel\n", len(r.contactos))
	return nil
}