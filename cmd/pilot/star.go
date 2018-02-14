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

	visZrot float32
	//const фиксируем при загрузке галактики и используем для синхронизации по глобальному времени
	startAngle float32
}

func (star *StarGameObject) GetID() string {
	return star.ID
}

func (star *StarGameObject) GetGravState() (pos V2.V2, Mass float32) {
	return star.Pos, star.Mass
}

func (s *StarGameObject) Update(dt float32) {

	s.visZrot += DEFVAL.StarRotationSpeed * dt

	if s.Parent == "" {
		//независимый объект
		//глобальное время не применяем, ибо звезды и точки покоятся,
		// иначе для расчета понадобится хранить начальное положение
		s.Pos = s.Pos.Add(s.Dir.Mul(dt))
	} else {
		//спутник
		s.Angle = s.startAngle + s.OrbSpeed*s.scene.NetSyncTime
		parentObj := s.scene.GetObjByID(s.Parent)
		if parentObj == nil {
			log.Panicln("Update of ", s.ID, "cant find the parent", s.Parent)
		}
		//TODO: полагаем что мы вращаемся ТОЛЬКО вокруг объекта с массой , а это HugeNass
		pp, _ := parentObj.(HugeMass).GetGravState()
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

		req := scene.NewRenderReq(s.tex, nil, camRect, scene.Z_GAME_OBJECT, float64(s.visZrot), nil, sdl.FLIP_NONE)
		res = append(res, req)
		//UI

		/* Пока убрали названия из пилота чтобы выделить навигатора
		cx, cy := s.scene.CameraTransformV2(rect.Center)
		destRect := &sdl.Rect{cx - s.UI_W/2, cy - s.UI_H/2, s.UI_W, s.UI_H}
		reqUI := scene.NewRenderReqSimple(s.UItex, nil, destRect, scene.Z_ABOVE_OBJECT)
		//убрать аппенд выше res = append(res, req, reqUI)
		*/
	}
	return res
}

func (star *StarGameObject) Init(s *scene.Scene) {
	star.scene = s
	star.tex = texture.Cache.GetTexture(star.TexName)
}
