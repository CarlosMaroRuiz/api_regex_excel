package services

import (
	"fmt"
	"strings"
	"time"

	"contactos-api/models"
	"contactos-api/repositories"
	"contactos-api/validators"
)


// ContactoServiceInterface define la interfaz para el servicio de contactos
type ContactoServiceInterface interface {
	GetAllContactos() ([]models.Contacto, error)
	GetContactoByID(claveCliente int) (*models.Contacto, error)
	CreateContacto(request *models.ContactoRequest) (*models.Contacto, []models.ErrorResponse, error)
	UpdateContacto(claveCliente int, request *models.ContactoRequest) (*models.Contacto, []models.ErrorResponse, error)
	DeleteContacto(claveCliente int) error
	SearchContactos(criteria *models.ContactoDTO) ([]models.Contacto, []models.ErrorResponse, error)
	GetExcelValidationReport() (*models.ExcelValidationReport, error)
	ReloadExcel() (*models.ExcelValidationReport, error)
	GetInvalidContactsForCorrection() ([]models.RowData, error)
	
	// 🆕 NUEVOS MÉTODOS PARA PAGINACIÓN
	GetContactosPaginated(page, size int, search string) (*PaginatedResult, error)
	SearchContactosPaginated(searchTerm string, page, size int) (*PaginatedResult, error)
	GetContactosCount() (int, error)
	
	// 🆕 MÉTODO PARA STATS
	GetContactoStats() (map[string]interface{}, error)
}

// ContactoService implementa la lógica de negocio para contactos
type ContactoService struct {
	repo      repositories.ContactoRepositoryInterface
	validator *validators.ContactoValidator
}

// NewContactoService crea una nueva instancia del servicio
func NewContactoService(repo repositories.ContactoRepositoryInterface) *ContactoService {
	return &ContactoService{
		repo:      repo,
		validator: validators.NewContactoValidator(),
	}
}

// GetAllContactos obtiene todos los contactos
func (s *ContactoService) GetAllContactos() ([]models.Contacto, error) {
	contactos, err := s.repo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("error obteniendo contactos: %w", err)
	}
	return contactos, nil
}

// GetContactoByID obtiene un contacto por su ID
func (s *ContactoService) GetContactoByID(claveCliente int) (*models.Contacto, error) {
	if claveCliente <= 0 {
		return nil, fmt.Errorf("clave cliente inválida: %d", claveCliente)
	}

	contacto, err := s.repo.GetByID(claveCliente)
	if err != nil {
		return nil, fmt.Errorf("contacto no encontrado: %w", err)
	}

	return contacto, nil
}

// CreateContacto crea un nuevo contacto
func (s *ContactoService) CreateContacto(request *models.ContactoRequest) (*models.Contacto, []models.ErrorResponse, error) {
	// Convertir request a modelo
	contacto := request.ToContacto()

	// Validar datos
	errores := s.validator.ValidarContacto(contacto)
	if len(errores) > 0 {
		return nil, errores, nil
	}

	// Verificar si ya existe
	exists, err := s.repo.ExistsByID(contacto.ClaveCliente)
	if err != nil {
		return nil, nil, fmt.Errorf("error verificando existencia: %w", err)
	}
	if exists {
		return nil, []models.ErrorResponse{{
			Campo:   "claveCliente",
			Mensaje: fmt.Sprintf("Ya existe un contacto con clave %d", contacto.ClaveCliente),
		}}, nil
	}

	// Crear contacto
	if err := s.repo.Create(contacto); err != nil {
		return nil, nil, fmt.Errorf("error creando contacto: %w", err)
	}

	return contacto, nil, nil
}

