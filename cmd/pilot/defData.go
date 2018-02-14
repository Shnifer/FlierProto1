//РАЗНЫЙ В РАЗНЫХ КЛИЕНТАХ ИЗ-ЗА РАЗНЫХ ПАРАМЕТРОВ
package main

import (
	"encoding/json"
	V2 "github.com/Shnifer/flierproto1/v2"
	"io/ioutil"
	"log"
)

type tDefVals struct {
	ServerName                    string
	tcpPort                       string
	MIN_FRAME_MS                  int
	MIN_PHYS_MS                   int
	MAX_FRAME_MS                  int
	MAX_PHYS_MS                   int
	TickerBalancerOverhead        float32
	FPS_UPDATE_MS                 int
	FullScreen                    bool
	WinW, WinH                    int32
	RENDERER_ACCELERATED          bool
	GravityConst                  float32
	GravityDepthSqr               float32
	GravityCalc3D                 bool
	StartLocationName             string
	StartLocationOffset           V2.V2
	ShowGizmoGravityForce         bool
	GizmoGravityForceK            float32
	ShowGizmoGravityRound         bool
	GizmoGravityRoundDotsInCirle  int
	GizmoGravityRoundLevels       []float32
	ShipSize                      float32
	ShipFixedSize                 int32
	ShipShowFixed                 bool
	ShipShowBotOffset             int32
	ShipThrustAxel                float32
	ShipMaxThrustForce            float32
	ShipAngAxel                   float32
	ShipMaxAngMomentum            float32
	MainEngineMaxParticles        int
	MainEngineParticlesLifetime   float32
	MainEngineParticlesRandStartK float32
	MainEngineParticlesRandSpeedK float32
	MainEngineParticlesMaxIntense float32
	StarRotationSpeed             float32
	//Обратное значение, не должно оказаться нулём!
	CameraMinScale float32
	CameraMaxScale float32
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
		WinW:                          1024,
		WinH:                          768,
		RENDERER_ACCELERATED:          true,
		GravityConst:                  0.01,
		GravityDepthSqr:               10,
		StartLocationName:             "magelan",
		ShowGizmoGravityForce:         false,
		ShowGizmoGravityRound:         false,
		GizmoGravityRoundDotsInCirle:  64,
		ShipSize:                      1,
		ShipFixedSize:                 30,
		ShipThrustAxel:                0.33,
		ShipMaxThrustForce:            100,
		ShipAngAxel:                   360,
		ShipMaxAngMomentum:            90,
		MainEngineMaxParticles:        1000,
		MainEngineParticlesLifetime:   1,
		MainEngineParticlesRandStartK: 0.2,
		MainEngineParticlesRandSpeedK: 0.2,
		MainEngineParticlesMaxIntense: 100,
		StarRotationSpeed:             30,
		CameraMaxScale:                1000,
		CameraMinScale:                1,
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
