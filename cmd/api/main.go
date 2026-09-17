// Commande api : point d'entrée du serveur backend.
//
// Ce fichier ne contient aucune logique métier. Son rôle est d'assembler les
// briques de l'application dans l'ordre (configuration, base de données,
// repositories, services, handlers, routes) puis de démarrer le serveur.
// L'assemblage est fait ici, explicitement, plutôt que masqué derrière un
// conteneur d'injection de dépendances : la lecture de ce fichier suffit à
// comprendre de quoi dépend quoi.
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
	// 1. Configuration : on refuse de démarrer si elle est incomplète.
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("configuration invalide : %v", err)
	}

	// 2. Base de données.
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connexion à la base de données : %v", err)
	}
	defer db.Close()
	log.Println("connexion à la base de données établie")

	// 3. Assemblage des couches, de la plus basse à la plus haute.
	tokenManager := auth.NewTokenManager(cfg.JWTSecret, cfg.JWTExpiration)

	userRepository := repository.NewUserRepository(db)
	spaceRepository := repository.NewSpaceRepository(db)

	authService := service.NewAuthService(userRepository, tokenManager)
	spaceService := service.NewSpaceService(spaceRepository)

	authHandler := handlers.NewAuthHandler(authService)
	spaceHandler := handlers.NewSpaceHandler(spaceService)

	// 4. Routeur HTTP.
	handlers.ConfigureValidation()

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	// Sonde de disponibilité, utile pour Docker et pour vérifier
	// rapidement que le serveur répond.
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := router.Group("/api")

	// Routes publiques : ce sont les seules accessibles sans jeton.
	api.POST("/auth/register", authHandler.Register)
	api.POST("/auth/login", authHandler.Login)

	// Routes protégées. Le middleware est appliqué au groupe entier : toute
	// route ajoutée ici est authentifiée par construction, il n'y a pas de
	// risque d'oublier la protection sur un nouvel endpoint.
	protected := api.Group("")
	protected.Use(middleware.Authenticate(tokenManager))
	{
		protected.GET("/me", authHandler.Me)

		// FT2 — gestion des espaces.
		protected.GET("/spaces", spaceHandler.List)
		protected.POST("/spaces", spaceHandler.Create)
		protected.GET("/spaces/:spaceID", spaceHandler.Get)
		protected.PUT("/spaces/:spaceID", spaceHandler.Update)
		protected.DELETE("/spaces/:spaceID", spaceHandler.Delete)
	}

	// 5. Démarrage du serveur.
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

	// 6. Arrêt propre : on attend un signal d'interruption, puis on laisse
	// aux requêtes en cours le temps de se terminer avant de fermer.
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
