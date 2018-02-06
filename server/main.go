package main

import (
	"bufio"
	"log"
	"net"
	"strings"
	MNT "github.com/Shnifer/flierproto1/mnt"
	"strconv"
	"math/rand"
	"time"
)

var NextID func() int

func CreateGenerator() func() int {
	i := 0
	return func() int {
		i++
		return i
	}
}

type inString struct {
	sender net.Conn
	text   string
}

//Комната символизирует корабль,
//Карта соединений по ролям
//Сами комнаты в карте по ИД

type Profile struct {
	Room string
	Role string
}

type Room struct {
	members map[string]net.Conn
}

func newRoom() Room{
	return Room{members:make(map[string]net.Conn)}
}

//ГО СИНГЛ сидит на listener.Accept()
func ConnAcepter(listener net.Listener, NewConns chan net.Conn) {
	log.Println("ConnAcepter START")
	for {
		conn, err := listener.Accept()
		if err != nil {
			break
		}
		NewConns <- conn
	}
	log.Println("ConnAcepter END")
}

//ГО 1 на ОТКРЫТОЕ СОЕДИНЕНИЕ сидит на reader.ReadString
func ListenConn(conn net.Conn, inStrs chan inString, DeadConns chan net.Conn) {
	id := NextID()
	log.Println("ListenConn #", id, "START for", conn)
	scaner := bufio.NewScanner(conn)
	for scaner.Scan() {
		str:=scaner.Text()
		inStrs <- inString{sender: conn, text: str}
	}
	DeadConns <- conn
	log.Println("ListenConn #", id, "END")
}

//ЗАПУСКАЕТ ГОРУТИНУ 1 на СОЕДИНЕНИЕ
//и возвращает канал, из которого принимает строки на отправку
//при закрытии канала выключается
func newSenderConn(conn net.Conn) chan string{
	log.Println("sender for",conn,"STARTED")
	c:=make(chan string,1)

	go func() {
		writer := bufio.NewWriter(conn)
		for msg:=range c{
			_,err:=writer.WriteString(msg+"\n")
			if err!=nil{
				log.Println("SEND to",conn,"ERR:",err)
				continue
			}
			writer.Flush()
			log.Println("sent", msg,"to",conn)
		}
	}()
	c<-MNT.RES_LOGIN
	return c
}

var	OutSender map[net.Conn]chan string

const serverDataPath = "res/server/"

//Главный тред
func main() {

	LoadDefVals(serverDataPath)

	rand.Seed(time.Now().Unix())

	if DEFVAL.LoadGalaxyFile!="" {
		LoadGalaxyFromFile()
	} else {
		GenerateRandomGalaxy()
	}

	NextID = CreateGenerator()

	NewConns := make(chan net.Conn, 1)
	DeadConns := make(chan net.Conn, 1)
//	Conns := make(map[net.Conn]bool)

	inStrs := make(chan inString, 1)

	Profiles := make(map[net.Conn]Profile)
	Rooms := make(map[string]Room)

	OutSender=make(map[net.Conn]chan string)

	listener, err := net.Listen("tcp", DEFVAL.tcpPort)
	if err != nil {
		log.Println(err)
	}
	defer listener.Close()

	go ConnAcepter(listener, NewConns)

	for {
		select {
		//Здесь для защиты доступа к Map Conns
		case conn := <-NewConns:
			//START LISTENER
			//Conns[conn] = true
			go ListenConn(conn, inStrs, DeadConns)
		case conn := <-DeadConns:
			//CLOSE & REMOVE FROM CONNS
			conn.Close()
			//delete(Conns, conn)
			prof, ok:= Profiles[conn]
			if ok {
				//CHECK OUT from ROOM
				role:=prof.Role
				room:=prof.Room
				delete(Rooms[room].members,role)
				if len(Rooms[room].members)==0 {
					//no empty rooms
					delete(Rooms,room)
				}
				//Закрываем канал рассылки и этим горутину
				close(OutSender[conn])
				delete(OutSender, conn)

				log.Println("Profile deleted for mr.",role,"from",room)
				delete(Profiles, conn)
			}

		case inMsg := <-inStrs:
			log.Println("inStrs",inMsg,"come")
			sender := inMsg.sender
			profile, ok := Profiles[sender]
			//Ещё не зарегистрирован, т.е. это должна быть строка логина
			if !ok {
				params := strings.SplitN(inMsg.text, " ", 3)
				if len(params) < 2 {
					//Хуйская строка пропускаем
					log.Println("Conn", sender, " incorrect login:", inMsg.text)
					DeadConns <- sender
					continue
				}

				reqRoom := params[0]
				reqRole := params[1]

				if room, ok := Rooms[reqRoom]; ok {
					if occ, ok := room.members[reqRole]; ok {
						//В этой комнате уже есть такая роль, считаем роли уникальными
						log.Println("Conn", sender, " FAILED login as", reqRole, "in", reqRoom, ". ALREADY OCCUPIED by", occ)
						DeadConns <- sender
						continue
					}
				}

				log.Println("Conn", sender, " SUCCESS login as", reqRole, "in", reqRoom)

				//CHECK IN
				Profiles[inMsg.sender] = Profile{Room: params[0], Role: params[1]}
				room, ok := Rooms[reqRoom];
				if !ok{
					//Если новая комната - добаляем
					room = newRoom()
					Rooms[reqRoom] = room
				}
				room.members[reqRole] = sender

				//Запускаем отправщика
				ch:=newSenderConn(sender)
				OutSender[sender] = ch
			} else {
				//Сообщение от зарегистрированного разбираем
				room:=profile.Room
				role:=profile.Role

				log.Println("in",room,"mr.",role,"says:",inMsg.text)


				params := strings.SplitN(inMsg.text, " ", 2)
				if len(params)<1 {
					continue
				}
				command:=params[0]
				param:=""
				if len(params)>1 {
					param = params[1]
				}
				log.Println("command ",command)

				HandleCommand(Rooms[room], role, command, param, OutSender[sender])
			}
		}
	}
}

func HandleCommand(Room Room, role string, command, params string, out chan string) {
	switch command {
	case MNT.CMD_BROADCAST:
		if len(params)<2 {
			break
		}
		for destRole,destConn:=range Room.members {
			msg:=params
			//не шлём себе
			if destRole!=role {
				OutSender[destConn] <- MNT.IN_MSG+" "+msg
			}
		}
	case MNT.CMD_CHECKROOM:
		out <- MNT.RES_CHECKROOM+" "+strconv.Itoa(len(Room.members))

	case MNT.CMD_GETGALAXY:
		res:= MNT.UploadGalaxy()
		for _,s:=range res {
			out <- s
		}
	default:
		log.Println("UNKNOWN COMMAND")
		out <- MNT.ERR_UNKNOWNCMD
	}
}