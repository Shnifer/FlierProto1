package main

import (
	"github.com/Shnifer/flierproto1/v2"
	"github.com/Shnifer/flierproto1/scene"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/Shnifer/flierproto1/texture"
)

type ShipGameObject struct {
	pos   V2.V2
	speed V2.V2
	angle      float32
	angleSpeed float32

	fixedSize int32

	scene *scene.Scene
	tex *sdl.Texture
}

func newShip() *ShipGameObject {
	return &ShipGameObject{
		fixedSize: DEFVAL.ShipSize,
	}
}

func (ship *ShipGameObject) Init(scene *scene.Scene) {
	ship.scene = scene
	ship.tex = texture.Cache.GetTexture("ship.png")
}

func (ship *ShipGameObject) GetID() string {
	return ""
}

func (ship *ShipGameObject) Update(dt float32) {
	//Оставляем его лететь и вращаться по инерции, для предсказания
	//регулярно получаем от пилота фактические данные
	ship.angle += ship.angleSpeed * dt
	ship.pos.DoAddMul(ship.speed, dt)
}

func (ship ShipGameObject) Draw(r *sdl.Renderer) (res scene.RenderReqList) {

	//Показ Корабля
	var camRect *sdl.Rect
	var inCamera bool
	camRect, inCamera = ship.scene.CameraRectByCenterAndScreenSize(ship.pos, ship.fixedSize)

	if inCamera {
		req := scene.NewRenderReq(ship.tex, nil, camRect, scene.Z_GAME_OBJECT,
			-float64(ship.angle+ship.scene.CameraAngle), nil, sdl.FLIP_NONE)
		res = append(res, req)
	}

	return res
}