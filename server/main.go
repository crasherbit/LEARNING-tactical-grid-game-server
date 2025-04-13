package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

// Strutture dati condivise con il client
type GridPosition struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type PlayerState struct {
	PlayerId    string       `json:"playerId"`
	Name        string       `json:"name"`
	Health      int          `json:"health"`
	MaxHealth   int          `json:"maxHealth"`
	ActionPoints int         `json:"actionPoints"`
	AbilityIds  []string     `json:"abilityIds"`
	Position    GridPosition `json:"position"`
}

type BaseEntity struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Position GridPosition `json:"position"`
	Health   int          `json:"health"`
	MaxHealth int         `json:"maxHealth"`
	OwnerId  string       `json:"ownerId"`
}

type GameState struct {
	GameId         string        `json:"gameId"`
	Players        []PlayerState `json:"players"`
	CurrentTurn    int           `json:"currentTurn"`
	CurrentPlayerId string       `json:"currentPlayerId"`
	Status         string        `json:"status"`
	Entities       []BaseEntity  `json:"entities"`
}

type GameInfo struct {
	ID          string `json:"id"`
	PlayerCount int    `json:"playerCount"`
	MaxPlayers  int    `json:"maxPlayers"`
	Status      string `json:"status"`
}

type GameAction struct {
	GameId        string       `json:"gameId"`
	PlayerId      string       `json:"playerId"`
	Type          string       `json:"type"`
	StartPosition GridPosition `json:"startPosition"`
	TargetPosition GridPosition `json:"targetPosition"`
	AbilityId     string       `json:"abilityId"`
	TargetIds     []string     `json:"targetIds"`
}

// Gestore delle partite
type GameManager struct {
	games map[string]*GameState
	mutex sync.RWMutex
}

func NewGameManager() *GameManager {
	return &GameManager{
		games: make(map[string]*GameState),
	}
}

// Crea una nuova partita
func (gm *GameManager) CreateGame() string {
	gm.mutex.Lock()
	defer gm.mutex.Unlock()
	
	// Genera un ID casuale per la partita
	rand.Seed(time.Now().UnixNano())
	gameId := fmt.Sprintf("game_%d", rand.Intn(10000))
	
	// Inizializza un nuovo stato di gioco
	game := &GameState{
		GameId:      gameId,
		Players:     []PlayerState{},
		CurrentTurn: 0,
		Status:      "WaitingForPlayers",
		Entities:    []BaseEntity{},
	}
	
	gm.games[gameId] = game
	return gameId
}

// Aggiunge un giocatore alla partita
func (gm *GameManager) JoinGame(gameId string) bool {
	gm.mutex.Lock()
	defer gm.mutex.Unlock()
	
	game, exists := gm.games[gameId]
	if !exists {
		return false
	}
	
	if game.Status != "WaitingForPlayers" || len(game.Players) >= 2 {
		return false
	}
	
	// Crea un nuovo giocatore
	playerId := fmt.Sprintf("player%d", len(game.Players)+1)
	
	// Posizione iniziale basata su quale giocatore è
	startX := 0
	startY := 0
	if playerId == "player2" {
		startX = 9 // Metti il secondo player dall'altra parte della griglia
		startY = 9
	}
	
	player := PlayerState{
		PlayerId:    playerId,
		Name:        playerId,
		Health:      100,
		MaxHealth:   100,
		ActionPoints: 5,
		AbilityIds:  []string{"fireball", "heal"},
		Position: GridPosition{
			X: startX,
			Y: startY,
		},
	}
	
	game.Players = append(game.Players, player)
	
	// Aggiungi il giocatore anche come entità
	entity := BaseEntity{
		ID:       playerId,
		Type:     "player",
		Position: player.Position,
		Health:   player.Health,
		MaxHealth: player.MaxHealth,
		OwnerId:  playerId,
	}
	
	game.Entities = append(game.Entities, entity)
	
	// Se ora abbiamo 2 giocatori, inizia la partita
	if len(game.Players) == 2 {
		game.Status = "InProgress"
		game.CurrentPlayerId = "player1"
		game.CurrentTurn = 1
	}
	
	return true
}

// Ottiene lo stato di una partita
func (gm *GameManager) GetGameState(gameId string) *GameState {
	gm.mutex.RLock()
	defer gm.mutex.RUnlock()
	
	game, exists := gm.games[gameId]
	if !exists {
		return nil
	}
	
	return game
}

