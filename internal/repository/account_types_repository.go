package repository

import (
	"context"
	"database/sql"
	"errors"
	"finance-service/internal/model"
	"finance-service/internal/utils"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

type AccountTypesRepository struct {
	DB *sql.DB
}

func NewAccountTypesRepository(db *sql.DB) *AccountTypesRepository {
	return &AccountTypesRepository{DB: db}
}

func (r *AccountTypesRepository) GetAccountTypes(ctx context.Context) ([]model.AccountTypes, error) {
	q := `select
			id,
			type_code,
			type_name,
			description
		from
			public.account_types
		where
			deleted_at is null
		order by id`
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	rows, err := r.DB.QueryContext(ctx, q)
	if err != nil {
		return nil, utils.AppError{
			Code:    http.StatusInternalServerError,
			Message: "Gagal mengambil tipe saldo. " + err.Error(),
		}
	}
	defer rows.Close()

	result := make([]model.AccountTypes, 0)
	for rows.Next() {
		var m model.AccountTypes
		if err := rows.Scan(
			&m.Id,
			&m.TypeCode,
			&m.TypeName,
			&m.Description,
		); err != nil {
			return nil, utils.AppError{
				Code:    http.StatusInternalServerError,
				Message: "Gagal membaca data tipe saldo. " + err.Error(),
			}
		}
		result = append(result, m)
	}
	if err := rows.Err(); err != nil {
		return nil, utils.AppError{
			Code:    http.StatusInternalServerError,
			Message: "Gagal mengambil tipe saldo",
		}
	}

	return result, nil
}

func (r *AccountTypesRepository) CreateAccountTypes(ctx context.Context, userId int, m model.AccountTypesModel) (*model.AccountTypesModel, error) {
	//membuat database transaction
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		slog.ErrorContext(ctx, "failed to begin transaction", "err", err)
		return nil, utils.AppError{
			Code:    http.StatusInternalServerError,
			Message: "Gagal memulai pembuatan tipe saldo",
		}
	}
	defer tx.Rollback()

	qType := `INSERT INTO 
			account_types
		(type_code, type_name, description, user_id) 
			VALUES 
		($1, $2, $3, $4)
		RETURNING id, created_at, updated_at`
	err = tx.QueryRowContext(ctx, qType, m.TypeCode, m.TypeName, m.Description, userId).
		Scan(&m.ID, &m.CreatedAt, &m.UpdatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, utils.AppError{
				Code:    http.StatusConflict,
				Message: "Kode tipe saldo sudah digunakan",
			}
		}
		slog.ErrorContext(ctx, "create account type failed", "err", err)
		return nil, utils.AppError{
			Code:    http.StatusInternalServerError,
			Message: "Gagal membuat tipe saldo",
		}
	}

	qBalance := `INSERT INTO
					account_balance
				(account_name, account_type, user_id, balance )
				values
				($1, $2, $3, 0.00)`
	_, err = tx.ExecContext(ctx, qBalance, m.TypeName, m.TypeCode, userId)
	if err != nil {
		slog.ErrorContext(ctx, "auto create account balance failed", "err", err)
		return nil, utils.AppError{
			Code:    http.StatusInternalServerError,
			Message: "Gagal membuat saldo awal",
		}
	}

	if err = tx.Commit(); err != nil {
		slog.ErrorContext(ctx, "commit transaction failed", "err", err)
		return nil, utils.AppError{
			Code:    http.StatusInternalServerError,
			Message: "Gagal menyelesaikan penyimpanan data",
		}
	}

	return &m, nil
}

