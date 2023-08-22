package main

import (
	"brms/config"
	api "brms/endpoints/api"
	"io"
	"os"
	"time"

	"brms/pkg/file"
	"brms/pkg/middlewares"
	"encoding/json"
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	FRecover "github.com/gofiber/fiber/v2/middleware/recover"
)

func setApp(file *os.File) *fiber.App { // setting up middlewares
	app := fiber.New()

	// recover middleware
	app.Use(FRecover.New(FRecover.Config{
		EnableStackTrace: true,
		StackTraceHandler: func(c *fiber.Ctx, e interface{}) {
			log.New(io.MultiWriter(os.Stdout, file), "[ERROR] ", log.Ldate|log.Ltime).Printf("%s %s: %s\n", c.Path(), c.Method(), e)
		},
	}))

	// cors middleware
	app.Use(cors.New())

	// logger middleware
	app.Use(logger.New(logger.Config{
		Format: "[${time}] [${severity}] ${path} ${method} (${ip}) ${status} ${latency} - ${message}\n",
		CustomTags: map[string]logger.LogFunc{
			"time": func(output logger.Buffer, c *fiber.Ctx, data *logger.Data, extraParam string) (int, error) {
				return output.WriteString(time.Now().Format("2006-01-02 15:04:05"))
			},
			"message": func(output logger.Buffer, c *fiber.Ctx, data *logger.Data, extraParam string) (int, error) {
				if bodyBytes := c.Response().Body(); bodyBytes != nil {
					var bodyData map[string]interface{}

					err := json.Unmarshal(bodyBytes, &bodyData)
					if err == nil {
						msgValue, _ := bodyData["message"].(string)
						return output.WriteString(msgValue)
					}
				}
				return 0, nil
			},
			"severity": func(output logger.Buffer, c *fiber.Ctx, data *logger.Data, extraParam string) (int, error) {
				status := c.Response().StatusCode()

				if status == fiber.StatusInternalServerError {
					return output.WriteString("WARNING")
				}
				return output.WriteString("INFO")
			},
		},
		Output: io.MultiWriter(os.Stdout, file),
	}))

	// caching middleware
	app.Use(cache.New(cache.Config{
		Expiration:   30 * time.Minute,
		CacheControl: true,
	}))

	// compression middleware
	app.Use(compress.New())

	// check valid routes middleware
	app.Use(middlewares.UndefinedRoutesMiddleware())

	// print error middleware for general usecase
	app.Use(middlewares.ErrorMiddleware())

	return app
}

func main() {
	file, err := file.OpenLogFile()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	defer file.Close()

	// recover panic not from routes
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintln(os.Stderr, r)
			log.New(file, "[ERROR] ", log.Ldate|log.Ltime).Println("Encountered a system error: ", r)
		}
	}()

	app := setApp(file)

	// register rule management routes
	api.Routes(app)

	if err := app.Listen(fmt.Sprintf(":%s", config.GetConfig().Port)); err != nil {
		log.New(file, "[ERROR] ", log.Ldate|log.Ltime).Println("Application failed to start running: ", err)
		os.Exit(1)
	}
}
