package main

import (
	V2 "github.com/Shnifer/flierproto1/v2"
	"github.com/veandco/go-sdl2/sdl"
)

type ShipGameObject struct{
	//Ship data включаем, пока нет других корабликов (значимых а не маркеров)

	//TODO: тут данные похожи на звезду. Выяснить что входит в абстрактый Transform/Rigidbody
	pos V2.V2
	speed V2.V2
	angle float32
	anglespeed float32

	//мощность маршевого движка от -1 до 1
	thrust float32
	//изменение величины аксель за единицу
	thrustaxel float32
	//максимальная сила движка
	maxThrustForce float32

	//изменение angAxel за 1 секунду
	angAxel float32


	//Сумма внешних сил за такт
	forceSum V2.V2

	//Радиус и коллизии и полупоперечник рисовки
	colRad float32

	scene *Scene
	tex *sdl.Texture
	ps *ParticleSystem

	MainEngineProducer *ProduceStats

	//TODO: абстрагировать в АНИМАЦИЮ
	animActive bool
	animTime float32
	//время на показ кадра
	animSlideTime float32
	MainEngineFlameTex *animTexture

	//TODO: абстрагировать в UI
	arrowTex *sdl.Texture
}

func newShip(ps *ParticleSystem) *ShipGameObject{

	res:= ShipGameObject{colRad: 40, thrustaxel:0.33, maxThrustForce:5000, angAxel:90, angle:180, ps: ps, animSlideTime: 10}

	res.MainEngineProducer = &ProduceStats{
		lifeTime:  3,
		randpos:   7,
		randspeed: 15,
	}

	return &res
}

func (ship *ShipGameObject) Init(scene *Scene) {
	ship.scene = scene
	ship.tex = TCache.GetTexture("ship.png")
	ship.arrowTex = TCache.GetTexture("arrow.png")
	flametex := TCache.GetTexture("flame_ani.png")
	flameani := newAnimTexture(flametex,5,4)
	ship.MainEngineFlameTex = flameani
}

func (ship *ShipGameObject) Update(dt float32) {
	CH:= ship.scene.ControlHandler

	//Поворачиваем на угол инрции
	ship.angle+=ship.anglespeed*dt

	if CH.GetKey(sdl.SCANCODE_A) {
		ship.anglespeed += (ship.angAxel * dt)
	}
	if CH.GetKey(sdl.SCANCODE_D) {
		ship.anglespeed -= (ship.angAxel * dt)
	}

	if CH.GetKey(sdl.SCANCODE_W) {
		ship.thrust+=ship.thrustaxel*dt
		if ship.thrust>1 {
			ship.thrust=1
		}
	}
	if CH.GetKey(sdl.SCANCODE_S) {
		ship.thrust-=ship.thrustaxel*dt
		if ship.thrust<(0) {
			ship.thrust=0
		}
	}

	//Расчитываем и добавляем силу движка к уже посчитаной гравитации
	ThrustForce := V2.InDir(ship.angle).Mul(ship.thrust*ship.maxThrustForce*dt)
	ship.ApplyForce(ThrustForce)

	//Добавили сумму сил к скорости и обнулили сумматор
	ship.speed.DoAddMul(ship.forceSum,dt)
	ship.forceSum =V2.V2{}

	//Пренесли скорость в координату
	ship.pos.DoAddMul(ship.speed,dt)

	//Отдельный по сути блок анимации
	const MaxIntense = 300
	produceMainEng:=(ship.thrust>0)
	if produceMainEng {
		thrust:=ship.thrust
		ship.MainEngineProducer.pos = ship.pos.Add(V2.InDir(ship.angle).Mul(-1*ship.colRad))
		psspeed:=10+thrust*70
		ship.MainEngineProducer.speed = V2.InDir(ship.angle).Mul(-psspeed).Add(ship.speed)
		ship.MainEngineProducer.Intense = MaxIntense * thrust
		ship.MainEngineProducer.color = sdl.Color{byte(100+155*thrust), byte(70*thrust), 0, 255}

		ship.ps.Produce(dt, ship.MainEngineProducer)

		ship.animActive = true
		ship.animTime+=dt
	}else {
		ship.animTime=0
		ship.animActive = false
	}
}

func (ship ShipGameObject) Draw(r *sdl.Renderer) {
	//Показ Корабля
	halfsize:=ship.colRad
	rect:=f32Rect{ship.pos.X-halfsize, ship.pos.Y-halfsize, 2*halfsize, 2*halfsize}
	camRect, inCamera := ship.scene.CameraTransformRect(rect)
	if inCamera {
		r.CopyEx(ship.tex, nil, camRect, float64(-ship.angle), nil, sdl.FLIP_VERTICAL)
	}
	//Показ анимации огня
		if inCamera && ship.animActive {
			ind:=int32(ship.animTime*1000/ship.animSlideTime)%ship.MainEngineFlameTex.totalcount
			flameRect := ship.MainEngineFlameTex.getRect(ind)
			flameCentre:=V2.V2{0.05*ship.colRad,-1.4*ship.colRad}.ApplyOnTransform(ship.pos,ship.angle)
			flamesize := float32(30)
			dRect:=newF32Rect(flameCentre,flamesize)
			cameraRect,_ := ship.scene.CameraTransformRect(dRect)
			r.CopyEx(ship.MainEngineFlameTex.tex, flameRect,cameraRect, float64(-ship.angle), nil, sdl.FLIP_NONE)
		}

	//Отдельный блок показа UI
	H:=int32(ship.thrust*float32(winH)*0.8)
	UIRect:=&sdl.Rect{X:60,
		Y:winH-30-H,
		W:40,
		H:H}
	ship.arrowTex.SetColorMod(255,0,0)
	r.CopyEx(ship.arrowTex, nil, UIRect,0,nil,sdl.FLIP_VERTICAL)
}

//Часть "физического движка", запускается непосредственно перед update
func (ship *ShipGameObject) ApplyForce(force V2.V2) {
	ship.forceSum = ship.forceSum.Add(force)
}