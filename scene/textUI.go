package scene

import (
	"github.com/Shnifer/flierproto1/texture"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
	"log"
)

const FROM_ANGLE bool = false
const FROM_CENTER bool = true

var WHITE = sdl.Color{255,255,255,255}
var RED = sdl.Color{255,0,0,255}

type TextUI struct {
	text         string
	font         *ttf.Font
	X, Y         int32
	FromCenter   bool
	color        sdl.Color
	Scale        float32
	Angle        float32

	Z ZLayer

	//Флаг что текстуру нужно пересобрать перед отрисовкой
	needReworkTex bool

	scene *Scene
	tex          *sdl.Texture
	tex_w, tex_h int32
}

func NewTextUI(text string, font *ttf.Font, color sdl.Color, z ZLayer, fromCenter bool) *TextUI {
	return &TextUI{
		text:       text,
		font:       font,
		color:      color,
		Scale:      1,
		Z:          z,
		FromCenter: fromCenter,
	}
}
func (t *TextUI) reworkTex() {
	if t.tex !=nil{
		t.tex.Destroy()
	}
	t.tex, t.tex_w, t.tex_h = texture.CreateTextTex(t.scene.R, t.text, t.font, t.color)
	t.tex.SetBlendMode(sdl.BLENDMODE_BLEND)
	t.tex.SetAlphaMod(t.color.A)
	t.needReworkTex = false
}

func (t *TextUI) ChangeFont(font *ttf.Font, color sdl.Color) {
	if t.font != font || t.color != color {
		t.font = font
		t.color = color
		t.needReworkTex = true
	}
}

func (t *TextUI) ChangeColor(color sdl.Color) {
	if t.color != color {
		t.color = color
		t.needReworkTex = true
	}
}

func (t *TextUI) ChangeText(text string) {
	if t.text != text {
		t.text = text
		t.needReworkTex = true
	}
}

func (t *TextUI) GetTexSize() (tex_w, tex_h int32) {
	if t.needReworkTex {
		log.Println("reworking texture before draw, cz of GetTexSize")
		t.reworkTex()
	}
	return t.tex_w, t.tex_h
}

func (t *TextUI) Init(s *Scene) {
	t.scene = s
	t.reworkTex()
}

func (t *TextUI) Update(dt float32) {
}

func (t *TextUI) Draw(r *sdl.Renderer) (res RenderReqList) {
	if t.needReworkTex {
		t.reworkTex()
	}

	w := int32(float32(t.tex_w) * t.Scale)
	h := int32(float32(t.tex_h) * t.Scale)

	var dx, dy int32
	if t.FromCenter {
		dx = -w / 2
		dy = -h / 2
	}

	req := NewRenderReq(t.tex, nil, &sdl.Rect{t.X + dx, t.Y + dy, w, h}, t.Z, -float64(t.Angle), nil, sdl.FLIP_NONE)
	res = append(res, req)
	return res
}

func (t TextUI) GetID() string {
	return ""
}
