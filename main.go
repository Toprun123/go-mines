package main

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
)

var start int
var pline int

var screen tcell.Screen

var symbols []rune
var numbers []rune

var mousey bool
var mouseThing string

var gameOver bool

type cell struct {
	err    bool
	mines  int
	flag   bool
	reveal bool
	isMine bool
	color  string
}

func main() {
	symbols = []rune("󰸞󰷚󰉀")
	numbers = []rune("🯰🯱🯲🯳🯴🯵🯶🯷🯸🯹")
	start = 0
	pline = 30

	screen_tmp, err := tcell.NewScreen()
	if err != nil {
		log.Fatalf("Failed to create screen: %v", err)
	}
	defer screen_tmp.Fini()
	if err := screen_tmp.Init(); err != nil {
		log.Fatalf("Failed to initialize screen: %v", err)
	}
	screen = screen_tmp
	screen.EnableMouse()
	screen.Clear()
	screen.Show()

	var fx, fy int
	board, _ := gen_board(10, 10, 15, fx, fy)
	print_it(fmt.Sprintf("Welcome to go-mines! %c  %c", symbols[4], symbols[5]))
	refresh(board)

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
		case *tcell.EventMouse:
			x, y := ev.Position()
			button := buttonify(ev.Buttons())
			boardX := x / 5
			boardY := y / 3
			if fx == 0 && fy == 0 {
				fx = boardX
				fx = boardY
				board, _ = gen_board(10, 10, 15, fx, fy)
			}
			if !gameOver {
				if (button == "L" || button == "R") && !mousey {
					mouseThing = button
					if mouseThing == "L" {
						_ = reveal(board, boardX, boardY, len(board[0]), len(board))
					} else if mouseThing == "R" {
						_ = flag(board, boardX, boardY, len(board[0]), len(board))
					}
					mousey = true
					mouseThing = ""
				} else if button != "L" && button != "R" {
					mousey = false
				}
			}
			screen.Show()
		case *tcell.EventResize:
			// screen.Sync()
			// screen.Clear()
			// refresh(board)
			// screen.Show()
		}
	}
}

func refresh(board [][]cell) {
	solved := true
	for i := 0; i < len(board); i++ {
		for j := 0; j < len(board[0]); j++ {
			if board[i][j].isMine && board[i][j].reveal {
				solved = false
				break
			} else if !board[i][j].isMine && !board[i][j].reveal {
				solved = false
				break
			}
		}
	}
	if solved {
		print_it("You won!")
		gameOver = true
	}
	print_board(board)
}

func reveal(board [][]cell, x, y, width, height int) error {
	if x < 0 || x >= width || y < 0 || y >= height {
		return nil
	}
	if board[y][x].reveal {
		return nil
	}
	board[y][x].flag = false
	board[y][x].reveal = true
	if board[y][x].isMine {
		board[y][x].err = true
		print_it("Game Over! You hit a mine!")
		gameOver = true
		for i := 0; i < height; i++ {
			for j := 0; j < width; j++ {
				if board[i][j].isMine {
					board[i][j].reveal = true
				} else if board[i][j].flag && !board[i][j].isMine {
					board[i][j].err = true
				}
			}
		}
		refresh(board)
		return nil
	}
	if board[y][x].mines == 0 {
		directions := []struct{ dx, dy int }{
			{-1, -1}, {0, -1}, {1, -1},
			{-1, 0}, {1, 0},
			{-1, 1}, {0, 1}, {1, 1},
		}
		for _, dir := range directions {
			nx, ny := x+dir.dx, y+dir.dy
			if nx >= 0 && nx < width && ny >= 0 && ny < height {
				if !board[ny][nx].reveal {
					reveal(board, nx, ny, width, height)
				}
			}
		}
	}
	refresh(board)
	return nil
}

func flag(board [][]cell, x, y, width, height int) error {
	if x < 0 || x >= width || y < 0 || y >= height {
		return nil
	}
	if board[y][x].reveal {
		return nil
	}
	board[y][x].flag = !board[y][x].flag
	refresh(board)
	return nil
}

