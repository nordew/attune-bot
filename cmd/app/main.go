package main

import (
	"attune/internal/app"
)

//func init() {
//	if err := godotenv.Load(); err != nil {
//		log.Fatalf("error loading .env file: %v", err)
//	}
//}

func main() {
	app.MustRun()
}
