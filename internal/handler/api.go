// @title			Stawberry API
// @version		1.0
// @description	Это API для управления сделками по продуктам.
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Bearer token for authentication. Format: "Bearer <token>"

package handler

import (
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"golang.org/x/text/currency"

	// Импорт сваггер-генератора
	"github.com/EM-Stawberry/Stawberry/docs"
	"github.com/EM-Stawberry/Stawberry/internal/handler/middleware"
	"github.com/EM-Stawberry/Stawberry/internal/handler/reviews"
	"github.com/EM-Stawberry/Stawberry/pkg/database"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"

	guesthandler "github.com/EM-Stawberry/Stawberry/internal/handler/guestoffer"
)

// @Summary		Получить статус сервера
// @Description	Возвращает статус сервера и текущее время
// @Tags			health
// @Produce		json
// @Success		200	{object}	map[string]interface{}	"Успешный ответ с данными"
// @Router			/health [get]
func SetupRouter(
	healthH *HealthHandler,
	productH *ProductHandler,
	offerH *OfferHandler,
	userH *UserHandler,
	notificationH *NotificationHandler,
	productReviewH *reviews.ProductReviewsHandler,
	sellerReviewH *reviews.SellerReviewsHandler,
	guestOfferH *guesthandler.Handler,
	userS middleware.UserGetter,
	tokenS middleware.TokenValidator,
	basePath string,
	logger *zap.Logger,
	auditMiddleware *middleware.AuditMiddleware,
	auditH *AuditHandler,
) *gin.Engine {
	router := gin.New()

	// Добавляет кастомные валидаторы для использования в json-тегах
	setupValitators()

	router.Use(auditMiddleware.Middleware())
	router.Use(middleware.ZapLogger(logger))
	router.Use(middleware.ZapRecovery(logger))
	router.Use(middleware.CORS())
	router.Use(middleware.Errors())
	router.Use(middleware.Timeout())

	// Swagger UI эндпоинт
	docs.SwaggerInfo.BasePath = basePath
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// base это эндпойнты без префикса версии
	base := router.Group("/")

	// public это эндпойнты с префиксом версии
	public := base.Group(basePath)

	// secured это эндпойнты, которые не сработают без авторизационного токера
	secured := public.Group("/").Use(middleware.AuthMiddleware(userS, tokenS))

	// healtcheck эндпойнты
	{
		base.GET("/health", healthH.health)
		public.GET("/health", healthH.health)
		secured.GET("/auth_required", healthH.authCheck)
	}

	// эндпойнты регистрации-авторизации
	auth := public.Group("/auth")
	{
		auth.POST("/reg", userH.Registration)
		auth.POST("/login", userH.Login)
		auth.POST("/logout", userH.Logout)
		auth.POST("/refresh", userH.Refresh)
	}
	// эндпойнты для продуктов
	{
		public.GET("/products", productH.GetProducts)
		public.GET("/products/:id", productH.GetProductByID)
	}

	// эндпойнты для гостевых заявок
	{
		base.POST("/guest/offers", guestOfferH.PostGuestOffer)
	}

	// эндпойнты запросов на покупку
	{
		secured.PATCH("offers/:offerID", offerH.PatchOfferStatus)
		secured.GET("offers", offerH.GetUserOffers)
		secured.POST("offers", offerH.PostOffer)
	}

	// эндпойнты отзывов
	{
		public.GET("/products/:id/reviews", productReviewH.GetReviews)
		public.GET("/sellers/:id/reviews", sellerReviewH.GetReviews)
		secured.POST("/products/:id/reviews", productReviewH.AddReview)
		secured.POST("/sellers/:id/reviews", sellerReviewH.AddReview)
	}

	// заглушка эндпоинта админа
	// admin := secured.Group("/admin", middleware.Admin)
	secured.GET("/audit", auditH.DisplayLogs)

	// Эндпоинты для бд
	{
		secured.POST("/dev/seed-db", seedDB)
		secured.POST("/dev/clear-db", clearDB)
	}

	// Эти заглушки можно убрать после реализации соответствующих хендлеров
	_ = productH
	_ = notificationH

	return router
}

func setupValitators() {
	// привязка валидатора кодов валюты (прим: USD, RUB и т.д.)
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		_ = v.RegisterValidation("iso4217", currencyValidator)
	}

}

// реализация валидатора кодов валюты
var currencyValidator validator.Func = func(fl validator.FieldLevel) bool {
	currencyCode := fl.Field().String()
	_, err := currency.ParseISO(currencyCode)
	return err == nil
}

func seedDB(c *gin.Context) {
	database.SeedDB()
}

func clearDB(c *gin.Context) {
	database.ClearDB()
}
