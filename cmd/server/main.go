package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"time"

	"github.com/cyandie/backend/internal/auth"
	"github.com/cyandie/backend/internal/core"
	"github.com/cyandie/backend/internal/core/cache"
	"github.com/cyandie/backend/internal/core/config"
	"github.com/cyandie/backend/internal/core/database"
	"github.com/cyandie/backend/internal/core/health"
	"github.com/cyandie/backend/internal/core/logger"
	"github.com/cyandie/backend/internal/core/middleware"
	"github.com/cyandie/backend/internal/core/server"
	"github.com/cyandie/backend/internal/db"
	"github.com/cyandie/backend/internal/users"
	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"
)

// redisLimiterAdapter adapts *redis.Client to the middleware.RedisLimiterClient interface.
type redisLimiterAdapter struct {
	client *redis.Client
}

func (a *redisLimiterAdapter) Incr(ctx context.Context, key string) *redis.IntCmd {
	return a.client.Incr(ctx, key)
}

func (a *redisLimiterAdapter) Expire(ctx context.Context, key string, ttl time.Duration) *redis.BoolCmd {
	return a.client.Expire(ctx, key, ttl)
}

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

	dbConn, err := database.New(cfg.Database)
	if err != nil {
		log.Error("connect database", "error", err)
		os.Exit(1)
	}
	defer dbConn.Close()
	log.Info("database connected")

	rdb, err := cache.New(cfg.Cache)
	if err != nil {
		log.Error("connect redis", "error", err)
		os.Exit(1)
	}
	defer rdb.Close()
	log.Info("redis connected")

	queries := db.New(dbConn)

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.Recovery)

	defaultTimeout, _ := time.ParseDuration(cfg.Timeout.Default)
	if defaultTimeout == 0 {
		defaultTimeout = 30 * time.Second
	}
	router.Use(middleware.Timeout(defaultTimeout))
	router.Use(middleware.Logger(log))

	httpSrv := server.NewHTTPServer(cfg.Server.HTTPAddr, router)

	app := core.NewApp()
	app.SetLogger(log)

	// Load JWT keys from config
	jwtKeys := make([]auth.JWTKey, len(cfg.Auth.JWTKeys))
	for i, k := range cfg.Auth.JWTKeys {
		jwtKeys[i] = auth.JWTKey{KID: k.KID, Secret: []byte(k.Secret)}
	}
	if len(jwtKeys) == 0 {
		jwtKeys = append(jwtKeys, auth.JWTKey{KID: "default", Secret: []byte("change-me-in-production-32byte")})
	}

	usersModule := users.NewModule(queries)
	authModule := auth.NewModule(queries, auth.NewKeyManager(jwtKeys), auth.NewSessionStore(auth.NewRedisAdapter(rdb.Client)))

	app.Register(usersModule)
	app.Register(authModule)

	healthHandler := health.NewHandler()
	healthHandler.RegisterRoutes(router)

	// Rate limited auth routes
	redisAdapter := &redisLimiterAdapter{client: rdb.Client}
	authLimiter := middleware.NewRateLimiter(redisAdapter, middleware.RateLimitConfig{
		Limit:  cfg.RateLimit.Auth.Limit,
		Window: cfg.RateLimit.Auth.Window,
	})
	authRouter := router.With(authLimiter.Middleware("auth"))
	authModule.RegisterRoutes(authRouter)

	// Rate limited user routes
	readLimiter := middleware.NewRateLimiter(redisAdapter, middleware.RateLimitConfig{
		Limit:  cfg.RateLimit.Read.Limit,
		Window: cfg.RateLimit.Read.Window,
	})
	usersRouter := router.With(readLimiter.Middleware("read"))
	usersModule.RegisterRoutes(usersRouter)

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
