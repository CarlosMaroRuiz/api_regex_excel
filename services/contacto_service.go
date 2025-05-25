package services

import (
	"fmt"
	"strings"

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

// isEmptySearch verifica si los criterios de búsqueda están vacíos
func (s *ContactoService) isEmptySearch(criteria *models.ContactoDTO) bool {
	return criteria.ClaveCliente == "" && 
		   criteria.Nombre == "" && 
		   criteria.Correo == "" && 
		   criteria.Telefono == ""
}

// GetContactoStats obtiene estadísticas de contactos
func (s *ContactoService) GetContactoStats() (map[string]interface{}, error) {
	contactos, err := s.repo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("error obteniendo contactos para estadísticas: %w", err)
	}

	stats := map[string]interface{}{
		"total_contactos": len(contactos),
		"dominios":        s.getDominioStats(contactos),
	}

	return stats, nil
}

// getDominioStats obtiene estadísticas por dominio de correo
func (s *ContactoService) getDominioStats(contactos []models.Contacto) map[string]int {
	dominios := make(map[string]int)
	
	for _, contacto := range contactos {
		// Extraer dominio del correo
		parts := strings.Split(contacto.Correo, "@")
		if len(parts) == 2 {
			dominio := parts[1]
			dominios[dominio]++
		}
	}
	
	return dominios
}