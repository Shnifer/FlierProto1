package main

import (
	"github.com/veandco/go-sdl2/sdl"
	"github.com/Shnifer/FlierProto1/V2"
	"github.com/Shnifer/FlierProto1/MNT"
)

type StarGameObject struct{
	*MNT.Star
	scene *Scene
	tex *sdl.Texture
}

func (star *StarGameObject) GetGravState() (pos V2.V2, Mass float32) {
	return star.Pos, star.Mass
}

func (s *StarGameObject) Update(dt float32) {
	s.Pos = s.Pos.Add(s.Dir.Mul(dt))
}

func (s *StarGameObject) Draw (r *sdl.Renderer) {
	halfsize:=s.ColRad
	rect:=f32Rect{s.Pos.X-halfsize, s.Pos.Y-halfsize, 2*halfsize, 2*halfsize}
	camRect, inCamera:=s.scene.CameraTransformRect(rect)
	if inCamera {
		r.Copy(s.tex, nil, camRect)
	}
}

func (star *StarGameObject) Init (scene *Scene) {
	star.scene = scene
	star.tex = TCache.GetTexture("planet.png")
}