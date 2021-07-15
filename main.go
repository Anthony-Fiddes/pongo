package main

import (
	"fmt"
	"image/color"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	screenWidth  = 320
	screenHeight = 240
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

type Point struct {
	X float64
	Y float64
}

func (p Point) Location() (x, y float64) {
	return p.X, p.Y
}

type Player struct {
	Sprite *ebiten.Image
	// Point is the top left corner of the player
	Point
	Controller
	Speed float64
}

func (p *Player) Height() float64 {
	return float64(p.Sprite.Bounds().Max.Y)
}

func (p *Player) Width() float64 {
	return float64(p.Sprite.Bounds().Max.X)
}

func (p *Player) Draw(screen *ebiten.Image) {
	options := &ebiten.DrawImageOptions{}
	options.GeoM.Translate(p.X, p.Y)
	screen.DrawImage(p.Sprite, options)
}

func (p *Player) Update() {
	if p.Controller != nil {
		dir := p.Input()
		if dir == Up && p.Y-p.Speed*2 >= 0 {
			p.Y -= p.Speed
		}
		if dir == Down && p.Y+p.Height()+p.Speed*2 <= screenHeight {
			p.Y += p.Speed
		}
	}
}

type Ball struct {
	Sprite *ebiten.Image
	Point
	XSpeed float64
	YSpeed float64
}

func (b *Ball) Height() float64 {
	return float64(b.Sprite.Bounds().Max.Y)
}

func (b *Ball) Width() float64 {
	return float64(b.Sprite.Bounds().Max.X)
}

type Collider interface {
	Height() float64
	Width() float64
	// Location returns the (x, y) coordinates of the top left point of the
	// collision box
	Location() (float64, float64)
}

func (b *Ball) Update(colliders ...Collider) {
	for _, c := range colliders {
		if ptr, ok := c.(*Ball); ok {
			if ptr == b {
				panic("ball.Update: ball cannot collide with itself")
			}
		}

		cX, cY := c.Location()
		inXArea := b.X+b.Width() >= cX && b.X <= cX+c.Width()
		inYArea := b.Y >= cY && b.Y+b.Height() <= cY+c.Height()
		inPaddle := inXArea && inYArea
		nextX := b.X + b.XSpeed
		nextY := b.Y + b.YSpeed
		inNextX := nextX+b.Width() >= cX && nextX <= cX+c.Width()
		inNextY := nextY >= cY && nextY+b.Height() <= cY+c.Height()
		inPaddleNextTick := inNextX && inNextY
		if !inPaddle && inPaddleNextTick {
			b.XSpeed *= -1
			// No need to check the rest once it has bounced
			break
		}
	}

	// Temporarily keeping the ball in the screen
	if b.X <= 0 || b.X+b.Width() >= screenWidth {
		//game over
		b.XSpeed = -b.XSpeed
	}
	if b.Y <= 0 || b.Y+b.Height() >= screenHeight {
		b.YSpeed = -b.YSpeed
	}
	b.X += b.XSpeed
	b.Y += b.YSpeed
}

func (b *Ball) Draw(screen *ebiten.Image) {
	options := &ebiten.DrawImageOptions{}
	options.GeoM.Translate(b.X, b.Y)
	screen.DrawImage(b.Sprite, options)
}

// Game implements ebiten.Game interface.
type Game struct {
	player1 Player
	player2 Player
	ball    Ball
}

// Update proceeds the game state.
// Update is called every tick (1/60 [s] by default).
func (g *Game) Update() error {
	// Write your game's logical update.
	g.player1.Update()
	g.player2.Update()
	g.ball.Update(&g.player1, &g.player2)
	if ebiten.CurrentTPS() < 55 {
		fmt.Println("TPS:", ebiten.CurrentTPS())
		fmt.Println("FPS:", ebiten.CurrentFPS())
	}
	return nil
}

// Draw draws the game screen.
// Draw is called every frame (typically 1/60[s] for 60Hz display).
func (g *Game) Draw(screen *ebiten.Image) {
	// Write your game's rendering.
	g.player1.Draw(screen)
	g.player2.Draw(screen)
	g.ball.Draw(screen)
}

// Layout takes the outside size (e.g., the window size) and returns the (logical) screen size.
// If you don't have to adjust the screen size with the outside size, just return a fixed size.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func setup(game *Game) {
	const playerWidth = 10
	const playerHeight = 50
	const playerSpeed = 2
	playerSprite := ebiten.NewImage(playerWidth, playerHeight)
	playerSprite.Fill(color.White)
	game.player1 = Player{
		Sprite: playerSprite, Controller: &ArrowKeys{}, Speed: playerSpeed,
	}
	game.player2 = Player{
		Sprite: playerSprite,
		Point:  Point{screenWidth - playerWidth, screenHeight - playerHeight},
		Speed:  playerSpeed,
	}

	const ballWidth = 10
	ballSprite := ebiten.NewImage(ballWidth, ballWidth)
	ballSprite.Fill(color.White)
	game.ball = Ball{
		Sprite: ballSprite,
		Point:  Point{X: screenWidth/2 - ballWidth/2, Y: screenHeight/2 - ballWidth/2},
		XSpeed: 1,
		YSpeed: 1,
	}

	fb := &FollowBall{Ball: &game.ball, Player: &game.player2}
	game.player2.Controller = fb
}

func main() {
	game := &Game{}
	// Specify the window size as you like. Here, a doubled size is specified.
	ebiten.SetWindowSize(screenWidth*2, screenHeight*2)
	ebiten.SetWindowTitle("Pong")
	setup(game)
	// Call ebiten.RunGame to start your game loop.
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
