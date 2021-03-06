package main

import (
	tl "github.com/joelotter/termloop"
	"math/rand"
	"strconv"
	"time"
)

////////////////////////
// Maze generation stuff
////////////////////////

type Point struct {
	x int
	y int
	p *Point
}

func (p *Point) Opposite() *Point {
	if p.x != p.p.x {
		return &Point{x: p.x + (p.x - p.p.x), y: p.y, p: p}
	}
	if p.y != p.p.y {
		return &Point{x: p.x, y: p.y + (p.y - p.p.y), p: p}
	}
	return nil
}

func adjacents(point *Point, maze [][]rune) []*Point {
	res := make([]*Point, 0)
	for i := -1; i < 2; i++ {
		for j := -1; j < 2; j++ {
			if (i == 0 && j == 0) || (i != 0 && j != 0) {
				continue
			}
			if !isInMaze(point.x+i, point.y+j, len(maze), len(maze[0])) {
				continue
			}
			if maze[point.x+i][point.y+j] == '*' {
				res = append(res, &Point{point.x + i, point.y + j, point})
			}
		}
	}
	return res
}

func isInMaze(x, y int, w, h int) bool {
	return x >= 0 && x < w &&
		y >= 0 && y < h
}

// Generates a maze using Prim's Algorithm
// https://en.wikipedia.org/wiki/Maze_generation_algorithm#Randomized_Prim.27s_algorithm
func generateMaze(w, h int) [][]rune {
	maze := make([][]rune, w)
	for row := range maze {
		maze[row] = make([]rune, h)
		for ch := range maze[row] {
			maze[row][ch] = '*'
		}
	}
	rand.Seed(time.Now().UnixNano())
	point := &Point{x: rand.Intn(w), y: rand.Intn(h)}
	maze[point.x][point.y] = 'S'
	var last *Point
	walls := adjacents(point, maze)
	for len(walls) > 0 {
		rand.Seed(time.Now().UnixNano())
		wall := walls[rand.Intn(len(walls))]
		for i, w := range walls {
			if w.x == wall.x && w.y == wall.y {
				walls = append(walls[:i], walls[i+1:]...)
				break
			}
		}
		opp := wall.Opposite()
		if isInMaze(opp.x, opp.y, w, h) && maze[opp.x][opp.y] == '*' {
			maze[wall.x][wall.y] = '.'
			maze[opp.x][opp.y] = '.'
			walls = append(walls, adjacents(opp, maze)...)
			last = opp
		}
	}
	maze[last.x][last.y] = 'L'
	bordered := make([][]rune, len(maze)+2)
	for r := range bordered {
		bordered[r] = make([]rune, len(maze[0])+2)
		for c := range bordered[r] {
			if r == 0 || r == len(maze)+1 || c == 0 || c == len(maze[0])+1 {
				bordered[r][c] = '*'
			} else {
				bordered[r][c] = maze[r-1][c-1]
			}
		}
	}
	return bordered
}

/////////////////
// Termloop stuff
/////////////////

type Block struct {
	r         *tl.Rectangle
	px        int // Previous x
	py        int // Previous y
	move      bool
	g         *tl.Game
	w         int // Width of maze
	h         int // Height of maze
	score     int
	scoretext *tl.Text
}

func NewBlock(x, y int, color tl.Attr, g *tl.Game, w, h, score int, scoretext *tl.Text) *Block {
	return &Block{
		r:         tl.NewRectangle(x, y, 1, 1, color),
		g:         g,
		w:         w,
		h:         h,
		score:     score,
		scoretext: scoretext,
	}
}

func (b *Block) Size() (int, int)     { return b.r.Size() }
func (b *Block) Position() (int, int) { return b.r.Position() }

func (b *Block) Draw(s *tl.Screen) {
	if l, ok := b.g.Level().(*tl.BaseLevel); ok {
		// Set the level offset so the player is always in the
		// center of the screen. This simulates moving the camera.
		sw, sh := s.Size()
		x, y := b.r.Position()
		l.SetOffset(sw/2-x, sh/2-y)
	}
	b.r.Draw(s)
}

func (b *Block) Tick(ev tl.Event) {
	// Enable arrow key movement
	if ev.Type == tl.EventKey {
		b.px, b.py = b.r.Position()
		switch ev.Key {
		case tl.KeyArrowRight:
			b.r.SetPosition(b.px+1, b.py)
			break
		case tl.KeyArrowLeft:
			b.r.SetPosition(b.px-1, b.py)
			break
		case tl.KeyArrowUp:
			b.r.SetPosition(b.px, b.py-1)
			break
		case tl.KeyArrowDown:
			b.r.SetPosition(b.px, b.py+1)
			break
		}
	}
}

func (b *Block) Collide(c tl.Physical) {
	if r, ok := c.(*tl.Rectangle); ok {
		if r.Color() == tl.ColorWhite {
			// Collision with walls
			b.r.SetPosition(b.px, b.py)
		} else if r.Color() == tl.ColorBlue {
			// Collision with end - new level!
			b.w += 1
			b.h += 1
			b.score += 1
			buildLevel(b.g, b.w, b.h, b.score)
		}
	}
}

func buildLevel(g *tl.Game, w, h, score int) {
	maze := generateMaze(w, h)
	l := tl.NewBaseLevel(tl.Cell{})
	g.SetLevel(l)
	g.Log("Building level with width %d and height %d", w, h)
	scoretext := tl.NewText(0, 1, "Levels explored: "+strconv.Itoa(score),
		tl.ColorBlue, tl.ColorBlack)
	g.AddEntity(tl.NewText(0, 0, "Pyramid!", tl.ColorBlue, tl.ColorBlack))
	g.AddEntity(scoretext)
	for i, row := range maze {
		for j, path := range row {
			if path == '*' {
				l.AddEntity(tl.NewRectangle(i, j, 1, 1, tl.ColorWhite))
			} else if path == 'S' {
				l.AddEntity(NewBlock(i, j, tl.ColorRed, g, w, h, score, scoretext))
			} else if path == 'L' {
				l.AddEntity(tl.NewRectangle(i, j, 1, 1, tl.ColorBlue))
			}
		}
	}
}

func main() {
	g := tl.NewGame()
	buildLevel(g, 6, 2, 0)
	g.SetDebugOn(true)
	g.Start()
}
