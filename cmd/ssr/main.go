package ssr

import (
	"fmt"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
	"time"
	"log"
	"path"
	"github.com/vit1251/go-fast-book-reader/internal"
)

type App struct {
	reactor     *internal.Reactor
	startTime   time.Time
	action      int
	running     bool
	surface     *sdl.Surface
	window      *sdl.Window
	format      uint32
	pixelFormat *sdl.PixelFormat
	fonts       []*Font
	wantRepaint bool
	pos         int
	speed       int
	reader      *internal.Reader
	pause       bool
}

type Font struct {
	name string
	size int
	font *ttf.Font
}

var (
	enableHighQuality bool   = true
	fontName          string = "SourceCodePro-Regular"
	fontSize          int    = 36
	readSpeed         int    = 240
	bookName          string = "example.txt"
)

func NewApplication() *App {

	log.Printf("Start reading speed %+v words per minute", readSpeed)

	app := &App{
		reactor:     internal.NewReactor(),
		running:     false,
		wantRepaint: true,
		speed:       readSpeed,
		startTime:   time.Now(),
		action:      0,
		pause:       true,
	}

	/* Monitoring events */
	app.reactor.CallAt(app.startTime, func() {
		app.monitorCounter()
	})

	/* Rendernig timer */
	var startShowDuration time.Duration = time.Duration(1) * time.Second
	app.reactor.CallLater(startShowDuration, func() {
		app.slide()
	})

	return app

}

func (self *App) slide() {
	var wordShowDuration time.Duration = time.Duration(60 * 1000 / readSpeed) * time.Millisecond
	log.Printf("Word %d duration on screen %+v", self.pos, wordShowDuration)
	//
	if self.pause == false {
		self.pos += 1
		self.action += 1
		self.wantRepaint = true
	}
	/* Register */
	self.reactor.CallLater(wordShowDuration, func() {
		self.slide()
	})
}

func (self *App) monitorCounter() {
	log.Printf("This is reactor CallAt every 1 sec. monitor")
	log.Printf("Action(s) %d per sec.", self.action)
	self.action = 0
	self.startTime = self.startTime.Add(time.Duration(1) * time.Second)
	log.Printf("Next monitor at %q", self.startTime)
	self.reactor.CallAt(self.startTime, func() {
		self.monitorCounter()
	})
}

func (self *Font) Close() error {
	self.font.Close()
	return nil
}

func (self *App) processKeyboardEvent(event sdl.KeyboardEvent) {

	log.Printf("key = %+v", event)

	if event.State == sdl.RELEASED {
	if (event.Keysym.Scancode == sdl.SCANCODE_ESCAPE) {
		self.running = false
	} else if (event.Keysym.Scancode == sdl.SCANCODE_SPACE) {
		log.Printf("Pause = %q", self.pause)
		self.pause = !self.pause
	} else if (event.Keysym.Scancode == sdl.SCANCODE_LEFT) {
		if self.pos > 0 {
			self.pos = self.pos - 1
			self.wantRepaint = true
		}
	} else if (event.Keysym.Scancode == sdl.SCANCODE_RIGHT) {
//		if self.pos ... {
			self.pos = self.pos + 1
			self.wantRepaint = true
//		}
	}
	}

}

func (self *App) ProcessEvent(event sdl.Event) {

	if keyboardEvent, ok := event.(*sdl.KeyboardEvent); ok {
		self.processKeyboardEvent(*keyboardEvent)
	} else if quitEvent, ok := event.(*sdl.QuitEvent); ok {
		log.Printf("quit = %+v", quitEvent)
		self.running = false
	}
}

func (self *App) processIteration() {

	/* Step 1. Process iteration */
	self.reactor.Process()

	/* Step 2. Wait next timer or event handling */
	var timeout int = 100
	
	if self.wantRepaint {
		timeout = 0
	}
	
	if event := sdl.WaitEventTimeout(timeout); event != nil {
		self.ProcessEvent(event)
	}

	if self.wantRepaint {
		self.Render()
		self.wantRepaint = false
	}

}

func (self *App) Run() {
	self.running = true
	for self.running {
		self.processIteration()
	}
}

