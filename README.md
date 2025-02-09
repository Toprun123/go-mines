# Minesweeper (Terminal Edition)

A terminal-based Minesweeper game written in Go, using the `tcell` library for handling terminal UI and events.

## Features
- Classic Minesweeper gameplay
- Intuitive keyboard controls
- Terminal-based UI
- Real-time updates with `tcell`

## Installation
### Prerequisites
- Go 1.18+ installed

### Steps
1. Clone this repository:
   ```sh
   git clone https://github.com/Toprun123/go-mines.git
   cd minesweeper-terminal
   ```
2. Install dependencies:
   ```sh
   go get ./...
   ```
3. Build the project:
   ```sh
   go build -o minesweeper
   ```
4. Run the game:
   ```sh
   ./minesweeper
   ```

## Controls
- `Arrow Keys` - Move cursor
- `Enter` - Reveal cell
- `F` - Flag/Unflag cell
- `Q` - Quit game

Or Mouse Clicks (Left Click to Reveal, Right Click to Flag/Unflag)

## Contributing
Feel free to open issues or submit pull requests!

## Author
Syed - https://github.com/Toprun123
