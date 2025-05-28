// services/contacto_service_paginated.go
package services

import (
	
	"contactos-api/models"
)


// 🆕 NUEVA ESTRUCTURA PARA PAGINACIÓN
type PaginatedResult struct {
	Data       []models.Contacto `json:"data"`
	Page       int               `json:"page"`
	Size       int               `json:"size"`
	Total      int               `json:"total"`
	TotalPages int               `json:"totalPages"`
	HasNext    bool              `json:"hasNext"`
	HasPrev    bool              `json:"hasPrev"`
}
