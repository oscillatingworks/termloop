package termloop

import "time"
import "github.com/nsf/termbox-go"

type Game struct {
	screen *Screen
	input  *Input
}

func NewGame() *Game {
	g := Game{screen: NewScreen(), input: NewInput()}
	return &g
}

func (g *Game) SetScreen(s *Screen) {
	g.screen = s
}

func (g *Game) CreateLevel(bg Attr) {
	g.screen.level = NewBaseLevel(Cell{Bg: bg})
}

func (g *Game) SetLevel(l Level) {
	g.screen.level = l
}

func (g *Game) AddEntity(d Drawable) {
	g.screen.entities = append(g.screen.entities, d)
}

func (g *Game) Start() {
	// Init Termbox
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()
	g.screen.Resize(termbox.Size())

	// Init input
	g.input.Start()
	defer g.input.Stop()
	clock := time.Now()

mainloop:
	for {
		select {
		case ev := <-g.input.eventQ:
			if ev.Key == g.input.endKey {
				break mainloop
			} else if EventType(ev.Type) == EventResize {
				g.screen.Resize(ev.Width, ev.Height)
			}
			g.screen.Tick(ConvertEvent(ev))
		default:
		}
		update := time.Now()
		g.screen.delta = update.Sub(clock).Seconds()
		clock = update
		g.screen.Draw()
	}
}
