package main

import (
	MNT "github.com/Shnifer/flierproto1/mnt"
	"github.com/Shnifer/flierproto1/v2"
	"github.com/veandco/go-sdl2/sdl"
	"math"
	"log"
)

type PilotScene struct {
	*Scene
	Ship *ShipGameObject
}

func NewPilotScene(r *sdl.Renderer, ch *controlHandler) *PilotScene {
	return &PilotScene{
		Scene: NewScene(r, ch),
	}
}

func (PilotScene *PilotScene) Init() {
	BackGround := newStaticImage("background.jpg")
	PilotScene.AddObject(SceneObject(BackGround))

	Particles := newParticleSystem(1000)
	PilotScene.AddObject(SceneObject(Particles))

	//DATA INIT
	for _, starData := range MNT.GalaxyData {
		StarGO := &StarGameObject{Star: starData}
		PilotScene.AddObject(SceneObject(StarGO))
	}
	log.Println("Stars on scene",len(MNT.GalaxyData))

	Ship := newShip(Particles)
	PilotScene.Ship = Ship
	PilotScene.AddObject(SceneObject(Ship))

	FrontCabin := newStaticImage("cabinBorder.png")
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

	//Сдвинули камеру
	ps.CameraCenter = ps.Ship.pos

}

func (ps PilotScene) Draw() {

	s := ps.Scene
	s.Draw()

	//Отрисовка "Гизмосов" гравитации
	sumForce := V2.V2{}
	for _, obj := range s.Objects {
		attractor, ok := obj.(HugeMass)
		if !ok {
			continue
		}
		// Гизмос Наш вектор
		pos, mass := attractor.GetGravState()
		ort := V2.Sub(pos, ps.Ship.pos).Normed()
		dist2 := V2.Sub(pos, ps.Ship.pos).LenSqr() + DepthSqr
		Amp := GravityConst * mass / dist2
		force := ort.Mul(Amp)
		s.R.SetDrawColor(0, 0, 255, 255)
		s.R.DrawLine(winW/2, winH/2, winW/2+int32(force.X), winH/2+int32(force.Y))
		sumForce = sumForce.Add(force)


		const GizmoGravConst = GravityConst * 0.001
		const dotsInCirle = 64
		var GizmoGravLevels = [...]float32{0.3, 0.1, 0.05}
		//Гизмос вокруг планеты

		levelsCount := len(GizmoGravLevels)

		points := make([]sdl.Point, dotsInCirle+1)

		for level := 0; level < levelsCount; level++ {

			GravRadSqr := mass / GizmoGravLevels[level] * GizmoGravConst
			GravRad := float32(math.Sqrt(float64(GravRadSqr)))
			rect := f32Rect{pos.X - GravRad, pos.Y - GravRad, 2 * GravRad, 2 * GravRad}
			_, inCamera := ps.CameraTransformRect(rect)
			if !inCamera {
				continue
			}

			//n+1 чтобы замкнуть круг
			for a := 0; a <= dotsInCirle; a++ {
				dot := pos.AddMul(V2.InDir(float32(a*360/dotsInCirle)), GravRad)
				x, y := ps.CameraTransformV2(dot)
				points[a] = sdl.Point{x, y}
			}
			s.R.SetDrawColor(128, 128, 128, 128)
			s.R.DrawLines(points)
		}

	}

	//Гизмос наш суммарный вектор
	s.R.SetDrawColor(0, 255, 0, 255)
	s.R.DrawLine(winW/2, winH/2, winW/2+int32(sumForce.X), winH/2+int32(sumForce.Y))
}
