package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/joanlopez/golang-multiplayer/player"
	"github.com/matryer/runner"
	"github.com/oklog/ulid"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

type Connection struct {
	Id string
	Socket *websocket.Conn
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Connections struct {
	items map[string]*Connection
	lock *sync.Mutex
}

type Players struct {
	items map[string]*player.Player
	lock *sync.Mutex
}

var connections = Connections{items: map[string]*Connection{}, lock: &sync.Mutex{}}
var players = Players{items: map[string]*player.Player{}, lock: &sync.Mutex{}}
var handlers = map[string]*runner.Task{}

func updatePositions() {
	// TODO: Handle positions on close (remove)
	players.lock.Lock()
	for _, p := range players.items {

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
	players.lock.Unlock()
}

func synchronizePositions() {
	connections.lock.Lock()
	for _, conn := range connections.items {
		err := conn.Socket.WriteJSON(players.items)
		if err != nil {
			log.Printf("Error while writing players position: %v\n", conn.Id)
			stopConnection(conn.Id)
		}
	}
	connections.lock.Unlock()
}

func newUlid() string {
	now := time.Now()
	entropy := ulid.Monotonic(rand.New(rand.NewSource(now.UnixNano())), 0)
	return ulid.MustNew(ulid.Timestamp(now), entropy).String()
}

func stopConnection(id string) {
	players.lock.Lock()

	delete(connections.items, id)
	delete(players.items, id)
	delete(handlers, id)

	players.lock.Unlock()
}

func main() {
	fmt.Println("Welcome to the server!")

	go func() {
		for {
			time.Sleep(10 * time.Millisecond)
			updatePositions()
		}
	}()

	go func() {
		for {
			synchronizePositions()
		}
	}()

	http.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) {
		connections.lock.Lock()
		players.lock.Lock()

		id := newUlid()

		conn, _ := upgrader.Upgrade(w, r, nil)

		// Adding new connection to the connections pool
		connections.items[id] = &Connection{Id: id, Socket: conn}

		// Adding new player to the players group
		p := player.Player{Id: id, X: 0, Y: 0, Speed: 5, Movements: player.Movements{false, false, false, false}}
		players.items[id] = &p

		handlers[id] = runner.Go(func(shouldStop runner.S) error {
			for {
				m := player.Movements{}
				err := conn.ReadJSON(&m)

				if err != nil {
					log.Printf("Error while reading movements of: %v\n", id)
					break
				}

				p.Movements = m

				if shouldStop() {
					break
				}
			}
			return nil
		})


		connections.lock.Unlock()
		players.lock.Unlock()
	})

	_ = http.ListenAndServe(":8080", nil)
}
