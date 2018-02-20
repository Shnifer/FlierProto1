package main

import (
	"github.com/Shnifer/flierproto1/control"
	MNT "github.com/Shnifer/flierproto1/mnt"
	"github.com/Shnifer/flierproto1/scene"
	"github.com/Shnifer/flierproto1/texture"
	"github.com/Shnifer/flierproto1/v2"
	"github.com/veandco/go-sdl2/sdl"
	"log"
)

type NaviCosmosScene struct {
	*scene.BScene
	ship          *ShipGameObject
	GlobalTime    float32
	camFollowShip bool

	scienceTex *sdl.Texture
	sc_pos     V2.V2
	sc_w, sc_h int32
	scShowTime float32

	fpsUI *scene.TextUI
}

func (s *NaviCosmosScene) Destroy() {
	s.scienceTex.Destroy()
	s.ship=nil
	s.BScene.Destroy()
}

func NewNaviCosmosScene(r *sdl.Renderer, ch *control.Handler) *NaviCosmosScene {
	return &NaviCosmosScene{
		BScene:        scene.NewScene(r, ch, winW, winH),
		camFollowShip: true,
	}
}

func (NaviScene *NaviCosmosScene) Init() {
	BackGround := scene.NewStaticImage("background.jpg", scene.Z_STAT_BACKGROUND)
	FrontCabin := scene.NewStaticImage("cabinBorder.png", scene.Z_STAT_HUD)
	NaviScene.AddObject(BackGround)
	NaviScene.AddObject(FrontCabin)

	Ship := newShip()
	NaviScene.ship = Ship
	NaviScene.AddObject(Ship)

	pf := texture.Cache.GetFont("phantom.ttf", 14)
	fpsUI := scene.NewTextUI("fps:", pf, sdl.Color{255, 0, 0, 255}, scene.Z_STAT_HUD, scene.FROM_ANGLE)
	fpsUI.X, fpsUI.Y = 10, 10

	NaviScene.AddObject(fpsUI)
	NaviScene.fpsUI = fpsUI

	f := texture.Cache.GetFont("interdim.ttf", 20)
	SceneCaption := scene.NewTextUI("NAVIGATOR scene", f, sdl.Color{200, 200, 200, 255}, scene.Z_STAT_HUD, scene.FROM_ANGLE)
	SceneCaption.X, SceneCaption.Y = 100, 100
	NaviScene.AddObject(SceneCaption)


	//DATA INIT
	for i := range MNT.GalaxyData {
		StarGO := &StarGameObject{Star: MNT.GalaxyData[i], startAngle: MNT.GalaxyData[i].Angle}
		NaviScene.AddObject(StarGO)
	}

	for i:=range NaviScene.Objects{
		NaviScene.Objects[i].Init(NaviScene)
	}

//	log.Println("Stars on scene", len(MNT.GalaxyData))
}

func (NaviScene *NaviCosmosScene) Update(dt float32) {

	NaviScene.cameraControlUpdate(dt)
	NaviScene.BScene.Update(dt)

	//Обновляем состояние здесь
	//Возможно вынести SCANER в отдельный объект
	if NaviScene.ship.ScanProgress >= 1 {
		NaviScene.ShowScienceData(NaviScene.ship.CurScanStar)
		NaviScene.ship.StopNaviScan()
	}
	if NaviScene.scShowTime > 0 {
		NaviScene.scShowTime -= dt
		if NaviScene.scShowTime < 0 {
			NaviScene.scShowTime = 0
		}
	}
	if NaviScene.camFollowShip {
		NaviScene.SetCameraCenter(NaviScene.ship.pos)
	}
}

func (NaviScene *NaviCosmosScene) cameraControlUpdate(dt float32) {
	сontrolHandler:=NaviScene.CH()
	scale:=NaviScene.CameraScale()
	if сontrolHandler.GetKey(sdl.SCANCODE_KP_PLUS) {
		scale*= (1 + dt)
	}
	if сontrolHandler.GetKey(sdl.SCANCODE_KP_MINUS) {
		scale *= (1 - dt)
	}
	if сontrolHandler.GetKey(sdl.SCANCODE_SPACE) {
		NaviScene.camFollowShip = true
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
	NaviScene.SetCameraScale(scale)

	ScrollSpeed := DEFVAL.CameraScrollSpeed

	delta := ScrollSpeed * dt / scale
	newCenter := NaviScene.CameraCenter()

	if сontrolHandler.GetKey(sdl.SCANCODE_W) {
		newCenter = newCenter.AddMul(V2.V2{0, 1}, delta)
	}
	if сontrolHandler.GetKey(sdl.SCANCODE_A) {
		newCenter = newCenter.AddMul(V2.V2{-1, 0}, delta)
	}
	if сontrolHandler.GetKey(sdl.SCANCODE_S) {
		newCenter = newCenter.AddMul(V2.V2{0, -1}, delta)
	}
	if сontrolHandler.GetKey(sdl.SCANCODE_D) {
		newCenter = newCenter.AddMul(V2.V2{1, 0}, delta)
	}
	if newCenter != NaviScene.CameraCenter() {
		NaviScene.camFollowShip = false
		NaviScene.SetCameraCenter(newCenter)
	}
}

//Обрабатываем по частоте IOtick~50 в секунду все события кликов мышки
func (s *NaviCosmosScene) UpdateClicks(clicks []*control.MouseClick) {
	for _, click := range clicks {
		//здесь маскируем клики в HUD и прочие скрытые элементы
		//Либо собираем всех, кто откликнулся на факт клика и анализируем
		for _, obj := range s.Objects {
			Clickable, ok := obj.(scene.Clickable)
			if !ok {
				continue
			}
			if Clickable.IsClicked(click.X, click.Y) {
				switch clicked := obj.(type) {
				case *StarGameObject:
					s.ship.StartNaviScan(clicked)
				default:
					log.Println("ЧОТ НАЖАЛИ, ОНО ОТКЛИКНУЛОСЬ, А ЧТО НЕ ПОНЯТНА!")
				}
			}
		}
	}
}

func (s *NaviCosmosScene) Draw() {
	s.BScene.Draw()

	if s.scShowTime > 0 {
		scR, inCamera := s.BScene.CameraRectByCenterAndScreenWH(s.sc_pos, int32(float32(s.sc_w)*s.scShowTime), int32(float32(s.sc_h)*s.scShowTime))
		if inCamera {
			s.R().Copy(s.scienceTex, nil, scR)
		}
	}

}

func (s *NaviCosmosScene) ShowScienceData(star *StarGameObject) {
	if s.scienceTex != nil {
		s.scienceTex.Destroy()
		s.scienceTex = nil
	}

	const startShowTime = 2

	f := texture.Cache.GetFont("furore.otf", 36)
	s.scienceTex, s.sc_w, s.sc_h = texture.CreateTextTex(s.R(), "Scanned data: "+star.ID, f, sdl.Color{150, 100, 255, 200})
	s.scShowTime = startShowTime
	s.sc_pos = star.Pos
}

func (ps *NaviCosmosScene) showFps(data string) {
	ps.fpsUI.ChangeText("fps: " + data)
}
