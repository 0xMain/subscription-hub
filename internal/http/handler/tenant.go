package handler

//import (
//	"app-go/internal/http/gen/tenantapi"
//	"app-go/internal/service"
//	"context"
//	"errors"
//	"log"
//	"net/http"
//
//	"github.com/gin-gonic/gin"
//	"github.com/go-playground/validator/v10"
//)
//
//type tenantService interface {
//	Create(ctx context.Context, req *tenantapi.CreateTenantRequest) (*tenantapi.Tenant, error)
//	ByID(ctx context.Context, id int64) (*tenantapi.Tenant, error)
//	FindAll(ctx context.Context, limit, offset int) ([]tenantapi.Tenant, int64, error)
//	Delete(ctx context.Context, id int64) error
//}
//
//type TenantHandler struct {
//	svc      tenantService
//	validate *validator.Validate
//}
//
//func NewTenantHandler(svc tenantService, validate *validator.Validate) tenantapi.ServerInterface {
//	return &TenantHandler{
//		svc:      svc,
//		validate: validate,
//	}
//}
//
//// CreateTenant - создание новой компании (POST /tenants)
//func (h *TenantHandler) CreateTenant(c *gin.Context) {
//	var req tenantapi.CreateTenantJSONRequestBody
//
//	if err := c.ShouldBindJSON(&req); err != nil {
//		c.JSON(http.StatusBadRequest, tenantapi.Error{Error: "неверный формат запроса"})
//		return
//	}
//
//	if err := h.validate.Struct(req); err != nil {
//		log.Printf("ошибка валидации: %v", err)
//		c.JSON(http.StatusBadRequest, tenantapi.Error{Error: "ошибка валидации данных"})
//		return
//	}
//
//	tenant, err := h.svc.Create(c.Request.Context(), &req)
//	if err != nil {
//		switch {
//		case errors.Is(err, service.ErrTenantAlreadyExists):
//			c.JSON(http.StatusConflict, tenantapi.Error{Error: "компания с таким названием уже существует"})
//		default:
//			log.Printf("ошибка создания компании: %v", err)
//			c.JSON(http.StatusInternalServerError, tenantapi.Error{Error: "внутренняя ошибка сервера"})
//		}
//		return
//	}
//
//	c.JSON(http.StatusCreated, tenant)
//}
//
//func (h *TenantHandler) GetTenantById(c *gin.Context, id int64) {
//	tenant, err := h.svc.ByID(c.Request.Context(), id)
//	if err != nil {
//		switch {
//		case errors.Is(err, service.ErrTenantNotFound):
//			c.JSON(http.StatusNotFound, tenantapi.Error{Error: "компания не найдена"})
//		default:
//			log.Printf("ошибка получения компании (id=%d): %v", id, err)
//			c.JSON(http.StatusInternalServerError, tenantapi.Error{Error: "внутренняя ошибка сервера"})
//		}
//		return
//	}
//
//	c.JSON(http.StatusOK, tenant)
//}
//
//func (h *TenantHandler) FindAllTenants(c *gin.Context, params tenantapi.FindAllTenantsParams) {
//
//	limit := 10
//	if params.Limit != nil && *params.Limit > 0 {
//		limit = *params.Limit
//	}
//	if limit > 100 {
//		limit = 100
//	}
//
//	offset := 0
//	if params.Offset != nil && *params.Offset > 0 {
//		offset = *params.Offset
//	}
//
//	tenants, total, err := h.svc.FindAll(c.Request.Context(), limit, offset)
//	if err != nil {
//		log.Printf("ошибка получения списка компаний: %v", err)
//		c.JSON(http.StatusInternalServerError, tenantapi.Error{Error: "внутренняя ошибка сервера"})
//		return
//	}
//
//	response := struct {
//		Items []tenantapi.Tenant `json:"items"`
//		Total int64              `json:"total"`
//	}{
//		Items: tenants,
//		Total: total,
//	}
//
//	c.JSON(http.StatusOK, response)
//}
//
//func (h *TenantHandler) DeleteTenant(c *gin.Context, id int64) {
//	err := h.svc.Delete(c.Request.Context(), id)
//	if err != nil {
//		switch {
//		case errors.Is(err, service.ErrTenantNotFound):
//			c.JSON(http.StatusNotFound, tenantapi.Error{Error: "компания не найдена"})
//		default:
//			log.Printf("ошибка удаления компании (id=%d): %v", id, err)
//			c.JSON(http.StatusInternalServerError, tenantapi.Error{Error: "внутренняя ошибка сервера"})
//		}
//		return
//	}
//
//	c.JSON(http.StatusNoContent, nil)
//}
