package main

import (
	"github.com/veandco/go-sdl2/sdl"
	"github.com/Shnifer/flierproto1/texture"
)

type Anim struct {
	Active bool
	time   float32
	//время на показ кадра
	SlideTime float32
	Tex       *texture.AnimTex
}

func NewAnim(TexName string, num_x, num_y int, SlideTime float32, Active bool) *Anim {
	statT := texture.Cache.GetTexture(TexName)
	AniT := texture.NewAnimTex(statT, num_x, num_y)

	return &Anim{
		Tex:       AniT,
		Active:    Active,
		SlideTime: SlideTime,
	}
}

func (anim *Anim) GetTexAndRect() (*sdl.Texture, *sdl.Rect) {
	ind := int32(anim.time*1000/anim.SlideTime) % anim.Tex.TotalCount()
	return anim.Tex.GetTexAndRect(ind)
}

//TODO: Если выделять в пакет то функции запуска, остановки, тика на dt
