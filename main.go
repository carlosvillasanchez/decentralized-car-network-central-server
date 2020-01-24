package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"math/rand"
	"encoding/json"
	"net"
	"github.com/dedis/protobuf"
)

type CentralServer struct {
	Cars 			map[string]Car
	Buildings 		[]Building
	ParkingSpots 	[]ParkingSpot
	CarCrashes 		[]CarCrash
	Map				[9][9]string
	conn 			*net.UDPConn
}

type Car struct {
	Id 				string
	IP 				string
	Port 			string
	X 				int
	Y 				int
	DestinationX 	int
	DestinationY  	int
	Messages 		[]MessageTrace
}

type Building struct {
	Id string
	X int
	Y int
}

type CarCrash struct {
	Id string
	X int
	Y int
}

type ParkingSpot struct {
	Id string
	X int
	Y int
}

type ServerNodeMessage struct {
	Position *Position
	Trace    *MessageTrace
}
type MessageTrace struct {
	Type string
	Text string
}
type Position struct { // TODO will probably be defined elsewhere
	X uint32
	Y uint32
}

type ServerMessage struct {
	Type string
}
	

func idAPI(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		w.Write([]byte("This is an example server.\n"))
	}
	
}

func (centralServer *CentralServer) setupAPI(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		fmt.Println("")
		if err := r.ParseForm(); err != nil {
            fmt.Fprintf(w, "ParseForm() err: %v", err)
            return
		}
		cars := r.FormValue("cars")
		buildings := r.FormValue("buildings")
		carcrashes := r.FormValue("carcrashes")
		parkingspots := r.FormValue("parkingspots")
		fmt.Println("Car", cars)
		fmt.Println("Buildings", buildings, buildings == "")
		fmt.Println("Carcrash", carcrashes)
		fmt.Println("ParkingSpot", parkingspots)
		centralServer.setupCentralServer(cars, buildings, carcrashes, parkingspots)
		w.Write([]byte("Everything ok.\n"))
		go centralServer.startProtocol()
	}	
}

func (centralServer *CentralServer) addCarAPI(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		fmt.Println("")
		if err := r.ParseForm(); err != nil {
            fmt.Fprintf(w, "ParseForm() err: %v", err)
            return
		}
		//[TODO: Check there is no car there]
		car := r.FormValue("car")
		fmt.Println("Car", car)
		carSplited := strings.Split(car, ",")
		x, _ := strconv.Atoi(carSplited[3])
		y, _ := strconv.Atoi(carSplited[4])
		destinationX, _ := strconv.Atoi(carSplited[5])
		destinationY, _ := strconv.Atoi(carSplited[6])
		newCar := Car{
			Id: carSplited[0],
			IP: carSplited[1],
			Port: carSplited[2],
			X: x,
			Y: y,
			DestinationX: destinationX,
			DestinationY: destinationY,
		}
		centralServer.Cars[carSplited[1] + ":" + carSplited[2]] = newCar
		centralServer.printMap()
		w.Write([]byte("Everything ok.\n"))
	}
}

func (centralServer *CentralServer) addParkingSpotAPI(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		fmt.Println("")
		if err := r.ParseForm(); err != nil {
            fmt.Fprintf(w, "ParseForm() err: %v", err)
            return
		}
		//[TODO: Check there it can be there]
		parkingSpot := r.FormValue("parkingSpot")
		fmt.Println("ParkingSpot", parkingSpot)
		parkingSpotSplitted := strings.Split(parkingSpot, ",")
		x, _ := strconv.Atoi(parkingSpotSplitted[1])
		y, _ := strconv.Atoi(parkingSpotSplitted[2])
		newParkingSpot := ParkingSpot{
			Id: parkingSpotSplitted[0],
			X: x,
			Y: y,
		}
		centralServer.ParkingSpots = append(centralServer.ParkingSpots, newParkingSpot)
		centralServer.Map[x][y] = "p"
		centralServer.printMap()
		w.Write([]byte("Everything ok.\n"))
	}
}


