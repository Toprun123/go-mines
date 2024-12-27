package main

import (
	"errors"
	"fmt"
	"log"
	"math"
	"math/rand"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
)

var start, pline int

var offx, offy int

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
	pline = 20

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

	var fx, fy, sx, sy, mines int
	fx = -1
	fy = -1
	mines = 10
	sx = 9
	sy = 9
	width, height := screen.Size()
	offx = width/2 - sx*5/2
	offy = height/2 - sy*3/2
	board, _ := gen_board(sx, sy, mines, fx, fy)
	refresh(board)
	screen.Show()

	for {
		ev := screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			if ev.Rune() == 'l' || ev.Rune() == 'q' {
				return
			}
			if ev.Key() == tcell.KeyCtrlC {
				return
			}
		case *tcell.EventMouse:
			x, y := ev.Position()
			button := buttonify(ev.Buttons())
			if x >= offx+(len(board[0])*5/2)-3 && x <= offx+(len(board[0])*5/2)+3 && y >= offy-4 && y <= offy-3 && button == "L" {
				fx = -1
				fy = -1
				gameOver = false
				board, _ := gen_board(sx, sy, mines, fx, fy)
				refresh(board)
				screen.Show()
				continue
			}
			x = x - offx
			y = y - offy
			boardX := int(math.Floor(float64(x) / 5))
			boardY := int(math.Floor(float64(y) / 3))
			if fx == -1 && fy == -1 && (button == "L" || button == "R") {
				fx = boardX
				fy = boardY
				board, _ = gen_board(sx, sy, mines, fx, fy)
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
			screen.Sync()
			screen.Clear()
			width, height := screen.Size()
			offx = width/2 - sx*5/2
			offy = height/2 - sy*3/2
			print_it(fmt.Sprintf("Welcome to go-mines! %c  %c", symbols[4], symbols[5]))
			refresh(board)
			screen.Show()
		}
	}
}

func refresh(board [][]cell) {
	solved := true
	state := 0
	for i := 0; i < len(board); i++ {
		for j := 0; j < len(board[0]); j++ {
			if board[i][j].isMine && board[i][j].reveal {
				solved = false
				state = -1
				break
			} else if !board[i][j].isMine && !board[i][j].reveal {
				solved = false
				break
			}
		}
	}
	if solved {
		for i := 0; i < len(board); i++ {
			for j := 0; j < len(board[0]); j++ {
				if board[i][j].isMine {
					board[i][j].flag = true
				}
			}
		}
		print_it("You won!")
		gameOver = true
		state = 1
	}
	print_board(board, state)
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
		board[y][x].isMine = true
		mineCount++
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

func print_board(board [][]cell, solved int) {
	style := tcell.StyleDefault.Foreground(tcell.GetColor("#464e56"))
	style2 := tcell.StyleDefault.Background(tcell.GetColor("#464e56")).Foreground(tcell.GetColor("#1e262e"))
	draw_box(offx-2, offy-1, len(board[0])*5+offx+1, len(board)*3+offy, style, "██████")
	draw_box(offx-3, offy-2, len(board[0])*5+offx+2, len(board)*3+offy+1, style, "██████")
	draw_box(offx-3, offy-3, len(board[0])*5+offx+2, len(board)*3+offy+1, style, "██████")
	draw_box(offx-3, offy-4, len(board[0])*5+offx+2, len(board)*3+offy+1, style, "██████")
	draw_box(offx-4, offy-5, len(board[0])*5+offx+3, len(board)*3+offy+1, style, "██████")
	draw_box(offx-5, offy-6, len(board[0])*5+offx+4, len(board)*3+offy+2, style2.Background(tcell.GetColor("#00000000")).Foreground(tcell.GetColor("#788088")), "🬭▌🬞🬁🬏🬀▐🬂")
	draw_box(offx-1, offy-1, len(board[0])*5+offx, len(board)*3+offy, style2, "🬭▌🬞🬁🬏🬀▐🬂")
	if solved == -1 {
		print_at(offx+(len(board[0])*5/2)-3, offy-3, " X _ X ", style2.Foreground(tcell.GetColor("#da9160")))
		draw_box(offx+(len(board[0])*5/2)-3, offy-4, offx+(len(board[0])*5/2)+3, offy-2, style2.Foreground(tcell.GetColor("#da9160")), "")
	} else if solved == 0 {
		print_at(offx+(len(board[0])*5/2)-3, offy-3, " • ‿ • ", style2.Foreground(tcell.GetColor("#f5f646")))
		draw_box(offx+(len(board[0])*5/2)-3, offy-4, offx+(len(board[0])*5/2)+3, offy-2, style2.Foreground(tcell.GetColor("#f5f646")), "")
	} else {
		print_at(offx+(len(board[0])*5/2)-3, offy-3, " • ‿ • ", style2.Foreground(tcell.GetColor("#66c266")))
		draw_box(offx+(len(board[0])*5/2)-3, offy-4, offx+(len(board[0])*5/2)+3, offy-2, style2.Foreground(tcell.GetColor("#66c266")), "")
	}
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
			print_at(j*5+offx, i*3+1+offy, fmt.Sprint("  ", symbol, "  "), style2)
			draw_box(j*5+offx, i*3+offy, j*5+4+offx, i*3+2+offy, style, chars_i)
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
