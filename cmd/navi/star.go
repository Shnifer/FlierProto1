package main

import (
	MNT "github.com/Shnifer/flierproto1/mnt"
	V2 "github.com/Shnifer/flierproto1/v2"
	"github.com/veandco/go-sdl2/sdl"

	"github.com/Shnifer/flierproto1/scene"
	"github.com/Shnifer/flierproto1/texture"
	"log"
)

type StarGameObject struct {
	*MNT.Star
	scene *scene.Scene
	tex   *sdl.Texture

	UItex      *sdl.Texture
	UI_H, UI_W int32
	visZRot    float32

	//const фиксируем при загрузке галактики и используем для синхронизации по глобальному времени
	startAngle float32

	nameUI *scene.TextUI
}

func (star *StarGameObject) IsClicked(x, y int32) bool {
	V := star.scene.CameraScrTransformV2(x, y)
	if V.Sub(star.Pos).LenSqr() <= star.ColRad*star.ColRad {
		return true
	}
	return false
}

func (star *StarGameObject) GetID() string {
	return star.ID
}

func (star *StarGameObject) GetGravState() (pos V2.V2, Mass float32) {
	return star.Pos, star.Mass
}

func (s *StarGameObject) Update(dt float32) {

	s.visZRot += DEFVAL.StarRotationSpeed * dt

	if s.Parent == "" {
		//независимый объект
		s.Pos = s.Pos.Add(s.Dir.Mul(dt))
	} else {
		//спутник
		s.Angle = s.startAngle + s.OrbSpeed*s.scene.NetSyncTime
		parentObj := s.scene.GetObjByID(s.Parent)
		if parentObj == nil {
			log.Panicln("Update of ", s.ID, "cant find the parent", s.Parent)
		}
		//TODO: полагаем что мы вращаемся ТОЛЬКО вокруг объекта с массой , а это HugeNass
		var pp V2.V2
		switch obj := parentObj.(type) {
		case *StarGameObject:
			pp = obj.Pos
		default:
			log.Panicln("STRANGE PARENT of ", s, "is", parentObj)
		}
		s.Pos = pp.AddMul(V2.InDir(s.Angle), s.OrbDist)
	}
}

func (s *StarGameObject) Draw(r *sdl.Renderer) (res scene.RenderReqList) {
	s.tex.SetColorMod(s.Color.R, s.Color.G, s.Color.B)
	halfsize := s.ColRad
	rect := scene.NewF32Sqr(s.Pos, halfsize)
	camRect, inCamera := s.scene.CameraTransformRect(rect)
	//log.Println("draw star #",s.N,inCamera)

	if inCamera {

		req := scene.NewRenderReq(s.tex, nil, camRect, scene.Z_GAME_OBJECT, float64(s.visZRot), nil, sdl.FLIP_NONE)
		res = append(res, req)
		//UI
		s.nameUI.X,s.nameUI.Y = s.scene.CameraTransformV2(s.Pos)
		reqUI:=s.nameUI.Draw(r)
		res = append(res, reqUI...)
	}
	return res
}

func (star *StarGameObject) Init(sc *scene.Scene) {
	star.scene = sc
	star.tex = texture.Cache.GetTexture(star.TexName)


	f := texture.Cache.GetFont("furore.otf", 9)
	star.nameUI = scene.NewTextUI(star.ID, f, sdl.Color{200, 255, 255, 200}, scene.Z_ABOVE_OBJECT, scene.FROM_CENTER)
	star.nameUI.Init(sc)
}
