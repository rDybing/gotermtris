/********************************
  gtt.go
  License: MIT
  Copyright (c) 2019 Roy Dybing
  github   : rDybing
  Linked In: Roy Dybing
  Full license text in README.md
*********************************/

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sort"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

type sizeT struct {
	x int
	y int
}

type hiScoreT struct {
	Name  string
	Score int
}

type screenT struct {
	ui           *widgets.Paragraph
	size         sizeT
	field        sizeT
	fieldXOffset int
	fieldYOffset int
	buffer       string
	fieldBuffer  []byte
	brickBuffer  []byte
}

type brickStateT struct {
	index    int
	rotation int
	posX     int
	posY     int
}

type gameStateT struct {
	score       int
	bricksCount int
	bricksTotal int
	ticksToDown int
	tickerCount int
	gameOver    bool
}

type brickT []byte

const hiScoreFile = "./hiScore.json"

func main() {
	if err := ui.Init(); err != nil {
		log.Fatalf("Failed to initialize terminal UI: %v", err)
	}
	defer ui.Close()

	hiScore, ok := loadScore()
	if !ok {
		hiScore = initScore()
		saveScore(hiScore)
	}

	rand.Seed(time.Now().UnixNano())
	screen := initScreen()
	brickPiece := initBricks()
	menuEvent := ui.PollEvents()
	screen.drawMenu(hiScore)
	quit := false
	for !quit {
		eventMenuLoop := <-menuEvent
		switch eventMenuLoop.ID {
		case "<Space>":
			score := screen.gameLoop(brickPiece)
			hiScore = screen.newHiScore(hiScore, score)
			screen.drawMenu(hiScore)
		case "<Escape>", "<C-c>":
			quit = true
		}
	}
}

func (screen screenT) drawMenu(hs []hiScoreT) {
	screen.clearField()
	menuText := "     GoTermTris\n\n"
	menuText += "       ʕ◔ϖ◔ʔ\n\n"
	menuText += "  2019 © Roy Dybing\n"
	menuText += "    License: MIT\n\n"
	menuText += "    High Scores:\n\n"
	menuText += displayScore(hs)
	menuText += "\n    ------------- \n\n"
	menuText += " Press Space to Start\n"
	menuText += " Press Escape to quit\n"
	screen.ui.Text = menuText
	ui.Render(screen.ui)
}

func (screen screenT) gameLoop(brickPiece []brickT) int {
	state := initGameState()
	var brickState brickStateT
	brickState.newBrick(screen, &state)
	ticker := time.NewTicker(time.Millisecond * 35).C
	keyEvent := ui.PollEvents()

	screen.setFieldBoundary()

	for !state.gameOver {
		brickTest := brickState
		select {
		case eventGameLoop := <-keyEvent:
			switch eventGameLoop.ID {
			case "<Left>":
				brickTest.posX--
				if brickTest.doBrickFit(screen, brickPiece[brickState.index]) {
					brickState.posX--
				}
			case "<Right>":
				brickTest.posX++
				if brickTest.doBrickFit(screen, brickPiece[brickState.index]) {
					brickState.posX++
				}
			case "<Down>":
				brickTest.posY++
				if brickTest.doBrickFit(screen, brickPiece[brickState.index]) {
					brickState.posY++
				}
			case "<Up>":
				if brickTest.rotation < 3 {
					brickTest.rotation++
				} else {
					brickTest.rotation = 0
				}
				if brickTest.doBrickFit(screen, brickPiece[brickState.index]) {
					brickState.rotation = brickTest.rotation
				}
			case "<Escape>", "<C-c>":
				state.gameOver = true
				return state.score
			}
		case <-ticker:
			if state.getPullDown() {
				brickTest.posY++
				if brickTest.doBrickFit(screen, brickPiece[brickState.index]) {
					brickState.posY++
				} else {
					screen.lockBrick(brickState, brickPiece[brickState.index])
					screen.checkLines(brickState, &state)
					brickState.newBrick(screen, &state)
					if !brickState.doBrickFit(screen, brickPiece[brickState.index]) {
						state.gameOver = true
						return state.score
					}
				}
			}
			screen.updateBrickBuffer(brickState, brickPiece[brickState.index])
			screen.drawScreenBuffer()
			bufTxt := screen.buffer
			screen.ui.Text = bufTxt + fmt.Sprintf(" -- Bricks: %06d --\n -- Score : %06d --", state.bricksTotal, state.score)
			ui.Render(screen.ui)
		}
	}
	return state.score
}

