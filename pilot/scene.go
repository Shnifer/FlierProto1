package main

import (
	"github.com/Shnifer/flierproto1/control"
	"github.com/Shnifer/flierproto1/v2"
	"github.com/veandco/go-sdl2/sdl"
	"sort"
)

//Слои вывода объектов на рендер
type ZLayer int

const (
	//Статический фон, обычно один и занимает весь экран
	Z_STAT_BACKGROUND ZLayer = iota * 100

	//Динамические изменения на фоне, например координатная сетка
	Z_BACKGROUND

	//Подложка под игровым объектом, например кружок выделения
	Z_UNDER_OBJECT

	//Сами игровые объекты
	Z_GAME_OBJECT

	//Сверху игровых объектов, надписи или гизмосы объектов
	Z_ABOVE_OBJECT

	//не привязанное к координатам игрового мира, например системы управления
	Z_HUD

	//обычно одна картинка с прозрачным центром отрисовывающая красивые края
	Z_STAT_HUD
)

type RenderReq struct {
	tex       *sdl.Texture
	src, dest *sdl.Rect
	z         ZLayer
	angle     float64
	pivot     *sdl.Point
	flip      sdl.RendererFlip
}

func NewRenderReq(tex *sdl.Texture, src, dest *sdl.Rect, z ZLayer, angle float64, pivot *sdl.Point, flip sdl.RendererFlip) RenderReq {
	return RenderReq{tex: tex,
		src:   src,
		dest:  dest,
		z:     z,
		angle: angle,
		pivot: pivot,
		flip:  flip,
	}
}

func NewRenderReqSimple(tex *sdl.Texture, src, dest *sdl.Rect, z ZLayer) RenderReq {
	return RenderReq{tex: tex,
		src:   src,
		dest:  dest,
		z:     z,
		angle: 0,
		pivot: nil,
		flip:  sdl.FLIP_NONE,
	}
}
type RenderReqList []RenderReq

func (r RenderReqList) Len() int {
	return len(r)
}
func (r RenderReqList) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

func (r RenderReqList) Less(i, j int) bool {
	return r[i].z < r[j].z
}

//Менеджер объектов, группирующий вызовы главного цикла
type SceneObject interface {
	Init(s *Scene)
	Update(dt float32)
	Draw(r *sdl.Renderer) RenderReqList
	GetID() string
}

type HugeMass interface {
	GetGravState() (pos V2.V2, Mass float32)
}

type Scene struct {
	//Рендерер запоминаем в сцену, CONST не меняем
	R              *sdl.Renderer
	ControlHandler *control.Handler

	//TODO: структура с сортировкой по Z-order
	Objects []SceneObject
	idmap   map[string]SceneObject

	//включаем камеру в структуру
	Camera
}

func NewScene(r *sdl.Renderer, ch *control.Handler) *Scene {
	return &Scene{R: r, Camera: newCamera(), ControlHandler: ch, idmap: make(map[string]SceneObject)}
}

func (s *Scene) AddObject(obj SceneObject) {
	s.Objects = append(s.Objects, obj)
	id := obj.GetID()
	if id != "" {
		s.idmap[id] = obj
	}
}

func (scene *Scene) Init() {
	for i := range scene.Objects {
		scene.Objects[i].Init(scene)
	}
}

//TODO: Возможно сделать UPDATE в горутинах, проверить на мутексы и отвутсствие вызовов SDL
func (s *Scene) Update(dt float32) {

	if s.ControlHandler.GetKey(sdl.SCANCODE_KP_PLUS) {
		s.CameraScale *= (1 + dt)
	}
	if s.ControlHandler.GetKey(sdl.SCANCODE_KP_MINUS) {
		s.CameraScale *= (1 - dt)
	}

	for i := range s.Objects {
		s.Objects[i].Update(dt)
	}
}

func (s Scene) Draw() {
	//TODO: возможно распараллелить
	var Reqs RenderReqList
	for i := range s.Objects {
		rs := s.Objects[i].Draw(s.R)

		Reqs = append(Reqs, rs...)
	}

	sort.Stable(Reqs)

	for _, v := range Reqs {
		s.R.CopyEx(v.tex, v.src, v.dest, v.angle, v.pivot, v.flip)
	}
}

func (s Scene) GetObjByID(name string) SceneObject {
	return s.idmap[name]
}