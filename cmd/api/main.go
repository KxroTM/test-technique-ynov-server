// Commande api : point d'entrée du serveur backend.
//
// Ce fichier ne contient aucune logique métier. Son unique rôle est
// d'assembler les briques de l'application (configuration, base de données,
// routeur HTTP) puis de démarrer le serveur.
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

	"github.com/KxroTM/test-technique-ynov/internal/config"
	"github.com/KxroTM/test-technique-ynov/internal/database"
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

	// 3. Routeur HTTP.
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	// Sonde de disponibilité, utile pour Docker et pour vérifier
	// rapidement que le serveur répond.
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// 4. Démarrage du serveur.
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

	// 5. Arrêt propre : on attend un signal d'interruption, puis on laisse
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
