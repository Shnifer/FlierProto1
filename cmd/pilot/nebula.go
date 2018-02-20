package main

import (
	"github.com/Shnifer/flierproto1/scene"
	"github.com/Shnifer/flierproto1/texture"
	"github.com/Shnifer/flierproto1/v2"
	"github.com/veandco/go-sdl2/sdl"
	"log"
	"sync"

	"math"
)

type metaballBase struct {
	star   *StarGameObject
	weight float32
}

//База будет вынесена в интерфейс, так что получения позиции - метод
func (b *metaballBase) Pos() V2.V2 {
	return b.star.Pos
}

//База отчитывается от максимальном радиусе, в котором может быть результат
//Из квадратов ограничивающих базы будут построены границы небулы
func (b metaballBase) maxR() float32 {
	return b.weight
}

//Расстояние (кв) на котором гарантируется область вокруг точки
func (b metaballBase) garantR2(totalW float32) float32 {
	return (1 - float32(math.Sqrt(float64(totalW)))) * b.weight * b.weight
}

//База возвращает потенциал для точки
//Должна работать для любой точки, сама отслеживая ограничения если нужно
func (b metaballBase) calcW(point V2.V2) float32 {
	v := b.Pos().Sub(point)
	if abs(v.X) > b.weight || abs(v.Y) > b.weight {
		return 0
	}
	dist2 := v.LenSqr() / b.weight / b.weight
	x := 1 - dist2
	if x <= 0 {
		return 0
	}
	return (x * x)
}

type metaball struct {
	base   []metaballBase
	totalW float32
}

type Nebula struct {
	metaball

	effect string

	//пока для переключателя 0 - не показывать, 1 - спрайты, 2 - по точкам
	drawMode int

	id string

	scene     scene.Scene
	tex       *sdl.Texture
	smokeTexs []*sdl.Texture
}

func (n *Nebula) Destroy() {
	n.tex.Destroy()
	//Текстуры дыма smokeTexs из кэша
}

func NewNebula(id string, stars []*StarGameObject, w float32) *Nebula {
	res := Nebula{id: id, effect: "Nebula effect string", metaball: metaball{totalW: 0.8}}
	for _, star := range stars {
		res.base = append(res.base, metaballBase{star: star, weight: w})
	}
	return &res
}

func (n *Nebula) Init(s scene.Scene) {
	n.scene = s
	n.smokeTexs = make([]*sdl.Texture, 3)
	n.smokeTexs[0] = texture.Cache.GetTexture("smoke1.png")
	n.smokeTexs[1] = texture.Cache.GetTexture("smoke2.png")
	n.smokeTexs[2] = texture.Cache.GetTexture("smoke3.png")
	for _, tex := range n.smokeTexs {
		tex.SetAlphaMod(60)
	}
	//	n.smokeTexs[3] = texture.Cache.GetTexture("smoke.png")
}

func (n *Nebula) Update(dt float32) {
}

func (n Nebula) isInside(point V2.V2) bool {
	var sum float32
	for _, base := range n.base {
		sum += base.calcW(point)
		if sum > n.totalW {
			return true
		}
	}
	return false
}

func (n Nebula) isInsideSum(point V2.V2) (bool, float32) {
	var sum float32
	for _, base := range n.base {
		sum += base.calcW(point)
	}
	return sum > n.totalW, sum
}

func (n *Nebula) Draw(r *sdl.Renderer) (res scene.RenderReqList) {
	//stop := timeCheck("nebuladraw")
	//defer stop()

	//ДРУГИМ ПУТЁМ
	switch n.drawMode {
	case 0:
		{
			return res
		}
	case 1:
		{
			return n.SmokedDraw(r)
		}
	}

	totalRect, pixels := n.calcPixels()

	if n.tex != nil {
		n.tex.Destroy()
	}

	tex, err := texture.PixelsToTexture(n.scene.R(), pixels, int(totalRect.W), int(totalRect.H))
	if err != nil {
		log.Panicln(err)
	}
	n.tex = tex
	req := scene.NewRenderReqSimple(n.tex, nil, &totalRect, scene.Z_UNDER_OBJECT)
	res = append(res, req)
	return res
}

