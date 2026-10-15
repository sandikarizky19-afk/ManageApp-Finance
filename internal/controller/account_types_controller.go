package controller

import (
	"context"
	"database/sql"
	"finance-service/internal/model"
	"finance-service/internal/repository"
	"finance-service/internal/utils"
	"net/http"
	"strings"
)

type AccountTypesController struct {
	DB   *sql.DB
	JWT  *utils.JWTManager
	Repo *repository.AccountTypesRepository
}

func NewAccountTypesController(db *sql.DB, jwt *utils.JWTManager) *AccountTypesController {
	return &AccountTypesController{
		DB:   db,
		JWT:  jwt,
		Repo: repository.NewAccountTypesRepository(db),
	}
}

func (c *AccountTypesController) GetAccountTypes(ctx context.Context) ([]model.AccountTypes, error) {
	return c.Repo.GetAccountTypes(ctx)
}

func (c *AccountTypesController) CreateAccountTypes(ctx context.Context, m model.AccountTypesModel) (*model.AccountTypesModel, error) {
	m.TypeCode = strings.ToLower(strings.TrimSpace(m.TypeCode))
	m.TypeName = strings.TrimSpace(m.TypeName)

	if m.TypeCode == "" || len(m.TypeCode) > 30 {
		return nil, utils.AppError{
			Code:    http.StatusBadRequest,
			Message: "type_code wajib diisi (maks 30 karakter)",
		}
	}
	if m.TypeName == "" || len(m.TypeName) > 100 {
		return nil, utils.AppError{
			Code:    http.StatusBadRequest,
			Message: "type_name wajib diisi (maks 100 karakter)",
		}
	}

	return c.Repo.CreateAccountTypes(ctx, m)
}

func (c *AccountTypesController) UpdateAccountTypes(ctx context.Context, id int64, m model.UpdateAccountTypeRequest) (*model.AccountTypeResponse, error) {
	return c.Repo.UpdateAccountTypes(ctx, id, m)
}

func (c *AccountTypesController) DeleteAccountTypes(ctx context.Context, id int64) error {
	return c.Repo.DeleteAccountTypes(ctx, id)
}
