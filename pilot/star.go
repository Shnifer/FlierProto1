package main

import (
	"github.com/veandco/go-sdl2/sdl"
	V2 "github.com/Shnifer/flierproto1/v2"
	MNT "github.com/Shnifer/flierproto1/mnt"

	"log"
)

type StarGameObject struct{
	*MNT.Star
	scene *Scene
	tex *sdl.Texture
}

func (star *StarGameObject) GetID () string{
	return star.ID
}

func (star *StarGameObject) GetGravState() (pos V2.V2, Mass float32) {
	return star.Pos, star.Mass
}

func (s *StarGameObject) Update(dt float32) {
	if s.Parent=="" {
		//независимый объект
		s.Pos = s.Pos.Add(s.Dir.Mul(dt))
	} else {
		//спутник
		s.Angle += s.OrbSpeed*dt
		parentObj := s.scene.GetObjByID(s.Parent)
		if parentObj==nil {
			log.Panicln("Update of ",s.ID,"cant find the parent", s.Parent)
		}
		//TODO: полагаем что мы вращаемся ТОЛЬКО вокруг объекта с массой , а это HugeNass
		pp, _:=parentObj.(HugeMass).GetGravState()
		s.Pos = pp.AddMul(V2.InDir(s.Angle), s.OrbDist)
	}

}

func (s *StarGameObject) Draw (r *sdl.Renderer) {
	s.tex.SetColorMod(s.Color.R,s.Color.G,s.Color.B)
	halfsize:=s.ColRad
	rect:=f32Rect{s.Pos.X-halfsize, s.Pos.Y-halfsize, 2*halfsize, 2*halfsize}
	camRect, inCamera:=s.scene.CameraTransformRect(rect)
	//log.Println("draw star #",s.N,inCamera)
	if inCamera {
		r.Copy(s.tex, nil, camRect)
	}
}

func (star *StarGameObject) Init (scene *Scene) {
	star.scene = scene
	star.tex = TCache.GetTexture("planet.png")
}