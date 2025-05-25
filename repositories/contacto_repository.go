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
	ReloadExcel() ([]models.RowError, error)
}

// ContactoRepository implementa el acceso a datos para contactos
type ContactoRepository struct {
	excelFile   string
	contactos   []models.Contacto
	loadErrors  []models.RowError
}

// NewContactoRepository crea una nueva instancia del repositorio
func NewContactoRepository(excelFile string) *ContactoRepository {
	repo := &ContactoRepository{
		excelFile:   excelFile,
		contactos:   []models.Contacto{},
		loadErrors:  []models.RowError{},
	}
	
	// Cargar datos al inicializar
	loadErrors, err := repo.loadFromExcel()
	if err != nil {
		fmt.Printf("⚠️  Error cargando Excel: %v. Iniciando con datos vacíos.\n", err)
	}
	
	repo.loadErrors = loadErrors
	
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

// ReloadExcel recarga el archivo Excel y retorna los errores encontrados
func (r *ContactoRepository) ReloadExcel() ([]models.RowError, error) {
	loadErrors, err := r.loadFromExcel()
	r.loadErrors = loadErrors
	return loadErrors, err
}

// loadFromExcel carga los contactos desde el archivo Excel con validación básica
func (r *ContactoRepository) loadFromExcel() ([]models.RowError, error) {
	file, err := xlsx.OpenFile(r.excelFile)
	if err != nil {
		return nil, fmt.Errorf("error abriendo archivo Excel: %w", err)
	}

	if len(file.Sheets) == 0 {
		return nil, fmt.Errorf("el archivo Excel no tiene hojas")
	}

	sheet := file.Sheets[0]
	r.contactos = []models.Contacto{}
	var loadErrors []models.RowError

	// Usar ForEachRow para xlsx/v3
	rowIndex := 0
	err = sheet.ForEachRow(func(row *xlsx.Row) error {
		if rowIndex == 0 { // Saltar encabezados
			rowIndex++
			return nil
		}

		currentRow := rowIndex + 1 // +1 porque empezamos desde 0 y queremos mostrar número real de fila

		// Verificar que la fila tenga al menos 4 celdas
		cellCount := 0
		row.ForEachCell(func(cell *xlsx.Cell) error {
			cellCount++
			return nil
		})

		if cellCount < 4 {
			// Si la fila tiene contenido pero menos de 4 columnas
			hasContent := false
			row.ForEachCell(func(cell *xlsx.Cell) error {
				if strings.TrimSpace(cell.String()) != "" {
					hasContent = true
				}
				return nil
			})
			
			if hasContent {
				loadErrors = append(loadErrors, models.RowError{
					Row:    currentRow,
					Column: "general",
					Field:  "estructura",
					Value:  "",
					Error:  "La fila debe contener exactamente 4 columnas: ClaveCliente, Nombre, Correo, TelefonoContacto",
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

		// Validar que no estén vacíos
		rowErrors := []models.RowError{}
		
		if claveStr == "" {
			rowErrors = append(rowErrors, models.RowError{
				Row:    currentRow,
				Column: "A",
				Field:  "claveCliente",
				Value:  claveStr,
				Error:  "La clave cliente no puede estar vacía",
			})
		}
		
		if nombre == "" {
			rowErrors = append(rowErrors, models.RowError{
				Row:    currentRow,
				Column: "B",
				Field:  "nombre",
				Value:  nombre,
				Error:  "El nombre no puede estar vacío",
			})
		}
		
		if correo == "" {
			rowErrors = append(rowErrors, models.RowError{
				Row:    currentRow,
				Column: "C",
				Field:  "correo",
				Value:  correo,
				Error:  "El correo no puede estar vacío",
			})
		}
		
		if telefono == "" {
			rowErrors = append(rowErrors, models.RowError{
				Row:    currentRow,
				Column: "D",
				Field:  "telefonoContacto",
				Value:  telefono,
				Error:  "El teléfono no puede estar vacío",
			})
		}

		// Si hay campos vacíos, agregar errores y continuar
		if len(rowErrors) > 0 {
			loadErrors = append(loadErrors, rowErrors...)
			rowIndex++
			return nil
		}

		// Validar formato de clave cliente
		clave, err := strconv.Atoi(claveStr)
		if err != nil {
			loadErrors = append(loadErrors, models.RowError{
				Row:    currentRow,
				Column: "A",
				Field:  "claveCliente",
				Value:  claveStr,
				Error:  "La clave cliente debe ser un número entero válido",
			})
			rowIndex++
			return nil
		}

		// Validaciones básicas sin usar el validador externo
		if clave <= 0 {
			loadErrors = append(loadErrors, models.RowError{
				Row:    currentRow,
				Column: "A",
				Field:  "claveCliente",
				Value:  claveStr,
				Error:  "La clave cliente debe ser un número mayor a 0",
			})
		}

		// Validar teléfono (10 dígitos)
		if len(telefono) != 10 {
			loadErrors = append(loadErrors, models.RowError{
				Row:    currentRow,
				Column: "D",
				Field:  "telefonoContacto",
				Value:  telefono,
				Error:  "El teléfono debe tener exactamente 10 dígitos",
			})
		}

		// Validar que teléfono sean solo números
		for _, char := range telefono {
			if char < '0' || char > '9' {
				loadErrors = append(loadErrors, models.RowError{
					Row:    currentRow,
					Column: "D",
					Field:  "telefonoContacto",
					Value:  telefono,
					Error:  "El teléfono debe contener solo números",
				})
				break
			}
		}

		// Validar formato básico de correo
		if !strings.Contains(correo, "@") {
			loadErrors = append(loadErrors, models.RowError{
				Row:    currentRow,
				Column: "C",
				Field:  "correo",
				Value:  correo,
				Error:  "El correo debe contener @",
			})
		}

		// Verificar duplicados de clave cliente
		for _, existingContacto := range r.contactos {
			if existingContacto.ClaveCliente == clave {
				loadErrors = append(loadErrors, models.RowError{
					Row:    currentRow,
					Column: "A",
					Field:  "claveCliente",
					Value:  claveStr,
					Error:  fmt.Sprintf("La clave cliente %d ya existe en el archivo", clave),
				})
				break
			}
		}

		// Si no hay errores críticos, agregar el contacto
		tempContacto := models.Contacto{
			ClaveCliente:     clave,
			Nombre:           nombre,
			Correo:           correo,
			TelefonoContacto: telefono,
		}

		// Solo agregar si no hay errores en esta fila
		hasErrors := false
		for _, err := range loadErrors {
			if err.Row == currentRow {
				hasErrors = true
				break
			}
		}

		if !hasErrors {
			r.contactos = append(r.contactos, tempContacto)
		}

		rowIndex++
		return nil
	})

	if err != nil {
		return loadErrors, fmt.Errorf("error iterando filas: %w", err)
	}

	fmt.Printf("✅ Procesadas %d filas del Excel\n", rowIndex-1)
	fmt.Printf("✅ Cargados %d contactos válidos\n", len(r.contactos))
	if len(loadErrors) > 0 {
		fmt.Printf("⚠️  Se encontraron %d errores de validación\n", len(loadErrors))
	}

	return loadErrors, nil
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