package main

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"
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

var state GameState
var mutex sync.Mutex

const (
	screenWidth  = 640
	screenHeight = 480
	paddleWidth  = 10
	paddleHeight = 100
	ballSize     = 10
	speed        = 5
)

func updateGame() {
	mutex.Lock()
	defer mutex.Unlock()

	state.Ball.X += state.Ball.VX
	state.Ball.Y += state.Ball.VY

	if state.Ball.Y < 0 || state.Ball.Y > screenHeight-ballSize {
		state.Ball.VY *= -1
	}

	if state.Ball.X < 0 || state.Ball.X > screenWidth-ballSize {
		state.Ball.X, state.Ball.Y = screenWidth/2, screenHeight/2
		state.Ball.VX, state.Ball.VY = 4, 4
	}

	for i := range state.Paddles {
		if state.Ball.X < state.Paddles[i].X+paddleWidth &&
			state.Ball.X+ballSize > state.Paddles[i].X &&
			state.Ball.Y < state.Paddles[i].Y+paddleHeight &&
			state.Ball.Y+ballSize > state.Paddles[i].Y {
			state.Ball.VX *= -1
		}
	}
}

func processInput(input Input) {
	mutex.Lock()
	defer mutex.Unlock()

	paddle := &state.Paddles[input.Player]
	if input.Up && paddle.Y > 0 {
		paddle.Y -= speed
	}
	if input.Down && paddle.Y < screenHeight-paddleHeight {
		paddle.Y += speed
	}
}

func handleConnection(conn net.Conn, player int, wg *sync.WaitGroup) {
	defer wg.Done()
	defer conn.Close()

	decoder := json.NewDecoder(conn)
	encoder := json.NewEncoder(conn)

	for {
		var input Input
		if err := decoder.Decode(&input); err != nil {
			fmt.Println("Player disconnected:", player)
			return
		}

		processInput(input)

		mutex.Lock()
		if err := encoder.Encode(state); err != nil {
			mutex.Unlock()
			fmt.Println("Failed to send game state to player:", player)
			return
		}
		mutex.Unlock()
	}
}

func main() {
	state = GameState{
		Paddles: [2]Paddle{
			{X: 20, Y: screenHeight / 2},
			{X: screenWidth - 40, Y: screenHeight / 2},
		},
		Ball: Ball{X: screenWidth / 2, Y: screenHeight / 2, VX: 4, VY: 4},
	}

	go func() {
		for {
			updateGame()
			time.Sleep(16 * time.Millisecond)
		}
	}()

	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Failed to start server:", err)
		return
	}
	defer ln.Close()

	fmt.Println("Server started on :8080")

	var wg sync.WaitGroup

	for i := 0; i < 2; i++ {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Failed to accept connection:", err)
			continue
		}

		wg.Add(1)
		go handleConnection(conn, i, &wg)
	}

	wg.Wait()
}
