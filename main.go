package main

import (
	"log"
	"path"
	"fmt"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

type App struct {
	running		bool
	surface		*sdl.Surface
	window		*sdl.Window
	format		uint32
	pixelFormat	*sdl.PixelFormat
	font1		*Font
	font2		*Font
	wasRepaint	bool
	pos		int
	speed		int
	reader		*Reader
}

type Font struct {
	name		string
	size		int
	font		*ttf.Font
}

var (
    enableHighQuality bool = true
)

func NewApplication() App {

	var speed int = 240
	log.Printf("Start reading speed %+v words per minute", speed)

	return App{
		running: false,
		wasRepaint: true,
		speed: speed,
	}
}

func (self *Font) Close() error {
	self.font.Close()
	return nil
}

func (self *App) ProcessEvent(event sdl.Event) {
	switch event.(type) {
		case *sdl.QuitEvent:
			println("Quit")
			self.running = false
	}
}

func (self *App) Run() {
	/* Step 1. Calculate spped */
	var timeout int = (1000*60) / self.speed
	/* Step 2. Perform main execution */
	self.running = true
	for self.running {
		if event := sdl.WaitEventTimeout(timeout); event != nil {
			self.ProcessEvent(event)
		} else {
			self.wasRepaint = true
			self.Render()
//			fmt.Printf("Event...")
			self.pos += 1
		}
	}
}

func (self *App) GetFont(name string, size int) (*Font, error) {

	var exists bool = false
	var result *Font

	/* Step 1. Search */
	// TODO - implement cache search ...

	/* Step 2. Create */
	if !exists {
		filename := fmt.Sprintf("%s.ttf", name)
		absolute := path.Join(".", "fonts", filename) // TODO - get starting program directory instead ...
		font, err := ttf.OpenFont(absolute, size)
		if err != nil {
			return nil, err
		}
		// TODO - register in index ...
		result = &Font{
			name: name,
			size: size,
			font: font,
		}
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

	pos_x := (self.surface.W - draw_surface.W) / 2
	pos_y := (self.surface.H - draw_surface.H) / 2

	src := sdl.Rect{0,0,self.surface.W,self.surface.H}
	dst := sdl.Rect{pos_x,pos_y,self.surface.W,self.surface.H}

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

	var book_name string = "example.txt"

	/* Step 0. Initialize reader */
	self.reader = NewReader()
	err = self.reader.Read(book_name)
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
	self.font1, _ = self.GetFont("SourceCodePro-Regular", 36)
	//self.font2, _ = self.GetFont("SourceCodePro-Italic", 36)

	return nil
}

func (self *App) Render() {

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
	self.surface.FillRect(nil/*&rect*/, color)

	/* Step 2. Draw text */
	//msg := fmt.Sprintf("%d", self.pos)
	msg := self.reader.Get(self.pos)
	self.drawText(self.font1, msg)

	/* Step 3. Update screen buffer */
	self.window.UpdateSurface()
	self.wasRepaint = false

}

func (self *App) Quit() {
	/* Step 0. Release fonts */
	self.font1.Close()
	self.font2.Close()
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