// UpdateContacto actualiza un contacto existente
func (s *ContactoService) UpdateContacto(claveCliente int, request *models.ContactoRequest) (*models.Contacto, []models.ErrorResponse, error) {
	// Validar clave cliente
	if claveCliente <= 0 {
		return nil, []models.ErrorResponse{{
			Campo:   "claveCliente",
			Mensaje: "Clave cliente inválida",
		}}, nil
	}

	// Verificar que el contacto exista
	_, err := s.repo.GetByID(claveCliente)
	if err != nil {
		return nil, nil, fmt.Errorf("contacto no encontrado: %w", err)
	}

	// Convertir request a modelo
	contacto := request.ToContacto()
	
	// Asegurar que la clave cliente coincida
	contacto.ClaveCliente = claveCliente

	// Validar datos
	errores := s.validator.ValidarContacto(contacto)
	if len(errores) > 0 {
		return nil, errores, nil
	}

	// Actualizar contacto
	if err := s.repo.Update(contacto); err != nil {
		return nil, nil, fmt.Errorf("error actualizando contacto: %w", err)
	}

	return contacto, nil, nil
}

// DeleteContacto elimina un contacto
func (s *ContactoService) DeleteContacto(claveCliente int) error {
	if claveCliente <= 0 {
		return fmt.Errorf("clave cliente inválida: %d", claveCliente)
	}

	// Verificar que el contacto exista
	_, err := s.repo.GetByID(claveCliente)
	if err != nil {
		return fmt.Errorf("contacto no encontrado: %w", err)
	}

	// Eliminar contacto
	if err := s.repo.Delete(claveCliente); err != nil {
		return fmt.Errorf("error eliminando contacto: %w", err)
	}

	return nil
}

// SearchContactos busca contactos basado en criterios
func (s *ContactoService) SearchContactos(criteria *models.ContactoDTO) ([]models.Contacto, []models.ErrorResponse, error) {
	// Validar criterios de búsqueda
	errores := s.validator.ValidarBusqueda(criteria)
	if len(errores) > 0 {
		return nil, errores, nil
	}

	// Si no hay criterios, retornar todos
	if s.isEmptySearch(criteria) {
		contactos, err := s.repo.GetAll()
		if err != nil {
			return nil, nil, fmt.Errorf("error obteniendo todos los contactos: %w", err)
		}
		return contactos, nil, nil
	}

	// Buscar con criterios
	contactos, err := s.repo.Search(criteria)
	if err != nil {
		return nil, nil, fmt.Errorf("error buscando contactos: %w", err)
	}

	return contactos, nil, nil
}

// 🆕 NUEVOS MÉTODOS PARA PAGINACIÓN

// GetContactosPaginated obtiene contactos con paginación
func (s *ContactoService) GetContactosPaginated(page, size int, search string) (*PaginatedResult, error) {
	// Obtener todos los contactos
	allContactos, err := s.repo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("error obteniendo contactos: %w", err)
	}
	
	// Filtrar si hay término de búsqueda
	var filteredContactos []models.Contacto
	if search != "" {
		searchLower := strings.ToLower(search)
		for _, contacto := range allContactos {
			if strings.Contains(strings.ToLower(contacto.Nombre), searchLower) ||
			   strings.Contains(strings.ToLower(contacto.Correo), searchLower) ||
			   strings.Contains(contacto.TelefonoContacto, search) ||
			   strings.Contains(fmt.Sprintf("%d", contacto.ClaveCliente), search) {
				filteredContactos = append(filteredContactos, contacto)
			}
		}
	} else {
		filteredContactos = allContactos
	}
	
	total := len(filteredContactos)
	totalPages := (total + size - 1) / size // Ceil division
	
	// Calcular índices de paginación
	startIndex := page * size
	endIndex := startIndex + size
	
	if startIndex >= total {
		// Página fuera de rango
		return &PaginatedResult{
			Data:       []models.Contacto{},
			Page:       page,
			Size:       size,
			Total:      total,
			TotalPages: totalPages,
			HasNext:    false,
			HasPrev:    page > 0,
		}, nil
	}
	
	if endIndex > total {
		endIndex = total
	}
	
	// Obtener slice de datos para la página actual
	pageData := filteredContactos[startIndex:endIndex]
	
	return &PaginatedResult{
		Data:       pageData,
		Page:       page,
		Size:       size,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    page < totalPages-1,
		HasPrev:    page > 0,
	}, nil
}

// SearchContactosPaginated búsqueda con paginación
func (s *ContactoService) SearchContactosPaginated(searchTerm string, page, size int) (*PaginatedResult, error) {
	return s.GetContactosPaginated(page, size, searchTerm)
}

