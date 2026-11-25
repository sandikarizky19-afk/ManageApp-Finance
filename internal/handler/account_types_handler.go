package handler

import (
	"finance-service/internal/controller"
	"finance-service/internal/model"
	"finance-service/internal/utils"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type AccountTypesHandler struct {
	controller *controller.AccountTypesController
}

func NewAccountTypesHandler(c *controller.AccountTypesController) *AccountTypesHandler {
	return &AccountTypesHandler{
		controller: c,
	}
}

func (h *AccountTypesHandler) GetAccountTypes(c *fiber.Ctx) error {
	accountTypes, err := h.controller.GetAccountTypes(c.UserContext())
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    accountTypes,
	})
}

func (h *AccountTypesHandler) CreateAccountTypes(c *fiber.Ctx) error {
	var req model.AccountTypesRequest
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
	createAccountType, err := h.controller.CreateAccountTypes(c.UserContext(), int(userId), req)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    createAccountType,
	})
}

func (h *AccountTypesHandler) UpdateAccountTypes(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		return utils.AppError{
			Code:    http.StatusBadRequest,
			Message: "ID account type tidak valid",
		}
	}

	userId, ok := c.Locals("user_id").(int64)
	if !ok {
		return utils.AppError{
			Code:    http.StatusUnauthorized,
			Message: "Unauthorized",
		}
	}

	var req model.UpdateAccountTypeRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.AppError{
			Code:    http.StatusBadRequest,
			Message: "Format request tidak valid",
		}
	}
	updateAccountTypes, err := h.controller.UpdateAccountTypes(c.UserContext(), userId, id, req)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    updateAccountTypes,
	})
}

func (h *AccountTypesHandler) DeleteAccountType(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		return utils.AppError{
			Code:    http.StatusBadRequest,
			Message: "ID account type tidak valid",
		}
	}

	err = h.controller.DeleteAccountType(
		c.UserContext(),
		id,
	)
	if err != nil {
		return err
	}

	return c.SendStatus(http.StatusNoContent)
}
