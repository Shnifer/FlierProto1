package main

import (
	"github.com/veandco/go-sdl2/sdl"
	"log"
	"sync"
)

const TexturePath = "textures/"

type TexCache struct {
	mu       sync.Mutex
	r        *sdl.Renderer
	textures map[string]*sdl.Texture
}

func newTexCache(r *sdl.Renderer) TexCache {
	return TexCache{
		r:        r,
		textures: make(map[string]*sdl.Texture)}
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
