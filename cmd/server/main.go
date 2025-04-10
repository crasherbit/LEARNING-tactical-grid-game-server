package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"tactical-game/pkg/api"
	"tactical-game/pkg/game"
)

func main() {
	// Ottieni la porta dalle variabili d'ambiente o usa la porta predefinita
	port := 8080
	if envPort := os.Getenv("PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil {
			port = p
		}
	}
	
	// Crea il game manager
	gameManager := game.NewGameManager()
	
	// Crea il server API
	server := api.NewServer(gameManager)
	
	// Configura le rotte
	router := server.SetupRoutes()
	
	// Avvia il server
	addr := fmt.Sprintf(":%d", port)
	log.Printf("Server in ascolto su http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, router))
}