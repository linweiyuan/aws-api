package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/favicon"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"

	"github.com/linweiyuan/aws-api/internal/aws"
	"github.com/linweiyuan/aws-api/internal/k8s"
	"github.com/linweiyuan/aws-api/internal/rds"
)

func main() {
	app := fiber.New()

	app.Use(cors.New())
	app.Use(favicon.New())
	app.Use(logger.New())
	app.Use(recover.New(recover.Config{
		EnableStackTrace: true,
	}))

	apiGroup := app.Group("/api")
	apiGroup.Post("/login", aws.Login)
	apiGroup.Post("/assume", aws.AssumeRole)
	apiGroup.Post("/token", rds.GetToken)
	apiGroup.Get("/k8s", k8s.GetToken)
	apiGroup.Get("/logs/:namespace/:pod/:container", k8s.GetLogs)

	if err := app.Listen(":" + os.Getenv("APP_PORT")); err != nil {
		log.Fatal("failed to start service", err.Error())
	}
}
