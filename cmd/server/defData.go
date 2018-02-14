package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type tDefVals struct {
	LoadGalaxyFile  string //если не пустой -- загружаем из файла, иначе генерим рандом
	LoadBSPFile     string //если не пустой -- загружаем из файла, иначе генерим рандом
	GalaxyRadius    float32
	NumStars        int
	Gen_minRadius   float32
	Gen_maxRadius   float32
	Gen_maxSpeed    float32
	Gen_maxMass     float32
	Gen_minMass     float32
	Load_K_Radius   float32
	Load_K_OrbSpeed float32
	Load_DefTexName string
	tcpPort         string
}

var DEFVAL tDefVals

func setDefDef() {
	DEFVAL = tDefVals{
		GalaxyRadius:    10000,
		NumStars:        100,
		Gen_minRadius:   20,
		Gen_maxRadius:   60,
		Gen_maxSpeed:    0,
		Gen_maxMass:     600,
		Gen_minMass:     100,
		Load_K_Radius:   1,
		Load_K_OrbSpeed: 1,
		Load_DefTexName: "planet.png",
		tcpPort:         ":6666",
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
