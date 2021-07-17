package main

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

type Direction int

const (
	Up Direction = iota
	Down
	Nothing
)

type Controller interface {
	Input() Direction
}

type ArrowKeys struct{}

func (ak *ArrowKeys) Input() Direction {
	upPressed := ebiten.IsKeyPressed(ebiten.KeyArrowUp)
	downPressed := ebiten.IsKeyPressed(ebiten.KeyArrowDown)
	if upPressed == downPressed {
		return Nothing
	}
	if upPressed {
		return Up
	}
	return Down
}

type FollowBall struct {
	Ball         *Ball
	Player       *Player
	lastDecision Direction
	count        int
}

// Input should be called in game.Update()
func (fb *FollowBall) Input() Direction {
	// decisionBuffer is the number of ticks for which the FollowBall controller
	// must keep the last decision that it made
	const decisionBuffer = 15

	ballX, ballY := fb.Ball.Location()
	x, y := fb.Player.Location()
	if fb.count > 0 {
		fb.count--
		return fb.lastDecision
	}
	if math.Abs(ballX-x) < 0.6*screenWidth {
		//the ball is close enough to see
		if ballY < y-fb.Player.Speed {
			fb.count += decisionBuffer
			fb.lastDecision = Up
			return Up
		} else if ballY > y+fb.Player.Speed {
			fb.count += decisionBuffer
			fb.lastDecision = Down
			return Down
		}
	}
	return Nothing
}
