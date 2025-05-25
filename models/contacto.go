// models/contacto.go
package models

// Contacto representa la estructura de un contacto
type Contacto struct {
	ClaveCliente     int    `json:"claveCliente"`
	Nombre           string `json:"nombre"`
	Correo           string `json:"correo"`
	TelefonoContacto string `json:"telefonoContacto"`
}

// ContactoDTO representa los datos de transferencia para búsquedas
type ContactoDTO struct {
	ClaveCliente string `json:"claveCliente,omitempty"`
	Nombre       string `json:"nombre,omitempty"`
	Correo       string `json:"correo,omitempty"`
	Telefono     string `json:"telefono,omitempty"`
}

// ErrorResponse representa un error de validación
type ErrorResponse struct {
	Campo   string `json:"campo"`
	Mensaje string `json:"mensaje"`
}

// APIResponse representa una respuesta estándar de la API
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Errors  []ErrorResponse `json:"errors,omitempty"`
}

// ContactoRequest representa los datos de entrada para crear/actualizar
type ContactoRequest struct {
	ClaveCliente     int    `json:"claveCliente"`
	Nombre           string `json:"nombre"`
	Correo           string `json:"correo"`
	TelefonoContacto string `json:"telefonoContacto"`
}

// ToContacto convierte ContactoRequest a Contacto
func (cr *ContactoRequest) ToContacto() *Contacto {
	return &Contacto{
		ClaveCliente:     cr.ClaveCliente,
		Nombre:           cr.Nombre,
		Correo:           cr.Correo,
		TelefonoContacto: cr.TelefonoContacto,
	}
}