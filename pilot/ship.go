package main

import (
	V2 "github.com/Shnifer/flierproto1/v2"
	"github.com/veandco/go-sdl2/sdl"
	"log"
)

type ShipGameObject struct {
	//Ship data включаем, пока нет других корабликов (значимых а не маркеров)

	//TODO: тут данные похожи на звезду. Выяснить что входит в абстрактый Transform/Rigidbody
	pos   V2.V2
	speed V2.V2

	angle      float32
	anglespeed float32

	//мощность маршевого движка от -1 до 1
	thrust float32
	//изменение величины аксель за единицу
	thrustAxel float32
	//максимальная сила движка
	maxThrustForce float32

	//изменение angAxel за 1 секунду
	angAxel float32

	//Сумма внешних сил за такт
	forceSum V2.V2

	//Радиус и коллизии и полупоперечник рисовки
	colRad float32

	scene *Scene
	tex   *sdl.Texture
	ps    *ParticleSystem

	MainEngineProducer *ProduceStats

	//TODO: абстрагировать в АНИМАЦИЮ
	animActive bool
	animTime   float32
	//время на показ кадра
	animSlideTime      float32
	MainEngineFlameTex *animTexture

	//TODO: абстрагировать в UI
	arrowTex *sdl.Texture

	showFixed bool
}

func newShip(ps *ParticleSystem) *ShipGameObject {

	res := ShipGameObject{
		colRad:         DEFVAL.ShipSize,
		thrustAxel:     DEFVAL.ShipThrustAxel,
		maxThrustForce: DEFVAL.ShipMaxThrustForce,
		angAxel:        DEFVAL.ShipAngAxel,
		showFixed:      DEFVAL.ShipShowFixed,
		angle:          0,
		ps:             ps,
		animSlideTime:  10}

	res.MainEngineProducer = &ProduceStats{
		lifeTime:  DEFVAL.MainEngineParticlesLifetime,
		randpos:   DEFVAL.MainEngineParticlesRandStartK * res.colRad,
		randspeed: DEFVAL.MainEngineParticlesRandSpeedK * res.colRad,
	}

	return &res
}

func (ship *ShipGameObject) GetID() string {
	return ""
}

func (ship *ShipGameObject) Init(scene *Scene) {
	ship.scene = scene
	ship.tex = TCache.GetTexture("ship.png")
	ship.arrowTex = TCache.GetTexture("arrow.png")
	flametex := TCache.GetTexture("flame_ani.png")
	flameani := newAnimTexture(flametex, 5, 4)
	ship.MainEngineFlameTex = flameani
}

func (ship *ShipGameObject) Update(dt float32) {
	CH := ship.scene.ControlHandler

	var cAngAxel float32

	//	if CH.Joystick == nil {
	//Поворачиваем на угол инрции

	if CH.GetKey(sdl.SCANCODE_A) {
		cAngAxel += ship.angAxel
	}
	if CH.GetKey(sdl.SCANCODE_D) {
		cAngAxel -= ship.angAxel
	}

	if CH.GetKey(sdl.SCANCODE_W) {
		ship.thrust += ship.thrustAxel * dt
	}
	if CH.GetKey(sdl.SCANCODE_S) {
		ship.thrust -= ship.thrustAxel * dt
	}
	//	}

	if CH.Joystick != nil {
		ship.thrust += ship.thrustAxel * CH.AxisY * dt
		cAngAxel -= ship.angAxel * CH.AxisX
	}

	if ship.thrust > 1 {
		ship.thrust = 1
	}
	if ship.thrust < (0) {
		ship.thrust = 0
	}

	ship.angle += ship.anglespeed * dt
	ship.anglespeed += cAngAxel * dt
	//Расчитываем и добавляем силу движка к уже посчитаной гравитации
	ThrustForce := V2.InDir(ship.angle).Mul(ship.thrust * ship.maxThrustForce * dt)
	ship.ApplyForce(ThrustForce)

	//Добавили сумму сил к скорости и обнулили сумматор
	ship.speed.DoAddMul(ship.forceSum, dt)
	ship.forceSum = V2.V2{}

	//Пренесли скорость в координату
	ship.pos.DoAddMul(ship.speed, dt)

	//Отдельный по сути блок расчёта анимации
	produceMainEng := (ship.thrust > 0)
	if produceMainEng {
		thrust := ship.thrust
		childPos := V2.V2{0, -1.4}.Mul(ship.colRad)
		ship.MainEngineProducer.pos = childPos.ApplyOnTransform(ship.pos, ship.angle)
		psspeed := 0.2 + thrust*3*ship.colRad
		ship.MainEngineProducer.speed = V2.InDir(180+ship.angle).Mul(psspeed).Add(ship.speed)
		log.Println(ship.angle, V2.InDir(180+ship.angle))
		ship.MainEngineProducer.Intense = DEFVAL.MainEngineParticlesMaxIntense * thrust
		ship.MainEngineProducer.color = sdl.Color{byte(100 + 155*thrust), byte(70 * thrust), 0, 255}

		ship.ps.Produce(dt, ship.MainEngineProducer)

		ship.animActive = true
		ship.animTime += dt
	} else {
		ship.animTime = 0
		ship.animActive = false
	}
}

