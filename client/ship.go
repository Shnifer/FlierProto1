package main

import (
	"github.com/Shnifer/FlierProto1/V2"
	"github.com/veandco/go-sdl2/sdl"
)

type ShipGameObject struct{
	//Ship data включаем, пока нет других корабликов (значимых а не маркеров)

	//TODO: тут данные похожи на звезду. Выяснить что входит в абстрактый Transform/Rigidbody
	pos V2.V2
	speed V2.V2
	angle float32
	anglespeed float32

	axel float32
	angAxel float32

	ForceSum V2.V2

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
}

func newShip(ps *ParticleSystem) *ShipGameObject{

	res:= ShipGameObject{colRad: 40, axel:50, angAxel:90, angle:180, ps: ps, animSlideTime: 10}

	res.MainEngineProducer = &ProduceStats{
		intense: 200,
		lifeTime: 3,
		color: sdl.Color{255,0,0,255},
		randpos:7,
		randspeed:15,
	}

	return &res
}

func (ship *ShipGameObject) Init(scene *Scene) {
	ship.scene = scene
	ship.tex = TCache.GetTexture("ship.png")
	flametex := TCache.GetTexture("flame_ani.png")
	flameani := newAnimTexture(flametex,5,4)
	ship.MainEngineFlameTex = flameani
}

func (ship *ShipGameObject) Update(dt float32) {
	CH:= ship.scene.ControlHandler

	//Поворачиваем на угол инрции
	ship.angle+=ship.anglespeed*dt

	produceMainEng:=false
	if CH.GetKey(sdl.SCANCODE_A) {
		ship.anglespeed += (ship.angAxel * dt)
	}
	if CH.GetKey(sdl.SCANCODE_D) {
		ship.anglespeed -= (ship.angAxel * dt)
	}

	if CH.GetKey(sdl.SCANCODE_W) {
		addSpeed:=V2.InDir(ship.angle).Mul(ship.axel*dt)
		ship.speed = V2.Add(ship.speed, addSpeed)
		produceMainEng = true
	}
	if CH.GetKey(sdl.SCANCODE_S) {
		addSpeed:=V2.InDir(ship.angle).Mul(-ship.axel*dt)
		ship.speed = V2.Add(ship.speed, addSpeed)
	}

	//Добавили постоянные силы гравитации и обнулили сумматор
	ship.speed.DoAddMul(ship.ForceSum,dt)
	ship.ForceSum=V2.V2{}

	//Пренесли скорость в координату
	ship.pos.DoAddMul(ship.speed,dt)

	//Сдвинули камеру
	ship.scene.CameraCenter = ship.pos

	//Отдельный по сути блок анимации
	if produceMainEng {
		ship.MainEngineProducer.pos = ship.pos.Add(V2.InDir(ship.angle).Mul(-1*ship.colRad))
		ship.MainEngineProducer.speed = V2.InDir(ship.angle).Mul(-50).Add(ship.speed)

		ship.ps.Produce(dt, ship.MainEngineProducer)

		ship.animActive = true
		ship.animTime+=dt
	}else {
		ship.animTime=0
		ship.animActive = false
	}
}

func (ship ShipGameObject) Draw(r *sdl.Renderer) {
	halfsize:=ship.colRad
	rect:=f32Rect{ship.pos.X-halfsize, ship.pos.Y-halfsize, 2*halfsize, 2*halfsize}
	camRect, inCamera := ship.scene.CameraTransformRect(rect)
	if inCamera {
		r.CopyEx(ship.tex, nil, camRect, float64(-ship.angle), nil, sdl.FLIP_VERTICAL)

		if ship.animActive {
			ind:=int32(ship.animTime*1000/ship.animSlideTime)%ship.MainEngineFlameTex.totalcount
			flameRect := ship.MainEngineFlameTex.getRect(ind)
			flameCentre:=V2.V2{0.05*ship.colRad,-1.4*ship.colRad}.ApplyOnTransform(ship.pos,ship.angle)
			flamesize := float32(30)
			dRect:=newF32Rect(flameCentre,flamesize)
			cameraRect,_ := ship.scene.CameraTransformRect(dRect)
			r.CopyEx(ship.MainEngineFlameTex.tex, flameRect,cameraRect, float64(-ship.angle), nil, sdl.FLIP_NONE)
		}
	}
}

//Часть "физического движка", запускается непосредственно перед update
func (ship *ShipGameObject) ApplyForce(force V2.V2) {
	ship.ForceSum = ship.ForceSum.Add(force)
}