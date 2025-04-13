package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"sync"

	"slices"

	"github.com/gorilla/websocket"
)

// Strutture dati per lobby e connessioni
type Lobby struct {
	ID          string
	Players     []string
	CurrentTurn int // Indice del giocatore il cui turno è attivo
}

type Message struct {
	EventName string `json:"eventName"`
	Data      any    `json:"data"`
}

var lobbies = make(map[string]*Lobby)
var playerConnections = make(map[string]*websocket.Conn)
var lobbyMutex sync.Mutex

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	http.HandleFunc("/ws", handleConnections)
	fmt.Println("Server avviato su :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Errore avviando il server:", err)
	}
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Errore durante l'upgrade:", err)
		return
	}
	defer ws.Close()

	playerID := r.RemoteAddr
	playerConnections[playerID] = ws
	fmt.Printf("Giocatore connesso: %s\n", playerID)

	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			fmt.Printf("Errore leggendo il messaggio dal giocatore %s: %v\n", playerID, err)
			delete(playerConnections, playerID)
			break
		}
		fmt.Printf("Messaggio ricevuto da %s: %s\n", playerID, string(msg))

		var message map[string]any
		if err := json.Unmarshal(msg, &message); err != nil {
			fmt.Printf("Errore nel parsing del messaggio JSON: %v\n", err)
			continue
		}

		switch message["eventName"] {
		case "matchmaking_request":
			handleMatchmakingRequest(playerID)
		case "end_turn":
			handleEndTurn(playerID)
		default:
			fmt.Printf("Messaggio non riconosciuto: %s\n", string(msg))
		}
	}
}

func handleMatchmakingRequest(playerID string) {
	lobbyMutex.Lock()
	defer lobbyMutex.Unlock()

	// Cerca una lobby disponibile
	for _, lobby := range lobbies {
		if len(lobby.Players) < 2 {
			// Aggiungi il giocatore alla lobby
			lobby.Players = append(lobby.Players, playerID)
			fmt.Printf("Giocatore %s unito alla lobby %s\n", playerID, lobby.ID)
			// randomizza chi inizia per primo, random da 0 a 1
			lobby.CurrentTurn = rand.Intn(2)
			for _, playerID := range lobby.Players {
				notifyPlayer(playerID, Message{EventName: "lobby_ready", Data: map[string]string{"lobby_id": lobby.ID}})
			}
			return
		}
	}

	// Crea una nuova lobby se non ne esistono di disponibili
	newLobby := &Lobby{
		ID:      generateLobbyID(),
		Players: []string{playerID},
	}
	lobbies[newLobby.ID] = newLobby
	fmt.Printf("Lobby %s creata per il giocatore %s\n", newLobby.ID, playerID)

	notifyPlayer(playerID, Message{EventName: "lobby_created", Data: map[string]string{"lobby_id": newLobby.ID}})
}

// Funzione per inviare notifiche ai giocatori
func notifyPlayer(playerID string, response Message) {
	conn, exists := playerConnections[playerID]
	if exists {

		responseJSON, _ := json.Marshal(response)
		err := conn.WriteMessage(websocket.TextMessage, responseJSON)
		if err != nil {
			fmt.Printf("Errore inviando il messaggio al giocatore %s: %v\n", playerID, err)
		} else {
			fmt.Printf("Notifica inviata al giocatore %s notifica: %s\n", playerID, response)
		}
	}
}

func generateLobbyID() string {
	return fmt.Sprintf("lobby-%d", len(lobbies)+1)
}

func handleEndTurn(playerID string) {
	lobbyMutex.Lock()
	defer lobbyMutex.Unlock()
	lobbyID := ""
	// Trova la lobby del giocatore
	for _, lobby := range lobbies {
		if slices.Contains(lobby.Players, playerID) {
			lobbyID = lobby.ID
		}
	}
	lobby, exists := lobbies[lobbyID]
	if !exists {
		fmt.Printf("Lobby %s non trovata\n", lobbyID)
		return
	}

	// Controlla se è il turno del giocatore
	if lobby.Players[lobby.CurrentTurn] != playerID {
		fmt.Printf("Non è il turno del giocatore %s\n", playerID)
		return
	}

	// Passa al prossimo turno
	lobby.CurrentTurn = (lobby.CurrentTurn + 1) % len(lobby.Players)
	fmt.Printf("Turno cambiato nella lobby %s. Ora è il turno del giocatore %s\n", lobbyID, lobby.Players[lobby.CurrentTurn])

	// Notifica i giocatori
	for _, playerID := range lobby.Players {
		notifyPlayer(playerID, Message{EventName: "turn_changed", Data: map[string]any{"current_turn": lobby.Players[lobby.CurrentTurn], "is_my_turn": playerID == lobby.Players[lobby.CurrentTurn]}})

	}
}