func (ship ShipGameObject) Draw(r *sdl.Renderer) {
	//Показ Корабля
	var camRect *sdl.Rect
	var inCamera bool
	showFixedSized := ship.showFixed && (DEFVAL.ShipFixedSize != 0)
	if showFixedSized {
		camRect, inCamera = ship.scene.CameraRectByCenterAndScreenSize(ship.pos, DEFVAL.ShipFixedSize)
	} else {
		rect := newF32Sqr(ship.pos, ship.colRad)
		camRect, inCamera = ship.scene.CameraTransformRect(rect)
	}

	if inCamera {
		r.CopyEx(ship.tex, nil, camRect, -float64(ship.angle+ship.scene.CameraAngle), nil, sdl.FLIP_NONE)
	}

	//Показ анимации огня
	//TODO: ЧИТАЕМЫЕ преобразования координат вложенных объектов
	if inCamera && ship.animActive {
		ind := int32(ship.animTime*1000/ship.animSlideTime) % ship.MainEngineFlameTex.totalcount
		flameRect := ship.MainEngineFlameTex.getRect(ind)

		var cameraRect *sdl.Rect
		childPos := V2.V2{0, -1.4}
		if showFixedSized {
			flameCentre := childPos.Mul(float32(DEFVAL.ShipFixedSize)/ship.scene.CameraScale).ApplyOnTransform(ship.pos, ship.angle)
			cameraRect, _ = ship.scene.CameraRectByCenterAndScreenSize(flameCentre, int32(float32(DEFVAL.ShipFixedSize)*0.75))
		} else {
			//физическая координата центра пламени
			flameCentre := childPos.Mul(ship.colRad).ApplyOnTransform(ship.pos, ship.angle)
			flamesize := float32(0.75 * ship.colRad)
			dRect := newF32Sqr(flameCentre, flamesize)
			cameraRect, _ = ship.scene.CameraTransformRect(dRect)
		}
		r.CopyEx(ship.MainEngineFlameTex.tex, flameRect, cameraRect, -float64(ship.angle+ship.scene.CameraAngle), nil, sdl.FLIP_VERTICAL)
	}

	//Отдельный блок показа UI
	H := int32(ship.thrust * float32(winH) * 0.8)
	UIRect := &sdl.Rect{X: 60,
		Y: winH - 30 - H,
		W: 40,
		H: H}
	ship.arrowTex.SetColorMod(255, 0, 0)
	r.CopyEx(ship.arrowTex, nil, UIRect, 0, nil, sdl.FLIP_VERTICAL)
}

//Часть "физического движка", запускается непосредственно перед update
func (ship *ShipGameObject) ApplyForce(force V2.V2) {
	ship.forceSum = ship.forceSum.Add(force)
}
