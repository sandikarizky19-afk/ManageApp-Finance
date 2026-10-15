package main

import (
	"context"
	"database/sql"
	"finance-service/internal/controller"
	"finance-service/internal/handler/server"
	"finance-service/internal/migration"
	"finance-service/internal/utils"
	"finance-service/repository"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg := controller.LoadConfig()

	//koneksi database
	db, err := repository.NewPostgresDB(cfg)
	if err != nil {
		log.Fatalf("Gagal terhubung ke database: %v", err)
	}
	defer db.Close()
	log.Println("Berhasil terhubung ke database PostgresSQL!")

	//matikan jika migrasi tidak digunakan
	// migrationDatabase(db)

	jwtManager, err := utils.NewJwtManager()
	if err != nil {
		log.Fatalf("Gagal inisialisasi JWT Manager: %v", err)
	}

	srv := server.New(db, jwtManager)
	host := cfg.AppHost
	if host == "" {
		host = "0.0.0.0"
	}
	address := net.JoinHostPort(host, cfg.AppPort)

	go func() {
		if err := srv.Listen(address); err != nil {
			log.Fatalf("Gagal menjalankan server: %v", err)
		}
	}()
	log.Printf("Server finance berjalan di %s\n", address)

	//shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Mematikan Server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.App.ShutdownWithContext(ctx); err != nil {
		log.Fatalf("Gagal shutdown server dengan rapi: %v", err)
	}
	log.Println("Server finance berhasil dimatikan.")
}

func migrationDatabase(db *sql.DB) {
	if err := migration.RunMigrations(db); err != nil {
		log.Fatalf("Gagal menjalankan migrasi database: %v", err)
	}

	log.Println("Migrasi database selesai.")
}
