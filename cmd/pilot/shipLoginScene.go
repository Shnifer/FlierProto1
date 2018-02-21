package main

import (
	"github.com/Shnifer/flierproto1/scene"
	"github.com/Shnifer/flierproto1/texture"
	"github.com/veandco/go-sdl2/sdl"
)

type shipLoginScene struct{
	*scene.BScene
}

func (s *shipLoginScene) Init(){
	f:=texture.Cache.GetFont()
}

//Переводит виртуальные координаты "проценты экрана" в экранный прямоугольник
func VirtualRect(x, y, w, h int) sdl.Rect {
	return sdl.Rect{
		X: winW * int32(x) / 100,
		Y: winH * int32(y) / 100,
		W: winW * int32(w) / 100,
		H: winH * int32(h) / 100,
	}
}