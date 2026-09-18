package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/KxroTM/test-technique-ynov/internal/auth"
	"github.com/KxroTM/test-technique-ynov/internal/config"
	"github.com/KxroTM/test-technique-ynov/internal/database"
	"github.com/KxroTM/test-technique-ynov/internal/handlers"
	"github.com/KxroTM/test-technique-ynov/internal/middleware"
	"github.com/KxroTM/test-technique-ynov/internal/repository"
	"github.com/KxroTM/test-technique-ynov/internal/service"
)

func main() {
	// 1. Configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("configuration invalide : %v", err)
	}

	// 2. Base de données
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connexion à la base de données : %v", err)
	}
	defer db.Close()
	log.Println("connexion à la base de données établie")

	// 3. Assemblage des couches
	tokenManager := auth.NewTokenManager(cfg.JWTSecret, cfg.JWTExpiration)

	userRepository := repository.NewUserRepository(db)
	spaceRepository := repository.NewSpaceRepository(db)
	noteRepository := repository.NewNoteRepository(db)

	var googleExchanger *auth.GoogleExchanger
	if cfg.GoogleEnabled() {
		googleExchanger = auth.NewGoogleExchanger(cfg.GoogleClientID, cfg.GoogleClientSecret)
		log.Println("connexion Google activée")
	}

	authService := service.NewAuthService(userRepository, tokenManager, googleExchanger)
	spaceService := service.NewSpaceService(spaceRepository)
	noteService := service.NewNoteService(noteRepository, spaceRepository)

	authHandler := handlers.NewAuthHandler(authService)
	spaceHandler := handlers.NewSpaceHandler(spaceService)
	noteHandler := handlers.NewNoteHandler(noteService)

	// 4. Routeur HTTP
	handlers.ConfigureValidation()

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	// endpoint de disponibilité
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := router.Group("/api")

	// Routes publiques
	api.POST("/auth/register", authHandler.Register)
	api.POST("/auth/login", authHandler.Login)
	api.POST("/auth/google", authHandler.LoginWithGoogle)

	// Routes protégées
	protected := api.Group("")
	protected.Use(middleware.Authenticate(tokenManager))
	{
		protected.GET("/me", authHandler.Me)

		protected.GET("/spaces", spaceHandler.List)
		protected.POST("/spaces", spaceHandler.Create)
		protected.GET("/spaces/:spaceID", spaceHandler.Get)
		protected.PUT("/spaces/:spaceID", spaceHandler.Update)
		protected.DELETE("/spaces/:spaceID", spaceHandler.Delete)

		protected.GET("/spaces/:spaceID/notes", noteHandler.ListBySpace)
		protected.POST("/spaces/:spaceID/notes", noteHandler.Create)
		protected.GET("/notes/:noteID", noteHandler.Get)
		protected.PUT("/notes/:noteID", noteHandler.Update)
		protected.DELETE("/notes/:noteID", noteHandler.Delete)
	}

	// 5. Démarrage du serveur
	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("serveur démarré sur le port %s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("arrêt inattendu du serveur : %v", err)
		}
	}()

	// 6. Arrêt
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("arrêt du serveur en cours...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("arrêt forcé du serveur : %v", err)
	}
	log.Println("serveur arrêté")
}
