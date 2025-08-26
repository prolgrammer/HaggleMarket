package main

import (
	"github.com/EM-Stawberry/Stawberry/internal/adapter/auth"
	"github.com/EM-Stawberry/Stawberry/internal/domain/service/audit"
	"github.com/EM-Stawberry/Stawberry/internal/domain/service/notification"
	"github.com/EM-Stawberry/Stawberry/internal/domain/service/reviews"
	"github.com/EM-Stawberry/Stawberry/internal/domain/service/token"
	"github.com/EM-Stawberry/Stawberry/internal/domain/service/user"
	"github.com/EM-Stawberry/Stawberry/internal/handler/middleware"
	"github.com/EM-Stawberry/Stawberry/internal/repository"
	"github.com/EM-Stawberry/Stawberry/pkg/database"
	"github.com/EM-Stawberry/Stawberry/pkg/email"
	"github.com/EM-Stawberry/Stawberry/pkg/logger"
	"github.com/EM-Stawberry/Stawberry/pkg/migrator"
	"github.com/EM-Stawberry/Stawberry/pkg/security"
	"github.com/EM-Stawberry/Stawberry/pkg/server"
	"github.com/jmoiron/sqlx"
	flag "github.com/spf13/pflag"
	"go.uber.org/zap"

	"github.com/EM-Stawberry/Stawberry/config"
	guestofferservice "github.com/EM-Stawberry/Stawberry/internal/domain/service/guestoffer"
	"github.com/EM-Stawberry/Stawberry/internal/domain/service/offer"
	"github.com/EM-Stawberry/Stawberry/internal/domain/service/product"
	"github.com/EM-Stawberry/Stawberry/internal/handler"
	guesthandler "github.com/EM-Stawberry/Stawberry/internal/handler/guestoffer"
	hdlr "github.com/EM-Stawberry/Stawberry/internal/handler/reviews"
	guestofferrepo "github.com/EM-Stawberry/Stawberry/internal/repository/guestoffer"
	repo "github.com/EM-Stawberry/Stawberry/internal/repository/reviews"
	"github.com/gin-gonic/gin"
)

const basePath = "/api/v1"

var enableMail bool

func init() {
	flag.BoolVarP(&enableMail, "mail", "m", false, "enable email notifications")
}

func main() {
	flag.Parse()

	cfg := config.LoadConfig()
	log := logger.SetupLogger(cfg.Environment)
	middleware.SetupGinWithZap(log)
	log.Info("Logger initialized")

	db, closer := database.InitDB(&cfg.DB, log)
	defer closer()

	migrator.RunMigrationsWithZap(db, "migrations", log)

	database.DefaultAdminAcc()

	router, mailer, auditMiddleware := initializeApp(cfg, db, log)

	if err := server.StartServer(router, mailer, &cfg.Server, log); err != nil {
		log.Fatal("Failed to start server", zap.Error(err))
	}

	auditMiddleware.Close()
}

func initializeApp(
	cfg *config.Config,
	db *sqlx.DB,
	log *zap.Logger,
) (
	*gin.Engine,
	email.MailerService,
	*middleware.AuditMiddleware) {
	mailer := email.NewMailer(log, &cfg.Email)
	log.Info("Mailer initialized")

	productRepository := repository.NewProductRepository(db)
	offerRepository := repository.NewOfferRepository(db)
	userRepository := repository.NewUserRepository(db)
	notificationRepository := repository.NewNotificationRepository(db)
	tokenRepository := repository.NewTokenRepository(db)
	productReviewsRepository := repo.NewProductReviewRepository(db, log)
	sellerReviewsRepository := repo.NewSellerReviewRepository(db, log)
	auditRepository := repository.NewAuditRepository(db)
	guestOfferRepository := guestofferrepo.NewRepository(db)
	log.Info("Repositories initialized")

	passwordManager := security.NewArgon2idPasswordManager()
	jwtManager := auth.NewJWTManager(cfg.Token.Secret)

	productService := product.NewService(productRepository)
	offerService := offer.NewService(offerRepository, mailer)
	tokenService := token.NewService(
		tokenRepository,
		jwtManager,
		cfg.Token.RefreshTokenDuration,
		cfg.Token.AccessTokenDuration,
	)
	userService := user.NewService(userRepository, tokenService, passwordManager, mailer)
	notificationService := notification.NewService(notificationRepository)
	productReviewsService := reviews.NewProductReviewService(productReviewsRepository, log)
	sellerReviewsService := reviews.NewSellerReviewService(sellerReviewsRepository, log)
	auditService := audit.NewAuditService(auditRepository)
	guestOfferService := guestofferservice.NewService(guestOfferRepository, mailer, log)
	log.Info("Services initialized")

	healthHandler := handler.NewHealthHandler()
	productHandler := handler.NewProductHandler(productService)
	offerHandler := handler.NewOfferHandler(offerService)
	userHandler := handler.NewUserHandler(cfg, userService)
	notificationHandler := handler.NewNotificationHandler(notificationService)
	productReviewsHandler := hdlr.NewProductReviewHandler(productReviewsService, log)
	sellerReviewsHandler := hdlr.NewSellerReviewsHandler(sellerReviewsService, log)
	auditHandler := handler.NewAuditHandler(auditService)
	guestOfferHandler := guesthandler.NewHandler(guestOfferService, log)
	log.Info("Handlers initialized")

	auditMiddleware := middleware.NewAuditMiddleware(&cfg.Audit, auditService, log)

	router := handler.SetupRouter(
		healthHandler,
		productHandler,
		offerHandler,
		userHandler,
		notificationHandler,
		productReviewsHandler,
		sellerReviewsHandler,
		guestOfferHandler,
		userService,
		tokenService,
		basePath,
		log,
		auditMiddleware,
		auditHandler,
	)

	return router, mailer, auditMiddleware
}
