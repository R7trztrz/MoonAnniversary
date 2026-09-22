// MoonAnniversary renders a looping anniversary scene in the terminal.
//
// The application intentionally uses only the Go standard library so release
// builds can be distributed as standalone executables.
//
// Author: Simon Tian
package main

import (
	"fmt"
	"io"
	"math"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unicode"
)

const (
	defaultWidth       = 80
	defaultHeight      = 24
	framesPerSecond    = 12
	anniversaryMessage = "祝最爱的月月周年快乐！！！"
	colorReset         = "\x1b[0m"
	colorStar          = "\x1b[38;5;153m"
	colorMoon          = "\x1b[38;5;229m"
	colorBird          = "\x1b[38;5;255m"
	colorHeart         = "\x1b[38;5;204m"
	colorText          = "\x1b[38;5;219m"
	colorLeaf          = "\x1b[38;5;114m"
	colorTrunk         = "\x1b[38;5;137m"
	colorIsland        = "\x1b[38;5;180m"
	colorFlag          = "\x1b[38;5;217m"
	colorFireworkPink  = "\x1b[38;5;213m"
	colorFireworkGold  = "\x1b[38;5;220m"
	colorFireworkBlue  = "\x1b[38;5;117m"
)

type cell struct {
	char         rune
	color        string
	continuation bool
}

type canvas struct {
	width  int
	height int
	cells  [][]cell
}

func newCanvas(width, height int) *canvas {
	cells := make([][]cell, height)
	for row := range cells {
		cells[row] = make([]cell, width)
		for column := range cells[row] {
			cells[row][column].char = ' '
		}
	}
	return &canvas{width: width, height: height, cells: cells}
}

func (c *canvas) put(x, y int, text, color string) {
	if y < 0 || y >= c.height {
		return
	}
	for _, char := range text {
		charWidth := runeDisplayWidth(char)
		if charWidth == 0 {
			// The cell model keeps uncommon combining marks visible in their own cell.
			charWidth = 1
		}
		if x >= 0 && x+charWidth <= c.width {
			c.cells[y][x] = cell{char: char, color: color}
			if charWidth == 2 {
				c.cells[y][x+1] = cell{continuation: true}
			}
		}
		x += charWidth
	}
}

func (c *canvas) render(withColor bool) string {
	var output strings.Builder
	activeColor := ""
	for row, line := range c.cells {
		lastVisible := -1
		for column, current := range line {
			if current.char != ' ' || current.continuation {
				lastVisible = column
			}
		}
		for column := 0; column <= lastVisible; column++ {
			current := line[column]
			if current.continuation {
				continue
			}
			if withColor && current.color != activeColor {
				if current.color == "" {
					output.WriteString(colorReset)
				} else {
					output.WriteString(current.color)
				}
				activeColor = current.color
			}
			output.WriteRune(current.char)
		}
		if row < c.height-1 {
			output.WriteByte('\n')
		}
	}
	if withColor {
		output.WriteString(colorReset)
	}
	return output.String()
}

func main() {
	if err := run(os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "moon:", err)
		os.Exit(1)
	}
}

