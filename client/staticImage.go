package main

import "github.com/veandco/go-sdl2/sdl"

type staticImage struct {
	scene *Scene
	texName string
	Tex *sdl.Texture
}

func newStaticImage(texName string) *staticImage{
	return &staticImage{texName:texName}
}

func (si *staticImage) Init(scene *Scene) {
	si.scene = scene
	//TODO: огут быть разные статичные картинки, а не только фон
	si.Tex = TCache.GetTexture(si.texName)
}

func (si *staticImage) Update(dt float32) {
	//nothing
}

func (si *staticImage) Draw(r *sdl.Renderer) {
	//На весь экран
	//TODO: определить порядок ФОн - объекты - Интерфейс
	r.Copy(si.Tex,nil,nil )
}

