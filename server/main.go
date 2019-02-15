package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/joanlopez/golang-multiplayer/player"
	"log"
	"net/http"
	"time"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var players [4]*player.Player
var currPlayers = 0
var connections [4]*websocket.Conn

func handlePlayer(conn *websocket.Conn, p *player.Player) {
	for {
		m := player.Movements{}
		err := conn.ReadJSON(&m)
		if err != nil {
			log.Println("Error while reading movements")
		}

		p.Movements = m
	}
}

func updatePositions(players *[4]*player.Player) {
	for _, p := range *players {
		if p == nil { // TODO: Improve management of players
			continue
		}

		if p.Movements.Up {
			p.Y += p.Speed
		}

		if p.Movements.Down {
			p.Y -= p.Speed
		}

		if p.Movements.Left {
			p.X -= p.Speed
		}

		if p.Movements.Right {
			p.X += p.Speed
		}

	}
}

func synchronizePositions(connections *[4]*websocket.Conn, players *[4]*player.Player) {
	for _, c := range *connections {
		if c == nil { // TODO: Improve management of connections
			continue
		}

		err := c.WriteJSON(players)
		if err != nil {
			log.Println("Error while writing players position")
		}

	}
}

func main() {
	fmt.Println("Welcome to the server!")

	go func() {
		for {
			time.Sleep(10 * time.Millisecond)
			updatePositions(&players)
		}
	}()

	go func() {
		for {
			synchronizePositions(&connections, &players)
		}
	}()

	http.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) {
		conn, _ := upgrader.Upgrade(w, r, nil)

		// Adding new connection to the connections pool
		connections[currPlayers] = conn

		// Adding new player to the players group
		p := player.Player{X: 0, Y: 0, Speed: 5, Movements: player.Movements{false, false, false, false}}
		players[currPlayers] = &p
		currPlayers++ // TODO: Handle players properly

		go handlePlayer(conn, &p)

		log.Printf("Handling new player: %v\n", currPlayers)

		// TODO: Set close handler
	})

	_ = http.ListenAndServe(":8080", nil)
}
