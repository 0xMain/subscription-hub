package app

import (
	"github.com/0xMain/subscription-hub/internal/http/gen/authapi"
	"github.com/0xMain/subscription-hub/internal/http/gen/profileapi"
	"github.com/0xMain/subscription-hub/internal/http/middleware"
	"github.com/0xMain/subscription-hub/internal/infra/postgres"
	"github.com/0xMain/subscription-hub/internal/infra/redis"
	"github.com/0xMain/subscription-hub/internal/pkg/tx"
	"github.com/0xMain/subscription-hub/internal/repository"

	"github.com/0xMain/subscription-hub/internal/http/handler"

	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/0xMain/subscription-hub/internal/service"

	"github.com/0xMain/subscription-hub/internal/config"
	"github.com/gin-contrib/cors"

	"github.com/gin-gonic/gin"
)

func Start(ctx context.Context) error {
	cfg, err := config.Load(ctx)
	if err != nil {
		return err
	}

	initCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	db, err := postgres.New(initCtx, postgres.Config{DSN: cfg.DSN()})
	if err != nil {
		return fmt.Errorf("подключение к Postgres: %w", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("закрытие Postgres: %v", err)
		}
	}()

	rdb, err := redis.New(initCtx, redis.Config{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err != nil {
		return fmt.Errorf("подключение к Redis: %w", err)
	}
	defer func() {
		if err := rdb.Close(); err != nil {
			log.Printf("закрытие Redis: %v", err)
		}
	}()

	txManager := tx.NewTransactor(db)
	userRepo := repository.NewUserRepository(db, txManager)
	userTenantRepo := repository.NewUserTenantRepository(db, txManager)

	// 3. Бизнес-логика (Services)
	authnService := service.NewAuthnService(userRepo, cfg.JWTSecret())
	profileService := service.NewProfileService(userRepo, userTenantRepo)

	// 4. Транспорт (Middleware & Handlers)
	authnMiddleware := middleware.NewAuthn(authnService)

	authHandler := handler.NewAuthHandler(authnService)
	profileHandler := handler.NewProfileHandler(profileService)

	// 5. Роутинг
	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:8082"}, // твой фронтенд
		AllowMethods:     []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Tenant-ID"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	authSwag, _ := authapi.GetSwagger()
	profileSwag, _ := profileapi.GetSwagger()

	authValidator, err := middleware.NewOpenApiValidator(authSwag)
	if err != nil {
		log.Fatalf("Ошибка инициализации AUTH валидатора: %v", err)
	}

	profileValidator, err := middleware.NewOpenApiValidator(profileSwag)
	if err != nil {
		log.Fatalf("Ошибка инициализации PROFILE валидатора: %v", err)
	}

	// 2. Группа AUTH (публичная, но с валидацией контракта)
	authGroup := router.Group("/")
	authGroup.Use(authValidator.Handler())
	authapi.RegisterHandlers(authGroup, authHandler)

	// 3. Группа PROFILE (защищенная + валидация контракта)
	profileGroup := router.Group("/")
	profileGroup.Use(profileValidator.Handler())
	profileapi.RegisterHandlersWithOptions(profileGroup, profileHandler, profileapi.GinServerOptions{
		Middlewares: []profileapi.MiddlewareFunc{
			profileapi.MiddlewareFunc(authnMiddleware.Verify()),
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
