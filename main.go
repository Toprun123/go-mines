package main

import (
	"log"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
)

var start int
var pline int

var symbols []rune

func main() {
	symbols = []rune("󰞔󰪸󰸞")
	// Initialize a new screen
	screen, err := tcell.NewScreen()
	if err != nil {
		log.Fatalf("Failed to create screen: %v", err)
	}
	defer screen.Fini()

	if err := screen.Init(); err != nil {
		log.Fatalf("Failed to initialize screen: %v", err)
	}

	// Clear the screen
	screen.Clear()
	// Show the screen
	screen.Show()

	// Wait for an event
	for {
		ev := screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			if ev.Rune() == 'q' {
				return
			}
			if ev.Key() == tcell.KeyCtrlC {
				return
			}
		case *tcell.EventResize:
			screen.Sync()
			screen.Clear()
			refresh(screen)
			screen.Show()
		}
	}
}

func refresh(screen tcell.Screen) {
	start = 0
	pline = 25
	// width, height := screen.Size()
	// Draw some text
	// style := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)
	print_it(screen, "Welcome to go-mines!")
	draw_box(screen, 7, 24, 47, pline, tcell.StyleDefault.Background(tcell.GetColor("#00000000")).Foreground(tcell.GetColor("#FFFF00")), "")
}

func print_it(s tcell.Screen, str string) {
	s.SetContent(9, pline, symbols[7], nil, tcell.StyleDefault.Background(tcell.GetColor("#00000000")).Foreground(tcell.GetColor("#FFFF00")))
	s.SetContent(10, pline, ' ', nil, tcell.StyleDefault.Background(tcell.GetColor("#00000000")).Foreground(tcell.GetColor("#FFFFFF")))
	for i, char := range []rune(str) {
		s.SetContent(12+i, pline, char, nil, tcell.StyleDefault.Background(tcell.GetColor("#00000000")).Foreground(tcell.GetColor("#FFFFFF")))
	}
	pline++
}

func draw_box(s tcell.Screen, x1, y1, x2, y2 int, style tcell.Style, chars_i string) {
	if chars_i == "" {
		chars_i = "─│╭╰╮╯" // "━┃┏┗┓┛"
	}
	// print_it(s, fmt.Sprintf("Result: %d", utf8.RuneCountInString(chars_i)), style)
	if utf8.RuneCountInString(chars_i) == 6 {
		chars := []rune(chars_i)
		for x := x1; x <= x2; x++ {
			s.SetContent(x, y1, chars[0], nil, style)
		}
		for x := x1; x <= x2; x++ {
			s.SetContent(x, y2, chars[0], nil, style)
		}
		for x := y1 + 1; x < y2; x++ {
			s.SetContent(x1, x, chars[1], nil, style)
		}
		for x := y1 + 1; x < y2; x++ {
			s.SetContent(x2, x, chars[1], nil, style)
		}
		s.SetContent(x1, y1, chars[2], nil, style)
		s.SetContent(x1, y2, chars[3], nil, style)
		s.SetContent(x2, y1, chars[4], nil, style)
		s.SetContent(x2, y2, chars[5], nil, style)
	} else {
		chars := []rune(chars_i)
		for x := x1; x <= x2; x++ {
			s.SetContent(x, y1, chars[0], nil, style)
		}
		for x := x1; x <= x2; x++ {
			s.SetContent(x, y2, chars[7], nil, style)
		}
		for x := y1 + 1; x < y2; x++ {
			s.SetContent(x1, x, chars[6], nil, style)
		}
		for x := y1 + 1; x < y2; x++ {
			s.SetContent(x2, x, chars[1], nil, style)
		}
		s.SetContent(x1, y1, chars[2], nil, style)
		s.SetContent(x1, y2, chars[3], nil, style)
		s.SetContent(x2, y1, chars[4], nil, style)
		s.SetContent(x2, y2, chars[5], nil, style)
	}
}
