package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

var players = make(map[string]*Player)
var playerConnections = make(map[string]*websocket.Conn)

func HandleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Errore durante l'upgrade:", err)
		return
	}
	defer ws.Close()

	playerID := r.RemoteAddr
	player := &Player{
		ID:             playerID,
		CurrentLobbyID: "",
		Name:           playerID,
		HealthPoints:   100,
		ActionPoints:   3,
		MovementPoints: 3,
		PositionX:      0,
		PositionY:      0,
		IsMyTurn:       false,
	}
	players[playerID] = player
	playerConnections[playerID] = ws

	fmt.Printf("Giocatore connesso: %s\n", playerID)

	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			fmt.Printf("Errore leggendo il messaggio dal giocatore %s: %v\n", playerID, err)
			delete(playerConnections, playerID)
			delete(players, playerID)
			break
		}

		var message Message
		if err := json.Unmarshal(msg, &message); err != nil {
			fmt.Printf("Errore nel parsing del messaggio JSON: %v\n", err)
			continue
		}

		switch message.EventName {
		case "matchmaking_request":
			JoinLobby(player)
		case "init_data_request":
			handleInitDataRequest(playerID)
		case "end_turn":
			handleEndTurn(playerID)
		default:
			fmt.Printf("Evento non riconosciuto: %s\n", message.EventName)
		}
	}
}

func handleInitDataRequest(playerID string) {
	player := players[playerID]
	lobby := lobbies[player.CurrentLobbyID]
	if lobby == nil {
		fmt.Printf("Errore: Nessuna lobby trovata per il giocatore %s\n", playerID)
		return
	}

	response := Message{
		EventName: "init_data_response",
		Data: map[string]any{
			"lobby_id":     lobby.ID,
			"players":      lobby.Players,
			"grid":         lobby.Grid,
			"my_turn":      player.IsMyTurn,
			"my_player":    player,
			"my_player_id": player.ID,
		},
	}
	NotifyPlayer(playerID, response)
}

func handleEndTurn(playerID string) {
	player := players[playerID]
	lobby := lobbies[player.CurrentLobbyID]

	if lobby.Players[lobby.CurrentTurnIndex].ID != playerID {
		fmt.Printf("Errore: Non è il turno del giocatore %s\n", playerID)
		return
	}

	lobby.CurrentTurnIndex = (lobby.CurrentTurnIndex + 1) % 2
	for i, p := range lobby.Players {
		p.IsMyTurn = (i == lobby.CurrentTurnIndex)
	}

	SendLobbyState(lobby)
}
