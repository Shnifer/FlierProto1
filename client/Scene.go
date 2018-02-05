package main

import (
	"github.com/veandco/go-sdl2/sdl"
	"github.com/Shnifer/flierproto1/v2"
)

//Менеджер объектов, группирующий вызовы главного цикла

const GravityConst = 5000
//Расстояние по Z для избежания нулевой дистанции гравитирования
const DepthSqr = 1000

type f32Rect struct{
	X,Y,W,H float32
}

func newF32Rect(center V2.V2, rad float32) f32Rect {
	return f32Rect{center.X-rad, center.Y-rad, 2*rad, 2*rad}
}

type SceneObject interface{
	Init(s *Scene)
	Update(dt float32)
	Draw(r *sdl.Renderer)
}

type HugeMass interface {
	GetGravState() (pos V2.V2, Mass float32)
}

type Scene struct {
	//Рендерер запоминаем в сцену, CONST не меняем
	R *sdl.Renderer
	ControlHandler *controlHandler
	//TODO: структура с сортировкой по Z-order
	Objects []SceneObject


	//Пока что камера -- свойство сцены

	//ЦентрКамеры в мировых координатах
	CameraCenter V2.V2
	//Масштаб пикселей/единицу мировых координат,
	//увеличение текстур там, где это надо
	CameraScale float32
}

func NewScene(r *sdl.Renderer, ch *controlHandler) *Scene{
	return &Scene{R: r, CameraScale:1, ControlHandler:ch}
}

func (s Scene) CameraTransformV2(v V2.V2) (x,y int32) {
	x = winW/2+int32((v.X-s.CameraCenter.X)*s.CameraScale)
	y = winH/2+int32((v.Y-s.CameraCenter.Y)*s.CameraScale)
	return
}

func (s Scene) CameraTransformRect(r f32Rect) (camRect *sdl.Rect, inCamera bool) {
	//TODO: положение камеры не по центру экрана
	x,y:= s.CameraTransformV2(V2.V2{r.X,r.Y})
	res:=sdl.Rect{
		x,
		y,
		int32(r.W*s.CameraScale),
		int32(r.H*s.CameraScale),
	}

	inCamera = !(res.X+res.W<0 || res.X >winW || res.Y>winH || res.Y+res.H<0)

	return &res, inCamera
}

func (s *Scene) AddObject(obj SceneObject) {
	s.Objects = append(s.Objects, obj)
}

func (scene *Scene) Init() {
	for i:=range scene.Objects{
		scene.Objects[i].Init(scene)
	}
}

//TODO: Возможно сделать UPDATE в горутинах, проверить на мутексы и отвутсствие вызовов SDL
func (s *Scene) Update(dt float32) {

	if s.ControlHandler.GetKey(sdl.SCANCODE_KP_PLUS) {
		s.CameraScale*=(1+dt)
	}
	if s.ControlHandler.GetKey(sdl.SCANCODE_KP_MINUS) {
		s.CameraScale*=(1-dt)
	}

	for i:=range s.Objects{
		s.Objects[i].Update(dt)
	}
}

func (s Scene) Draw() {
	for i:=range s.Objects{
		s.Objects[i].Draw(s.R)
	}
}