func (n *Nebula) calcPixels() (total sdl.Rect, pixels []byte) {
	var bases []sdl.Rect
	baseIndNebulaInd := make(map[int]int)
	for nInd, base := range n.base {

		fsqr := scene.NewF32Sqr(base.Pos(), base.maxR())

		camrect, inCamera := n.scene.CameraTransformRect(fsqr)
		if inCamera {
			bases = append(bases, *camrect)
			total = total.Union(camrect)
			baseIndNebulaInd[len(bases)-1] = nInd
		}
	}

	if total.Empty() {
		//Вся небула не попадает в экран -- больше не считаем
		return total, []byte{}
	}

	//TODO: распараллелить!
	pixels = make([]byte, total.W*total.H*4)
	wg := sync.WaitGroup{}
	tX, tY := total.X, total.Y //угол total в экранных координатах
	tW := total.W
	const maxInTimeN = 3
	wcontrol := make(chan bool, maxInTimeN)
	for baseInd, base := range bases {
		wg.Add(1)
		go func(baseInd int, base sdl.Rect) {
			wcontrol <- true
			//======
			garantR2 := n.base[baseIndNebulaInd[baseInd]].garantR2(n.totalW)
			bpoint := n.base[baseIndNebulaInd[baseInd]].Pos()
			var intRects []sdl.Rect
			var myInd int
			for sInd := 0; sInd < len(bases); sInd++ {
				if base.HasIntersection(&bases[sInd]) {
					intRects = append(intRects, bases[sInd])
					if sInd == baseInd {
						myInd = len(intRects) - 1
					}
				}
			}
			bX, bY := base.X-tX, base.Y-tY //угол base относительно начала total для расчёта индекса
			for y := int32(0); y < base.W; y++ {
			loop:
				for x := int32(0); x < base.H; x++ {
					//Вдруг точка есть в более ранних прямоугольниках?
					aX, aY := x+base.X, y+base.Y //координаты точки в экранных координатах
					isMoreRects := false
					for r := 0; r < len(intRects); r++ {
						if r == myInd {
							continue
						}
						cr := &intRects[r]
						if cr.X <= aX && cr.Y <= aY &&
							cr.X+cr.W > aX && cr.Y+cr.H > aY {
							if r < myInd {
								continue loop
							} else {
								isMoreRects = true
								break
							}
						}
					}

					point := n.scene.CameraScrTransformV2(aX, aY)
					D2 := (point.X-bpoint.X)*(point.X-bpoint.X) + (point.Y-bpoint.Y)*(point.Y-bpoint.Y)
					draw := false
					if D2 <= garantR2 {
						draw = true
					} else if !isMoreRects && garantR2 > 0 {
						continue
					} else {
						draw = n.isInside(point)
					}

					if draw {
						//ПИШЕМ БЕЗ МУТЕКСОВ, как пидоры!
						ind := (bY+y)*(tW*4) + (bX+x)*4
						pixels[ind+0] = 255
						pixels[ind+1] = 255
						pixels[ind+2] = 255
						pixels[ind+3] = 255

					}
				}
			}
			<-wcontrol
			wg.Done()
			//======
		}(baseInd, base)
	}
	wg.Wait()
	return total, pixels
}

func (n Nebula) GetID() string {
	return n.id
}

func (n *Nebula) SmokedDraw(r *sdl.Renderer) (res scene.RenderReqList) {
	var bases []sdl.Rect
	var total sdl.Rect
	baseIndNebulaInd := make(map[int]int)

	for nInd, base := range n.base {

		fsqr := scene.NewF32Sqr(base.Pos(), base.maxR())

		camrect, inCamera := n.scene.CameraTransformRect(fsqr)
		if inCamera {
			bases = append(bases, *camrect)
			total = total.Union(camrect)
			baseIndNebulaInd[len(bases)-1] = nInd
		}
	}

	if total.Empty() {
		//Вся небула не попадает в экран -- больше не считаем
		return res
	}

	wg := sync.WaitGroup{}
	//tX, tY := total.X, total.Y //угол total в экранных координатах
	const maxInTimeN = 3
	wcontrol := make(chan bool, maxInTimeN)
	reschan := make(chan scene.RenderReqList, len(bases))
	for baseInd, base := range bases {
		wg.Add(1)
		go func(baseInd int, base sdl.Rect) {
			wcontrol <- true
			chanResult := make(scene.RenderReqList, 0)
			//======
			//garantR2 := n.base[baseIndNebulaInd[baseInd]].garantR2(n.totalW)
			//bpoint := n.base[baseIndNebulaInd[baseInd]].Pos()
			var intRects []sdl.Rect
			var myInd int
			for sInd := 0; sInd < len(bases); sInd++ {
				if base.HasIntersection(&bases[sInd]) {
					intRects = append(intRects, bases[sInd])
					if sInd == baseInd {
						myInd = len(intRects) - 1
					}
				}
			}
			//bX, bY := base.X-tX, base.Y-tY //угол base относительно начала total для расчёта индекса
			const dotStep = 30
			const halfDraw = 30
			bx := ((base.X-1)/dotStep+1)*dotStep - base.X
			by := ((base.Y-1)/dotStep+1)*dotStep - base.Y

			for y := by; y < base.W; y += dotStep {
			loop:
				for x := bx; x < base.H; x += dotStep {
					//Вдруг точка есть в более ранних прямоугольниках?
					aX, aY := x+base.X, y+base.Y //координаты точки в экранных координатах
					//isMoreRects := false
					for r := 0; r < len(intRects); r++ {
						if r == myInd {
							continue
						}
						cr := &intRects[r]
						if cr.X <= aX && cr.Y <= aY &&
							cr.X+cr.W > aX && cr.Y+cr.H > aY {
							if r < myInd {
								continue loop
							} else {
								//isMoreRects = true
								break
							}
						}
					}

					point := n.scene.CameraScrTransformV2(aX, aY)
					//D2 := (point.X-bpoint.X)*(point.X-bpoint.X) + (point.Y-bpoint.Y)*(point.Y-bpoint.Y)
					draw, sum := n.isInsideSum(point)
					sizeK := (sum - n.totalW) / n.totalW
					if sizeK > 2 {
						sizeK = 2
					}

					if draw {
						smokeInd := int(aX/dotStep+aY/dotStep) % len(n.smokeTexs)
						if smokeInd < 0 {
							smokeInd = -smokeInd
						}
						smokeTex := n.smokeTexs[smokeInd]
						size := int32(halfDraw * sizeK)
						camrect := sdl.Rect{aX - size, aY - size, 2 * size, 2 * size}
						req := scene.NewRenderReqSimple(smokeTex, nil, &camrect, scene.Z_UNDER_OBJECT+scene.ZLayer(baseInd%10))
						chanResult = append(chanResult, req)
					}
				}
			}
			//======
			reschan <- chanResult
			<-wcontrol
			wg.Done()
		}(baseInd, base)
	}
	wg.Wait()
	for range bases {
		chanResult := <-reschan
		res = append(res, chanResult...)
	}

	return res
}
