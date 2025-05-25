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