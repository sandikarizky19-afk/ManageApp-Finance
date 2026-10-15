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
	var req model.AccountTypesModel
	if err := c.BodyParser(&req); err != nil {
		return utils.AppError{
			Code:    http.StatusBadRequest,
			Message: "format request tidak valid",
		}
	}
	createAccountTypes, err := h.controller.CreateAccountTypes(c.UserContext(), req)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    createAccountTypes,
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

	var req model.UpdateAccountTypeRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.AppError{
			Code:    http.StatusBadRequest,
			Message: "Format request tidak valid",
		}
	}
	updateAccountTypes, err := h.controller.UpdateAccountTypes(c.UserContext(), id, req)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    updateAccountTypes,
	})
}

func (h *AccountTypesHandler) DeleteAccountTypes(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		return utils.AppError{
			Code:    http.StatusBadRequest,
			Message: "ID account type tidak valid",
		}
	}

	err = h.controller.DeleteAccountTypes(
		c.UserContext(),
		id,
	)
	if err != nil {
		return err
	}

	return c.SendStatus(http.StatusNoContent)
}
