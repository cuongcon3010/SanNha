package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"sannha/be/internal/app"
	"sannha/be/internal/modules/identity"
	"sannha/be/internal/platform/database"
)

func run() error {
	cfg, err := app.ConfigFromEnv()
	if err != nil {
		return err
	}
	email, password := os.Getenv("ADMIN_EMAIL"), os.Getenv("ADMIN_PASSWORD")
	if email == "" || password == "" {
		return fmt.Errorf("set ADMIN_EMAIL and ADMIN_PASSWORD; no default credentials are provided")
	}
	db, err := database.Open(cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer db.Close()
	accounts, err := identity.New(db)
	if err != nil {
		return err
	}
	_, err = accounts.BootstrapAdmin(context.Background(), email, password)
	if err != nil {
		return err
	}
	slog.Info("Administrator created; remove ADMIN_PASSWORD from the environment after use")
	return nil
}
func main() {
	if err := run(); err != nil {
		slog.Error("Cannot create administrator", "error", err)
		os.Exit(1)
	}
}
