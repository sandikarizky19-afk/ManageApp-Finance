package handler

import (
	"finance-service/internal/controller"
	"finance-service/internal/model"
	"finance-service/internal/utils"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

type SavingsHandler struct {
	controller *controller.SavingsController
}

func NewSavingsHandler(c *controller.SavingsController) *SavingsHandler {
	return &SavingsHandler{
		controller: c,
	}
}

func (h *SavingsHandler) CreateSavings(c *fiber.Ctx) error {
	var req model.CreateSavingsRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.AppError{
			Code:    http.StatusBadRequest,
			Message: "format request tidak valid",
		}
	}

	userId, ok := c.Locals("user_id").(int64)
	if !ok {
		return utils.AppError{
			Code:    http.StatusUnauthorized,
			Message: "Unauthorized",
		}
	}

	createSavings, err := h.controller.CreateSavings(c.UserContext(), int(userId), req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Tabungan berhasil dibuat",
		"data":    createSavings,
	})
}

func (h *SavingsHandler) GetAllSavings(c *fiber.Ctx) error {
	userId, ok := c.Locals("user_id").(int64)
	if !ok {
		return utils.AppError{
			Code:    http.StatusUnauthorized,
			Message: "Unauthorized",
		}
	}

	responses, err := h.controller.GetAllSavings(c.Context(), userId)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Berhasil mengambil data tabungan",
		"data":    responses,
	})
}
