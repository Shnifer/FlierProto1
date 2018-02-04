package main

import (
	"github.com/Shnifer/FlierProto1/V2"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/Shnifer/FlierProto1/MNT"
)

type PilotScene struct {
	Scene
	Ship *ShipGameObject
}

func NewPilotScene(r *sdl.Renderer, ch *controlHandler) *PilotScene{
	return &PilotScene{
		Scene: *NewScene(r, ch),
		}
}

func (PilotScene *PilotScene) Init() {
	BackGround:=newStaticImage("background.jpg")
	PilotScene.AddObject(SceneObject(BackGround))

	Particles:=newParticleSystem(1000)
	PilotScene.AddObject(SceneObject(Particles))

	//DATA INIT
	for _,starData:=range MNT.GalaxyData{
		StarGO:= &StarGameObject{Star:starData}
		PilotScene.AddObject(SceneObject(StarGO))
	}

	Ship:=newShip(Particles)
	PilotScene.Ship = Ship
	PilotScene.AddObject(SceneObject(Ship))

	FrontCabin:=newStaticImage("cabinBorder.png")
	PilotScene.AddObject(SceneObject(FrontCabin))


	PilotScene.Scene.Init()
}

func (ps *PilotScene) Update(dt float32) {
	//ФИЗИКА
	s := ps.Scene
	for _, obj := range s.Objects {
		attractor, ok := obj.(HugeMass)
		if !ok {
			continue
		}
		pos, mass := attractor.GetGravState()
		ort := V2.Sub(pos, ps.Ship.pos).Normed()
		dist2 := V2.Sub(pos, ps.Ship.pos).LenSqr() + DepthSqr
		Amp := GravityConst * mass / dist2
		force := ort.Mul(Amp)
		ps.Ship.ApplyForce(force)
	}

	s.Update(dt)
}

func(ps PilotScene) Draw() {

	s:=ps.Scene
	s.Draw()

	//Отрисовка "Гизмосов" гравитации
	sumForce := V2.V2{}
	for _, obj := range s.Objects {
		attractor, ok := obj.(HugeMass)
		if !ok {
			continue
		}

		pos, mass := attractor.GetGravState()
		ort := V2.Sub(pos, ps.Ship.pos).Normed()
		dist2 := V2.Sub(pos, ps.Ship.pos).LenSqr() + DepthSqr
		Amp := GravityConst * mass / dist2
		force := ort.Mul(Amp)
		s.R.SetDrawColor(0, 0, 255, 255)
		s.R.DrawLine(winW/2, winH/2, winW/2+int32(force.X), winH/2+int32(force.Y))
		sumForce = sumForce.Add(force)
	}
	s.R.SetDrawColor(0, 255, 0, 255)
	s.R.DrawLine(winW/2, winH/2, winW/2+int32(sumForce.X), winH/2+int32(sumForce.Y))
}
