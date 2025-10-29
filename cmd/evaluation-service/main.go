package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/offer-hunt/oh-evaluation/internal/app"
	"github.com/offer-hunt/oh-evaluation/internal/config"
)

func main() {
	// Для локальной разработки можно использовать .env файл
	_ = godotenv.Load()

	cfg, err := config.New()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Запускаем приложение
	if err := app.Run(cfg); err != nil {
		log.Printf("server shut down with error: %v", err)
		os.Exit(1)
	}

	log.Println("server stopped gracefully")
}
