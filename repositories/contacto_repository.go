// repositories/contacto_repository.go
package repositories

import (
	"fmt"
	"strconv"
	"strings"

	"contactos-api/models"

	"github.com/tealeg/xlsx/v3"
)

// ContactoRepository define la interfaz para el repositorio de contactos
type ContactoRepositoryInterface interface {
	GetAll() ([]models.Contacto, error)
	GetByID(claveCliente int) (*models.Contacto, error)
	Create(contacto *models.Contacto) error
	Update(contacto *models.Contacto) error
	Delete(claveCliente int) error
	Search(criteria *models.ContactoDTO) ([]models.Contacto, error)
	ExistsByID(claveCliente int) (bool, error)
}

// ContactoRepository implementa el acceso a datos para contactos
type ContactoRepository struct {
	excelFile string
	contactos []models.Contacto
}

// NewContactoRepository crea una nueva instancia del repositorio
func NewContactoRepository(excelFile string) *ContactoRepository {
	repo := &ContactoRepository{
		excelFile: excelFile,
		contactos: []models.Contacto{},
	}
	
	// Cargar datos al inicializar
	if err := repo.loadFromExcel(); err != nil {
		fmt.Printf("⚠️  Error cargando Excel: %v. Iniciando con datos vacíos.\n", err)
	}
	
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

// loadFromExcel carga los contactos desde el archivo Excel
func (r *ContactoRepository) loadFromExcel() error {
	file, err := xlsx.OpenFile(r.excelFile)
	if err != nil {
		return fmt.Errorf("error abriendo archivo Excel: %w", err)
	}

	if len(file.Sheets) == 0 {
		return fmt.Errorf("el archivo Excel no tiene hojas")
	}

	sheet := file.Sheets[0]
	r.contactos = []models.Contacto{}

	// Usar ForEachRow para xlsx/v3
	rowIndex := 0
	err = sheet.ForEachRow(func(row *xlsx.Row) error {
		if rowIndex == 0 { // Saltar encabezados
			rowIndex++
			return nil
		}

		// Verificar que la fila tenga al menos 4 celdas
		cellCount := 0
		row.ForEachCell(func(cell *xlsx.Cell) error {
			cellCount++
			return nil
		})

		if cellCount < 4 {
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

		claveStr := cells[0].String()
		nombre := cells[1].String()
		correo := cells[2].String()
		telefono := cells[3].String()

		// Validar que no estén vacíos
		if claveStr == "" || nombre == "" || correo == "" || telefono == "" {
			rowIndex++
			return nil
		}

		clave, err := strconv.Atoi(claveStr)
		if err != nil {
			fmt.Printf("⚠️  Fila %d: clave inválida '%s', saltando\n", rowIndex+1, claveStr)
			rowIndex++
			return nil
		}

		contacto := models.Contacto{
			ClaveCliente:     clave,
			Nombre:           nombre,
			Correo:           correo,
			TelefonoContacto: telefono,
		}

		r.contactos = append(r.contactos, contacto)
		rowIndex++
		return nil
	})

	if err != nil {
		return fmt.Errorf("error iterando filas: %w", err)
	}

	fmt.Printf("✅ Cargados %d contactos desde Excel\n", len(r.contactos))
	return nil
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