package migration

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func RunMigrations(db *sql.DB) error {

	// 1. Buat tabel pencatatan migration
	if err := createMigrationTable(db); err != nil {
		return err
	}

	// 2. Cari direktori migration
	migrationDir, err := resolveMigrationDir()
	if err != nil {
		return err
	}

	// 3. Baca seluruh file migration
	entries, err := os.ReadDir(migrationDir)
	if err != nil {
		return fmt.Errorf(
			"gagal membaca direktori migrasi: %w", err,
		)
	}

	files := make([]string, 0, len(entries))

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if strings.HasSuffix(entry.Name(), ".sql") {
			files = append(files, entry.Name())
		}
	}

	// 4. Urutkan file migration
	sort.Strings(files)

	// 5. Jalankan setiap migration
	for _, fileName := range files {

		// Cek apakah migration sudah dijalankan
		executed, err := isMigrationExecuted(db, fileName)
		if err != nil {
			return fmt.Errorf(
				"gagal memeriksa migrasi %s: %w",
				fileName, err,
			)
		}

		// Jika sudah, lewati
		if executed {
			log.Printf(
				"Migrasi dilewati: %s (sudah dijalankan)",
				fileName,
			)
			continue
		}

		log.Printf("Menjalankan migrasi: %s", fileName)

		// Baca isi file SQL
		filePath := filepath.Join(migrationDir, fileName)

		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf(
				"gagal membaca file migrasi %s: %w",
				fileName, err,
			)
		}

		// Mulai transaksi
		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf(
				"gagal memulai transaksi: %w", err,
			)
		}

		// Jalankan SQL migration
		_, err = tx.Exec(string(content))
		if err != nil {
			tx.Rollback()

			return fmt.Errorf(
				"gagal menjalankan migrasi %s: %w",
				fileName, err,
			)
		}

		// Catat migration yang berhasil
		_, err = tx.Exec(`
			INSERT INTO schema_migrations (migration_name)
			VALUES ($1)
		`, fileName)

		if err != nil {
			tx.Rollback()

			return fmt.Errorf(
				"gagal mencatat migrasi %s: %w",
				fileName, err,
			)
		}

		// Simpan transaksi
		if err := tx.Commit(); err != nil {
			return fmt.Errorf(
				"gagal commit migrasi %s: %w",
				fileName, err,
			)
		}

		log.Printf(
			"Migrasi berhasil dijalankan: %s",
			fileName,
		)
	}

	log.Println("Seluruh migrasi database berhasil diperiksa!")

	return nil
}

// Membuat tabel pencatatan migration
func createMigrationTable(db *sql.DB) error {

	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			id BIGSERIAL PRIMARY KEY,
			migration_name VARCHAR(255) NOT NULL UNIQUE,
			executed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
	`

	if _, err := db.Exec(query); err != nil {
		return fmt.Errorf(
			"gagal membuat tabel schema_migrations: %w",
			err,
		)
	}

	return nil
}

// Memeriksa apakah migration sudah pernah dijalankan
func isMigrationExecuted(
	db *sql.DB,
	migrationName string,
) (bool, error) {

	var exists bool

	query := `
		SELECT EXISTS (
			SELECT 1
			FROM schema_migrations
			WHERE migration_name = $1
		)
	`

	err := db.QueryRow(query, migrationName).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

// Mencari direktori migration
func resolveMigrationDir() (string, error) {

	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf(
			"gagal mendapatkan working directory: %w",
			err,
		)
	}

	for i := 0; i < 10; i++ {

		candidate := filepath.Join(dir, "migrations", ".up")

		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate, nil
		}

		parent := filepath.Dir(dir)

		if parent == dir {
			break
		}

		dir = parent
	}

	return "", fmt.Errorf(
		"direktori migrasi tidak ditemukan: %s",
		filepath.Join("migrations", ".up"),
	)
}
