package main

import (
	"context"
	"fileshare-be/internal/auth"
	"fileshare-be/internal/handlers"
	"fileshare-be/internal/middleware"
	"fileshare-be/internal/models"
	"fileshare-be/internal/services"
	"fileshare-be/pkg/config"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic("failed to load config: " + err.Error())
	}

	logger, err := config.NewLogger(cfg.AppEnv)
	if err != nil {
		panic("failed to create logger: " + err.Error())
	}
	defer logger.Sync()

	db, err := config.NewDatabase(cfg)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}

	s3Client, presignClient, err := config.NewS3Client(cfg)
	if err != nil {
		logger.Fatal("failed to create S3 client", zap.Error(err))
	}

	jwtMgr := auth.NewJWTManager(cfg.JWTSecret, cfg.RefreshSecret)

	auditService := services.NewAuditService(db, logger)
	authService := services.NewAuthService(db, jwtMgr, auditService, logger)
	mfaService := services.NewMFAService(db, auditService, logger)
	viewService := services.NewViewService(db, logger)
	docService := services.NewDocumentService(db, s3Client, presignClient, cfg.S3Bucket, cfg.AppKey, auditService, viewService, logger)
	adminService := services.NewAdminService(db)
	groupService := services.NewGroupService(db, auditService, logger)
	shareService := services.NewShareService(db, auditService, logger)

	authHandler := handlers.NewAuthHandler(authService, mfaService)
	docHandler := handlers.NewDocumentHandler(docService, viewService)
	adminHandler := handlers.NewAdminHandler(adminService)
	groupHandler := handlers.NewGroupHandler(groupService)
	shareHandler := handlers.NewShareHandler(shareService)

	r := chi.NewRouter()

	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(middleware.RequestLogger(logger))
	r.Use(chimw.Recoverer)
	r.Use(middleware.SecureHeaders)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{cfg.FrontendOrigin},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	r.Route("/api/auth", func(r chi.Router) {
		rateLimiter := httprate.LimitByIP(10, time.Minute)

		r.With(rateLimiter).Post("/register", authHandler.Register)
		r.With(rateLimiter).Post("/login", authHandler.Login)
		r.Post("/refresh", authHandler.RefreshToken)

		r.Group(func(r chi.Router) {
			r.Use(middleware.Authenticate(jwtMgr))
			r.Post("/logout", authHandler.Logout)
			r.Post("/mfa/setup", authHandler.SetupMFA)
			r.Post("/mfa/verify", authHandler.VerifyMFA)
		})
	})

	r.Route("/api/documents", func(r chi.Router) {
		r.Use(middleware.Authenticate(jwtMgr))
		r.Post("/upload", docHandler.InitiateUpload)
		r.Get("/", docHandler.ListDocuments)

		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", docHandler.GetDownloadURL)
			r.Delete("/", docHandler.DeleteDocument)
			r.Put("/secret-key", docHandler.SetSecretKey)
			r.Delete("/secret-key", docHandler.RemoveSecretKey)
			r.Get("/views", docHandler.GetDocumentViews)

			r.Get("/shares", shareHandler.ListDocumentShares)
			r.Post("/shares", shareHandler.ShareWithUser)
			r.Delete("/shares/{shareId}", shareHandler.RevokeUserShare)

			r.Get("/group-shares", shareHandler.ListDocumentGroupShares)
			r.Post("/group-shares", shareHandler.ShareWithGroup)
			r.Delete("/group-shares/{shareId}", shareHandler.RevokeGroupShare)
		})
	})

	r.Route("/api/groups", func(r chi.Router) {
		r.Use(middleware.Authenticate(jwtMgr))
		r.Get("/", groupHandler.ListGroups)
		r.Post("/", groupHandler.CreateGroup)

		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", groupHandler.GetGroup)
			r.Delete("/", groupHandler.DeleteGroup)
			r.Post("/members", groupHandler.AddMember)
			r.Delete("/members/{userId}", groupHandler.RemoveMember)
		})
	})

	r.Route("/api/admin", func(r chi.Router) {
		r.Use(middleware.Authenticate(jwtMgr))
		r.Use(middleware.RequireRole(models.RoleStaff, models.RoleAdmin))
		r.Get("/users", adminHandler.ListUsers)
		r.Get("/audit-logs", adminHandler.GetAuditLogs)
	})

	c := cron.New()
	c.AddFunc("0 3 * * *", func() {
		ctx := context.Background()
		if err := docService.CleanupExpired(ctx); err != nil {
			logger.Error("expired document cleanup failed", zap.Error(err))
		}
	})
	c.Start()

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("server starting", zap.String("port", cfg.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server failed", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server...")

	c.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("server forced to shutdown", zap.Error(err))
	}

	logger.Info("server stopped")
}