func (centralServer *CentralServer) addCarCrashAPI(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		fmt.Println("")
		if err := r.ParseForm(); err != nil {
            fmt.Fprintf(w, "ParseForm() err: %v", err)
            return
		}
		//[TODO: Check there it can be there]
		carCrash := r.FormValue("carCrash")
		fmt.Println("CarCrash", carCrash)
		carCrashSplitted := strings.Split(carCrash, ",")
		x, _ := strconv.Atoi(carCrashSplitted[1])
		y, _ := strconv.Atoi(carCrashSplitted[2])
		newCarCrash := CarCrash{
			Id: carCrashSplitted[0],
			X: x,
			Y: y,
		}
		centralServer.CarCrashes = append(centralServer.CarCrashes, newCarCrash)
		centralServer.Map[x][y] = "cc"
		centralServer.printMap()
		w.Write([]byte("Everything ok.\n"))
	}
}

type UpdateUI struct {
	Pos map[string][]int
	Messages map[string][]MessageTrace
}

func (centralServer *CentralServer) updateAPI(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		toSendPos := make(map[string][]int)
		toSendMessages := make(map[string][]MessageTrace)
		for k, v := range centralServer.Cars {
			v.X = (v.X +1)%9
			types := []string{"parking", "crash", "police", "other"}
			newMessage := MessageTrace{
				Type: types[rand.Intn(len(types))],
				Text: strconv.Itoa(len(v.Messages)),
			}
			v.Messages = append(v.Messages, newMessage)
			toSendPos[v.Id] = []int{v.X, v.Y, v.DestinationX, v.DestinationY}
			toSendMessages[v.Id] = v.Messages
			v.Messages = []MessageTrace{}
			centralServer.Cars[k] = v
		}
		toSend := UpdateUI{
			Pos: toSendPos,
			Messages: toSendMessages,
		}
		js, err := json.Marshal(toSend)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	}
}

func main() {
	
	udpAddr, _ := net.ResolveUDPAddr("udp4", "127.0.0.1:5999")
	udpConn, _ := net.ListenUDP("udp4", udpAddr)
	centralServer := CentralServer{
		conn: udpConn,
	}
	http.HandleFunc("/id", idAPI)
	http.HandleFunc("/setup", centralServer.setupAPI)
	http.HandleFunc("/addCarCrash", centralServer.addCarCrashAPI)
	http.HandleFunc("/addParkingSpot", centralServer.addParkingSpotAPI)
	http.HandleFunc("/addCar", centralServer.addCarAPI)
	http.HandleFunc("/update", centralServer.updateAPI)
	http.ListenAndServe(":" + strconv.Itoa(8086), nil)
}

