package main

import (
	"github.com/Shnifer/flierproto1/control"
	MNT "github.com/Shnifer/flierproto1/mnt"
	"github.com/Shnifer/flierproto1/scene"
	"github.com/Shnifer/flierproto1/texture"
	"github.com/Shnifer/flierproto1/v2"
	"github.com/veandco/go-sdl2/sdl"
	"log"
	"math"
	"strings"
)

type HugeMass interface {
	GetGravState() (pos V2.V2, Mass float32)
}

type PilotScene struct {
	*scene.BScene
	Ship          *PlayerShipGameObject
	gravityCalc3D bool
	showgizmos    bool

	//корабль ниже центра, процент от WinH
	shipBack       int32
	camFollowAngle bool

	fpsUI *scene.TextUI
}

func NewPilotScene(r *sdl.Renderer, ch *control.Handler) *PilotScene {
	return &PilotScene{
		BScene:         scene.NewScene(r, ch, winW, winH),
		gravityCalc3D:  DEFVAL.GravityCalc3D,
		shipBack:       DEFVAL.ShipShowBotOffset,
		camFollowAngle: true,
		showgizmos:     true,
	}
}

func (PilotScene *PilotScene) Init() {
	BackGround := scene.NewStaticImage("background.jpg", scene.Z_STAT_BACKGROUND)
	FrontCabin := scene.NewStaticImage("cabinBorder.png", scene.Z_STAT_HUD)
	PilotScene.AddObject(BackGround)
	PilotScene.AddObject(FrontCabin)

	Particles := newParticleSystem(DEFVAL.MainEngineMaxParticles)
	PilotScene.AddObject(Particles)

	//TODO: УБРАТЬ РУЧНУЮ НЕБУЛУ
	var nebula1Points []*StarGameObject
	//DATA INIT
	for _, starData := range MNT.GalaxyData {
		StarGO := &StarGameObject{Star: starData, startAngle: starData.Angle}
		if strings.HasPrefix(starData.ID, "asteroid") {
			nebula1Points = append(nebula1Points, StarGO)
		}
		PilotScene.AddObject(StarGO)
	}
	log.Println("Stars on scene", len(MNT.GalaxyData))

	Nebula1 := NewNebula("nebula1", nebula1Points, 150)
	PilotScene.AddObject(Nebula1)

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

	f := texture.Cache.GetFont("interdim.ttf", 20)
	SceneCaption := scene.NewTextUI("PILOT scene", f, sdl.Color{200, 200, 200, 255}, scene.Z_STAT_HUD, scene.FROM_ANGLE)
	SceneCaption.X, SceneCaption.Y = 100, 100
	PilotScene.AddObject(SceneCaption)

	pf := texture.Cache.GetFont("phantom.ttf", 14)
	fpsUI := scene.NewTextUI("fps:", pf, sdl.Color{255, 0, 0, 255}, scene.Z_STAT_HUD, scene.FROM_ANGLE)
	fpsUI.X, fpsUI.Y = 10, 10

	PilotScene.AddObject(fpsUI)
	PilotScene.fpsUI = fpsUI

	for i := range PilotScene.Objects {
		PilotScene.Objects[i].Init(PilotScene)
	}
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
	scale:=ps.CameraScale()
	if ps.CH().GetKey(sdl.SCANCODE_KP_PLUS) {
		scale *= (1 + dt)
	}
	if ps.CH().GetKey(sdl.SCANCODE_KP_MINUS) {
		scale *= (1 - dt)
	}
	min := DEFVAL.CameraMaxScale
	if min == 0 {
		min = 100000
	} else {
		min = 1 / min
	}
	max := DEFVAL.CameraMinScale
	if max == 0 {
		max = 100000
	} else {
		max = 1 / max
	}
	Clamp(&scale, min, max)
	ps.SetCameraScale(scale)

	controlHandler:=ps.CH()

	//ФИЗИКА - ГРАВИТАЦИЯ
	for _, obj := range ps.Objects {
		attractor, ok := obj.(HugeMass)
		if !ok {
			continue
		}
		force := GravityForce(attractor, ps.Ship.pos, ps.gravityCalc3D)
		ps.Ship.ApplyForce(force)
	}
	//ФИЗИКА - ТРЕНИЕ
	const kTens = 0.2
	tensForce := ps.Ship.speed.Mul(-kTens)
	neb := ps.GetObjByID("nebula1").(*Nebula)
	if neb.isInside(ps.Ship.pos) {
		ps.Ship.ApplyForce(tensForce)
	}

	//ВНЕШНИЕ ПРЯМЫЕ ВОЗДЕЙСТВИЯ НИ КИНЕМАТИКУ КОРАБЛЯ
	if controlHandler.GetKey(sdl.SCANCODE_SPACE) {
		ps.Ship.speed = V2.V2{}
		ps.Ship.angleSpeed = 0
		ps.SetCameraAngle(0)
		ps.camFollowAngle = true
	}

	if controlHandler.GetKey(sdl.SCANCODE_KP_ENTER) {
		ps.Ship.speed = V2.V2{}
		ps.Ship.angleSpeed = 0
		ps.SetCameraAngle(0)
		ps.camFollowAngle = true
		startLoc := ps.GetObjByID(DEFVAL.StartLocationName)
		if startLoc != nil {
			pos, _ := startLoc.(HugeMass).GetGravState()
			ps.Ship.pos = pos.Add(DEFVAL.StartLocationOffset)
		} else {
			ps.Ship.pos = DEFVAL.StartLocationOffset
		}
	}

	if controlHandler.WasKey(sdl.SCANCODE_1) {
		ps.Ship.showFixed = !ps.Ship.showFixed
		log.Printf("Fixed ship size mode = %v\n", ps.Ship.showFixed)
	}

	if controlHandler.WasKey(sdl.SCANCODE_2) {
		ps.gravityCalc3D = !ps.gravityCalc3D
		log.Printf("3D gravity mode = %v\n", ps.gravityCalc3D)
	}

	if controlHandler.WasKey(sdl.SCANCODE_3) {
		ps.showgizmos = !ps.showgizmos
		log.Printf("Show Gizmos = %v", ps.showgizmos)
	}

	if controlHandler.WasKey(sdl.SCANCODE_4) {
		ps.GetObjByID("nebula1").(*Nebula).drawMode =
			(ps.GetObjByID("nebula1").(*Nebula).drawMode + 1) % 3
		log.Println("Nebula Mod Changed")
	}

	//АПДЕЙТ СЦЕНЫ
	ps.BScene.Update(dt)

	angle := ps.CameraAngle()
	if controlHandler.GetKey(sdl.SCANCODE_Q) {
		angle += 180 * dt
		ps.SetCameraAngle(angle)
		ps.camFollowAngle = false
	}
	if controlHandler.GetKey(sdl.SCANCODE_E) {
		angle -= 180 * dt
		ps.SetCameraAngle(angle)
		ps.camFollowAngle = false
	}

	//Сдвинули камеру
	if ps.camFollowAngle {
		ps.SetCameraAngle(-ps.Ship.angle)
	}

	scrOff := float32(winW * ps.shipBack / 100)
	offset := V2.InDir(-ps.CameraAngle()).Mul(scrOff / ps.CameraScale())
	ps.CameraCenter = ps.Ship.pos.Add(offset)
}

