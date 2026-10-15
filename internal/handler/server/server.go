package server

import (
	"database/sql"
	"errors"
	"finance-service/internal/controller"
	"finance-service/internal/handler"
	"finance-service/internal/middleware"
	"finance-service/internal/utils"
	"log"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

type Server struct {
	App *fiber.App
	DB  *sql.DB
}

func New(db *sql.DB, jwtManager *utils.JWTManager) *Server {
	app := fiber.New(fiber.Config{
		AppName:      "Finance Service",
		ErrorHandler: errorHandler,
	})

	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3000, http://127.0.0.1:3000, http://localhost:5173, http://127.0.0.1:5173",
		AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS, PATCH",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: true,
	}))

	s := &Server{
		App: app,
		DB:  db,
	}

	s.setupRoutes(jwtManager)

	return s
}

func errorHandler(c *fiber.Ctx, err error) error {
	var appErr utils.AppError
	if errors.As(err, &appErr) {
		return c.Status(appErr.Code).JSON(fiber.Map{
			"success": false,
			"message": appErr.Message,
		})
	}

	var fe *fiber.Error
	if errors.As(err, &fe) {
		return c.Status(fe.Code).JSON(fiber.Map{
			"success": false,
			"message": fe.Message,
		})
	}

	log.Printf("unhandled error: %v", err)
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"success": false,
		"message": "terjadi kesalahan pada server",
	})
}

func (s *Server) setupRoutes(jwtManager *utils.JWTManager) {

	accountTypesController := controller.NewAccountTypesController(s.DB, jwtManager)
	accountTypesHandler := handler.NewAccountTypesHandler(accountTypesController)

	api := s.App.Group("/api/v1")

	//protected
	finance := api.Group("/finance", middleware.AuthRequired(jwtManager))
	//testing
	finance.Get("/me", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"user_id": c.Locals("user_id"),
			"role":    c.Locals("role"),
		})
	})

	finance.Get("/account-types", accountTypesHandler.GetAccountTypes)
	finance.Post("/account-types", accountTypesHandler.CreateAccountTypes)
	finance.Put("/account-types/:id", accountTypesHandler.UpdateAccountTypes)
	finance.Delete("/account-types/:id", accountTypesHandler.DeleteAccountTypes)

}

func (s *Server) Listen(addr string) error {
	return s.App.Listen(addr)
}

func AdminOnly() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if role, _ := c.Locals("role").(string); role != "admin" {
			return utils.AppError{Code: http.StatusForbidden, Message: "akses ditolak"}
		}
		return c.Next()
	}
}