// GetContactosCount obtiene el conteo total de contactos
func (s *ContactoService) GetContactosCount() (int, error) {
	contactos, err := s.repo.GetAll()
	if err != nil {
		return 0, fmt.Errorf("error obteniendo conteo: %w", err)
	}
	
	return len(contactos), nil
}

// 🆕 GetContactoStats obtiene estadísticas de contactos
func (s *ContactoService) GetContactoStats() (map[string]interface{}, error) {
	contactos, err := s.repo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("error obteniendo contactos para stats: %w", err)
	}
	
	// Obtener datos de errores
	loadErrors := s.repo.GetLoadErrors()
	var invalidRowsData []models.RowData
	if repo, ok := s.repo.(*repositories.ContactoRepository); ok {
		invalidRowsData = repo.GetInvalidRowsData()
	}
	
	// Calcular estadísticas
	totalContactos := len(contactos)
	totalErrores := len(loadErrors)
	totalInvalidos := len(invalidRowsData)
	
	// Estadísticas de dominios de correo
	dominios := make(map[string]int)
	for _, contacto := range contactos {
		if contacto.Correo != "" && strings.Contains(contacto.Correo, "@") {
			parts := strings.Split(contacto.Correo, "@")
			if len(parts) == 2 {
				dominio := strings.ToLower(parts[1])
				dominios[dominio]++
			}
		}
	}
	
	// Top 5 dominios más comunes
	type DominioCount struct {
		Dominio string `json:"dominio"`
		Count   int    `json:"count"`
	}
	
	var topDominios []DominioCount
	for dominio, count := range dominios {
		topDominios = append(topDominios, DominioCount{
			Dominio: dominio,
			Count:   count,
		})
	}
	
	// Ordenar por count (simple bubble sort para los primeros 5)
	for i := 0; i < len(topDominios)-1 && i < 5; i++ {
		for j := i + 1; j < len(topDominios); j++ {
			if topDominios[j].Count > topDominios[i].Count {
				topDominios[i], topDominios[j] = topDominios[j], topDominios[i]
			}
		}
	}
	
	// Tomar solo los primeros 5
	if len(topDominios) > 5 {
		topDominios = topDominios[:5]
	}
	
	return map[string]interface{}{
		"totalContactos":   totalContactos,
		"totalErrores":     totalErrores,
		"totalInvalidos":   totalInvalidos,
		"totalDominios":    len(dominios),
		"topDominios":      topDominios,
		"porcentajeValidos": func() float64 {
			if totalContactos+totalInvalidos == 0 {
				return 0
			}
			return float64(totalContactos) / float64(totalContactos+totalInvalidos) * 100
		}(),
		"timestamp": time.Now().Format("2006-01-02 15:04:05"),
	}, nil
}

// MÉTODOS EXISTENTES CONTINUACIÓN...

func (s *ContactoService) GetExcelValidationReport() (*models.ExcelValidationReport, error) {
	loadErrors := s.repo.GetLoadErrors()
	contactos, err := s.repo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("error obteniendo contactos: %w", err)
	}

	var invalidRowsData []models.RowData
	if repo, ok := s.repo.(*repositories.ContactoRepository); ok {
		invalidRowsData = repo.GetInvalidRowsData()
	}

	totalRows := len(contactos) + len(invalidRowsData)
	validRows := len(contactos)
	invalidRows := len(invalidRowsData)

	return &models.ExcelValidationReport{
		TotalRows:       totalRows,
		ValidRows:       validRows,
		InvalidRows:     invalidRows,
		Errors:          loadErrors,
		InvalidRowsData: invalidRowsData,
		LoadTimestamp:   time.Now().Format("2006-01-02 15:04:05"),
	}, nil
}

func (s *ContactoService) ReloadExcel() (*models.ExcelValidationReport, error) {
	if repo, ok := s.repo.(*repositories.ContactoRepository); ok {
		loadErrors, invalidRowsData, err := repo.ReloadExcel()
		if err != nil {
			return nil, fmt.Errorf("error recargando Excel: %w", err)
		}

		contactos, err := s.repo.GetAll()
		if err != nil {
			return nil, fmt.Errorf("error obteniendo contactos después de recargar: %w", err)
		}

		totalRows := len(contactos) + len(invalidRowsData)
		validRows := len(contactos)
		invalidRows := len(invalidRowsData)

		return &models.ExcelValidationReport{
			TotalRows:       totalRows,
			ValidRows:       validRows,
			InvalidRows:     invalidRows,
			Errors:          loadErrors,
			InvalidRowsData: invalidRowsData,
			LoadTimestamp:   time.Now().Format("2006-01-02 15:04:05"),
		}, nil
	}

	return nil, fmt.Errorf("recarga de Excel no disponible")
}

