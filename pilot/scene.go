package main

import (
	"github.com/Shnifer/flierproto1/v2"
	"github.com/veandco/go-sdl2/sdl"
)

//Менеджер объектов, группирующий вызовы главного цикла

type SceneObject interface {
	Init(s *Scene)
	Update(dt float32)
	Draw(r *sdl.Renderer)
	GetID() string
}

type HugeMass interface {
	GetGravState() (pos V2.V2, Mass float32)
}

type Scene struct {
	//Рендерер запоминаем в сцену, CONST не меняем
	R              *sdl.Renderer
	ControlHandler *controlHandler

	//TODO: структура с сортировкой по Z-order
	Objects []SceneObject
	idmap   map[string]SceneObject

	//включаем камеру в структуру
	Camera
}

func NewScene(r *sdl.Renderer, ch *controlHandler) *Scene {
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
	for i := range s.Objects {
		s.Objects[i].Draw(s.R)
	}
}

func (s Scene) GetObjByID(name string) SceneObject {
	return s.idmap[name]
}
