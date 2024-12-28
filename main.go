package main

import (
	"errors"
	"flag"
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
var fx, fy, sx, sy, mines int
var screen tcell.Screen
var gameTime time.Time
var timr int
var symbols []rune
var numbers []rune
var mousey bool
var mouseThing string
var asciiDigits map[rune][]string
var gameOver bool
var flags int

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
	asciiDigits = map[rune][]string{
		'0': {
			"┏━┓",
			"┃ ┃",
			"┗━┛",
		},
		'1': {
			"┏┓ ",
			" ┃ ",
			"━┻━",
		},
		'2': {
			"┏━┓",
			"┏━┛",
			"┗━┛",
		},
		'3': {
			"┏━┓",
			" ━┫",
			"┗━┛",
		},
		'4': {
			"╻  ",
			"┗━╋",
			"  ╹",
		},
		'5': {
			"┏━┓",
			"┗━┓",
			"┗━┛",
		},
		'6': {
			"┏━┓",
			"┣━┓",
			"┗━┛",
		},
		'7': {
			"━━┓",
			"  ┃",
			"  ╹",
		},
		'8': {
			"┏━┓",
			"┣━┫",
			"┗━┛",
		},
		'9': {
			"┏━┓",
			"┗━┫",
			"┗━┛",
		},
		'-': {
			"   ",
			"╺━╸",
			"   ",
		},
	}
	sx_o := flag.Int("w", 9, "Width of board")
	sy_o := flag.Int("h", 9, "Height of board")
	m_o := flag.Int("m", 10, "No. of mines on board")
	flag.Parse()
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
	start = 0
	pline = 7
	fx = -1
	fy = -1
	mines = *m_o
	sx = *sx_o
	sy = *sy_o
	flags = mines
	width, height := screen.Size()
	if width <= sx*5+5 || height <= sy*3+5 {
		screen.Suspend()
		fmt.Println("Sorry screen size not enough for playing this size of board!")
		return
	} else if sx <= 7 || sy <= 2 {
		screen.Suspend()
		fmt.Println("Sorry size provided is too small - minimum width: 6 & height: 2!")
		return
	}
	offx = width/2 - (sx*5)/2
	offy = height/2 - (sy*3-3)/2
	board, _ := gen_board(sx, sy, mines, fx, fy)
	refresh(board)
	screen.Show()
	go func() {
		for {
			updateTime()
			time.Sleep(500 * time.Millisecond)
		}
	}()
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
			doit := true
			button := buttonify(ev.Buttons())
			if x >= offx+(sx*5/2)-3 && x <= offx+(sx*5/2)+3 && y >= offy-4 && y <= offy-3 && button == "L" {
				fx = -1
				fy = -1
				gameOver = false
				flags = mines
				gameTime = time.Time{}
				board, _ := gen_board(sx, sy, mines, fx, fy)
				refresh(board)
				screen.Show()
				doit = false
				continue
			}
			x = x - offx
			y = y - offy
			boardX := int(math.Floor(float64(x) / 5))
			boardY := int(math.Floor(float64(y) / 3))
			if fx == -1 && fy == -1 && (button == "L" || button == "R") && doit {
				fx = boardX
				fy = boardY
				board, _ = gen_board(sx, sy, mines, fx, fy)
				gameTime = time.Now()
			}
			if !gameOver {
				if (button == "L" || button == "R") && !mousey {
					mouseThing = button
					if mouseThing == "L" && !(boardX < 0 || boardX >= sx || boardY < 0 || boardY >= sy) && !board[boardY][boardX].flag {
						_ = reveal(board, boardX, boardY, sx, sy)
					} else if mouseThing == "R" {
						_ = flag_it(board, boardX, boardY, sx, sy)
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
			offy = height/2 - (sy*3-3)/2
			if height-sy*3+5 > 5 {
				str := `
┏┓         ┳┳┓ •
┃┓ ┏┓  ━━  ┃┃┃ ┓ ┏┓ ┏┓ ┏
┗┛ ┗┛      ┛ ┗ ┗ ┛┗ ┗  ┛
				`
				print_at(width/2-12, 1, strings.Split(str, "\n")[1], tcell.StyleDefault)
				print_at(width/2-12, 2, strings.Split(str, "\n")[2], tcell.StyleDefault)
				print_at(width/2-12, 3, strings.Split(str, "\n")[3], tcell.StyleDefault)
			}
			print_it(fmt.Sprintf("Welcome to go-mines! %c  %c", symbols[4], symbols[5]))
			refresh(board)
			screen.Show()
		}
	}
}

func updateTime() {
	if !gameTime.IsZero() {
		elapsedTime := time.Since(gameTime)
		seconds := int(elapsedTime.Seconds())
		if seconds > 999 {
			seconds = 999
		}
		secondsStr := fmt.Sprintf("%03d", seconds)
		timr = seconds
		timeLines := []string{"", "", ""}
		for _, digit := range secondsStr {
			digitArt := asciiDigits[digit]
			for i := 0; i < 3; i++ {
				timeLines[i] += " " + digitArt[i] + " "
			}
		}
		for i, line := range timeLines {
			print_at(sx*5+offx-15, offy-4+i, line, tcell.StyleDefault.Foreground(tcell.GetColor("#cc0000")).Background(tcell.GetColor("#000000")))
		}
		screen.Show()
	} else {
		secondsStr := fmt.Sprintf("%03d", timr)
		timeLines := []string{"", "", ""}
		for _, digit := range secondsStr {
			digitArt := asciiDigits[digit]
			for i := 0; i < 3; i++ {
				timeLines[i] += " " + digitArt[i] + " "
			}
		}
		for i, line := range timeLines {
			print_at(sx*5+offx-15, offy-4+i, line, tcell.StyleDefault.Foreground(tcell.GetColor("#cc0000")).Background(tcell.GetColor("#000000")))
		}
		screen.Show()
	}
}

func refresh(board [][]cell) {
	solved := true
	state := 0
	for i := 0; i < sy; i++ {
		for j := 0; j < sx; j++ {
			if board[i][j].isMine && board[i][j].reveal {
				solved = false
				gameTime = time.Time{}
				state = -1
				break
			} else if !board[i][j].isMine && !board[i][j].reveal {
				solved = false
				break
			}
		}
	}
	if solved {
		for i := 0; i < sy; i++ {
			for j := 0; j < sx; j++ {
				if board[i][j].isMine {
					board[i][j].flag = true
				}
			}
		}
		print_it("You won!")
		gameOver = true
		gameTime = time.Time{}
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

func flag_it(board [][]cell, x, y, width, height int) error {
	if x < 0 || x >= width || y < 0 || y >= height {
		return nil
	}
	if board[y][x].reveal {
		return nil
	}
	if board[y][x].flag {
		flags++
		board[y][x].flag = false
	} else {
		flags--
		board[y][x].flag = true
	}
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
			board[y][x].mines = count
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

func print_board(board [][]cell, solved int) {
	style := tcell.StyleDefault.Foreground(tcell.GetColor("#464e56"))
	style2 := tcell.StyleDefault.Background(tcell.GetColor("#464e56")).Foreground(tcell.GetColor("#1e262e"))
	draw_box(offx-2, offy-1, sx*5+offx+1, sy*3+offy, style, "██████")
	draw_box(offx-3, offy-2, sx*5+offx+2, sy*3+offy+1, style, "██████")
	draw_box(offx-3, offy-3, sx*5+offx+2, sy*3+offy+1, style, "██████")
	draw_box(offx-3, offy-4, sx*5+offx+2, sy*3+offy+1, style, "██████")
	draw_box(offx-4, offy-5, sx*5+offx+3, sy*3+offy+1, style, "██████")
	draw_box(offx-5, offy-6, sx*5+offx+4, sy*3+offy+2, style2.Background(tcell.GetColor("#00000000")).Foreground(tcell.GetColor("#788088")), "🬭▌🬞🬁🬏🬀▐🬂")
	draw_box(offx-1, offy-5, offx+15, offy-1, style2, "🬭▌🬞🬁🬏🬀▐🬂")
	draw_box(offx-1, offy-1, sx*5+offx, sy*3+offy, style2, "🬭▌🬞🬁🬏🬀▐🬂")
	print_at(offx-1, offy-1, "🬠🬰🬰🬰🬰🬰🬰🬰🬰🬰🬰🬰🬰🬰🬰🬰🬮", style2)
	draw_box(sx*5+offx-16, offy-5, sx*5+offx, offy-1, style2, "🬭▌🬞🬁🬏🬀▐🬂")
	print_at(sx*5+offx-16, offy-1, "🬯🬰🬰🬰🬰🬰🬰🬰🬰🬰🬰🬰🬰🬰🬰🬰🬐", style2)
	if solved == -1 {
		print_at(offx+(sx*5/2)-3, offy-3, " X _ X ", style2.Foreground(tcell.GetColor("#da9160")))
		draw_box(offx+(sx*5/2)-3, offy-4, offx+(sx*5/2)+3, offy-2, style2.Foreground(tcell.GetColor("#da9160")), "")
	} else if solved == 0 {
		print_at(offx+(sx*5/2)-3, offy-3, " • ‿ • ", style2.Foreground(tcell.GetColor("#f5f646")))
		draw_box(offx+(sx*5/2)-3, offy-4, offx+(sx*5/2)+3, offy-2, style2.Foreground(tcell.GetColor("#f5f646")), "")
	} else {
		print_at(offx+(sx*5/2)-3, offy-3, " ^ o ^ ", style2.Foreground(tcell.GetColor("#66c266")))
		draw_box(offx+(sx*5/2)-3, offy-4, offx+(sx*5/2)+3, offy-2, style2.Foreground(tcell.GetColor("#66c266")), "")
	}
	chars_i := "▔▕🭽🭼🭾🭿▏▁"
	symbol := "H"
	for i := 0; i < sy; i++ {
		for j := 0; j < sx; j++ {
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
	finalStr := fmt.Sprintf("%03d", flags)
	lines := []string{"", "", ""}
	for _, digit := range finalStr {
		digitArt := asciiDigits[digit]
		for i := 0; i < 3; i++ {
			lines[i] += " " + digitArt[i] + " "
		}
	}
	for i, line := range lines {
		print_at(offx, offy-4+i, line, tcell.StyleDefault.Foreground(tcell.GetColor("#cc0000")).Background(tcell.GetColor("#000000")))
	}
	updateTime()
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
