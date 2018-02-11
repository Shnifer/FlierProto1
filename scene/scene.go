package scene

import (
	"github.com/Shnifer/flierproto1/control"
	"github.com/veandco/go-sdl2/sdl"
	"sort"
)

//Менеджер объектов, группирующий вызовы главного цикла
type SceneObject interface {
	Init(s *Scene)
	Update(dt float32)
	Draw(r *sdl.Renderer) RenderReqList
	GetID() string
}

type Scene struct {
	//Рендерер запоминаем в сцену, CONST не меняем
	R              *sdl.Renderer
	ControlHandler *control.Handler

	//TODO: структура с сортировкой по Z-order
	Objects []SceneObject
	idmap   map[string]SceneObject

	//включаем камеру в структуру
	Сamera
}

func NewScene(r *sdl.Renderer, ch *control.Handler, camW,camH int32) *Scene {
	return &Scene{R: r, Сamera: newCamera(camW,camH,1), ControlHandler: ch, idmap: make(map[string]SceneObject)}
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
		switch req := v.(type) {
		case RenderCopyReq:
			s.R.CopyEx(req.tex, req.src, req.dest, req.angle, req.pivot, req.flip)
		case RenderDrawLinesReq:
			s.R.SetDrawColor(req.color.R, req.color.G, req.color.B, req.color.A)
			s.R.SetDrawBlendMode(sdl.BLENDMODE_BLEND)
			s.R.DrawLines(req.points)
		}
	}
}

func (s Scene) GetObjByID(name string) SceneObject {
	return s.idmap[name]
}
