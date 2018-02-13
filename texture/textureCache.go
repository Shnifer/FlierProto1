package texture

import (
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
	"log"
	"math"
	"sync"
)

type fontType struct {
	name string
	size int
}

type TexCache struct {
	mu           sync.Mutex
	r            *sdl.Renderer
	textures     map[string]*sdl.Texture
	fonts        map[fontType]*ttf.Font
	texfilepath  string
	fontfilepath string
}

//Создаётся из вне после инициализации рендерера
func NewTexCache(r *sdl.Renderer, texfilepath, fontfilepath string) TexCache {

	return TexCache{
		r:            r,
		textures:     make(map[string]*sdl.Texture),
		fonts:        make(map[fontType]*ttf.Font),
		texfilepath:  texfilepath,
		fontfilepath: fontfilepath,
	}
}

var Cache TexCache

//Загружает текстуру из файла в хранилище, если её там ещё нет
//TODO: Асинхронная загрузка из файла в пиксели и передача в главный тред на компоновку в текстуру
func (tc *TexCache) PreloadTextureNoSync(name string) {
	if _, ok := tc.textures[name]; ok {
		//уже есть с таким именем
		return
	}
	pixels, w, h, err := loadFileToPixels(tc.texfilepath + name)
	if err != nil {
		log.Panicln(err)
	}

	tex, err := PixelsToTexture(tc.r, pixels, w, h)
	if err != nil {
		log.Panicln("can't load tex:", err)
	}
	tc.textures[name] = tex
}

func (tc *TexCache) PreloadTexture(name string) {
	tc.mu.Lock()
	tc.PreloadTextureNoSync(name)
	tc.mu.Unlock()
}

func (tc *TexCache) GetTexture(name string) *sdl.Texture {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tex := tc.textures[name]
	if tex != nil {
		return tex
	}

	tc.PreloadTextureNoSync(name)
	return tc.textures[name]
}

func (tc *TexCache) GetFont(name string, size int) *ttf.Font {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	ft := fontType{name: name, size: size}
	font := tc.fonts[ft]
	if font != nil {
		return font
	}

	font, err := ttf.OpenFont(tc.fontfilepath+name, size)
	if err != nil {
		log.Panicln(err)
	}
	tc.fonts[ft] = font
	return font
}

//НЕ ЗАБУДЬ УДАЛИТЬ ПРЕДЫДУЩУЮ В НАЧАЛЕ
func CreateTextTex(r *sdl.Renderer, text string, font *ttf.Font, color sdl.Color) (T *sdl.Texture,
	w, h int32) {
	surf, err := font.RenderUTF8Blended(text, color)
	if err != nil {
		log.Panicln(err)
	}
	defer surf.Free()
	w = surf.W
	h = surf.H

	tex, err := r.CreateTextureFromSurface(surf)

	if err != nil {
		log.Panicln(err)
	}
	return tex, w, h
}

func CreateFilledCirle(r *sdl.Renderer, rad int32, color sdl.Color) *sdl.Texture {
	pixels := make([]byte, (2*rad+1)*(2*rad+1)*4)
	pt := (2*rad + 1) * 4
	r2 := rad * rad
	for y := -rad; y <= +rad; y++ {
		for x := -rad; x <= +rad; x++ {
			if x*x+y*y <= r2 {
				ind := (y+rad)*pt + (x+rad)*4
				pixels[ind+0] = color.R
				pixels[ind+1] = color.G
				pixels[ind+2] = color.B
				pixels[ind+3] = color.A
			}
		}
	}
	tex, err := PixelsToTexture(r, pixels, int(2*rad+1), int(2*rad+1))
	if err != nil {
		log.Panicln(err)
	}

	return tex
}

func CreateFilledPie(r *sdl.Renderer, rad, inrad, start, end int32, color sdl.Color) *sdl.Texture {
	if inrad > rad {
		inrad = rad
	}
	start = angClamp(start)
	end = angClamp(end)
	pixels := make([]byte, (2*rad+1)*(2*rad+1)*4)
	pt := (2*rad + 1) * 4
	r2 := rad * rad
	inr2 := inrad * inrad

	//Завернуть циклы в горутину
	for y := -rad; y <= +rad; y++ {
		for x := -rad; x <= +rad; x++ {
			d2 := x*x + y*y
			deg := getDeg(x, y)
			if d2 <= r2 && d2 >= inr2 {
				if end >= start {
					if deg < start || deg > end {
						continue
					}
				} else {
					if deg > start && deg < end {
						continue
					}
				}
				ind := (y+rad)*pt + (x+rad)*4
				pixels[ind+0] = color.R
				pixels[ind+1] = color.G
				pixels[ind+2] = color.B
				pixels[ind+3] = color.A
			}
		}
	}

	tex, err := PixelsToTexture(r, pixels, int(2*rad+1), int(2*rad+1))
	if err != nil {
		log.Panicln(err)
	}

	return tex
}

func getDeg(x, y int32) (deg int32) {
	//1-использовать втроенную функцию, 2 - смотреть по табличке
	const Pi = math.Pi
	if x == 0 {
		deg = 90
		if y < 0 {
			deg *= -1
		}
	} else {
		t := float64(y) / float64(x)
		ang := math.Atan(t)
		deg = int32(ang / Pi * 180)
	}
	if x < 0 {
		deg += 180
	}
	if deg < 0 {
		deg += 360
	}
	return deg
}

//копия в main/cogitator
func angClamp(ang int32) int32 {
	switch {
	case ang < 0:
		{
			return angClamp(ang + 360)
		}
	case ang >= 360:
		{
			return angClamp(ang - 360)
		}
	default:
		return ang
	}
}
