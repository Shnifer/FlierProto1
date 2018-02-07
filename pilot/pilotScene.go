package main

import (
	MNT "github.com/Shnifer/flierproto1/mnt"
	"github.com/Shnifer/flierproto1/v2"
	"github.com/veandco/go-sdl2/sdl"
	"log"
	"math"
)

type PilotScene struct {
	*Scene
	Ship          *ShipGameObject
	gravityCalc3D bool
	showgizmos bool
}

func NewPilotScene(r *sdl.Renderer, ch *controlHandler) *PilotScene {
	return &PilotScene{
		Scene:         NewScene(r, ch),
		gravityCalc3D: DEFVAL.GravityCalc3D,
		showgizmos: true,
	}
}

func (PilotScene *PilotScene) Init() {
	BackGround := newStaticImage("background.jpg")
	PilotScene.AddObject(SceneObject(BackGround))

	Particles := newParticleSystem(DEFVAL.MainEngineMaxParticles)
	PilotScene.AddObject(SceneObject(Particles))

	//DATA INIT
	for _, starData := range MNT.GalaxyData {
		StarGO := &StarGameObject{Star: starData}
		PilotScene.AddObject(SceneObject(StarGO))
	}
	log.Println("Stars on scene", len(MNT.GalaxyData))

	Ship := newShip(Particles)
	PilotScene.Ship = Ship
	PilotScene.AddObject(SceneObject(Ship))

	startLoc := PilotScene.GetObjByID(DEFVAL.StartLocationName)
	if startLoc != nil {
		pos, _ := startLoc.(HugeMass).GetGravState()
		Ship.pos = pos.Add(DEFVAL.StartLocationOffset)
	} else {
		Ship.pos = DEFVAL.StartLocationOffset
	}

	FrontCabin := newStaticImage("cabinBorder.png")
	PilotScene.AddObject(SceneObject(FrontCabin))

	PilotScene.Scene.Init()
}

//Возвращает силу тяжести, точнее ускорение для заданной массы и заданного пробного положения
func GravityForce(attractor HugeMass, body V2.V2, Calc3D bool) V2.V2 {
	pos, mass := attractor.GetGravState()
	ort := V2.Sub(pos, body).Normed()
	dist2 := V2.Sub(pos, body).LenSqr()
	distFull2 := dist2 + DEFVAL.GravityDepthSqr
	Amp := DEFVAL.GravityConst * mass / distFull2
	if Calc3D {
		Amp = Amp * float32(math.Sqrt(float64(dist2/distFull2)))
	}
	force := ort.Mul(Amp)
	return force
}

func (ps *PilotScene) Update(dt float32) {
	//ФИЗИКА
	s := ps.Scene
	for _, obj := range s.Objects {
		attractor, ok := obj.(HugeMass)
		if !ok {
			continue
		}
		force := GravityForce(attractor, ps.Ship.pos, ps.gravityCalc3D)
		ps.Ship.ApplyForce(force)
	}

	//ВНЕШНИЕ ПРЯМЫЕ ВОЗДЕЙСТВИЯ НИ КИНЕМАТИКУ КОРАБЛЯ
	if ps.ControlHandler.GetKey(sdl.SCANCODE_SPACE) {
		ps.Ship.speed = V2.V2{}
		ps.Ship.anglespeed = 0
		ps.CameraAngle = 0
	}

	if ps.ControlHandler.GetKey(sdl.SCANCODE_KP_ENTER) {
		ps.Ship.speed = V2.V2{}
		ps.Ship.anglespeed = 0
		ps.CameraAngle = 0
		startLoc := ps.GetObjByID(DEFVAL.StartLocationName)
		if startLoc != nil {
			pos, _ := startLoc.(HugeMass).GetGravState()
			ps.Ship.pos = pos.Add(DEFVAL.StartLocationOffset)
		} else {
			ps.Ship.pos = DEFVAL.StartLocationOffset
		}
	}

	if ps.ControlHandler.GetKey(sdl.SCANCODE_1) {
		ps.Ship.showFixed = true
	}
	if ps.ControlHandler.GetKey(sdl.SCANCODE_2) {
		ps.Ship.showFixed = false
	}
	if ps.ControlHandler.GetKey(sdl.SCANCODE_3) {
		log.Println("disable 3D gravity")
		ps.gravityCalc3D = false
	}
	if ps.ControlHandler.GetKey(sdl.SCANCODE_4) {
		log.Println("enable 3D gravity")
		ps.gravityCalc3D = true
	}

	if ps.ControlHandler.GetKey(sdl.SCANCODE_5) {
		log.Println("disable 3D gravity")
		ps.showgizmos= false
	}
	if ps.ControlHandler.GetKey(sdl.SCANCODE_6) {
		log.Println("enable 3D gravity")
		ps.showgizmos= true
	}

	//АПДЕЙТ СЦЕНЫ
	s.Update(dt)

	//Сдвинули камеру
	ps.CameraCenter = ps.Ship.pos

	if ps.ControlHandler.GetKey(sdl.SCANCODE_Q) {
		ps.CameraAngle += 180 * dt
	}
	if ps.ControlHandler.GetKey(sdl.SCANCODE_E) {
		ps.CameraAngle -= 180 * dt
	}
}

func (ps PilotScene) Draw() {

	s := ps.Scene
	s.Draw()

	GizmoGravityForceK := DEFVAL.GizmoGravityForceK
	//Отрисовка "Гизмосов" гравитации
	if ps.showgizmos && (DEFVAL.ShowGizmoGravityRound || DEFVAL.ShowGizmoGravityForce) {
		sumForce := V2.V2{}

		for _, obj := range s.Objects {
			attractor, ok := obj.(HugeMass)
			if !ok {
				continue
			}

			pos, mass := attractor.GetGravState()

			if DEFVAL.ShowGizmoGravityForce {
				// Гизмос Наш вектор
				force := GravityForce(attractor, ps.Ship.pos, ps.gravityCalc3D).Mul(GizmoGravityForceK)

				s.R.SetDrawColor(0, 0, 255, 255)
				s.R.DrawLine(winW/2, winH/2, winW/2+int32(force.Rotate(s.CameraAngle).X), winH/2-int32(force.Rotate(s.CameraAngle).Y))
				sumForce = sumForce.Add(force)
			}

			if DEFVAL.ShowGizmoGravityRound {
				dotsInCirle := DEFVAL.GizmoGravityRoundDotsInCirle
				var GizmoGravLevels = DEFVAL.GizmoGravityRoundLevels
				//Гизмос вокруг планеты

				levelsCount := len(GizmoGravLevels)

				points := make([]sdl.Point, dotsInCirle+1)

				for level := 0; level < levelsCount; level++ {

					//GizmoLevel - сила(ускорение)
					//GizmoLevel = GravityConst*mass/RadSqr
					GravRadSqr := DEFVAL.GravityConst * mass / GizmoGravLevels[level]
					GravRad := float32(math.Sqrt(float64(GravRadSqr)))
					rect := newF32Sqr(pos, GravRad)
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
		}

		if DEFVAL.ShowGizmoGravityForce {
			//Гизмос наш суммарный вектор
			s.R.SetDrawColor(0, 255, 0, 255)
			s.R.DrawLine(winW/2, winH/2, winW/2+int32(sumForce.Rotate(s.CameraAngle).X), winH/2-int32(sumForce.Rotate(s.CameraAngle).Y))
		}
	}
}
