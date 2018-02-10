package main

import (
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
	"log"
	"sync"
)

const TexturePath = "textures/"

type fontType struct {
	name string
	size int
}

type TexCache struct {
	mu       sync.Mutex
	r        *sdl.Renderer
	textures map[string]*sdl.Texture
	fonts    map[fontType]*ttf.Font
}

func newTexCache(r *sdl.Renderer) TexCache {
	return TexCache{
		r:        r,
		textures: make(map[string]*sdl.Texture),
		fonts:    make(map[fontType]*ttf.Font),
	}
}

var TCache TexCache

//Загружает текстуру из файла в хранилище, если её там ещё нет
//TODO: Асинхронная загрузка из файла в пиксели и передача в главный тред на компоновку в текстуру
func (tc *TexCache) preloadTextureNoSync(name string) {
	if _, ok := tc.textures[name]; ok {
		//уже есть с таким именем
		return
	}
	pixels, w, h, err := loadFileToPixels(ResourcePath + TexturePath + name)
	if err != nil {
		log.Panicln(err)
	}

	tex, err := pixelsToTexture(tc.r, pixels, w, h)
	if err != nil {
		log.Panicln("can't load tex:", err)
	}
	tc.textures[name] = tex
}

func (tc *TexCache) PreloadTexture(name string) {
	tc.mu.Lock()
	tc.preloadTextureNoSync(name)
	tc.mu.Unlock()
}

func (tc *TexCache) GetTexture(name string) *sdl.Texture {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tex := tc.textures[name]
	if tex != nil {
		return tex
	}

	tc.preloadTextureNoSync(name)
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

	font, err := ttf.OpenFont(ResourcePath+name, size)
	if err != nil {
		log.Panicln(err)
	}
	tc.fonts[ft] = font
	return font
}

func (tc *TexCache) CreateTextTex(r *sdl.Renderer, text string, font *ttf.Font, color sdl.Color) (T *sdl.Texture,
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
