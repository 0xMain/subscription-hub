package app

import (
	"github.com/0xMain/subscription-hub/internal/http/gen/authapi"
	"github.com/0xMain/subscription-hub/internal/http/gen/profileapi"
	"github.com/0xMain/subscription-hub/internal/http/middleware"
	"github.com/0xMain/subscription-hub/internal/infra/postgres"
	"github.com/0xMain/subscription-hub/internal/pkg/tx"
	"github.com/0xMain/subscription-hub/internal/repository"

	"github.com/0xMain/subscription-hub/internal/http/handler"

	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/0xMain/subscription-hub/internal/service"

	"github.com/gin-contrib/cors"
	"github.com/go-playground/validator/v10"

	"github.com/0xMain/subscription-hub/internal/config"

	"github.com/gin-gonic/gin"
)

func Start(ctx context.Context) error {
	cfg, err := config.Load(ctx)
	if err != nil {
		return err
	}

	db, err := postgres.New(cfg.DSN())
	if err != nil {
		return fmt.Errorf("ошибка подключения к БД: %w", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("ошибка при закрытии БД: %v", err)
		}
	}()

	txManager := tx.NewTransactor(db)
	userRepo := repository.NewUserRepository(db, txManager)
	userTenantRepo := repository.NewUserTenantRepository(db, txManager)

	// 3. Бизнес-логика (Services)
	authnService := service.NewAuthnService(userRepo, cfg.JWTSecret())
	profileService := service.NewProfileService(userRepo, userTenantRepo)

	// 4. Транспорт (Middleware & Handlers)
	authnMiddleware := middleware.NewAuthnMiddleware(authnService)
	validate := validator.New()

	authHandler := handler.NewAuthHandler(authnService, validate)
	profileHandler := handler.NewProfileHandler(profileService, validate)

	// 5. Роутинг
	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:8082"}, // твой фронтенд
		AllowMethods:     []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Tenant-ID"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	authapi.RegisterHandlers(router, authHandler)

	// Защищенные эндпоинты (Profile)
	profileapi.RegisterHandlersWithOptions(router, profileHandler, profileapi.GinServerOptions{
		Middlewares: []profileapi.MiddlewareFunc{
			profileapi.MiddlewareFunc(authnMiddleware.Verify()), // Прямая передача без анонимной функции
		},
	})

	server := &http.Server{
		Addr:    cfg.Server.Host + ":" + cfg.Server.Port,
		Handler: router,
	}

	go func() {
		<-ctx.Done()
		log.Println("завершение работы сервера...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("ошибка при завершении работы сервера: %v", err)
		}
	}()

	log.Printf("сервер запускается на %s", server.Addr)
	return server.ListenAndServe()
}