// Processa un'azione di gioco
func (gm *GameManager) ProcessAction(action GameAction) (bool, string) {
	gm.mutex.Lock()
	defer gm.mutex.Unlock()
	
	game, exists := gm.games[action.GameId]
	if !exists {
		return false, "Partita non trovata"
	}
	
	// Verifica che sia il turno del giocatore
	if game.CurrentPlayerId != action.PlayerId {
		return false, "Non è il tuo turno"
	}
	
	// Trova il giocatore nell'array
	var playerIndex = -1
	for i, p := range game.Players {
		if p.PlayerId == action.PlayerId {
			playerIndex = i
			break
		}
	}
	
	if playerIndex == -1 {
		return false, "Giocatore non trovato"
	}
	
	player := &game.Players[playerIndex]
	
	// Gestisci i diversi tipi di azioni
	switch action.Type {
	case "Move":
		return gm.handleMoveAction(game, player, action)
	case "UseAbility":
		return gm.handleAbilityAction(game, player, action)
	case "EndTurn":
		return gm.handleEndTurnAction(game, player)
	default:
		return false, "Tipo di azione non riconosciuto"
	}
}

// Gestisce l'azione di movimento
func (gm *GameManager) handleMoveAction(game *GameState, player *PlayerState, action GameAction) (bool, string) {
	// Verifica punti azione
	if player.ActionPoints < 1 {
		return false, "Punti azione insufficienti"
	}
	
	// Verifica che il movimento sia valido (solo celle adiacenti)
	startX, startY := player.Position.X, player.Position.Y
	targetX, targetY := action.TargetPosition.X, action.TargetPosition.Y
	
	deltaX := abs(targetX - startX)
	deltaY := abs(targetY - startY)
	
	if (deltaX == 1 && deltaY == 0) || (deltaX == 0 && deltaY == 1) {
		// Movimento valido
		player.Position.X = targetX
		player.Position.Y = targetY
		player.ActionPoints--
		
		// Aggiorna anche l'entità corrispondente
		for i := range game.Entities {
			if game.Entities[i].ID == player.PlayerId {
				game.Entities[i].Position = player.Position
				break
			}
		}
		
		return true, "Movimento completato"
	}
	
	return false, "Movimento non valido"
}

// Gestisce l'uso di un'abilità
func (gm *GameManager) handleAbilityAction(game *GameState, player *PlayerState, action GameAction) (bool, string) {
	// In una versione completa qui avremmo la logica per verificare validità, range, ecc.
	// Questo è un esempio semplificato
	
	// Costo fisso per tutte le abilità in questo esempio
	abilityCost := 2
	
	if player.ActionPoints < abilityCost {
		return false, "Punti azione insufficienti"
	}
	
	// Gestisci tipi di abilità differenti
	if action.AbilityId == "fireball" {
		// Cerca un bersaglio in quella posizione
		for i := range game.Entities {
			if game.Entities[i].Position.X == action.TargetPosition.X &&
				game.Entities[i].Position.Y == action.TargetPosition.Y &&
				game.Entities[i].ID != player.PlayerId {
				// Colpisci il bersaglio
				game.Entities[i].Health -= 20
				player.ActionPoints -= abilityCost
				return true, "Fireball lanciata con successo"
			}
		}
		return false, "Nessun bersaglio valido trovato"
	} else if action.AbilityId == "heal" {
		// Cerca un'entità alleata
		for i := range game.Entities {
			if game.Entities[i].Position.X == action.TargetPosition.X &&
				game.Entities[i].Position.Y == action.TargetPosition.Y {
				// Cura il bersaglio
				game.Entities[i].Health += 25
				if game.Entities[i].Health > game.Entities[i].MaxHealth {
					game.Entities[i].Health = game.Entities[i].MaxHealth
				}
				// Aggiorna anche il player state se è un giocatore
				for j := range game.Players {
					if game.Players[j].PlayerId == game.Entities[i].ID {
						game.Players[j].Health = game.Entities[i].Health
						break
					}
				}
				player.ActionPoints -= abilityCost
				return true, "Cura completata con successo"
			}
		}
		return false, "Nessun bersaglio valido trovato"
	}
	
	return false, "Abilità non riconosciuta"
}