func (ps *PilotScene) Draw() {

	ps.BScene.Draw()

	GizmoGravityForceK := DEFVAL.GizmoGravityForceK
	//TODO: Вынести гизмосы в отдельный объект сцены
	//Отрисовка "Гизмосов" гравитации
	if ps.showgizmos && (DEFVAL.ShowGizmoGravityRound || DEFVAL.ShowGizmoGravityForce) {
		sumForce := V2.V2{}
		for _, obj := range ps.Objects {
			attractor, ok := obj.(HugeMass)
			if !ok {
				continue
			}

			pos, mass := attractor.GetGravState()

			if DEFVAL.ShowGizmoGravityForce {
				// Гизмос Наш вектор
				force := GravityForce(attractor, ps.Ship.pos, ps.gravityCalc3D).Mul(GizmoGravityForceK)
				forceCamRot:=force.Rotate(ps.CameraAngle())
				ps.R().SetDrawColor(0, 0, 255, 255)
				ps.R().DrawLine(winW/2, winH/2, winW/2+int32(forceCamRot.X), winH/2-int32(forceCamRot.Y))
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
					ps.R().SetDrawColor(128, 128, 128, 128)
					ps.R().DrawLines(points)
				}
			}
		}

		if DEFVAL.ShowGizmoGravityForce {
			//Гизмос суммарный вектор гравитации

			sumForceCamRot:=sumForce.Rotate(ps.CameraAngle())

			ps.R().SetDrawColor(0, 200, 0, 255)
			ps.R().DrawLine(winW/2, winH/2, winW/2+int32(sumForceCamRot.X), winH/2-int32(sumForceCamRot.Y))

			//Вектор тяги
			thrustForce := V2.InDir(ps.Ship.angle).Mul(ps.Ship.mainThrust * ps.Ship.maxThrustForce)
			thrustForceCamRot:=thrustForce.Rotate(ps.CameraAngle())
			ps.R().SetDrawColor(200, 0, 0, 255)
			ps.R().DrawLine(winW/2, winH/2, winW/2+int32(thrustForceCamRot.X), winH/2-int32(thrustForceCamRot.Y))

			//тяга + гравитация
			sumForce = sumForce.Add(thrustForce)
			sumForceCamRot = sumForce.Rotate(ps.CameraAngle())
			ps.R().SetDrawColor(200, 200, 200, 255)
			ps.R().DrawLine(winW/2, winH/2, winW/2+int32(sumForceCamRot.X), winH/2-int32(sumForceCamRot.Y))
		}
	}
}

func (ps *PilotScene) showFps(data string) {
	ps.fpsUI.ChangeText("fps: " + data)
}