func (r *AccountTypesRepository) UpdateAccountType(ctx context.Context, userId int64, id int64, m model.UpdateAccountTypeRequest) (*model.AccountTypesResponse, error) {
	// 1. Mulai Transaksi
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("[ERROR] Gagal memulai transaksi UpdateAccountType: %v", err)
		return nil, utils.AppError{
			Code:    http.StatusInternalServerError,
			Message: "Gagal memulai pembaruan data",
		}
	}
	defer tx.Rollback()

	// 2. Update data di account_types
	qType := `
		UPDATE account_types 
		SET 
			type_code = $3,
			type_name = $4,
			description = $5,
			updated_at = CURRENT_TIMESTAMP
		WHERE
			id = $1 AND user_id = $2 AND deleted_at IS NULL
		RETURNING
			id, type_code, type_name, description, created_at, updated_at
	`
	var response model.AccountTypesResponse

	err = tx.QueryRowContext(ctx, qType, id, userId, m.TypeCode, m.TypeName, m.Description).
		Scan(
			&response.ID,
			&response.TypeCode,
			&response.TypeName,
			&response.Description,
			&response.CreatedAt,
			&response.UpdatedAt,
		)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, utils.AppError{
				Code:    http.StatusNotFound,
				Message: "Account type tidak ditemukan",
			}
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" { // Unique violation
				return nil, utils.AppError{
					Code:    http.StatusConflict,
					Message: "Kode tipe saldo sudah digunakan",
				}
			}
			if pgErr.Code == "23503" { // Foreign key violation (jika lupa set CASCADE)
				return nil, utils.AppError{
					Code:    http.StatusConflict,
					Message: "Tidak dapat mengubah kode karena masih ada transaksi terkait",
				}
			}
		}

		log.Printf("[ERROR] Update account type (ID: %d): %v", id, err)
		return nil, utils.AppError{
			Code:    http.StatusInternalServerError,
			Message: "Gagal memperbarui account type",
		}
	}

	// 3. Sinkronisasi nama (account_name) di account_balance
	// (Note: type_code pada balance berubah otomatis berkat ON UPDATE CASCADE di DB)
	qBalance := `
		UPDATE account_balance
		SET
			account_name = $1,
			updated_at = CURRENT_TIMESTAMP
		WHERE
			account_type = $2 AND deleted_at IS NULL
	`
	_, err = tx.ExecContext(ctx, qBalance, response.TypeName, response.TypeCode)
	if err != nil {
		log.Printf("[ERROR] Gagal sinkronisasi account_balance saat update (TypeCode: %s): %v", response.TypeCode, err)
		return nil, utils.AppError{
			Code:    http.StatusInternalServerError,
			Message: "Gagal menyinkronkan saldo terkait",
		}
	}

	// 4. Selesaikan Transaksi
	if err = tx.Commit(); err != nil {
		log.Printf("[ERROR] Gagal commit UpdateAccountType: %v", err)
		return nil, utils.AppError{
			Code:    http.StatusInternalServerError,
			Message: "Gagal menyimpan pembaruan data",
		}
	}

	return &response, nil
}

func (r *AccountTypesRepository) DeleteAccountType(ctx context.Context, id int64) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("[ERROR] Gagal memulai transaksi DeleteAccountType: %v", err)
		return utils.AppError{
			Code:    http.StatusInternalServerError,
			Message: "Gagal memulai penghabusan data balance",
		}
	}
	defer tx.Rollback()

	qType := `
		UPDATE account_types
		SET
			deleted_at = CURRENT_TIMESTAMP,
			updated_at = CURRENT_TIMESTAMP
		WHERE
			id=$1
		AND deleted_at is null
	`
	var typeCode string
	err = tx.QueryRowContext(ctx, qType, id).Scan(&typeCode)

	if err != nil {
		// Jika data tidak ditemukan (sudah terhapus atau ID salah)
		if errors.Is(err, sql.ErrNoRows) {
			return utils.AppError{
				Code:    http.StatusNotFound,
				Message: "Account type tidak ditemukan",
			}
		}

		log.Printf("[ERROR] Gagal DeleteAccountType (ID: %d): %v", id, err)
		return utils.AppError{
			Code:    http.StatusInternalServerError,
			Message: "Gagal menonaktifkan account type.",
		}
	}

	qBalance := `
        UPDATE account_balance
        SET
            deleted_at = CURRENT_TIMESTAMP,
            updated_at = CURRENT_TIMESTAMP
        WHERE
            account_type = $1
        AND deleted_at IS NULL
    `

	_, err = tx.ExecContext(ctx, qBalance, typeCode)
	if err != nil {
		log.Printf("[ERROR] Gagal menonaktifkan account balance terkait (TypeCode: %s): %v", typeCode, err)
		return utils.AppError{
			Code:    http.StatusInternalServerError,
			Message: "Gagal menonaktifkan balance terkait.",
		}
	}

	// 4. Commit transaksi jika kedua query di atas berhasil
	if err = tx.Commit(); err != nil {
		log.Printf("[ERROR] Gagal commit DeleteAccountType: %v", err)
		return utils.AppError{
			Code:    http.StatusInternalServerError,
			Message: "Gagal menyelesaikan proses penghapusan data",
		}
	}

	return nil
}
