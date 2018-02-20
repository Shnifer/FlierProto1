package texture

import (
	"github.com/veandco/go-sdl2/sdl"
	"image"
	"image/jpeg"
	"image/png"
	"log"
	"os"
	"strings"
)

type AnimTex struct {
	tex *sdl.Texture
	//Количество и размер частей
	part_w, part_h int32
	num_x, num_y   int32
	totalcount     int32
}

func (a *AnimTex) TotalCount() int32 {
	return a.totalcount
}

func (a *AnimTex) GetTexAndRect(i int32) (*sdl.Texture, *sdl.Rect) {
	return a.tex, a.getRect(i)
}

func NewAnimTex(tex *sdl.Texture, num_x, num_y int) *AnimTex {
	_, _, w, h, err := tex.Query()
	if err != nil {
		log.Panicln(err)
	}
	nx, ny := int32(num_x), int32(num_y)

	return &AnimTex{
		tex:        tex,
		num_x:      nx,
		num_y:      ny,
		part_w:     w / nx,
		part_h:     h / ny,
		totalcount: nx * ny,
	}
}

func (at *AnimTex) getRect(i int32) *sdl.Rect {

	if i < 0 || i >= at.totalcount {
		log.Panicln("animtexture.getRect Out of index!", i, "of total", at.totalcount)
	}

	px := i % at.num_x
	py := i / at.num_x
	return &sdl.Rect{px * at.part_w, py * at.part_h, at.part_w, at.part_h}
}

//
//func setPixel(x, y int, c sdl.Color, pixels []byte) {
//	if x < 0 || y < 0 || x >= winW || y >= winH {
//		return
//	}
//	index := (y*winW + x) * 4
//
//	if index < len(pixels) && index >= 0 {
//		pixels[index+0] = c.r
//		pixels[index+1] = c.G
//		pixels[index+2] = c.B
//	}
//}

func PixelsToTexture(renderer *sdl.Renderer, pixels []byte, w, h int) (*sdl.Texture, error) {
	tex, err := renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STREAMING, int32(w), int32(h))
	if err != nil {
		return nil, err
	}
	tex.Update(nil, pixels, w*4)
	if err := tex.SetBlendMode(sdl.BLENDMODE_BLEND); err != nil {
		return nil, err
	}
	return tex, nil
}

func loadFileToPixels(fn string) ([]byte, int, int, error) {
	f, err := os.Open(fn)
	if err != nil {
		return nil, 0, 0, err
	}
	defer f.Close()

	var img image.Image
	switch {
	case strings.HasSuffix(fn, ".png"):
		img, err = png.Decode(f)
		if err != nil {
			return nil, 0, 0, err
		}
	case strings.HasSuffix(fn, ".jpg") ||
		strings.HasSuffix(fn, ".jpeg"):
		img, err = jpeg.Decode(f)
		if err != nil {
			return nil, 0, 0, err
		}

	default:
		log.Panicln("Unknown img format ", fn)
	}

	w := img.Bounds().Max.X
	h := img.Bounds().Max.Y

	Pixels := make([]byte, w*h*4)
	bIndex := 0
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			Pixels[bIndex+0] = byte(r / 256)
			Pixels[bIndex+1] = byte(g / 256)
			Pixels[bIndex+2] = byte(b / 256)
			Pixels[bIndex+3] = byte(a / 256)
			bIndex += 4
		}
	}
	return Pixels, w, h, nil
}
