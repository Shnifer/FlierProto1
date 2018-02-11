package scene

import (
	"github.com/veandco/go-sdl2/sdl"
	"github.com/Shnifer/flierproto1/texture"
)

type StaticImage struct {
	scene   *Scene
	texName string
	Tex     *sdl.Texture
	ZLayer  ZLayer
}

func NewStaticImage(texName string, ZLayer ZLayer) *StaticImage {
	return &StaticImage{texName: texName, ZLayer: ZLayer}
}

func (si *StaticImage) GetID() string {
	return ""
}

func (si *StaticImage) Init(scene *Scene) {
	si.scene = scene
	si.Tex = texture.Cache.GetTexture(si.texName)
}

func (si *StaticImage) Update(dt float32) {
	//nothing
}

func (si *StaticImage) Draw(r *sdl.Renderer) RenderReqList {
	//На весь экран
	//TODO: определить порядок ФОн - объекты - Интерфейс
	return RenderReqList{NewRenderReqSimple(si.Tex, nil, nil, si.ZLayer)}
}