func gen_board(width, height, mines, fx, fy int) ([][]cell, error) {
	board := make([][]cell, height)
	for i := range board {
		board[i] = make([]cell, width)
	}
	if mines >= width*height {
		return board, errors.New("too many mines")
	}
	source := rand.NewSource(time.Now().UnixNano())
	rand := rand.New(source)
	mineCount := 0
	for mineCount < mines {
		x := rand.Intn(width)
		y := rand.Intn(height)
		if (x == fx && y == fy) || board[y][x].isMine {
			continue
		}
		if !board[y][x].isMine {
			board[y][x].isMine = true
			mineCount++
		}
	}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if board[y][x].isMine {
				continue
			}
			board[y][x].mines = count_mines(board, x, y, width, height)
			colors := map[int]string{
				0: "#41e8b0",
				1: "#7cc7ff",
				2: "#66c266",
				3: "#ff7788",
				4: "#ee88ff",
				5: "#ddaa22",
				6: "#66cccc",
				7: "#999999",
				8: "#d0d8e0",
			}
			board[y][x].color = colors[board[y][x].mines]
		}
	}
	return board, nil
}

func count_mines(board [][]cell, x, y, width, height int) int {
	directions := []struct{ dx, dy int }{
		{-1, -1}, {0, -1}, {1, -1},
		{-1, 0}, {1, 0},
		{-1, 1}, {0, 1}, {1, 1},
	}
	count := 0
	for _, dir := range directions {
		nx, ny := x+dir.dx, y+dir.dy
		if nx >= 0 && nx < width && ny >= 0 && ny < height && board[ny][nx].isMine {
			count++
		}
	}
	return count
}

func print_board(board [][]cell) {
	style := tcell.StyleDefault.Foreground(tcell.ColorWhite)
	style2 := tcell.StyleDefault.Foreground(tcell.ColorWhite)
	chars_i := "▔▕🭽🭼🭾🭿▏▁"
	symbol := "H"
	for i := 0; i < len(board); i++ {
		for j := 0; j < len(board[0]); j++ {
			chars_i = "▔▕🭽🭼🭾🭿▏▁"
			if board[i][j].mines == 0 && board[i][j].reveal && !board[i][j].isMine {
				symbol = ""
				style = tcell.StyleDefault.Background(tcell.GetColor("#384048")).Foreground(tcell.GetColor("#1f272f"))
				style2 = tcell.StyleDefault.Background(tcell.GetColor("#384048")).Foreground(tcell.GetColor("#1f272f"))
			} else if board[i][j].flag && !board[i][j].err {
				symbol = string(symbols[5])
				style = tcell.StyleDefault.Background(tcell.GetColor("#4c545c")).Foreground(tcell.GetColor("#707880"))
				style2 = tcell.StyleDefault.Background(tcell.GetColor("#4c545c")).Foreground(tcell.GetColor("#ff7d7d"))
				chars_i = "🬂▐🬕🬲🬨🬷▌🬭"
			} else if board[i][j].flag && board[i][j].err {
				symbol = string(symbols[5])
				style = tcell.StyleDefault.Background(tcell.GetColor("#883333")).Foreground(tcell.GetColor("#671212"))
				style2 = tcell.StyleDefault.Background(tcell.GetColor("#883333")).Foreground(tcell.GetColor("#ff7d7d"))
				chars_i = "🬂▐🬕🬲🬨🬷▌🬭"
			} else if board[i][j].isMine && board[i][j].reveal && !board[i][j].err {
				symbol = string(symbols[4])
				style = tcell.StyleDefault.Background(tcell.GetColor("#464e56")).Foreground(tcell.GetColor("#1f272f"))
				style2 = tcell.StyleDefault.Background(tcell.GetColor("#464e56")).Foreground(tcell.GetColor("#000000"))
			} else if board[i][j].isMine && board[i][j].reveal && board[i][j].err {
				symbol = string(symbols[4])
				style = tcell.StyleDefault.Background(tcell.GetColor("#ee6666")).Foreground(tcell.GetColor("#1f272f"))
				style2 = tcell.StyleDefault.Background(tcell.GetColor("#ee6666")).Foreground(tcell.GetColor("#000000"))
			} else if !board[i][j].reveal {
				symbol = ""
				style = tcell.StyleDefault.Background(tcell.GetColor("#4c545c")).Foreground(tcell.GetColor("#707880")).Bold(true)
				style2 = tcell.StyleDefault.Background(tcell.GetColor("#4c545c")).Foreground(tcell.GetColor("#707880")).Bold(true)
				chars_i = "🬂▐🬕🬲🬨🬷▌🬭"
			} else {
				symbol = string(numbers[board[i][j].mines])
				style = tcell.StyleDefault.Background(tcell.GetColor("#384048")).Foreground(tcell.GetColor("#1f272f")).Bold(true)
				style2 = tcell.StyleDefault.Background(tcell.GetColor("#384048")).Foreground(tcell.GetColor(board[i][j].color)).Bold(true)
			}
			print_at(j*5, i*3+1, fmt.Sprint("  ", symbol, "  "), style2)
			draw_box(j*5, i*3, j*5+4, i*3+2, style, chars_i)
		}
	}
}

