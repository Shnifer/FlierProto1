package main

import (
	"github.com/Shnifer/flierproto1/control"
	MNT "github.com/Shnifer/flierproto1/mnt"
	"github.com/Shnifer/flierproto1/scene"
	"github.com/veandco/go-sdl2/sdl"
	"log"
	"github.com/Shnifer/flierproto1/v2"
)

type NaviCosmosScene struct {
	*scene.Scene
	ship *ShipGameObject
}

func NewNaviCosmosScene(r *sdl.Renderer, ch *control.Handler) *NaviCosmosScene {
	return &NaviCosmosScene{
		Scene: scene.NewScene(r, ch, winW, winH),
	}
}

func (NaviScene *NaviCosmosScene) Init() {
	BackGround := scene.NewStaticImage("background.jpg", scene.Z_STAT_BACKGROUND)
	FrontCabin := scene.NewStaticImage("cabinBorder.png", scene.Z_STAT_HUD)
	NaviScene.AddObject(BackGround)
	NaviScene.AddObject(FrontCabin)

	//DATA INIT
	for _, starData := range MNT.GalaxyData {
		StarGO := &StarGameObject{Star: starData, startAngle:starData.Angle}
		NaviScene.AddObject(StarGO)
	}
	log.Println("Stars on scene", len(MNT.GalaxyData))

	Ship := newShip()
	NaviScene.ship = Ship
	NaviScene.AddObject(Ship)

	NaviScene.Scene.Init()
}

func (NaviScene *NaviCosmosScene) Update(dt float32) {

	NaviScene.cameraControlUpdate(dt)
	NaviScene.Scene.Update(dt)
}

func (NaviScene *NaviCosmosScene) cameraControlUpdate(dt float32) {
	if NaviScene.ControlHandler.GetKey(sdl.SCANCODE_KP_PLUS) {
		NaviScene.CameraScale *= (1 + dt)
	}
	if NaviScene.ControlHandler.GetKey(sdl.SCANCODE_KP_MINUS) {
		NaviScene.CameraScale *= (1 - dt)
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
	Clamp(&NaviScene.CameraScale, min, max)

	ScrollSpeed:=DEFVAL.CameraScrollSpeed

	delta:= ScrollSpeed*dt/NaviScene.CameraScale
	newCenter:=NaviScene.CameraCenter

	if NaviScene.ControlHandler.GetKey(sdl.SCANCODE_W) {
		newCenter = newCenter.AddMul(V2.V2{0,1}, delta)
	}
	if NaviScene.ControlHandler.GetKey(sdl.SCANCODE_A) {
		newCenter = newCenter.AddMul(V2.V2{-1,0}, delta)
	}
	if NaviScene.ControlHandler.GetKey(sdl.SCANCODE_S) {
		newCenter = newCenter.AddMul(V2.V2{0,-1}, delta)
	}
	if NaviScene.ControlHandler.GetKey(sdl.SCANCODE_D) {
		newCenter = newCenter.AddMul(V2.V2{1,0}, delta)
	}
	NaviScene.CameraCenter = newCenter
}
