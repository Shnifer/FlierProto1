package main

import (
	"github.com/Shnifer/flierproto1/v2"
	"github.com/veandco/go-sdl2/sdl"
)

type particle struct {
	active bool

	pos   V2.V2
	speed V2.V2

	color    sdl.Color
	lifeTime float32
	restTime float32
}

type ProduceStats struct {
	Intense        float32
	intenseCounter float32
	color          sdl.Color
	lifeTime       float32
	pos, speed     V2.V2

	//случайные отклонения, круговой вектор заданого радиуса
	randpos   float32
	randspeed float32
}

//Частицы могут часто создаваться, поэтому делаем постоянный срез, и управляем чеез поле active
type ParticleSystem struct {
	scene    *Scene
	maxCount int
	curCount int

	//ходим массив по кругу
	cursor    int
	particles []particle
}

func newParticleSystem(maxCount int) *ParticleSystem {
	return &ParticleSystem{
		maxCount:  maxCount,
		particles: make([]particle, maxCount),
	}
}

func (ps *ParticleSystem) GetID() string {
	return ""
}

func (ps *ParticleSystem) Init(scene *Scene) {
	ps.scene = scene
}

func (ps *ParticleSystem) Update(dt float32) {
	//TODO:всяко стоит распараллелить
	for i, v := range ps.particles {
		if !v.active {
			continue
		}
		if v.restTime <= dt {
			ps.particles[i].active = false
			ps.particles[i].restTime = 0
			ps.curCount--
		}
		ps.particles[i].restTime -= dt
		ps.particles[i].pos = ps.particles[i].pos.AddMul(ps.particles[i].speed, dt)
	}
}

func (ps *ParticleSystem) Draw(r *sdl.Renderer) {

	for _, v := range ps.particles {
		if !v.active {
			continue
		}
		r.SetDrawBlendMode(sdl.BLENDMODE_BLEND)
		r.SetDrawColor(v.color.R, v.color.G, v.color.B, uint8(255*v.restTime/v.lifeTime))
		x, y := ps.scene.CameraTransformV2(v.pos)
		points := make([]sdl.Point, 5*ps.maxCount+5)
		points[0] = sdl.Point{x, y}
		points[1] = sdl.Point{x + 1, y + 1}
		points[2] = sdl.Point{x - 1, y + 1}
		points[3] = sdl.Point{x + 1, y - 1}
		points[4] = sdl.Point{x - 1, y - 1}
		r.DrawPoints(points)
	}
}

func (ps *ParticleSystem) Produce(dt float32, pStats *ProduceStats) {
	pStats.intenseCounter += pStats.Intense * dt
	numToProduce := int(pStats.intenseCounter)
	if numToProduce == 0 {
		return
	}
	pStats.intenseCounter -= float32(numToProduce)

	for i := 0; i < numToProduce; i++ {
		ps.Spawn(pStats)
	}
}

func (ps *ParticleSystem) Spawn(pStats *ProduceStats) {
	if ps.curCount == ps.maxCount {
		//TODO: убирать самые старые
		return
	}
	ps.curCount++
	for ; ps.particles[ps.cursor].active; ps.cursor = (ps.cursor + 1) % ps.maxCount {
		//цикл для позиционирования на неактивный элемент
	}

	p := particle{
		active:   true,
		pos:      pStats.pos.Add(V2.RandomInCircle(pStats.randpos)),
		speed:    pStats.speed.Add(V2.RandomInCircle(pStats.randspeed)),
		color:    pStats.color,
		lifeTime: pStats.lifeTime,
		restTime: pStats.lifeTime,
	}

	ps.particles[ps.cursor] = p
}
