// Tests for the MoonAnniversary terminal renderer.
//
// Author: Simon Tian
package main

import (
	"math"
	"strings"
	"testing"
)

func TestSceneContainsRequiredFixedElements(t *testing.T) {
	frame := renderFrame(100, 30, 80, false)

	for _, expected := range []string{"月岛", anniversaryMessage, "&&&", "●"} {
		if !strings.Contains(frame, expected) {
			t.Fatalf("rendered frame does not contain %q", expected)
		}
	}
}

func TestAnimationChangesAcrossFrames(t *testing.T) {
	first := renderFrame(80, 24, 0, false)
	later := renderFrame(80, 24, 12, false)
	if first == later {
		t.Fatal("animation frames should change over time")
	}
}

func TestCrescentRotatesAcrossFrames(t *testing.T) {
	first := newCanvas(30, 12)
	tilted := newCanvas(30, 12)
	drawRotatingCrescent(first, 0)
	drawRotatingCrescent(tilted, 48)

	if first.render(false) == tilted.render(false) {
		t.Fatal("crescent should change orientation at its maximum tilt")
	}
}

func TestCrescentUsesLeftHalf(t *testing.T) {
	scene := newCanvas(30, 12)
	drawRotatingCrescent(scene, 0)
	left, right := 0, 0
	for _, row := range scene.cells {
		for column, current := range row {
			if current.color != colorMoon {
				continue
			}
			if column < 9 {
				left++
			} else if column > 9 {
				right++
			}
		}
	}
	if left <= right {
		t.Fatalf("left crescent has %d left-side cells and %d right-side cells", left, right)
	}
}

func TestFireworksHaveVisibleBurstParticles(t *testing.T) {
	scene := newCanvas(80, 24)
	drawFireworks(scene, 30)
	visibleParticles := 0
	for _, row := range scene.cells {
		for _, current := range row {
			if current.color == colorFireworkPink || current.color == colorFireworkGold || current.color == colorFireworkBlue {
				visibleParticles++
			}
		}
	}
	if visibleParticles < 10 {
		t.Fatalf("firework frame has only %d visible particles", visibleParticles)
	}
}

func TestCrescentAngleOscillatesBetweenZeroAndMinusFortyFiveDegrees(t *testing.T) {
	if angle := crescentAngle(0); angle != 0 {
		t.Fatalf("starting angle = %f, want 0", angle)
	}
	if angle := crescentAngle(48); angle != -math.Pi/4 {
		t.Fatalf("maximum tilt = %f, want %f", angle, -math.Pi/4)
	}
	if angle := crescentAngle(96); angle != 0 {
		t.Fatalf("return angle = %f, want 0", angle)
	}
}

func TestCanvasClipsTextOutsideBounds(t *testing.T) {
	frame := newCanvas(5, 2)
	frame.put(-2, 0, "abcdefghi", "")

	firstLine := strings.Split(frame.render(false), "\n")[0]
	if firstLine != "cdefg" {
		t.Fatalf("unexpected clipped line: %q", firstLine)
	}
}

func TestPositiveModulo(t *testing.T) {
	if actual := positiveModulo(-1, 5); actual != 4 {
		t.Fatalf("positiveModulo(-1, 5) = %d, want 4", actual)
	}
}

func TestTextDisplayWidthSupportsChinese(t *testing.T) {
	if actual := textDisplayWidth("月岛A"); actual != 5 {
		t.Fatalf("textDisplayWidth(\"月岛A\") = %d, want 5", actual)
	}
}
