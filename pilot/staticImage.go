package main

import "github.com/veandco/go-sdl2/sdl"

type staticImage struct {
	scene   *Scene
	texName string
	Tex     *sdl.Texture
	ZLayer  ZLayer
}

func newStaticImage(texName string, ZLayer ZLayer) *staticImage {
	return &staticImage{texName: texName, ZLayer: ZLayer}
}

func (si *staticImage) GetID() string {
	return ""
}

func (si *staticImage) Init(scene *Scene) {
	si.scene = scene
	si.Tex = TCache.GetTexture(si.texName)
}

func (si *staticImage) Update(dt float32) {
	//nothing
}

func (si *staticImage) Draw(r *sdl.Renderer) RenderReqList {
	//На весь экран
	//TODO: определить порядок ФОн - объекты - Интерфейс
	return RenderReqList{NewRenderReqSimple(si.Tex, nil, nil, si.ZLayer)}
}
