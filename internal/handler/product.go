package handler

import (
	"context"
	"encoding/json"
	"math"
	"net/http"
	"strconv"

	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"

	"github.com/EM-Stawberry/Stawberry/internal/app/apperror"

	"github.com/EM-Stawberry/Stawberry/internal/repository/model"

	"github.com/gin-gonic/gin"
)

type ProductService interface {
	GetFilteredProducts(ctx context.Context, filter model.ProductFilter, limit, offset int) ([]entity.Product, int, error)
	GetProductByID(ctx context.Context, id string) (entity.Product, error)
}

type ProductHandler struct {
	productService ProductService
}

func NewProductHandler(productService ProductService) *ProductHandler {
	return &ProductHandler{productService: productService}
}

// GetProductByID godoc
// @Summary      Получить продукт по его ID
// @Description  Возвращает один продукт по его идентификатору
// @Tags         products
// @Param        id   path      int  true  "ID продукта"
// @Success      200  {object}  entity.Product
// @Failure      400  {object}  apperror.Error "Некорректный ID"
// @Failure      500  {object}  apperror.Error "Ошибка сервера при получении продукта"
// @Router       /products/{id} [get]
func (h *ProductHandler) GetProductByID(c *gin.Context) {
	id := c.Param("id")

	valid, err := strconv.Atoi(id)
	if err != nil || valid < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    apperror.BadRequest,
			"message": "Invalid id)",
		})
		return
	}

	product, err := h.productService.GetProductByID(context.Background(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    apperror.DatabaseError,
			"message": "failed to fetch product)",
		})
		return
	}

	c.JSON(http.StatusOK, product)
}

// GetProducts godoc
// @Summary      Получить список продуктов с фильтрацией и пагинацией
// @Description  Возвращает список продуктов по фильтру (категория, цена, магазин, имя, атрибуты) с поддержкой пагинации
// @Tags         products
// @Accept       json
// @Produce      json
// @Param        page         query     int     false  "Номер страницы (по умолчанию 1)"
// @Param        limit        query     int     false  "Размер страницы (по умолчанию 10, максимум 100)"
// @Param        name         query     string  false  "Фильтр по названию продукта (поиск по подстроке)"
// @Param        min_price    query     int     false  "Минимальная цена (в копейках)"
// @Param        max_price    query     int     false  "Максимальная цена (в копейках)"
// @Param        category_id  query     int     false  "ID категории (с учетом подкатегорий)"
// @Param        shop_id      query     int     false  "ID магазина"
// @Param        attributes   query     string  false  "JSON-строка с фильтрами по атрибутам (exmpl: {"color":"Black"})"
// @Success      200  {object}  map[string]interface{} "Список продуктов и метаинформация"
// @Failure      400  {object}  apperror.Error "Некорректный запрос"
// @Failure      500  {object}  apperror.Error "Ошибка сервера при получении продуктов"
// @Router       /products [get]
func (h *ProductHandler) GetProducts(c *gin.Context) {
	var filter model.ProductFilter

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		_ = c.Error(apperror.New(apperror.BadRequest, "Invalid page number", err))
		c.Abort()
		return
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 || limit > 100 {
		_ = c.Error(apperror.New(apperror.BadRequest, "Invalid limit value (should be between 1 and 100)", err))
		c.Abort()
		return
	}

	offset := (page - 1) * limit

	if err := c.ShouldBindQuery(&filter); err != nil {
		_ = c.Error(apperror.New(apperror.BadRequest, "Invalid query parameters", err))
		c.Abort()
		return
	}

	attrParam := c.Query("attributes")
	if attrParam != "" {
		var attrs map[string]string
		if err := json.Unmarshal([]byte(attrParam), &attrs); err != nil {
			_ = c.Error(apperror.New(apperror.BadRequest, "Invalid attributes json", err))
			c.Abort()
			return
		}
		filter.Attributes = attrs
	}

	products, total, err := h.productService.GetFilteredProducts(c.Request.Context(), filter, limit, offset)
	if err != nil {
		_ = c.Error(apperror.New(apperror.DatabaseError, "Failed to get products", err))
		c.Abort()
		return
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	c.JSON(http.StatusOK, gin.H{
		"data": products,
		"meta": gin.H{
			"current_page": page,
			"per_page":     limit,
			"total_items":  total,
			"total_pages":  totalPages,
		},
	})
}
