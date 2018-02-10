package main

import (
	V2 "github.com/Shnifer/flierproto1/v2"
	"github.com/veandco/go-sdl2/sdl"
)

type ShipGameObject struct {
	//Ship data включаем, пока нет других корабликов (значимых а не маркеров)

	//TODO: тут данные похожи на звезду. Выяснить что входит в абстрактый Transform/Rigidbody
	pos   V2.V2
	speed V2.V2

	angle      float32
	angleSpeed float32

	//Сумма внешних сил за такт
	forceSum  V2.V2
	momentSum float32

	//Радиус и коллизии и полупоперечник рисовки
	colRad float32

	scene *Scene
	tex   *sdl.Texture
	ps    *ParticleSystem

	MainEngineProducer *ProduceStats

	//TODO: абстрагировать в АНИМАЦИЮ
	anim      *Anim
	showFixed bool
}

func newShip(ps *ParticleSystem) *ShipGameObject {

	res := ShipGameObject{
		colRad:    DEFVAL.ShipSize,
		showFixed: DEFVAL.ShipShowFixed,
		angle:     0,
		ps:        ps}

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

	ship.anim = NewAnim("flame_ani.png", 5, 4, 10, false)
}

func (ship *ShipGameObject) Update(dt float32) {

	//Добавили суммарный момент и обнулили сумматор
	ship.angleSpeed += ship.momentSum * dt
	ship.angle += ship.angleSpeed * dt
	ship.momentSum = 0

	//МАГИЧЕСКИ СТАБИЛИЗИРОВАЛИ угол, если angleSpeed мал
	if abs(ship.angleSpeed)<0.5 {
		ship.angleSpeed=0
	}

	//Добавили сумму сил к скорости и обнулили сумматор
	ship.speed.DoAddMul(ship.forceSum, dt)
	ship.forceSum = V2.V2{}

	//Пренесли скорость в координату
	ship.pos.DoAddMul(ship.speed, dt)
}

func (ship ShipGameObject) Draw(r *sdl.Renderer) RenderReqList {
	//Показ Корабля
	var camRect *sdl.Rect
	var inCamera bool
	var res RenderReqList
	showFixedSized := ship.showFixed && (DEFVAL.ShipFixedSize != 0)
	if showFixedSized {
		camRect, inCamera = ship.scene.CameraRectByCenterAndScreenSize(ship.pos, DEFVAL.ShipFixedSize)
	} else {
		rect := newF32Sqr(ship.pos, ship.colRad)
		camRect, inCamera = ship.scene.CameraTransformRect(rect)
	}

	if inCamera {
		req := NewRenderReq(ship.tex, nil, camRect, Z_GAME_OBJECT,
			-float64(ship.angle+ship.scene.CameraAngle), nil, sdl.FLIP_NONE)
		res = append(res, req)
	}

	return res
}

//TODO: когда-нибудь это тоже будет частью RigidBody
//Часть "физического движка", запускается непосредственно перед update
//ПАРАМЕТР БЕЗ ДОМНОЖЕНИЯ НА ДТ
func (ship *ShipGameObject) ApplyForce(force V2.V2) {
	ship.forceSum = ship.forceSum.Add(force)
}

//ПАРАМЕТР БЕЗ ДОМНОЖЕНИЯ НА ДТ
func (ship *ShipGameObject) ApplyMoment(momentum float32) {
	ship.momentSum += momentum
}
