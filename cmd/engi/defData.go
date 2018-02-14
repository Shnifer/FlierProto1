//РАЗНЫЙ В РАЗНЫХ КЛИЕНТАХ ИЗ-ЗА РАЗНЫХ ПАРАМЕТРОВ
package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type tDefVals struct {
	ServerName             string
	tcpPort                string
	MIN_FRAME_MS           int
	MIN_PHYS_MS            int
	MAX_FRAME_MS           int
	MAX_PHYS_MS            int
	TickerBalancerOverhead float32
	FPS_UPDATE_MS          int
	FullScreen             bool
	WinW, WinH             int32
	RENDERER_ACCELERATED   bool
	SSDFontName            string
	SSDFontSize            int
}

var DEFVAL tDefVals

func setDefDef() {
	DEFVAL = tDefVals{
		ServerName:             "localhost",
		tcpPort:                ":6666",
		MIN_FRAME_MS:           10,
		MIN_PHYS_MS:            3,
		MAX_FRAME_MS:           30,
		MAX_PHYS_MS:            10,
		FPS_UPDATE_MS:          1000,
		TickerBalancerOverhead: 2,
		WinW:                 1024,
		WinH:                 768,
		RENDERER_ACCELERATED: true,
		SSDFontName:          "furore.otf",
		SSDFontSize:          14,
	}
}

func LoadDefVals(filepath string) {
	setDefDef()

	exfn := filepath + "example_defdata.json"
	exbuf, err := json.Marshal(DEFVAL)
	if err := ioutil.WriteFile(exfn, exbuf, 0); err != nil {
		log.Println("can't even write ", exfn)
	}

	fn := filepath + "defdata.json"

	buf, err := ioutil.ReadFile(filepath + "defdata.json")
	if err != nil {
		log.Println("cant read ", fn, "using default")
	}
	json.Unmarshal(buf, &DEFVAL)
}
