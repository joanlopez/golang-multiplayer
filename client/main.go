package main

import (
	"fmt"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/gorilla/websocket"
	"github.com/joanlopez/golang-multiplayer/player"
	"golang.org/x/image/colornames"
	"log"
	"net/url"
)


const width float64 = 1024
const height float64= 768

var players []player.Player

var playerHeight float64 = 50
var playerWidth float64 = 50


var conn *websocket.Conn
var movements = player.Movements{false, false, false, false}

func playerImd() (imd *imdraw.IMDraw) {
	imd = imdraw.New(nil)
	imd.Color = pixel.RGB(255, 0, 0)

	for _, p := range players {
		imd.Push(
			pixel.V(p.X-playerWidth, p.Y-playerHeight),
			pixel.V(p.X+playerWidth, p.Y+playerHeight),
		)

		imd.Rectangle(0)
	}

	return
}

func run() {
	cfg := pixelgl.WindowConfig{
		Title:  "Pixel Rocks!",
		Bounds: pixel.R(0, 0, width, height),
		//Monitor: pixelgl.PrimaryMonitor(),
		VSync: true,
	}

	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	win.SetSmooth(true)

	for !win.Closed() {
		win.Clear(colornames.White)

		if win.Pressed(pixelgl.KeyEscape) {
			win.Destroy()
		}

		movements.Up = win.Pressed(pixelgl.KeyUp)
		movements.Down = win.Pressed(pixelgl.KeyDown)
		movements.Left = win.Pressed(pixelgl.KeyLeft)
		movements.Right = win.Pressed(pixelgl.KeyRight)

		err := conn.WriteJSON(&movements)
		if err != nil {
			log.Println("Error while writing movements")
		}

		imd := playerImd()

		imd.Draw(win)

		win.Update()
	}
}

func synchronizePositions(conn *websocket.Conn) {
	for {
		err := conn.ReadJSON(&players)
		if err != nil {
			log.Println("Error while reading positions")
		}
	}
}

func main() {
	fmt.Println("Welcome to the client!")

	u := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/echo"}

	conn, _, _ = websocket.DefaultDialer.Dial(u.String(), nil)

	defer conn.Close()

	go synchronizePositions(conn)

	pixelgl.Run(run)
}