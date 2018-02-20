package mnt

//Команды обмена Клиент-Сервер
const (
	CMD_LOGIN        = "login"
	RES_LOGIN        = "LoginAccepted"
	ERR_UNKNOWNCMD   = "unknownCmd"
	CMD_BROADCAST    = "broadcast"
	IN_MSG           = "msg"
	CMD_GETGALAXY    = "getGalaxy"
	RES_GALAXY       = "galaxy"
	CMD_READYFORCHAT = "readyForChat"
	CMD_STOPCHAT     = "stopChat"
	RDY_BSP			= "rdyBSP"
	CMD_GETBSP       = "getBSP"
	RES_BSP          = "BSP"
	SET_SHIPID       = "setShipID"
	//	CMD_CHECKROOM    = "checkRoom"
	//	RES_CHECKROOM    = "resCheckRoom"
)

//Коменды обмена клиент-клиент (для сервера IN_MSG)
const (
	SHIP_POS         = "shipPos"
	SESSION_TIME     = "sessionTime"
	UPD_SSS          = "updSSS"
)

const (
	ROLE_PILOT     = "pilot"
	ROLE_ENGINEER  = "engineer"
	ROLE_NAVIGATOR = "navigator"
	ROLE_CARGO     = "cargo"
)

const (
	STATE_NOSHIP = "stateNoShip"
	STATE_LANDED = "stateLanded"
	STATE_COSMOS = "stateCosmos"
	STATE_GALAXY = "stateGalaxy"
	STATE_DIED   = "stateDied"
)