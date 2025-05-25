// validators/contacto_validator.go
package validators

import (
	"regexp"
	"strings"
	"contactos-api/models"
)

// ContactoValidator maneja las validaciones de contactos
type ContactoValidator struct {
	telefonoRegex *regexp.Regexp
	correoRegex   *regexp.Regexp
	nombreRegex   *regexp.Regexp
}

// NewContactoValidator crea una nueva instancia del validador
func NewContactoValidator() *ContactoValidator {
	return &ContactoValidator{
		telefonoRegex: regexp.MustCompile(`^\d{10}$`),
		correoRegex:   regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@(gmail\.com|yahoo\.com|hotmail\.com|outlook\.com|live\.com|icloud\.com|protonmail\.com)$`),
		nombreRegex:   regexp.MustCompile(`^[a-zA-ZáéíóúÁÉÍÓÚñÑ\s]+$`),
	}
}

// ValidarContacto valida un contacto completo
func (v *ContactoValidator) ValidarContacto(contacto *models.Contacto) []models.ErrorResponse {
	var errores []models.ErrorResponse

	// Validar teléfono
	if err := v.ValidarTelefono(contacto.TelefonoContacto); err != nil {
		errores = append(errores, *err)
	}

	// Validar correo
	if err := v.ValidarCorreo(contacto.Correo); err != nil {
		errores = append(errores, *err)
	}

	// Validar nombre
	if err := v.ValidarNombre(contacto.Nombre); err != nil {
		errores = append(errores, *err)
	}

	// Validar clave cliente
	if err := v.ValidarClaveCliente(contacto.ClaveCliente); err != nil {
		errores = append(errores, *err)
	}

	return errores
}

// ValidarTelefono valida el formato del teléfono
func (v *ContactoValidator) ValidarTelefono(telefono string) *models.ErrorResponse {
	if !v.telefonoRegex.MatchString(telefono) {
		return &models.ErrorResponse{
			Campo:   "telefonoContacto",
			Mensaje: "El teléfono debe tener exactamente 10 dígitos sin letras",
		}
	}
	return nil
}

// ValidarCorreo valida el formato del correo
func (v *ContactoValidator) ValidarCorreo(correo string) *models.ErrorResponse {
	if !v.correoRegex.MatchString(correo) {
		return &models.ErrorResponse{
			Campo:   "correo",
			Mensaje: "El correo debe ser de un proveedor conocido (gmail, yahoo, hotmail, outlook, live, icloud, protonmail)",
		}
	}
	return nil
}

// ValidarNombre valida el formato del nombre
func (v *ContactoValidator) ValidarNombre(nombre string) *models.ErrorResponse {
	if !v.nombreRegex.MatchString(nombre) || strings.TrimSpace(nombre) == "" {
		return &models.ErrorResponse{
			Campo:   "nombre",
			Mensaje: "El nombre no debe contener números ni estar vacío",
		}
	}
	return nil
}

// ValidarClaveCliente valida la clave del cliente
func (v *ContactoValidator) ValidarClaveCliente(clave int) *models.ErrorResponse {
	if clave <= 0 {
		return &models.ErrorResponse{
			Campo:   "claveCliente",
			Mensaje: "La clave cliente debe ser un número mayor a 0",
		}
	}
	return nil
}

// ValidarBusqueda valida los parámetros de búsqueda
func (v *ContactoValidator) ValidarBusqueda(dto *models.ContactoDTO) []models.ErrorResponse {
	var errores []models.ErrorResponse

	// Las validaciones de búsqueda son más permisivas
	// Solo validamos formato si se proporcionan valores

	if dto.ClaveCliente != "" {
		// Validar que sea numérico si se proporciona
		if !regexp.MustCompile(`^\d+$`).MatchString(dto.ClaveCliente) {
			errores = append(errores, models.ErrorResponse{
				Campo:   "claveCliente",
				Mensaje: "La clave cliente debe ser numérica",
			})
		}
	}

	return errores
}