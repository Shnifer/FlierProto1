package main

import (
	"github.com/Shnifer/flierproto1/v2"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/Shnifer/flierproto1/scene"
)

type PlayerShipGameObject struct {
	ShipGameObject

	//максимальная сила движка
	maxThrustForce float32
	maxAngMomentum float32
	//мощность маршевого движка от -1 до 1
	mainThrust float32
	angThrust  float32
	//РАзгон и торможение самого движка:
	// максимальное изменение величины mainThrust angThrust  за единицу
	thrustAxel float32
	angAxel    float32

	*Cogitator
}

func NewPlayerShip(ps *ParticleSystem) *PlayerShipGameObject {
	ship := newShip(ps)
	res := &PlayerShipGameObject{
		ShipGameObject: *ship,
		maxThrustForce: DEFVAL.ShipMaxThrustForce,
		maxAngMomentum: DEFVAL.ShipMaxAngMomentum,
		thrustAxel:     DEFVAL.ShipThrustAxel,
		angAxel:        DEFVAL.ShipAngAxel,
		Cogitator:      &Cogitator{},
	}
	return res
}

func (ship *PlayerShipGameObject) GetID() string {
	return ""
}

func (ship *PlayerShipGameObject) Init(scene *scene.Scene) {
	ship.ShipGameObject.Init(scene)
}

//Изменяет mainThrust angThrust по расчётам Когитатора в соответствии с ограничениями движка
func (ship *PlayerShipGameObject) ApplyCO(co CogitatorOutput, dt float32) {
	dThurst := co.wantedMainThrust - ship.mainThrust
	Clamp(&dThurst, -ship.thrustAxel*dt, +ship.thrustAxel*dt)
	ship.mainThrust += dThurst

	dMomentum := co.wantedAngThrust - ship.angThrust
	Clamp(&dMomentum, -ship.angAxel*dt, +ship.angAxel*dt)
	ship.angThrust += dMomentum
	Clamp(&ship.mainThrust, 0, 1)
	Clamp(&ship.angThrust, -1, 1)
}

func (ship *PlayerShipGameObject) Update(dt float32) {
	ship.GetInputs(ship.scene.ControlHandler)
	ship.GetShipStates(ship)
	Output := ship.Cogitate(dt)
	ship.ApplyCO(Output, dt)

	//Расчитываем и добавляем силу движка к уже посчитаной гравитации
	ship.ApplyMoment(ship.angThrust * ship.maxAngMomentum)
	ThrustForce := V2.InDir(ship.angle).Mul(ship.mainThrust * ship.maxThrustForce)
	ship.ApplyForce(ThrustForce)

	ship.ShipGameObject.Update(dt)

	//Установка Генератора частиц, Переключение Анимации
	if ship.mainThrust > 0 {
		thrust := ship.mainThrust
		childPos := V2.V2{0, -1.4}.Mul(ship.colRad)
		ship.MainEngineProducer.pos = childPos.ApplyOnTransform(ship.pos, ship.angle)
		psspeed := 0.2 + thrust*3*ship.colRad
		ship.MainEngineProducer.speed = V2.InDir(180 + ship.angle).Mul(psspeed).Add(ship.speed)
		ship.MainEngineProducer.Intense = DEFVAL.MainEngineParticlesMaxIntense * thrust
		ship.MainEngineProducer.color = sdl.Color{byte(100 + 155*thrust), byte(70 * thrust), 0, 255}

		ship.ps.Produce(dt, ship.MainEngineProducer)

	}
	ship.anim.time += dt

}

func (ship PlayerShipGameObject) Draw(r *sdl.Renderer) scene.RenderReqList {
	res := ship.ShipGameObject.Draw(r)

	//Показ анимации пламени
	//Карабль что-то показывает, считаем что это основной спрайт, а значи он в камере
	//хакаем на проверку inCamera
	if len(res) > 0 {
		showFixedSized := ship.showFixed && (DEFVAL.ShipFixedSize != 0)
		flameTex, flameRect := ship.anim.GetTexAndRect()
		const mainFlameSize=0.75
		if ship.mainThrust>0 {
			var cameraRect *sdl.Rect
			childPos := V2.V2{0, -1.4}
			if showFixedSized {
				flameCentre := childPos.Mul(float32(DEFVAL.ShipFixedSize) / ship.scene.CameraScale).ApplyOnTransform(ship.pos, ship.angle)
				cameraRect, _ = ship.scene.CameraRectByCenterAndScreenSize(flameCentre, int32(float32(DEFVAL.ShipFixedSize)*mainFlameSize))
			} else {
				//физическая координата центра пламени
				flameCentre := childPos.Mul(ship.colRad).ApplyOnTransform(ship.pos, ship.angle)
				flamesize := float32(mainFlameSize * ship.colRad)
				dRect := scene.NewF32Sqr(flameCentre, flamesize)
				cameraRect, _ = ship.scene.CameraTransformRect(dRect)
			}
			req := scene.NewRenderReq(flameTex, flameRect, cameraRect, scene.Z_UNDER_OBJECT,
				-float64(ship.angle + ship.scene.CameraAngle), nil, sdl.FLIP_VERTICAL)
			res = append(res, req)
		}
		if ship.angThrust!=0 {
			var cameraRect *sdl.Rect
			var childPos V2.V2
			const sideFlameSize=0.45
			if ship.angThrust>0 {
				childPos = V2.V2{+0.55, -1.1}
			} else {
				childPos = V2.V2{-0.55, -1.1}
			}
			if showFixedSized {
				flameCentre := childPos.Mul(float32(DEFVAL.ShipFixedSize) / ship.scene.CameraScale).ApplyOnTransform(ship.pos, ship.angle)
				cameraRect, _ = ship.scene.CameraRectByCenterAndScreenSize(flameCentre, int32(float32(DEFVAL.ShipFixedSize)*sideFlameSize))
			} else {
				//физическая координата центра пламени
				flameCentre := childPos.Mul(ship.colRad).ApplyOnTransform(ship.pos, ship.angle)
				flamesize := float32(sideFlameSize * ship.colRad)
				dRect := scene.NewF32Sqr(flameCentre, flamesize)
				cameraRect, _ = ship.scene.CameraTransformRect(dRect)
			}
			req := scene.NewRenderReq(flameTex, flameRect, cameraRect, scene.Z_UNDER_OBJECT,
				-float64(ship.angle + ship.scene.CameraAngle), nil, sdl.FLIP_VERTICAL)
			res = append(res, req)
		}
	}

	HUD:=ship.Cogitator.Draw(r)
	res = append(res,HUD...)
	return res
}

func Clamp(v *float32, min, max float32) {
	if *v < min {
		*v = min
	}
	if *v > max {
		*v = max
	}
}
