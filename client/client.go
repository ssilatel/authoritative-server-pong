package main

import (
	"encoding/json"
	"fmt"
	"image/color"
	"log"
	"net"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Paddle struct {
	X, Y float64
}

type Ball struct {
	X, Y, VX, VY float64
}

type GameState struct {
	Paddles [2]Paddle
	Ball    Ball
}

type Input struct {
	Player int
	Up     bool
	Down   bool
}

var (
	conn         net.Conn
	state        GameState
	player       int
	screenWidth  = 640
	screenHeight = 480
	paddleWidth  = 10
	paddleHeight = 100
	ballSize     = 10
)

type Game struct{}

func (g *Game) Update() error {
	up := ebiten.IsKeyPressed(ebiten.KeyW)
	down := ebiten.IsKeyPressed(ebiten.KeyS)

	input := Input{Player: player, Up: up, Down: down}

	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(input); err != nil {
		return err
	}

	decoder := json.NewDecoder(conn)
	if err := decoder.Decode(&state); err != nil {
		return err
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	ebitenutil.DrawRect(screen, float64(state.Paddles[0].X), float64(state.Paddles[0].Y), float64(paddleWidth), float64(paddleHeight), color.RGBA{255, 255, 255, 255})
	ebitenutil.DrawRect(screen, float64(state.Paddles[1].X), float64(state.Paddles[1].Y), float64(paddleWidth), float64(paddleHeight), color.RGBA{255, 255, 255, 255})
	ebitenutil.DrawRect(screen, state.Ball.X, state.Ball.Y, float64(ballSize), float64(ballSize), color.RGBA{255, 255, 255, 255})
}

func (g *Game) Layout(w, h int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	var err error
	conn, err = net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("Failed to connect to server:", err)
		return
	}
	defer conn.Close()

	fmt.Print("Enter player number (0 or 1): ")
	fmt.Scanf("%d", &player)

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Pong")

	if err := ebiten.RunGame(&Game{}); err != nil {
		log.Fatal(err)
	}
}