// Gestisce la fine del turno
func (gm *GameManager) handleEndTurnAction(game *GameState, player *PlayerState) (bool, string) {
	// Cambia il giocatore corrente
	if game.CurrentPlayerId == "player1" {
		game.CurrentPlayerId = "player2"
	} else {
		game.CurrentPlayerId = "player1"
		game.CurrentTurn++
	}
	
	// Ripristina i punti azione per il nuovo giocatore
	for i := range game.Players {
		if game.Players[i].PlayerId == game.CurrentPlayerId {
			game.Players[i].ActionPoints = 5
			break
		}
	}
	
	return true, "Turno terminato"
}

// Ottiene la lista delle partite
func (gm *GameManager) GetGamesList() []GameInfo {
	gm.mutex.RLock()
	defer gm.mutex.RUnlock()
	
	var games []GameInfo
	
	for id, game := range gm.games {
		info := GameInfo{
			ID:          id,
			PlayerCount: len(game.Players),
			MaxPlayers:  2,
			Status:      game.Status,
		}
		games = append(games, info)
	}
	
	return games
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func main() {
	gameManager := NewGameManager()
	
	// API per creare una nuova partita
	http.HandleFunc("/api/game/create", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Metodo non supportato", http.StatusMethodNotAllowed)
			return
		}
		
		gameId := gameManager.CreateGame()
		w.Write([]byte(gameId))
	})
	
	// API per unirsi a una partita
	http.HandleFunc("/api/game/", func(w http.ResponseWriter, r *http.Request) {
		// Estrai l'ID della partita e l'azione dall'URL
		// URL formato: /api/game/{gameId}/action
		path := r.URL.Path[9:] // Rimuovi "/api/game/"
		
		if path == "" {
			http.Error(w, "Percorso non valido", http.StatusBadRequest)
			return
		}
		
		// Trova separatore
		var gameId string
		var action string
		
		for i := 0; i < len(path); i++ {
			if path[i] == '/' {
				gameId = path[:i]
				action = path[i+1:]
				break
			}
		}
		
		if gameId == "" {
			http.Error(w, "ID partita non valido", http.StatusBadRequest)
			return
		}
		
		switch action {
		case "join":
			if r.Method != http.MethodPost {
				http.Error(w, "Metodo non supportato", http.StatusMethodNotAllowed)
				return
			}
			
			success := gameManager.JoinGame(gameId)
			if !success {
				http.Error(w, "Impossibile unirsi alla partita", http.StatusBadRequest)
				return
			}
			
			w.WriteHeader(http.StatusOK)
			
		case "state":
			if r.Method != http.MethodGet {
				http.Error(w, "Metodo non supportato", http.StatusMethodNotAllowed)
				return
			}
			
			game := gameManager.GetGameState(gameId)
			if game == nil {
				http.Error(w, "Partita non trovata", http.StatusNotFound)
				return
			}
			
			// Converte lo stato in JSON
			jsonData, err := json.Marshal(game)
			if err != nil {
				http.Error(w, "Errore interno del server", http.StatusInternalServerError)
				return
			}
			
			w.Header().Set("Content-Type", "application/json")
			w.Write(jsonData)
			
		case "action":
			if r.Method != http.MethodPost {
				http.Error(w, "Metodo non supportato", http.StatusMethodNotAllowed)
				return
			}
			
			// Decodifica l'azione dal corpo della richiesta
			var action GameAction
			decoder := json.NewDecoder(r.Body)
			if err := decoder.Decode(&action); err != nil {
				http.Error(w, "Richiesta non valida", http.StatusBadRequest)
				return
			}
			
			// Verifica che l'ID partita corrisponda
			if action.GameId != gameId {
				http.Error(w, "ID partita non corrispondente", http.StatusBadRequest)
				return
			}
			
			// Processa l'azione
			success, message := gameManager.ProcessAction(action)
			if !success {
				http.Error(w, message, http.StatusBadRequest)
				return
			}
			
			// Risposta di successo
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(message))
			
		default:
			http.Error(w, "Azione non riconosciuta", http.StatusBadRequest)
		}
	})
	
	// API per ottenere la lista delle partite
	http.HandleFunc("/api/games", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Metodo non supportato", http.StatusMethodNotAllowed)
			return
		}
		
		games := gameManager.GetGamesList()
		
		// Converte la lista in JSON
		jsonData, err := json.Marshal(games)
		if err != nil {
			http.Error(w, "Errore interno del server", http.StatusInternalServerError)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)
	})
	
	// Abilita CORS per lo sviluppo
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Authorization")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		http.Error(w, "Endpoint non trovato", http.StatusNotFound)
	})
	
	fmt.Println("Server avviato sulla porta 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}