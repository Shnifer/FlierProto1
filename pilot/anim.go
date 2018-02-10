package main

import (
	"github.com/veandco/go-sdl2/sdl"
)

type Anim struct {
	Active bool
	time   float32
	//время на показ кадра
	SlideTime float32
	Tex       *animTexture
}

func NewAnim(TexName string, num_x, num_y int, SlideTime float32, Active bool) *Anim {
	statT := TCache.GetTexture(TexName)
	AniT := newAnimTexture(statT, num_x, num_y)

	return &Anim{
		Tex:       AniT,
		Active:    Active,
		SlideTime: SlideTime,
	}
}

func (anim *Anim) GetTexAndRect() (*sdl.Texture, *sdl.Rect) {
	ind := int32(anim.time*1000/anim.SlideTime) % anim.Tex.totalcount
	return anim.Tex.tex, anim.Tex.getRect(ind)
}

//TODO: Если выделять в пакет то функции запуска, остановки, тика на dt
