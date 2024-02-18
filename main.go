package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"log"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

const (
	debug        = false
	screenX      = 640
	screenY      = 480
	baseX        = 100
	groundY      = 400
	speed        = 6
	jumpingPower = 15
	gravity      = 1
	interval     = 10
	minTreeDist  = 50
	maxTreeCount = 3
	fontSize     = 10

	// game modes
	modeTitle    = 0
	modeGame     = 1
	modeGameover = 2

	// tree kinds
	kindTreeSmall = 0
	kindTreeBig   = 1

	// image sizes
	dinosaurHeight  = 50
	dinosaurWidth   = 100
	groundHeight    = 50
	groundWidth     = 50
	treeSmallWidth  = 25
	treeSmallHeghit = 50
	treeBigWidth    = 25
	treeBigHeghit   = 75
)

//go:embed resources/images/dinosaur_01.png
var byteDinosaur1Img []byte

//go:embed resources/images/dinosaur_02.png
var byteDinosaur2Img []byte

//go:embed resources/images/tree_small.png
var byteTreeSmallImg []byte

//go:embed resources/images/tree_big.png
var byteTreeBigImg []byte

//go:embed resources/images/ground.png
var byteGroundImg []byte

var (
	dinosaur1Img *ebiten.Image
	dinosaur2Img *ebiten.Image
	treeSmallImg *ebiten.Image
	treeBigImg   *ebiten.Image
	groundImg    *ebiten.Image
	arcadeFont   font.Face
)

func init() {
	rand.Seed(time.Now().UnixNano())

	img, _, err := image.Decode(bytes.NewReader(byteDinosaur1Img))
	if err != nil {
		log.Fatal(err)
	}
	dinosaur1Img = ebiten.NewImageFromImage(img)

	img, _, err = image.Decode(bytes.NewReader(byteDinosaur2Img))
	if err != nil {
		log.Fatal(err)
	}
	dinosaur2Img = ebiten.NewImageFromImage(img)

	img, _, err = image.Decode(bytes.NewReader(byteTreeSmallImg))
	if err != nil {
		log.Fatal(err)
	}
	treeSmallImg = ebiten.NewImageFromImage(img)

	img, _, err = image.Decode(bytes.NewReader(byteTreeBigImg))
	if err != nil {
		log.Fatal(err)
	}
	treeBigImg = ebiten.NewImageFromImage(img)

	img, _, err = image.Decode(bytes.NewReader(byteGroundImg))
	if err != nil {
		log.Fatal(err)
	}
	groundImg = ebiten.NewImageFromImage(img)

	tt, err := opentype.Parse(fonts.PressStart2P_ttf)
	if err != nil {
		log.Fatal(err)
	}
	const dpi = 72
	arcadeFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    fontSize,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
}

type tree struct {
	x       int
	y       int
	kind    int
	visible bool
}

func (t *tree) move(speed int) {
	t.x -= speed
}

func (t *tree) show() {
	t.kind = rand.Intn(2)
	t.x = screenX
	switch t.kind {
	case kindTreeSmall:
		t.y = groundY - treeSmallHeghit
	case kindTreeBig:
		t.y = groundY - treeBigHeghit
	}
	t.visible = true
}

func (t *tree) hide() {
	t.visible = false
}

func (t *tree) isOutOfScreen() bool {
	return t.x < -50
}

type ground struct {
	x int
	y int
}

func (g *ground) move(speed int) {
	g.x -= speed
	if g.x < -groundWidth {
		g.x = g.x + groundWidth
	}
}

// Game struct
type Game struct {
	mode      int
	count     int
	score     int
	hiscore   int
	dinosaurX int
	dinosaurY int
	gy        int
	jumpFlg   bool
	trees     [maxTreeCount]*tree
	lastTreeX int
	ground    *ground
}

// NewGame method
func NewGame() *Game {
	g := &Game{}
	g.init()
	return g
}

// Init method
func (g *Game) init() {
	// Update hiscore
	if g.hiscore < g.score {
		g.hiscore = g.score
	}
	g.count = 0
	g.score = 0
	g.lastTreeX = 0
	g.gy = 0
	g.dinosaurX = baseX
	g.dinosaurY = groundY - dinosaurHeight
	for i := 0; i < maxTreeCount; i++ {
		g.trees[i] = &tree{}
	}
	g.ground = &ground{y: groundY - 30}
}

