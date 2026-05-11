package main

import (
	"flag"
	"log/slog"
	"os"

	"github.com/cyandie/backend/internal/core"
	"github.com/cyandie/backend/internal/core/cache"
	"github.com/cyandie/backend/internal/core/config"
	"github.com/cyandie/backend/internal/core/database"
	"github.com/cyandie/backend/internal/core/health"
	"github.com/cyandie/backend/internal/core/logger"
	"github.com/cyandie/backend/internal/core/middleware"
	"github.com/cyandie/backend/internal/core/server"
	"github.com/go-chi/chi/v5"
)

func main() {
	configPath := flag.String("config", "", "path to config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		slog.Error("load config", "error", err)
		os.Exit(1)
	}

	log := logger.New(cfg.Logger)
	slog.SetDefault(log)

	db, err := database.New(cfg.Database)
	if err != nil {
		log.Error("connect database", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	log.Info("database connected")

	rdb, err := cache.New(cfg.Cache)
	if err != nil {
		log.Error("connect redis", "error", err)
		os.Exit(1)
	}
	defer rdb.Close()
	log.Info("redis connected")

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.Recovery)
	router.Use(middleware.Logger(log))

	httpSrv := server.NewHTTPServer(cfg.Server.HTTPAddr, router)

	app := core.NewApp()
	app.SetLogger(log)

	healthHandler := health.NewHandler()
	app.Register(healthHandler)
	healthHandler.RegisterRoutes(router)

	log.Info("starting server", "addr", cfg.Server.HTTPAddr)

	go func() {
		if err := httpSrv.ListenAndServe(); err != nil {
			slog.Error("http server error", "error", err)
		}
	}()

	if err := app.Run(); err != nil {
		log.Error("app run", "error", err)
		os.Exit(1)
	}

	httpSrv.Close()
	log.Info("server stopped")
}