func (s *gameStateT) getPullDown() bool {
	s.tickerCount++
	// make faster every X bricks
	if s.bricksCount >= 10 {
		if s.ticksToDown > 3 {
			s.ticksToDown--
		}
		s.bricksCount = 0
	}
	// set pull brick down
	if s.tickerCount >= s.ticksToDown {
		s.tickerCount = 0
		return true
	}
	return false
}

// ******************************************* Brick Handling Stuff ****************************************************

func (bs brickStateT) doBrickFit(s screenT, brick brickT) bool {
	for brickPixelX := 0; brickPixelX < 4; brickPixelX++ {
		for brickPixelY := 0; brickPixelY < 4; brickPixelY++ {
			brickPixelIndex := rotateBrick(brickPixelX, brickPixelY, bs.rotation)
			fieldPixelIndex := (bs.posY+brickPixelY)*s.field.x + (bs.posX + brickPixelX)
			if bs.posX+brickPixelX >= 0 && bs.posX+brickPixelX < s.field.x {
				if bs.posY+brickPixelY >= 0 && bs.posY+brickPixelY < s.field.y {
					if brick[brickPixelIndex] != 0 && s.fieldBuffer[fieldPixelIndex] != 0 {
						return false
					}
				}
			}
		}
	}
	return true
}

func (bs *brickStateT) newBrick(s screenT, gs *gameStateT) {
	bs.index = rand.Intn(7)
	bs.rotation = 0
	bs.posX = (s.field.x / 2) - 2
	bs.posY = 0
	gs.bricksCount++
	gs.bricksTotal++
	if gs.bricksTotal > 1 {
		gs.score += 15
	}
}

func rotateBrick(brickPixelX, brickPixelY, brickRotation int) int {
	brickPixelIndex := 0
	switch brickRotation {
	case 0:
		brickPixelIndex = brickPixelY*4 + brickPixelX
	case 1:
		brickPixelIndex = 12 + brickPixelY - (brickPixelX * 4)
	case 2:
		brickPixelIndex = 15 - (brickPixelY * 4) - brickPixelX
	case 3:
		brickPixelIndex = 3 - brickPixelY + (brickPixelX * 4)
	}
	return brickPixelIndex
}

// ******************************************* Playfield Handling Stuff ************************************************

func (screen *screenT) checkLines(bs brickStateT, gs *gameStateT) {
	var lines int
	for brickY := 0; brickY < 4; brickY++ {
		if bs.posY+brickY < screen.field.y-1 {
			line := true
			for brickX := 1; brickX < screen.field.x-1; brickX++ {
				if screen.fieldBuffer[(bs.posY+brickY)*screen.field.x+brickX] == 0 {
					line = false
				}
			}
			if line {
				for brickX := 1; brickX < screen.field.x-1; brickX++ {
					screen.fieldBuffer[(bs.posY+brickY)*screen.field.x+brickX] = 8
				}
				lines++
			}
		}
	}
	if lines > 0 {
		go screen.deleteLines(lines, gs)
	}
}

func (screen *screenT) deleteLines(l int, gs *gameStateT) {
	for i := 0; i < l; i++ {
		time.Sleep(time.Millisecond * 51)
		for y := 0; y < screen.field.y; y++ {
			for x := 1; x < screen.field.x-1; x++ {
				if screen.fieldBuffer[x+(y*screen.field.x)] == 8 {
					for revY := y; revY > 1; revY-- {
						screen.fieldBuffer[x+(revY*screen.field.x)] = screen.fieldBuffer[x+((revY-1)*screen.field.x)]
					}
				}
			}
		}
		gs.score += 50 + (l * 50)
	}
}

func (screen *screenT) lockBrick(bs brickStateT, brick brickT) {
	for brickX := 0; brickX < 4; brickX++ {
		for brickY := 0; brickY < 4; brickY++ {
			if brick[rotateBrick(brickX, brickY, bs.rotation)] != 0 {
				screen.fieldBuffer[(bs.posY+brickY)*screen.field.x+(bs.posX+brickX)] = byte(bs.index + 1)
			}
		}
	}
}