func buttonify(button tcell.ButtonMask) string {
	var parts []string
	if button&tcell.Button1 != 0 {
		parts = append(parts, "L")
	}
	if button&tcell.Button2 != 0 {
		parts = append(parts, "R")
	}
	if button&tcell.Button3 != 0 {
		parts = append(parts, "M")
	}
	if button&tcell.WheelUp != 0 {
		parts = append(parts, "U")
	}
	if button&tcell.WheelDown != 0 {
		parts = append(parts, "D")
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, "")
}

func print_it(str string) {
	screen.SetContent(9, pline, symbols[3], nil, tcell.StyleDefault.Background(tcell.GetColor("#00000000")).Foreground(tcell.GetColor("#FFFF00")))
	screen.SetContent(10, pline, ' ', nil, tcell.StyleDefault.Background(tcell.GetColor("#00000000")).Foreground(tcell.GetColor("#FFFFFF")))
	for i, char := range []rune(str) {
		screen.SetContent(12+i, pline, char, nil, tcell.StyleDefault.Background(tcell.GetColor("#00000000")).Foreground(tcell.GetColor("#FFFFFF")))
	}
	pline++
}

func print_at(x, y int, str string, style tcell.Style) {
	for i, char := range []rune(str) {
		screen.SetContent(x+i, y, char, nil, style)
	}
}

func draw_box(x1, y1, x2, y2 int, style tcell.Style, chars_i string) {
	if chars_i == "" {
		chars_i = "─│╭╰╮╯" // "━┃┏┗┓┛"
	}
	// print_it(s, fmt.Sprintf("Result: %d", utf8.RuneCountInString(chars_i)), style)
	if utf8.RuneCountInString(chars_i) == 6 {
		chars := []rune(chars_i)
		for x := x1; x <= x2; x++ {
			screen.SetContent(x, y1, chars[0], nil, style)
		}
		for x := x1; x <= x2; x++ {
			screen.SetContent(x, y2, chars[0], nil, style)
		}
		for x := y1 + 1; x < y2; x++ {
			screen.SetContent(x1, x, chars[1], nil, style)
		}
		for x := y1 + 1; x < y2; x++ {
			screen.SetContent(x2, x, chars[1], nil, style)
		}
		screen.SetContent(x1, y1, chars[2], nil, style)
		screen.SetContent(x1, y2, chars[3], nil, style)
		screen.SetContent(x2, y1, chars[4], nil, style)
		screen.SetContent(x2, y2, chars[5], nil, style)
	} else {
		chars := []rune(chars_i)
		for x := x1; x <= x2; x++ {
			screen.SetContent(x, y1, chars[0], nil, style)
		}
		for x := x1; x <= x2; x++ {
			screen.SetContent(x, y2, chars[7], nil, style)
		}
		for x := y1 + 1; x < y2; x++ {
			screen.SetContent(x1, x, chars[6], nil, style)
		}
		for x := y1 + 1; x < y2; x++ {
			screen.SetContent(x2, x, chars[1], nil, style)
		}
		screen.SetContent(x1, y1, chars[2], nil, style)
		screen.SetContent(x1, y2, chars[3], nil, style)
		screen.SetContent(x2, y1, chars[4], nil, style)
		screen.SetContent(x2, y2, chars[5], nil, style)
	}
}
