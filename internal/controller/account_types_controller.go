package controller

import (
	"context"
	"database/sql"
	"finance-service/internal/model"
	"finance-service/internal/repository"
	"finance-service/internal/utils"
	"net/http"
	"strings"
	"time"
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

func (c *AccountTypesController) CreateAccountTypes(ctx context.Context, userId int, req model.AccountTypesRequest) (*model.AccountTypesResponse, error) {
	req.TypeCode = strings.ToLower(strings.TrimSpace(req.TypeCode))
	req.TypeName = strings.TrimSpace(req.TypeName)

	if req.TypeCode == "" || len(req.TypeCode) > 30 {
		return nil, utils.AppError{
			Code:    http.StatusBadRequest,
			Message: "type_code wajib diisi (maks 30 karakter)",
		}
	}
	if req.TypeName == "" || len(req.TypeName) > 100 {
		return nil, utils.AppError{
			Code:    http.StatusBadRequest,
			Message: "type_name wajib diisi (maks 100 karakter)",
		}
	}

	accountType := model.AccountTypesModel{
		TypeCode:    req.TypeCode,
		TypeName:    req.TypeName,
		Description: req.Description,
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	createdModel, err := c.Repo.CreateAccountTypes(ctx, userId, accountType)
	if err != nil {
		return nil, err
	}

	response := &model.AccountTypesResponse{
		ID:          createdModel.ID,
		TypeCode:    createdModel.TypeCode,
		TypeName:    createdModel.TypeName,
		Description: createdModel.Description,
		CreatedAt:   createdModel.CreatedAt,
		UpdatedAt:   createdModel.UpdatedAt,
	}

	return response, nil
}

func (c *AccountTypesController) UpdateAccountTypes(ctx context.Context, userId int64, id int64, req model.UpdateAccountTypeRequest) (*model.AccountTypesResponse, error) {
	req.TypeCode = strings.ToLower(strings.TrimSpace(req.TypeCode)) // membersihkan data yang masuk agar tidak ada spasi diawal dan diakhir kemudian diubah menjadi huruf kecil semua
	req.TypeName = strings.TrimSpace(req.TypeName)                  //ini hanya membersihkan spasi diawal dan diakhir

	if req.TypeCode == "" || len(req.TypeCode) > 30 {
		return nil, utils.AppError{
			Code:    http.StatusBadRequest,
			Message: "Jenis balance harus diisi (maks 30 karakter)",
		}
	}

	if req.TypeName == "" || len(req.TypeName) > 100 {
		return nil, utils.AppError{
			Code:    http.StatusBadRequest,
			Message: "Nama balance harus diisi (maks 100 karakter)",
		}
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.Repo.UpdateAccountType(ctx, userId, id, req)
}

func (c *AccountTypesController) DeleteAccountType(ctx context.Context, id int64) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.Repo.DeleteAccountType(ctx, id)
}
