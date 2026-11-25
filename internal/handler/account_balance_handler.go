package handler

import (
	"finance-service/internal/controller"
	"finance-service/internal/model"
	"finance-service/internal/utils"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type AccountBalanceHandler struct {
	controller *controller.AccountBalanceController
	validate   *validator.Validate
}

func NewAccountBalanceHandler(c *controller.AccountBalanceController) *AccountBalanceHandler {
	return &AccountBalanceHandler{
		controller: c,
	}
}

func (h *AccountBalanceHandler) CreateAccountBalance(c *fiber.Ctx) error {
	userId, ok := c.Locals("user_id").(int64)
	if !ok {
		return utils.AppError{
			Code:    http.StatusUnauthorized,
			Message: "Unauthorized",
		}
	}

	var req model.CreateAccountBalanceRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.AppError{
			Code:    http.StatusBadRequest,
			Message: "format request tidak valid",
		}
	}

	if err := h.validate.Struct(req); err != nil {
		return utils.AppError{
			Code:    http.StatusBadRequest,
			Message: "data tidak valid: pastikan account_name, account_type, dan balance terisi dengan benar",
		}
	}

	result, err := h.controller.CreateAccountBalance(c.UserContext(), req, userId)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    result,
	})
}

func (h *AccountBalanceHandler) AddBalance(c *fiber.Ctx) error {
	userId, ok := c.Locals("user_id").(int64)
	if !ok {
		return utils.AppError{
			Code:    http.StatusUnauthorized,
			Message: "Unauthorized",
		}
	}

	var req model.AddBalanceRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.AppError{
			Code:    http.StatusBadRequest,
			Message: "Format request tidak valid",
		}
	}

	if req.Balance.IsNegative() || req.Balance.IsZero() {
		return utils.AppError{
			Code:    http.StatusBadRequest,
			Message: "Jumlah saldo yang ditambahkan harus lebih dari 0",
		}
	}
	if req.AccountType == "" {
		return utils.AppError{
			Code:    http.StatusBadRequest,
			Message: "Tipe akun (Account Type) tidak boleh kosong",
		}
	}
	addBalance, err := h.controller.AddBalance(c.UserContext(), int(userId), req)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    addBalance,
	})

}

func (h *AccountBalanceHandler) GetBalance(c *fiber.Ctx) error {
	userId, ok := c.Locals("user_id").(int64)
	if !ok {
		return utils.AppError{
			Code:    http.StatusUnauthorized,
			Message: "Unauthorized",
		}
	}

	accountType := c.Query("account_type")

	if accountType == "" {
		return utils.AppError{
			Code:    http.StatusBadRequest,
			Message: "Tipe akun (Account Type) tidak boleh kosong",
		}
	}

	req := model.GetBalanceRequest{
		AccountType: accountType,
	}

	getBalance, err := h.controller.GetBalance(c.UserContext(), int(userId), req)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    getBalance,
	})

}

func (h *AccountBalanceHandler) TransferBalance(c *fiber.Ctx) error {
	userId, ok := c.Locals("user_id").(int64)
	if !ok {
		return utils.AppError{
			Code:    http.StatusUnauthorized,
			Message: "Unauthorized",
		}
	}

	var req model.TransferBalanceRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.AppError{
			Code:    http.StatusBadRequest,
			Message: "Format request tidak valid",
		}
	}

	if req.FromAccountType == req.ToAccountType {
		return utils.AppError{
			Code:    http.StatusBadRequest,
			Message: "Rekening asal dan tujuan tidak boleh sama",
		}
	}

	if req.Amount.IsNegative() || req.Amount.IsZero() {
		return utils.AppError{
			Code:    http.StatusBadRequest,
			Message: "Jumlah transfer harus lebih dari 0",
		}
	}

	transferBalance, err := h.controller.TransferBalance(c.UserContext(), int(userId), req)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    transferBalance,
	})

}
