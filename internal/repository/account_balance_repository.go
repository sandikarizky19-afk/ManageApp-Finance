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

type AccountBalanceRepository struct {
	DB *sql.DB
}

func NewAccountBalanceRepository(db *sql.DB) *AccountBalanceRepository {
	return &AccountBalanceRepository{DB: db}
}

func (r *AccountBalanceRepository) CreateAccountBalance(ctx context.Context, m model.AccountBalanceModel) (*model.AccountBalanceModel, error) {
	const q = `
		INSERT INTO account_balance
			(user_id, account_name, account_type, balance)
		VALUES
			($1, $2, $3, $4)
		RETURNING id, created_at, updated_at`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := r.DB.QueryRowContext(ctx, q,
		m.UserId,
		m.AccountName,
		m.AccountType,
		m.Balance,
	).Scan(&m.ID, &m.CreatedAt, &m.UpdatedAt) // pointer wajib untuk semua field
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			return nil, utils.AppError{
				Code:    http.StatusConflict,
				Message: "Account balance sudah pernah terbuat",
			}
		}

		slog.ErrorContext(ctx, "create account balance failed",
			"err", err,
			"user_id", m.UserId,
		)
		return nil, utils.AppError{
			Code:    http.StatusInternalServerError,
			Message: "Gagal membuat account balance",
		}
	}

	return &m, nil
}

func (r *AccountBalanceRepository) AddBalance(ctx context.Context, userId int, m model.AddBalanceRequest) (*model.AccountBalanceModel, error) {
	q := `
		UPDATE account_balance
		SET 
			balance = balance + $1,
			updated_at = CURRENT_TIMESTAMP
		WHERE
			user_id = $2 AND account_type = $3 AND deleted_at IS NULL
		RETURNING
			id, user_id, account_name, account_type, balance, created_at, updated_at
	`
	var response model.AccountBalanceModel

	// 2. Eksekusi Query dan Scan Hasilnya
	// Asumsi model.AddBalanceRequest memiliki field Amount dan AccountType
	err := r.DB.QueryRowContext(
		ctx,
		q,
		m.Balance,     // $1
		userId,        // $2
		m.AccountType, // $3
	).Scan(
		&response.ID,
		&response.UserId,
		&response.AccountName,
		&response.AccountType,
		&response.Balance,
		&response.CreatedAt,
		&response.UpdatedAt,
	)

	// 3. Penanganan Error
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, utils.AppError{
				Code:    http.StatusNotFound,
				Message: "Jenis balance tidak ditemukan",
			}
		}

		log.Printf("[ERROR] Gagal AddBalance (UserID: %d, Type: %s): %v", userId, m.AccountType, err)
		return nil, utils.AppError{
			Code:    http.StatusInternalServerError,
			Message: "Gagal menambahkan saldo",
		}
	}

	return &response, nil
}

func (r *AccountBalanceRepository) GetBalance(ctx context.Context, userId int, req model.GetBalanceRequest) (*model.GetBalanceResponse, error) {
	q := `SELECT
			id,
			user_id,
			account_type,
			account_name,
			balance,
			updated_at
		FROM
			account_balance
		WHERE
			user_id = $1
			and account_type = $2
			and deleted_at is null
		ORDER BY id`

	var response model.GetBalanceResponse

	err := r.DB.QueryRowContext(ctx, q, userId, req.AccountType).Scan(
		&response.ID,
		&response.UserId,
		&response.AccountType,
		&response.AccountName,
		&response.Balance,
		&response.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, utils.AppError{
				Code:    http.StatusNotFound,
				Message: "Data balance tidak ditemukan",
			}
		}
		log.Printf("[ERROR] Gagal mengambil data balance (UserID: %d, Type: %s): %v", userId, req.AccountType, err)
		return nil, utils.AppError{
			Code:    http.StatusInternalServerError,
			Message: "Gagal mengambil data balance",
		}
	}

	return &response, nil
}

func (r *AccountBalanceRepository) TransferBalance(ctx context.Context, userId int, req model.TransferBalanceRequest) (*model.TransferBalanceResponse, error) {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("[ERROR] Gagal memulai transfer balance : %v", err)
		return nil, utils.AppError{
			Code:    http.StatusInternalServerError,
			Message: "Gagal memulai proses transfer",
		}
	}
	defer tx.Rollback()

	qDeduct := `
			UPDATE account_balance
			SET
				balance = balance - $1,
				updated_at = CURRENT_TIMESTAMP
			WHERE
				user_id = $2
				and account_type = $3
				and balance >= $1
				and deleted_at is null`
	resultDeduct, err := tx.ExecContext(ctx, qDeduct, req.Amount, userId, req.FromAccountType)
	if err != nil {
		log.Printf("[ERROR] Gagal memotong saldo (UserID: %d, Type: %s): %v", userId, req.FromAccountType, err)
		return nil, utils.AppError{
			Code:    http.StatusInternalServerError,
			Message: "Gagal mengurangi saldo pengirim",
		}
	}

	rowsAffected, _ := resultDeduct.RowsAffected()
	if rowsAffected == 0 {
		return nil, utils.AppError{
			Code:    http.StatusBadRequest,
			Message: "Saldo tidak mencukupi atau jenis balance asal tidak valid",
		}
	}

	//tambah saldo tujuan
	qAdd := `UPDATE account_balance
			SET
				balance = balance + $1,
				updated_at = CURRENT_TIMESTAMP
			WHERE
				user_id= $2
				and account_type = $3	
				and deleted_at is null`
	resultAdd, err := tx.ExecContext(ctx, qAdd, req.Amount, userId, req.ToAccountType)
	if err != nil {
		log.Printf("[ERROR] Gagal menambah saldo (UserID: %d, Type: %s): %v", userId, req.ToAccountType, err)
		return nil, utils.AppError{
			Code:    http.StatusInternalServerError,
			Message: "Gagal menambahkan saldo ke tujuan",
		}
	}

	rowsAffectedAdd, _ := resultAdd.RowsAffected()
	if rowsAffectedAdd == 0 {
		return nil, utils.AppError{
			Code:    http.StatusNotFound,
			Message: "Jenis balance tujuan tidak ditemukan",
		}
	}

	if err = tx.Commit(); err != nil {
		log.Printf("[ERROR] Gagal commit transaksi transfer: %v", err)
		return nil, utils.AppError{
			Code:    http.StatusInternalServerError,
			Message: "Gagal menyelesaikan proses transfer",
		}
	}

	var response model.TransferBalanceResponse
	response.FromAccountType = req.FromAccountType
	response.ToAccountType = req.ToAccountType
	response.Amount = req.Amount

	return &response, nil
}
