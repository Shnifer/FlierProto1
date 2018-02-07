package main

import (
	MNT "github.com/Shnifer/flierproto1/mnt"
	V2 "github.com/Shnifer/flierproto1/v2"
	"github.com/veandco/go-sdl2/sdl"

	"log"
)

type StarGameObject struct {
	*MNT.Star
	scene *Scene
	tex   *sdl.Texture

	visZrot float32
}

func (star *StarGameObject) GetID() string {
	return star.ID
}

func (star *StarGameObject) GetGravState() (pos V2.V2, Mass float32) {
	return star.Pos, star.Mass
}

func (s *StarGameObject) Update(dt float32) {
	s.visZrot += DEFVAL.StarRotationSpeed * dt

	if s.Parent == "" {
		//независимый объект
		s.Pos = s.Pos.Add(s.Dir.Mul(dt))
	} else {
		//спутник
		s.Angle += s.OrbSpeed * dt
		parentObj := s.scene.GetObjByID(s.Parent)
		if parentObj == nil {
			log.Panicln("Update of ", s.ID, "cant find the parent", s.Parent)
		}
		//TODO: полагаем что мы вращаемся ТОЛЬКО вокруг объекта с массой , а это HugeNass
		pp, _ := parentObj.(HugeMass).GetGravState()
		s.Pos = pp.AddMul(V2.InDir(s.Angle), s.OrbDist)
	}
}

func (s *StarGameObject) Draw(r *sdl.Renderer) {
	s.tex.SetColorMod(s.Color.R, s.Color.G, s.Color.B)
	halfsize := s.ColRad
	rect := newF32Sqr(s.Pos, halfsize)
	camRect, inCamera := s.scene.CameraTransformRect(rect)
	//log.Println("draw star #",s.N,inCamera)
	if inCamera {
		r.CopyEx(s.tex, nil, camRect, float64(s.visZrot), nil, sdl.FLIP_NONE)
	}
}

func (star *StarGameObject) Init(scene *Scene) {
	star.scene = scene
	star.tex = TCache.GetTexture(star.TexName)
}
