package api

import (
	"encoding/json"
	"net/http"

	"tactical-game/pkg/game"
)

// Server gestisce le richieste HTTP
type Server struct {
	gameManager *game.GameManager
}

// NewServer crea un nuovo server API
func NewServer(gm *game.GameManager) *Server {
	return &Server{
		gameManager: gm,
	}
}

// CreateGameHandler gestisce la richiesta di creazione di una nuova partita
func (s *Server) CreateGameHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Metodo non supportato", http.StatusMethodNotAllowed)
		return
	}
	
	gameID := s.gameManager.CreateGame()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(gameID))
}

// JoinGameHandler gestisce la richiesta di unirsi a una partita
func (s *Server) JoinGameHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Metodo non supportato", http.StatusMethodNotAllowed)
		return
	}
	
	// Estrai gameID dall'URL
	gameID := r.URL.Path[len("/api/game/"):len(r.URL.Path)-len("/join")]
	if len(gameID) == 0 {
		http.Error(w, "ID partita mancante", http.StatusBadRequest)
		return
	}
	
	// Per ora generiamo un ID giocatore casuale
	// In una implementazione reale, dovresti usare l'autenticazione
	playerID := "player" + gameID[len(gameID)-1:]
	playerName := "Player " + playerID
	
	err := s.gameManager.AddPlayer(gameID, playerID, playerName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	w.WriteHeader(http.StatusOK)
}

// GameStateHandler gestisce la richiesta di ottenere lo stato di una partita
func (s *Server) GameStateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Metodo non supportato", http.StatusMethodNotAllowed)
		return
	}
	
	// Estrai gameID dall'URL
	gameID := r.URL.Path[len("/api/game/"):len(r.URL.Path)-len("/state")]
	
	gameState, err := s.gameManager.GetGame(gameID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(gameState)
}

// ActionHandler gestisce le azioni di gioco
func (s *Server) ActionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Metodo non supportato", http.StatusMethodNotAllowed)
		return
	}
	
	// Estrai gameID dall'URL
	gameID := r.URL.Path[len("/api/game/"):len(r.URL.Path)-len("/action")]
	
	var action game.GameAction
	err := json.NewDecoder(r.Body).Decode(&action)
	if err != nil {
		http.Error(w, "JSON non valido: "+err.Error(), http.StatusBadRequest)
		return
	}
	
	// Impostiamo l'ID della partita dall'URL
	action.GameID = gameID
	
	err = s.gameManager.ProcessAction(action)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	w.WriteHeader(http.StatusOK)
}

// ListGamesHandler gestisce la richiesta di elenco partite
func (s *Server) ListGamesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Metodo non supportato", http.StatusMethodNotAllowed)
		return
	}
	
	gamesList := s.gameManager.ListGames()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(gamesList)
}

// SetupRoutes configura i gestori delle rotte
func (s *Server) SetupRoutes() http.Handler {
	mux := http.NewServeMux()
	
	// Imposta CORS per consentire richieste da Unity
	corsMiddleware := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}
			
			next(w, r)
		}
	}
	
	mux.HandleFunc("/api/games", corsMiddleware(s.ListGamesHandler))
	mux.HandleFunc("/api/game/create", corsMiddleware(s.CreateGameHandler))
	mux.HandleFunc("/api/game/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		
		if len(path) > len("/api/game/") {
			if path[len(path)-len("/state"):] == "/state" {
				corsMiddleware(s.GameStateHandler)(w, r)
			} else if path[len(path)-len("/action"):] == "/action" {
				corsMiddleware(s.ActionHandler)(w, r)
			} else if path[len(path)-len("/join"):] == "/join" {
				corsMiddleware(s.JoinGameHandler)(w, r)
			} else {
				http.NotFound(w, r)
			}
		} else {
			http.NotFound(w, r)
		}
	})
	
	return mux
}