package main

import (
	"github.com/Shnifer/flierproto1/scene"
	"github.com/Shnifer/flierproto1/texture"
	"github.com/Shnifer/flierproto1/v2"
	"github.com/veandco/go-sdl2/sdl"
	"log"
)

type ShipGameObject struct {
	pos        V2.V2
	speed      V2.V2
	angle      float32
	angleSpeed float32

	fixedSize int32

	scene *scene.BScene
	tex   *sdl.Texture

	maxScanRange float32
	CurScanStar  *StarGameObject
	ScanProgress float32
	ScanSpeed    float32

	ScanRadTex *sdl.Texture
}

func newShip() *ShipGameObject {
	return &ShipGameObject{
		fixedSize:    DEFVAL.ShipSize,
		maxScanRange: BSP.ScanRange,
		ScanSpeed:    BSP.ScanSpeed,
	}
}

func (ship *ShipGameObject) Init(scene *scene.BScene) {
	transGreen := sdl.Color{0, 200, 0, 100}

	ship.scene = scene
	ship.tex = texture.Cache.GetTexture("ship.png")
	ship.ScanRadTex = texture.CreateFilledCirle(scene.r, int32(ship.maxScanRange), transGreen)
}

func (ship *ShipGameObject) GetID() string {
	return ""
}

func (ship *ShipGameObject) Update(dt float32) {
	//Оставляем его лететь и вращаться по инерции, для предсказания
	//регулярно получаем от пилота фактические данные
	ship.angle += ship.angleSpeed * dt
	ship.pos.DoAddMul(ship.speed, dt)

	if ship.CurScanStar != nil {
		if ship.pos.Sub(ship.CurScanStar.Pos).Len() > ship.maxScanRange+ship.CurScanStar.ColRad {
			log.Println("scaning range BROCKEN")
			ship.StopNaviScan()
		} else {
			ship.ScanProgress += ship.ScanSpeed * dt
			if ship.ScanProgress >= 1 {
				ship.ScanProgress = 1
			}
		}
	}
}

func (ship ShipGameObject) Draw(r *sdl.Renderer) (res scene.RenderReqList) {

	//Показ Корабля
	var camRect *sdl.Rect
	var inCamera bool

	camRect, inCamera = ship.scene.CameraRectByCenterAndScreenSize(ship.pos, ship.fixedSize)

	if inCamera {
		req := scene.NewRenderReq(ship.tex, nil, camRect, scene.Z_GAME_OBJECT,
			-float64(ship.angle+ship.scene.сameraAngle), nil, sdl.FLIP_NONE,nil)
		res = append(res, req)
	}

	camRect, inCamera = ship.scene.CameraTransformRect(scene.NewF32Sqr(ship.pos, ship.maxScanRange))

	if inCamera {
		req := scene.NewRenderReqSimple(ship.ScanRadTex, nil, camRect, scene.Z_UNDER_OBJECT)
		res = append(res, req)
	}

	if ship.CurScanStar != nil {
		transYellow := sdl.Color{255, 200, 0, 255}

		x, y := ship.scene.CameraTransformV2(ship.CurScanStar.Pos)
		inrad := int32(ship.CurScanStar.ColRad*ship.scene.cameraScale) + 3
		rad := inrad + 10
		req := scene.NewFilledPieReq(x, y, rad, inrad, 0, int32(ship.ScanProgress*360), transYellow, scene.Z_UNDER_OBJECT)
		res = append(res, req)
	}

	return res
}

func (ship *ShipGameObject) StartNaviScan(star *StarGameObject) {
	if ship.pos.Sub(star.Pos).Len() > ship.maxScanRange+star.ColRad {
		log.Println("target too far")
		return
	}
	if ship.CurScanStar == star {
		log.Println("already scanning this star!")
		return
	}
	ship.CurScanStar = star
	ship.ScanProgress = 0
	log.Println("Started scanning", star.ID)
}

func (ship *ShipGameObject) StopNaviScan() {
	ship.CurScanStar = nil
	ship.ScanProgress = 0
	log.Println("Stopped scanning")
}
