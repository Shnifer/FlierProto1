package main

import (
	"github.com/veandco/go-sdl2/sdl"
	V2 "github.com/Shnifer/flierproto1/v2"
	MNT "github.com/Shnifer/flierproto1/mnt"

)

type StarGameObject struct{
	*MNT.Star
	scene *Scene
	tex *sdl.Texture
	color sdl.Color
	N int
}

func (star *StarGameObject) GetGravState() (pos V2.V2, Mass float32) {
	return star.Pos, star.Mass
}

func (s *StarGameObject) Update(dt float32) {
	s.Pos = s.Pos.Add(s.Dir.Mul(dt))
}

func (s *StarGameObject) Draw (r *sdl.Renderer) {
	s.tex.SetColorMod(s.color.R,s.color.G,s.color.B)
	halfsize:=s.ColRad
	rect:=f32Rect{s.Pos.X-halfsize, s.Pos.Y-halfsize, 2*halfsize, 2*halfsize}
	camRect, inCamera:=s.scene.CameraTransformRect(rect)
	//log.Println("draw star #",s.N,inCamera)
	if inCamera {
		r.Copy(s.tex, nil, camRect)
	}
}

var starIDgen func() int
func init() {
	i:=0;
	starIDgen = func()int{
		i++
		return i
	}
}

func getid100() int{
	N:=starIDgen()
	return N/100
}

func (star *StarGameObject) Init (scene *Scene) {
	star.scene = scene
	star.N=starIDgen()
	N:=star.N/100
	star.color = sdl.Color{
		byte(N/3)*100,
		byte(N*2/3)*100,
		byte(N*5/3)*100,
		255,
		}
	star.tex = TCache.GetTexture("planet.png")
}