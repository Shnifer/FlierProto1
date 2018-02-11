package main

import (
	"github.com/Shnifer/flierproto1/control"
	MNT "github.com/Shnifer/flierproto1/mnt"
	"github.com/Shnifer/flierproto1/v2"
	"github.com/veandco/go-sdl2/sdl"
	"log"
	"math"
	"github.com/Shnifer/flierproto1/scene"
	"github.com/Shnifer/flierproto1/texture"
)

type HugeMass interface {
	GetGravState() (pos V2.V2, Mass float32)
}


type PilotScene struct {
	*scene.Scene
	Ship          *PlayerShipGameObject
	gravityCalc3D bool
	showgizmos    bool
}

func NewPilotScene(r *sdl.Renderer, ch *control.Handler) *PilotScene {
	return &PilotScene{
		Scene:         scene.NewScene(r, ch, winW,winH),
		gravityCalc3D: DEFVAL.GravityCalc3D,
		showgizmos:    true,
	}
}

func (PilotScene *PilotScene) Init() {
	BackGround := newStaticImage("background.jpg", scene.Z_STAT_BACKGROUND)
	FrontCabin := newStaticImage("cabinBorder.png", scene.Z_STAT_HUD)
	PilotScene.AddObject(BackGround)
	PilotScene.AddObject(FrontCabin)

	Particles := newParticleSystem(DEFVAL.MainEngineMaxParticles)
	PilotScene.AddObject(Particles)

	//DATA INIT
	for _, starData := range MNT.GalaxyData {
		StarGO := &StarGameObject{Star: starData}
		PilotScene.AddObject(StarGO)
	}
	log.Println("Stars on scene", len(MNT.GalaxyData))

	Ship := NewPlayerShip(Particles)
	PilotScene.Ship = Ship
	PilotScene.AddObject(Ship)

	startLoc := PilotScene.GetObjByID(DEFVAL.StartLocationName)
	if startLoc != nil {
		pos, _ := startLoc.(HugeMass).GetGravState()
		Ship.pos = pos.Add(DEFVAL.StartLocationOffset)
	} else {
		Ship.pos = DEFVAL.StartLocationOffset
	}

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
		ps.Ship.angleSpeed = 0
		ps.CameraAngle = 0
	}

	if ps.ControlHandler.GetKey(sdl.SCANCODE_KP_ENTER) {
		ps.Ship.speed = V2.V2{}
		ps.Ship.angleSpeed = 0
		ps.CameraAngle = 0
		startLoc := ps.GetObjByID(DEFVAL.StartLocationName)
		if startLoc != nil {
			pos, _ := startLoc.(HugeMass).GetGravState()
			ps.Ship.pos = pos.Add(DEFVAL.StartLocationOffset)
		} else {
			ps.Ship.pos = DEFVAL.StartLocationOffset
		}
	}

	if ps.ControlHandler.WasKey(sdl.SCANCODE_1) {
		ps.Ship.showFixed = !ps.Ship.showFixed
		log.Printf("Fixed ship size mode = %v\n", ps.Ship.showFixed)
	}

	if ps.ControlHandler.WasKey(sdl.SCANCODE_2) {
		ps.gravityCalc3D = !ps.gravityCalc3D
		log.Printf("3D gravity mode = %v\n", ps.gravityCalc3D)
	}

	if ps.ControlHandler.WasKey(sdl.SCANCODE_3) {
		ps.showgizmos = !ps.showgizmos
		log.Printf("Show Gizmos = %v", ps.showgizmos)
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
					rect := scene.NewF32Sqr(pos, GravRad)
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
			//Гизмос суммарный вектор гравитации
			s.R.SetDrawColor(0, 200, 0, 255)
			s.R.DrawLine(winW/2, winH/2, winW/2+int32(sumForce.Rotate(s.CameraAngle).X), winH/2-int32(sumForce.Rotate(s.CameraAngle).Y))

			//Вектор тяги
			thrustForce := V2.InDir(ps.Ship.angle).Mul(ps.Ship.mainThrust * ps.Ship.maxThrustForce)
			s.R.SetDrawColor(200, 0, 0, 255)
			s.R.DrawLine(winW/2, winH/2, winW/2+int32(thrustForce.Rotate(s.CameraAngle).X), winH/2-int32(thrustForce.Rotate(s.CameraAngle).Y))
			sumForce = sumForce.Add(thrustForce)

			//тяга + гравитация
			s.R.SetDrawColor(200, 200, 200, 255)
			s.R.DrawLine(winW/2, winH/2, winW/2+int32(sumForce.Rotate(s.CameraAngle).X), winH/2-int32(sumForce.Rotate(s.CameraAngle).Y))
		}
	}

	f := texture.Cache.GetFont("interdim.ttf", 20)
	t, w, h := texture.Cache.CreateTextTex(s.R, "PILOT scene", f, sdl.Color{200, 200, 200, 255})
	rect := &sdl.Rect{100, 100, w, h}
	s.R.Copy(t, nil, rect)
}