// Update method
func (g *Game) Update() error {
	switch g.mode {
	case modeTitle:
		if g.isKeyJustPressed() {
			g.mode = modeGame
		}
	case modeGame:
		g.count++
		g.score = g.count / 5

		if !g.jumpFlg && g.isKeyJustPressed() {
			g.jumpFlg = true
			g.gy = -jumpingPower
		}

		if g.jumpFlg {
			g.dinosaurY += g.gy
			g.gy += gravity
		}

		if g.dinosaurY >= groundY-dinosaurHeight {
			g.jumpFlg = false
		}

		for _, t := range g.trees {
			if t.visible {
				t.move(speed)
				if t.isOutOfScreen() {
					t.hide()
				}
			} else {
				if g.count-g.lastTreeX > minTreeDist && g.count%interval == 0 && rand.Intn(10) == 0 {
					g.lastTreeX = g.count
					t.show()
					break
				}
			}
		}

		g.ground.move(speed)

		if g.hit() {
			g.mode = modeGameover
		}
	case modeGameover:
		if g.isKeyJustPressed() {
			g.init()
			g.mode = modeGame
		}
	}

	return nil
}

// Draw method
func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.White)
	text.Draw(screen, fmt.Sprintf("Hisore: %d", g.hiscore), arcadeFont, 300, 20, color.Black)
	text.Draw(screen, fmt.Sprintf("Score: %d", g.score), arcadeFont, 500, 20, color.Black)
	var xs [3]int
	var ys [3]int

	if len(g.trees) > 0 {
		for i, t := range g.trees {
			xs[i] = t.x
			ys[i] = t.y
		}
	}

	if debug {
		ebitenutil.DebugPrint(screen, fmt.Sprintf(
			"g.y: %d\nTree1 x:%d, y:%d\nTree2 x:%d, y:%d\nTree3 x:%d, y:%d",
			g.dinosaurY,
			xs[0],
			ys[0],
			xs[1],
			ys[1],
			xs[2],
			ys[2],
		))
	}

	g.drawGround(screen)
	g.drawTrees(screen)
	g.drawDinosaur(screen)

	switch g.mode {
	case modeTitle:
		text.Draw(screen, "PRESS SPACE KEY", arcadeFont, 245, 240, color.Black)
	case modeGameover:
		text.Draw(screen, "GAME OVER", arcadeFont, 275, 240, color.Black)
	}
}

func (g *Game) drawDinosaur(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(baseX, float64(g.dinosaurY))
	op.Filter = ebiten.FilterLinear
	if (g.count/5)%2 == 0 {
		screen.DrawImage(dinosaur1Img, op)
		return
	}
	screen.DrawImage(dinosaur2Img, op)
}

func (g *Game) drawTrees(screen *ebiten.Image) {
	for _, t := range g.trees {
		if t.visible {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(t.x), float64(t.y))
			op.Filter = ebiten.FilterLinear
			switch t.kind {
			case kindTreeSmall:
				screen.DrawImage(treeSmallImg, op)
			case kindTreeBig:
				screen.DrawImage(treeBigImg, op)
			}
		}
	}
}

func (g *Game) drawGround(screen *ebiten.Image) {
	for i := 0; i < 14; i++ {
		x := float64(groundWidth * i)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(x, float64(g.ground.y))
		op.GeoM.Translate(float64(g.ground.x), 0.0)
		op.Filter = ebiten.FilterLinear
		screen.DrawImage(groundImg, op)
	}
}

// Layout method
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return screenX, screenY
}

func (g *Game) isKeyJustPressed() bool {
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		return true
	}
	return false
}

func (g *Game) hit() bool {
	hitDinosaurMinX := g.dinosaurX + 20
	hitDinosaurMaxX := g.dinosaurX + dinosaurWidth - 15
	hitDinosaurMaxY := g.dinosaurY + dinosaurHeight - 10

	for _, t := range g.trees {
		hitTreeMinX := t.x + 5
		hitTreeMaxX := t.x + treeSmallWidth - 5
		hitTreeMinY := t.y + 5

		if t.visible {
			if hitDinosaurMaxX-hitTreeMinX > 0 && hitTreeMaxX-hitDinosaurMinX > 0 && hitDinosaurMaxY-hitTreeMinY > 0 {
				return true
			}
		}
	}
	return false
}

func main() {
	ebiten.SetWindowSize(screenX, screenY)
	ebiten.SetWindowTitle("Dinosaur Jump")
	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
