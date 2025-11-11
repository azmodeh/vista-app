package main

import (
	"log"

	"vista-app/internal/app/runner"
)

func main() {
	svc := runner.NewService()
	if err := svc.Run("data/config/app_config.yml"); err != nil {
		log.Fatalf("Application failed: %v", err)
	}
}
