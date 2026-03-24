package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"proxy-service/internal/database"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var (
	ErrVersionMismatch = errors.New("Version mismatch")
)

type AdminHandlers struct {
	DB *gorm.DB
}

func (h *AdminHandlers) GetServices(c *gin.Context) {
	var services []database.Service
	result := h.DB.Preload("Targets").Find(&services)

	if result.Error != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": result.Error.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, services)
}

func (h *AdminHandlers) GetService(c *gin.Context) {
	var service database.Service

	err := h.DB.Transaction(func(tx *gorm.DB) error {
		return tx.Preload("Targets").First(&service, "name = ?", c.Param("name")).Error
	})

	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("Service '%s' not found", c.Param("name"))})
		return
	}
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, service)
}

func (h *AdminHandlers) CreateService(c *gin.Context) {
	var service database.Service

	if err := c.BindJSON(&service); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	err := h.DB.Transaction(func(tx *gorm.DB) error {
		return tx.Create(&service).Error
	})
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		c.IndentedJSON(http.StatusConflict, gin.H{"message": fmt.Sprintf("Service '%s' already exists", service.Name)})
		return
	}
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "ok"})
}

func (h *AdminHandlers) UpdateService(c *gin.Context) {
	var service database.Service

	if err := c.BindJSON(&service); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	if service.Name != c.Param("name") {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Connot update name"})
		return
	}

	err := h.DB.Transaction(func(tx *gorm.DB) error {
		var tmpService database.Service
		if err := tx.First(&tmpService, "name = ?", c.Param("name")).Error; err != nil {
			return err
		}

		if tmpService.Version != service.Version {
			return ErrVersionMismatch
		}

		service.Version = tmpService.Version + 1
		if err := tx.Omit("Targets").Save(&service).Error; err != nil {
			return err
		}

		if err := tx.Where("service_name = ?", service.Name).Delete(&database.Target{}).Error; err != nil {
			return err
		}

		if len(service.Targets) > 0 {
			for i := range service.Targets {
				service.Targets[i].ServiceName = service.Name
			}

			if err := tx.Create(&service.Targets).Error; err != nil {
				return err
			}
		}

		return nil
	})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("Service '%s' not found", c.Param("name"))})
		return
	}
	if errors.Is(err, ErrVersionMismatch) {
		c.IndentedJSON(http.StatusConflict, gin.H{"message": err.Error()})
		return
	}
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "ok"})
}

func (h *AdminHandlers) DeleteService(c *gin.Context) {
	result := h.DB.Delete(&database.Service{Name: c.Param("name")})
	if result.Error != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": result.Error.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"message": "ok"})
}