func run(output io.Writer) error {
	width, height := terminalDimensions()
	if !isTerminal(os.Stdout) {
		_, err := fmt.Fprintln(output, renderFrame(width, height, 80, false))
		return err
	}

	interrupted := make(chan os.Signal, 1)
	signal.Notify(interrupted, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(interrupted)

	// The alternate screen and hidden cursor keep the user's terminal history clean.
	fmt.Fprint(output, "\x1b[?1049h\x1b[?25l")
	defer fmt.Fprint(output, "\x1b[0m\x1b[?25h\x1b[?1049l")

	ticker := time.NewTicker(time.Second / framesPerSecond)
	defer ticker.Stop()
	for frame := 0; ; frame++ {
		fmt.Fprintf(output, "\x1b[H\x1b[2J%s", renderFrame(width, height, frame, true))
		select {
		case <-interrupted:
			return nil
		case <-ticker.C:
		}
	}
}

func renderFrame(width, height, frame int, withColor bool) string {
	width = max(60, width)
	height = max(22, height)
	scene := newCanvas(width, height)
	drawStars(scene, frame)
	drawFireworks(scene, frame)
	drawRotatingCrescent(scene, frame)
	drawFlyingMessage(scene, frame)
	drawIsland(scene)
	return scene.render(withColor)
}

func drawStars(scene *canvas, frame int) {
	sparkles := []rune{'.', '.', '+', '*'}
	for index := 0; index < scene.width/4; index++ {
		x := positiveModulo(index*29+11, scene.width)
		y := 1 + positiveModulo(index*13+3, max(1, scene.height-8))
		phase := positiveModulo(index+frame/4, len(sparkles))
		scene.put(x, y, string(sparkles[phase]), colorStar)
	}
}

func drawRotatingCrescent(scene *canvas, frame int) {
	const (
		spriteWidth  = 13
		spriteHeight = 7
	)
	angle := crescentAngle(frame)
	cosine, sine := math.Cos(angle), math.Sin(angle)
	for row := 0; row < spriteHeight; row++ {
		for column := 0; column < spriteWidth; column++ {
			x := (float64(column) - float64(spriteWidth-1)/2) / (float64(spriteWidth-1) / 2)
			y := (float64(row) - float64(spriteHeight-1)/2) / (float64(spriteHeight-1) / 2)
			// Inverse rotation samples the same crescent while its orientation changes.
			sourceX := x*cosine + y*sine
			sourceY := -x*sine + y*cosine
			outerCircle := sourceX*sourceX+sourceY*sourceY <= 1
			innerX := sourceX - 0.38
			innerCircle := innerX*innerX+sourceY*sourceY <= 0.72
			if outerCircle && !innerCircle {
				scene.put(3+column, 1+row, "●", colorMoon)
			}
		}
	}
}

func drawFireworks(scene *canvas, frame int) {
	const cycleLength = 72
	shows := []struct {
		xPercent int
		targetY  int
		offset   int
		color    string
	}{
		{xPercent: 34, targetY: 4, offset: 0, color: colorFireworkPink},
		{xPercent: 64, targetY: 6, offset: 24, color: colorFireworkGold},
		{xPercent: 84, targetY: 3, offset: 48, color: colorFireworkBlue},
	}

	for _, show := range shows {
		phase := positiveModulo(frame+show.offset, cycleLength)
		x := scene.width * show.xPercent / 100
		drawFirework(scene, x, show.targetY, phase, show.color)
	}
}

func drawFirework(scene *canvas, centerX, targetY, phase int, color string) {
	const (
		launchFrames = 18
		burstEnd     = 44
		fadeEnd      = 58
	)
	if phase < launchFrames {
		startY := scene.height - 5
		progress := float64(phase) / float64(launchFrames-1)
		rocketY := startY - int(math.Round(float64(startY-targetY)*progress))
		scene.put(centerX, rocketY, "|", color)
		scene.put(centerX, rocketY+1, ".", color)
		return
	}
	if phase >= fadeEnd {
		return
	}

	radius := 7
	particle := "."
	if phase < burstEnd {
		radius = 1 + (phase-launchFrames)/4
		particle = "*"
		scene.put(centerX, targetY, "+", color)
	}
	for ray := 0; ray < 16; ray++ {
		angle := float64(ray) * 2 * math.Pi / 16
		// Horizontal distance is doubled to compensate for tall terminal cells.
		x := centerX + int(math.Round(math.Cos(angle)*float64(radius)*2))
		y := targetY + int(math.Round(math.Sin(angle)*float64(radius)))
		scene.put(x, y, particle, color)
	}
}

func crescentAngle(frame int) float64 {
	const halfCycle = 48
	phase := positiveModulo(frame, halfCycle*2)
	if phase > halfCycle {
		phase = halfCycle*2 - phase
	}
	return -math.Pi / 4 * float64(phase) / halfCycle
}

func drawFlyingMessage(scene *canvas, frame int) {
	messageWidth := textDisplayWidth(anniversaryMessage)
	birdWidth := 8
	trailerWidth := messageWidth + 7
	cycleLength := scene.width + birdWidth + trailerWidth
	birdX := positiveModulo(frame, cycleLength) - birdWidth
	birdY := max(8, scene.height/3)

	wingUp := []string{"  \\ /   ", " --(o)> "}
	wingDown := []string{" --(o)> ", "  / \\   "}
	bird := wingUp
	if frame/3%2 == 1 {
		bird = wingDown
	}
	for row, line := range bird {
		scene.put(birdX, birdY+row, line, colorBird)
	}

	heartX := birdX - 4
	messageX := heartX - messageWidth - 2
	scene.put(heartX, birdY+1, "♥", colorHeart)
	scene.put(messageX, birdY+1, anniversaryMessage, colorText)
}

func drawIsland(scene *canvas) {
	groundY := scene.height - 1
	centerX := scene.width / 2

	// Three trees sit directly on the raised island.
	drawTree(scene, centerX-17, groundY-3, 1)
	drawTree(scene, centerX-7, groundY-3, 0)
	drawTree(scene, centerX, groundY-3, 0)

	flag := []string{
		"┌──────┐",
		"│ 月岛 │",
		"└──────┤",
		"       │",
		"       │",
		"       │",
		"       │",
		"       │",
	}
	for row, line := range flag {
		scene.put(centerX+9, groundY-10+row, line, colorFlag)
	}

	island := []string{
		".-~~~~~~~~~~-.",
		"_.-'              '-._",
		"_..-'                      '-.._",
		"____.-'______________________________'-.____",
	}
	for row, line := range island {
		x := centerX - textDisplayWidth(line)/2
		scene.put(x, groundY-len(island)+1+row, line, colorIsland)
	}
}

func drawTree(scene *canvas, x, baseY, variant int) {
	large := []string{"   &&&   ", "  &&&&&  ", " &&&&&&& ", "   |||   "}
	small := []string{"  &&&  ", " &&&&& ", "  |||  "}
	tree := large
	if variant == 0 {
		tree = small
	}
	for row, line := range tree {
		color := colorLeaf
		if row == len(tree)-1 {
			color = colorTrunk
		}
		scene.put(x, baseY-len(tree)+1+row, line, color)
	}
}

func terminalDimensions() (int, int) {
	width := envInt("COLUMNS", defaultWidth)
	height := envInt("LINES", defaultHeight)
	return min(160, max(60, width)), min(60, max(22, height))
}

func envInt(name string, fallback int) int {
	value, err := strconv.Atoi(os.Getenv(name))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func isTerminal(file *os.File) bool {
	info, err := file.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

func positiveModulo(value, modulus int) int {
	return ((value % modulus) + modulus) % modulus
}

func textDisplayWidth(text string) int {
	width := 0
	for _, char := range text {
		width += runeDisplayWidth(char)
	}
	return width
}

func runeDisplayWidth(char rune) int {
	if unicode.Is(unicode.Mn, char) || unicode.Is(unicode.Me, char) || char == '\u200d' {
		return 0
	}
	// These ranges cover the CJK and emoji characters commonly rendered as two columns.
	if char >= 0x1100 && (char <= 0x115f || char == 0x2329 || char == 0x232a ||
		(char >= 0x2e80 && char <= 0xa4cf) || (char >= 0xac00 && char <= 0xd7a3) ||
		(char >= 0xf900 && char <= 0xfaff) || (char >= 0xfe10 && char <= 0xfe19) ||
		(char >= 0xfe30 && char <= 0xfe6f) || (char >= 0xff00 && char <= 0xff60) ||
		(char >= 0xffe0 && char <= 0xffe6) || (char >= 0x1f300 && char <= 0x1faff) ||
		(char >= 0x20000 && char <= 0x3fffd)) {
		return 2
	}
	return 1
}
