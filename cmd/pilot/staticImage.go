package main

import (
	"github.com/veandco/go-sdl2/sdl"
	"github.com/Shnifer/flierproto1/scene"
	"github.com/Shnifer/flierproto1/texture"
)

type staticImage struct {
	scene   *scene.Scene
	texName string
	Tex     *sdl.Texture
	ZLayer  scene.ZLayer
}

func newStaticImage(texName string, ZLayer scene.ZLayer) *staticImage {
	return &staticImage{texName: texName, ZLayer: ZLayer}
}

func (si *staticImage) GetID() string {
	return ""
}

func (si *staticImage) Init(scene *scene.Scene) {
	si.scene = scene
	si.Tex = texture.Cache.GetTexture(si.texName)
}

func (si *staticImage) Update(dt float32) {
	//nothing
}

func (si *staticImage) Draw(r *sdl.Renderer) scene.RenderReqList {
	//На весь экран
	//TODO: определить порядок ФОн - объекты - Интерфейс
	return scene.RenderReqList{scene.NewRenderReqSimple(si.Tex, nil, nil, si.ZLayer)}
}
