package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	http.HandleFunc("/ws", HandleConnections)
	fmt.Println("Server avviato su :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Errore avviando il server:", err)
	}
}