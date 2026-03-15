package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/0xMain/subscription-hub/internal/app"
)

func main() {
	//Создание контекста
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	//Запуск
	if err := app.Start(ctx); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			log.Println("ошибка запуска приложения:", err)
		}
	}

	log.Println("приложение корректно остановлено")
}
