package scene

import (
	"github.com/Shnifer/flierproto1/control"
	"github.com/Shnifer/flierproto1/texture"
	"github.com/veandco/go-sdl2/sdl"
	"sort"
	"log"
)

//Менеджер объектов, группирующий вызовы главного цикла
type SceneObject interface {
	Init(s Scene)
	Update(dt float32)
	Draw(r *sdl.Renderer) RenderReqList
	GetID() string
	Destroy()
}
type Clickable interface {
	IsClicked(x, y int32) bool
}

type Scene interface{
	//Основные функции главного цикла
	Init()
	Update(dt float32)
	Draw()
	Destroy()
	//Ссылки базовой сцены
	R() *sdl.Renderer
	CH() *control.Handler
	NetSyncTime() float32
	SetNetSyncTime(newTime float32)
	UpdateClicks([]*control.MouseClick)
	GetObjByID(name string) SceneObject
	//Камера
	CameraInterface
}

type BScene struct {
	//Рендерер запоминаем в сцену, CONST не меняем
	//Рендерер и Контроль внешние объекты, не удаляем их при уничтожении сцены
	r              *sdl.Renderer
	сontrolHandler *control.Handler

	//TODO: структура с сортировкой по Z-order
	Objects []SceneObject
	idmap   map[string]SceneObject

	//включаем камеру в структуру
	Сamera

	//Внешне задаваемое время сессии, к нему могу обращаться звёзды для синхронизации вращения
	netSyncTime float32
}

func (s *BScene) UpdateClicks([]*control.MouseClick) {
}

func (s *BScene) Destroy() {
	for i:=range s.Objects {
		s.Objects[i].Destroy()
	}
	s.idmap = map[string]SceneObject{}
}

func NewScene(r *sdl.Renderer, ch *control.Handler, camW, camH int32) *BScene {
	return &BScene{r: r, Сamera: newCamera(camW, camH, 1), сontrolHandler: ch, idmap: make(map[string]SceneObject)}
}

func (s *BScene) AddObject(obj SceneObject) {
	s.Objects = append(s.Objects, obj)
	id := obj.GetID()
	if id != "" {
		s.idmap[id] = obj
	}
}

func (scene *BScene) Init() {
	//Убираем, метод виртуален, каждый инит свои объекты как хочет
	/*for i := range scene.Objects {
		scene.Objects[i].Init(Scene(scene))
	}*/
	log.Panicln("BScene.Init() is totally virtual! do scene.Objects[i].Init(scene) in specific Scene")
}

//TODO: Возможно сделать UPDATE в горутинах, проверить на мутексы и отвутсствие вызовов SDL
func (s *BScene) Update(dt float32) {

	s.netSyncTime +=dt

	for i := range s.Objects {
		s.Objects[i].Update(dt)
	}
}

func (s *BScene) Draw() {
	//TODO: возможно распараллелить
	var Reqs RenderReqList
	for i := range s.Objects {
		rs := s.Objects[i].Draw(s.r)

		Reqs = append(Reqs, rs...)
	}

	sort.Stable(Reqs)

	for _, v := range Reqs {
		switch req := v.(type) {
		case RenderCopyReq:
			if req.color!=nil {
				req.tex.SetColorMod(req.color.R, req.color.G, req.color.B)
			}
			s.r.CopyEx(req.tex, req.src, req.dest, req.angle, req.pivot, req.flip)
		case RenderDrawLinesReq:
			s.r.SetDrawColor(req.color.R, req.color.G, req.color.B, req.color.A)
			s.r.SetDrawBlendMode(sdl.BLENDMODE_BLEND)
			s.r.DrawLines(req.points)
		case RenderFilledCircleReq:
			t := texture.CreateFilledCirle(s.r, req.rad, req.color)
			defer t.Destroy()
			s.r.Copy(t, nil, &sdl.Rect{req.x - req.rad, req.y - req.rad, 2 * req.rad, 2 * req.rad})
		case RenderFilledPieReq:
			t := texture.CreateFilledPie(s.r, req.rad, req.inrad, req.start, req.end, req.color)
			defer t.Destroy()
			s.r.Copy(t, nil, &sdl.Rect{req.x - req.rad, req.y - req.rad, 2 * req.rad, 2 * req.rad})
		case RenderRectsReq:
			s.r.SetDrawColor(req.color.R, req.color.G, req.color.B, req.color.A)
			s.r.SetDrawBlendMode(sdl.BLENDMODE_BLEND)
			s.r.DrawRects(req.rects)
		}
	}
}

func (s *BScene) GetObjByID(name string) SceneObject {
	return s.idmap[name]
}

func (s *BScene) SetNetSyncTime(newTime float32) {
	s.netSyncTime = newTime
}

func (s *BScene) NetSyncTime() float32{
	return s.netSyncTime
}

func (s *BScene) R() *sdl.Renderer{
	return s.r
}

func (s *BScene) CH() *control.Handler{
	return s.сontrolHandler
}