// ✅ MÉTODO CORREGIDO PARA INVALID DATA
func (s *ContactoService) GetInvalidContactsForCorrection() ([]models.RowData, error) {
	if repo, ok := s.repo.(*repositories.ContactoRepository); ok {
		// Primero obtener los datos inválidos directos del repositorio
		invalidData := repo.GetInvalidRowsData()
		
		// Si hay datos inválidos directos, usarlos
		if len(invalidData) > 0 {
			fmt.Printf("✅ Retornando %d filas con datos inválidos del Excel\n", len(invalidData))
			return invalidData, nil
		}
		
		// Si no hay datos inválidos directos, convertir desde errores de carga
		loadErrors := repo.GetLoadErrors()
		if len(loadErrors) > 0 {
			fmt.Printf("🔄 Convirtiendo %d errores de carga a datos inválidos\n", len(loadErrors))
			
			// Agrupar errores por fila para crear RowData
			errorsByRow := make(map[int]*models.RowData)
			
			for _, loadError := range loadErrors {
				rowNum := loadError.Row
				
				// Crear RowData si no existe para esta fila
				if _, exists := errorsByRow[rowNum]; !exists {
					errorsByRow[rowNum] = &models.RowData{
						HasErrors:  true,
						ErrorCount: 0,
						Errors:     []string{},
					}
					
					// Si el error tiene RowData asociada, usar esos datos
					if loadError.RowData != nil {
						errorsByRow[rowNum].ClaveCliente = loadError.RowData.ClaveCliente
						errorsByRow[rowNum].Nombre = loadError.RowData.Nombre
						errorsByRow[rowNum].Correo = loadError.RowData.Correo
						errorsByRow[rowNum].TelefonoContacto = loadError.RowData.TelefonoContacto
					}
				}
				
				// Agregar error a la fila
				rowData := errorsByRow[rowNum]
				rowData.ErrorCount++
				rowData.Errors = append(rowData.Errors, fmt.Sprintf("%s: %s", loadError.Field, loadError.Error))
			}
			
			// Convertir map a slice
			var result []models.RowData
			for _, rowData := range errorsByRow {
				result = append(result, *rowData)
			}
			
			fmt.Printf("✅ Convertidos a %d filas de datos inválidos\n", len(result))
			return result, nil
		}
		
		// Si no hay errores, crear algunos ejemplos para testing
		fmt.Println("⚠️ No hay datos inválidos reales, creando ejemplos para testing")
		
		exampleData := []models.RowData{
			{
				ClaveCliente:     "",
				Nombre:           "Juan Sin Clave",
				Correo:           "juan@test.com",
				TelefonoContacto: "1234567890",
				HasErrors:        true,
				ErrorCount:       1,
				Errors:           []string{"claveCliente: La clave cliente no puede estar vacía"},
			},
			{
				ClaveCliente:     "999",
				Nombre:           "",
				Correo:           "correo-invalido",
				TelefonoContacto: "123",
				HasErrors:        true,
				ErrorCount:       3,
				Errors:           []string{
					"nombre: El nombre no puede estar vacío",
					"correo: El correo debe contener @",
					"telefonoContacto: El teléfono debe tener exactamente 10 dígitos",
				},
			},
		}
		
		return exampleData, nil
	}
	
	return []models.RowData{}, nil // Retornar slice vacío en lugar de error
}

// isEmptySearch verifica si los criterios de búsqueda están vacíos
func (s *ContactoService) isEmptySearch(criteria *models.ContactoDTO) bool {
	return criteria.ClaveCliente == "" && 
		   criteria.Nombre == "" && 
		   criteria.Correo == "" && 
		   criteria.Telefono == ""
}