func (centralServer *CentralServer) setupCentralServer(cars string, buildings string, carCrashes string, parkingSpots string) {
	if cars != ""{
		carsSplited := strings.Split(cars, ",")
		carsDict := make(map[string]Car)
		for i := 0; i < len(carsSplited)/7; i++ {
			x, _ := strconv.Atoi(carsSplited[i*7+3])
			y, _ := strconv.Atoi(carsSplited[i*7+4])
			destinationX, _ := strconv.Atoi(carsSplited[i*7+5])
			destinationY, _ := strconv.Atoi(carsSplited[i*7+6])
			newCar := Car{
				Id: carsSplited[i*7],
				IP: carsSplited[i*7+1],
				Port: carsSplited[i*7+2],
				X: x,
				Y: y,
				DestinationX: destinationX,
				DestinationY: destinationY,
			}
			carsDict[newCar.IP + ":" + newCar.Port] = newCar
		}
		centralServer.Cars = carsDict
	}
	if buildings != ""{
		buildingsSplited := strings.Split(buildings, ",")
		var buildingsArray []Building
		for i := 0; i < len(buildingsSplited)/3; i++ {
			x, _ := strconv.Atoi(buildingsSplited[i*3+1])
			y, _ := strconv.Atoi(buildingsSplited[i*3+2])
			newBuilding := Building{
				Id: buildingsSplited[i*3],
				X: x,
				Y: y,
			}
			buildingsArray = append(buildingsArray, newBuilding)
		}
		centralServer.Buildings = buildingsArray
	}
	if carCrashes != ""{
		carCrashesSplited := strings.Split(carCrashes, ",")
		var carCrashesArray []CarCrash
		for i := 0; i < len(carCrashesSplited)/3; i++ {
			x, _ := strconv.Atoi(carCrashesSplited[i*3+1])
			y, _ := strconv.Atoi(carCrashesSplited[i*3+2])
			newCarCrash := CarCrash{
				Id: carCrashesSplited[i*3],
				X: x,
				Y: y,
			}
			carCrashesArray = append(carCrashesArray, newCarCrash)
		}
		centralServer.CarCrashes = carCrashesArray
	}
	if parkingSpots != ""{
		parkingSpotsSplited := strings.Split(parkingSpots, ",")
		var parkingSpotsArray []ParkingSpot
		for i := 0; i < len(parkingSpotsSplited)/3; i++ {
			x, _ := strconv.Atoi(parkingSpotsSplited[i*3+1])
			y, _ := strconv.Atoi(parkingSpotsSplited[i*3+2])
			newParkingSpot := ParkingSpot{
				Id: parkingSpotsSplited[i*3],
				X: x,
				Y: y,
			}
			parkingSpotsArray = append(parkingSpotsArray, newParkingSpot)
		}
		centralServer.ParkingSpots = parkingSpotsArray
	}
	fmt.Println(centralServer)
}

func (centralServer *CentralServer) startProtocol(){
	centralServer.mapAddBuildings()
	centralServer.startNodes()
	centralServer.mapAddCarCrashes()
	centralServer.mapAddParkingSpots()
}

func (centralServer *CentralServer) mapAddBuildings(){
	for _, building := range centralServer.Buildings {
		centralServer.Map[building.X][building.Y] = "b"
	}
	centralServer.printMap()
}

func (centralServer *CentralServer) mapAddCarCrashes(){
	for _, carCrash := range centralServer.CarCrashes {
		centralServer.Map[carCrash.X][carCrash.Y] = "cc"
	}
	centralServer.printMap()
}

func (centralServer *CentralServer) mapAddParkingSpots(){
	for _, parkingSpot := range centralServer.ParkingSpots {
		centralServer.Map[parkingSpot.X][parkingSpot.Y] = "p"
	}
	centralServer.printMap()
}

func (centralServer *CentralServer) startNodes(){
	
}

func (centralServer *CentralServer) printMap(){
	fmt.Println("MAP", centralServer.Map)
}


func (centralServer *CentralServer) readNodes(){
	for {
		buffer := make([]byte, 9000)
		n, addr, _ := centralServer.conn.ReadFromUDP(buffer)
		// Dcoding the Packet
		packet := &ServerNodeMessage{}
		err := protobuf.Decode(buffer, packet)
		if(err != nil){
			//fmt.Printf("Error dcoding from peer, reason: %s\n", err)
		}
		protobuf.Decode(buffer[0:n], packet)
		addrString := addr.String()
		if(packet.Position != nil){
			fmt.Println("IT IS A POSITION FROM", addrString, "NAME", centralServer.Cars[addrString].Id, ": X", packet.Position.X, "Y", packet.Position.Y)
			c, ok := centralServer.Cars[addrString]
			if ok {
				c.X = int(packet.Position.X)
				c.Y = int(packet.Position.Y)
			}
			centralServer.Cars[addrString] = c
		}else if(packet.Trace != nil){
			fmt.Println("IT IS A TRACE FROM", addrString, "NAME", centralServer.Cars[addrString].Id, "TYPE", packet.Trace.Type, "TEXT", packet.Trace.Text)
			c, ok := centralServer.Cars[addrString]
			if ok {
				c.Messages = append(c.Messages, *packet.Trace)
			}
		}
	}
	
}