func (screen *screenT) setFieldBoundary() {
	for x := 0; x < screen.field.x; x++ {
		for y := 0; y < screen.field.y; y++ {
			if x == 0 || x == screen.field.x-1 || y == screen.field.y-1 {
				screen.fieldBuffer[y*screen.field.x+x] = 9
			} else {
				screen.fieldBuffer[y*screen.field.x+x] = 0
			}
		}
	}
}

func (screen *screenT) clearField() {
	for i := range screen.fieldBuffer {
		screen.fieldBuffer[i] = 0
	}
}

func (screen *screenT) updateBrickBuffer(bs brickStateT, brick brickT) {
	for i := range screen.brickBuffer {
		screen.brickBuffer[i] = 0
	}
	for brickX := 0; brickX < 4; brickX++ {
		for brickY := 0; brickY < 4; brickY++ {
			if brick[rotateBrick(brickX, brickY, bs.rotation)] != 0 {
				screen.brickBuffer[(bs.posY+brickY)*screen.field.x+(bs.posX+brickX)] = byte(bs.index + 1)
			}
		}
	}
}

func (screen *screenT) drawScreenBuffer() {
	screen.buffer = ""
	fieldIndex := 0

	for y := 0; y < screen.size.y; y++ {
		for x := 0; x < screen.size.x; x++ {
			if (x >= screen.fieldXOffset && x < screen.field.x+screen.fieldXOffset) &&
				(y >= screen.fieldYOffset && y < screen.field.y+screen.fieldYOffset) {
				newPixel := string(getRune(screen.fieldBuffer[fieldIndex]))
				if newPixel == " " {
					newPixel = string(getRune(screen.brickBuffer[fieldIndex]))
				}
				screen.buffer += newPixel
				fieldIndex++
			} else {
				screen.buffer += " "
			}
		}
		screen.buffer += "\n"
	}
}

func getRune(index byte) rune {
	out := []rune{' ', '0', 'O', 'Ø', 'H', 'W', 'M', 'X', '-', '*'}
	return out[index]
}

func (screen *screenT) newHiScore(hs []hiScoreT, score int) []hiScoreT {
	if score > hs[4].Score {
		menuText := "  You made it into\n"
		menuText += "  the top five list!\n"
		menuText += "   with a score of\n"
		menuText += fmt.Sprintf("      %6d\n\n", score)
		menuText += "   Enter your name:\n"
		screen.ui.Text = menuText
		ui.Render(screen.ui)
		scoreEvent := ui.PollEvents()
		var nameOut string
		back := false
		for !back {
			eventScore := <-scoreEvent
			if eventScore.ID == "<Enter>" {
				back = true
			} else if eventScore.ID != "<Up>" &&
				eventScore.ID != "<Left>" &&
				eventScore.ID != "<Right>" &&
				eventScore.ID != "<Down>" &&
				eventScore.ID != "<Escape>" {
				nameOut += eventScore.ID
				screen.ui.Text = menuText + "\n  -> " + nameOut + " <-"
				ui.Render(screen.ui)
			}
		}
		hsTemp := hs
		hs = nil
		var hsAdd hiScoreT
		hsAdd.Name = nameOut
		hsAdd.Score = score
		hsTemp = append(hsTemp, hsAdd)
		sort.Slice(hsTemp, func(i, j int) bool {
			return hsTemp[i].Score > hsTemp[j].Score
		})
		for i := 0; i < 5; i++ {
			hs = append(hs, hsTemp[i])
		}
		saveScore(hs)
	}
	return hs
}

// ******************************************* HiScore Stuff ***********************************************************

func loadScore() ([]hiScoreT, bool) {
	var hs []hiScoreT
	ok := true
	f, err := os.Open(hiScoreFile)
	if err != nil {
		ok = false
	}
	defer f.Close()

	hsJSON := json.NewDecoder(f)
	if err = hsJSON.Decode(&hs); err != nil {
		ok = false
	}
	return hs, ok
}

func initScore() []hiScoreT {
	var hs []hiScoreT
	temp := []struct {
		Name  string
		Score int
	}{
		{Name: "Roy", Score: 10000},
		{Name: "Maurice", Score: 7500},
		{Name: "Jen", Score: 5000},
		{Name: "Douglas", Score: 2500},
		{Name: "Denholm", Score: 1000},
	}
	for i := range temp {
		hs = append(hs, temp[i])
	}
	return hs
}

