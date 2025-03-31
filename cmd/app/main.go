package main

import (
	"attune/internal/app"
	"log"

	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("error loading .env file: %v", err)
	}
}

func main() {
	app.MustRun()
}
