package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lancer/log/internal/auth"
	"github.com/lancer/log/internal/config"
	"github.com/lancer/log/internal/db"
	"github.com/lancer/log/internal/migrate"
	"github.com/lancer/log/internal/repo"
	"github.com/lancer/log/internal/server"
)

func main() {
	log.SetFlags(log.Ltime | log.Lmsgprefix)
	log.SetPrefix("lancer.log • ")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	// 1) apply migrations
	if err := migrate.Run(ctx, cfg.DatabaseURL); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	// 2) bootstrap admin user from env (idempotent)
	if err := bootstrapAdmin(ctx, cfg.DatabaseURL); err != nil {
		log.Printf("bootstrap admin: %v (skipping)", err)
	}
	cancel()

	// 3) build server
	srv, err := server.New(context.Background(), cfg)
	if err != nil {
		log.Fatalf("server: %v", err)
	}
	defer srv.Close()

	log.Printf("listening on %s — admin at /admin, api at /api", cfg.HTTPAddr)
	go func() { _ = srv.Run() }()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	log.Println("shutting down…")
}

func bootstrapAdmin(ctx context.Context, dsn string) error {
	d, err := db.New(ctx, dsn)
	if err != nil {
		return err
	}
	defer d.Close()
	n, err := repo.CountUsers(ctx, d.Pool)
	if err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	user := os.Getenv("ADMIN_USERNAME")
	pass := os.Getenv("ADMIN_PASSWORD")
	if user == "" {
		user = "admin"
	}
	if pass == "" {
		pass = randomPassword()
		log.Printf("no ADMIN_PASSWORD set; created admin %q with password %q (change immediately!)", user, pass)
	}
	hash, err := auth.HashPassword(pass)
	if err != nil {
		return err
	}
	_, err = repo.CreateUser(ctx, d.Pool, user, hash, user)
	if err != nil {
		return err
	}
	log.Printf("created initial admin user %q", user)
	return nil
}

func randomPassword() string {
	return fmt.Sprintf("change-me-%d", time.Now().Unix()%100000)
}

var _ = errors.New