func saveScore(hs []hiScoreT) {
	f, err := os.Create(hiScoreFile)
	if err != nil {
		log.Fatalf("Could not save hi score file: %v\n", err)
	}
	defer f.Close()
	tempJSON := json.NewEncoder(f)
	tempJSON.SetIndent("", "    ")
	if err := tempJSON.Encode(hs); err != nil {
		log.Fatalf("Could not encode hi score file: %v\n", err)
	}
}

func displayScore(hs []hiScoreT) string {
	var score string
	for i := range hs {
		score += fmt.Sprintf("  %8s - %6d\n", hs[i].Name, hs[i].Score)
	}
	return score
}

// ******************************************* Initial Setup Stuff *****************************************************

func initScreen() screenT {
	var s screenT
	s.size.x = 24
	s.size.y = 21
	s.field.x = 12
	s.field.y = 18
	s.fieldXOffset = 5
	s.fieldYOffset = 1
	s.ui = widgets.NewParagraph()
	s.ui.Text = ""
	s.ui.Border = true
	s.ui.SetRect(0, 0, s.size.x, s.size.y+5)
	s.buffer = ""
	s.fieldBuffer = make([]byte, s.field.x*s.field.y)
	s.brickBuffer = make([]byte, s.field.x*s.field.y)

	return s
}

func initBricks() []brickT {
	var out []brickT
	var temp [4][4]byte
	// line
	temp[0] = [4]byte{0, 0, 1, 0}
	temp[1] = [4]byte{0, 0, 1, 0}
	temp[2] = [4]byte{0, 0, 1, 0}
	temp[3] = [4]byte{0, 0, 1, 0}
	out = append(out, buildByteSlice(temp))
	// pyramid
	temp[0] = [4]byte{0, 0, 1, 0}
	temp[1] = [4]byte{0, 1, 1, 0}
	temp[2] = [4]byte{0, 0, 1, 0}
	temp[3] = [4]byte{0, 0, 0, 0}
	out = append(out, buildByteSlice(temp))
	// box
	temp[0] = [4]byte{0, 0, 0, 0}
	temp[1] = [4]byte{0, 1, 1, 0}
	temp[2] = [4]byte{0, 1, 1, 0}
	temp[3] = [4]byte{0, 0, 0, 0}
	out = append(out, buildByteSlice(temp))
	// stepLeft
	temp[0] = [4]byte{0, 0, 1, 0}
	temp[1] = [4]byte{0, 1, 1, 0}
	temp[2] = [4]byte{0, 1, 0, 0}
	temp[3] = [4]byte{0, 0, 0, 0}
	out = append(out, buildByteSlice(temp))
	// stepRight
	temp[0] = [4]byte{0, 1, 0, 0}
	temp[1] = [4]byte{0, 1, 1, 0}
	temp[2] = [4]byte{0, 0, 1, 0}
	temp[3] = [4]byte{0, 0, 0, 0}
	out = append(out, buildByteSlice(temp))
	// hookRight
	temp[0] = [4]byte{0, 1, 0, 0}
	temp[1] = [4]byte{0, 1, 0, 0}
	temp[2] = [4]byte{0, 1, 1, 0}
	temp[3] = [4]byte{0, 0, 0, 0}
	out = append(out, buildByteSlice(temp))
	// hookLeft
	temp[0] = [4]byte{0, 0, 1, 0}
	temp[1] = [4]byte{0, 0, 1, 0}
	temp[2] = [4]byte{0, 1, 1, 0}
	temp[3] = [4]byte{0, 0, 0, 0}
	out = append(out, buildByteSlice(temp))

	return out
}

func buildByteSlice(t [4][4]byte) []byte {
	var tOut []byte
	for i := range t {
		for j := range t[i] {
			tOut = append(tOut, t[i][j])
		}
	}
	return tOut
}

func initGameState() gameStateT {
	var gs gameStateT
	gs.score = 0
	gs.ticksToDown = 20
	gs.tickerCount = 0
	gs.bricksCount = 0
	gs.bricksTotal = 0
	gs.gameOver = false
	return gs
}