func (self *App) GetFont(name string, size int) (*Font, error) {

	var exists bool = false
	var result *Font

	/* Step 1. Search */
	for _, font := range self.fonts {
		if font.name == name && font.size == size {
			result = font
			exists = true
		}
	}

	/* Step 2. Create */
	if !exists {
		filename := fmt.Sprintf("%s.ttf", name)
		absolute := path.Join(".", "fonts", filename) // TODO - get starting program directory instead ...
		font, err := ttf.OpenFont(absolute, size)
		if err != nil {
			return nil, err
		}
		/* Register font */
		newFont := Font{
			name: name,
			size: size,
			font: font,
		}
		self.fonts = append(self.fonts, &newFont)
		result = &newFont
	}

	return result, nil

}

func (self *App) drawText(font *Font, msg string) error {

	var draw_surface *sdl.Surface
	var err error

	color := sdl.Color{0, 128, 50, 0}
	if enableHighQuality {
		draw_surface, err = font.font.RenderUTF8Blended(msg, color)
	} else {
		draw_surface, err = font.font.RenderUTF8Solid(msg, color)
	}
	if err != nil {
		return err
	}
	defer draw_surface.Free()

	posX := (self.surface.W - draw_surface.W) / 2
	posY := (self.surface.H - draw_surface.H) / 2

	src := sdl.Rect{0, 0, self.surface.W, self.surface.H}
	dst := sdl.Rect{posX, posY, self.surface.W, self.surface.H}

	err = draw_surface.Blit(&src, self.surface, &dst)
	if err != nil {
		return err
	}

	//    texture, err3 := renderer.CreateTextureFromSurface(draw_surface)
	//    if err3 != nil {
	//        panic(err3)
	//    }
	//    renderer.Copy(texture, nil, nil)
	//    texture.Destroy()

	return nil
}

func (self *App) Init() error {

	var err error

	/* Step 0. Initialize reader */
	self.reader = internal.NewReader()
	err = self.reader.Read(bookName)
	if err != nil {
		return err
	}

	/* Step 1. Initialize SDL2 library */
	err = sdl.Init(sdl.INIT_EVERYTHING)
	if err != nil {
		return err
	}

	/* Step 2. Initialize TTF library */
	err = ttf.Init()
	if err != nil {
		return err
	}

	/* Step 3. Create application window */
	self.window, err = sdl.CreateWindow("FastBookReader v1.0.0", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, 800, 600, sdl.WINDOW_SHOWN)
	if err != nil {
		return err
	}

	/* Step 4. Create renderer */
	//	self.renderer, err2 := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)

	/* Step 5. Get main surface */
	self.surface, err = self.window.GetSurface()
	if err != nil {
		return err
	}

	/* Step 6. Save pixel information */
	self.format, err = self.window.GetPixelFormat()
	if err != nil {
		return err
	}
	self.pixelFormat, err = sdl.AllocFormat(uint(self.format))
	if err != nil {
		return err
	}

	/* Step 7. Preload fonts */
	font1, err := self.GetFont(fontName, fontSize)
	self.fonts = append(self.fonts, font1)

	return nil
}

func (self *App) Render() {

	log.Printf("Render")

	/* Step 1. Clear background */
	//var rect sdl.Rect
	bg := sdl.Color{}
	bg.R = 0
	bg.G = 0
	bg.B = 0
	bg.A = 0
	//	log.Printf("background: color = %+v", bg)
	color := sdl.MapRGBA(self.pixelFormat, bg.R, bg.G, bg.B, bg.A)
	//	log.Printf("background: int = 0x%X", color)
	self.surface.FillRect(nil /*&rect*/, color)

	/* Step 2. Draw text */
	msg := self.reader.Get(self.pos)
	log.Printf("Show %d is %s", self.pos, msg)
	self.drawText(self.fonts[0], msg)

	/* Step 3. Update screen buffer */
	self.window.UpdateSurface()

}

func (self *App) Quit() {
	/* Step 0. Release fonts */
	for _, font := range self.fonts {
		if font != nil {
			font.Close()
		}
	}
	/* Step 1. Release window pixel information */
	self.pixelFormat.Free()
	/* Step 2. Release window */
	self.window.Destroy()
	/* Step 3. Release TTF library */
	sdl.Quit()
	/* Step 4. Release SDL library */
	ttf.Quit()

}

func main() {

	/* Step 2. Application run */
	app := NewApplication()
	if err := app.Init(); err != nil {
		panic(err)
	}
	defer app.Quit()
	app.Run()

}
