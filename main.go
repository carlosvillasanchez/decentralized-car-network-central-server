package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	//"math/rand"
	//"encoding/json"
	"net"
	"github.com/dedis/protobuf"
	"sync"
)

const (
    Parking = "parking" // For traces regarding parking spots
    Crash =   "crash"   // For traces regarding the avoidance of crashes
    Police =  "police"  // For traces regarding crash handling
    Other =   "other"   // All rest of traces
)

// STRUCTURES: 

type CentralServer struct {
	Cars 			map[string]Car
	carsMutex		sync.RWMutex
	Buildings 		[]Building
	ParkingSpots 	[]ParkingSpot
	CarCrashes 		[]CarCrash
	Map				[9][9]string
	mapMutex		sync.RWMutex
	conn 			*net.UDPConn
	Police  		bool
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
	
// MAIN FUNCTION. Starting the server.
func main() {
	
	udpAddr, _ := net.ResolveUDPAddr("udp4", "127.0.0.1:5999")
	udpConn, _ := net.ListenUDP("udp4", udpAddr)
	centralServer := CentralServer{
		conn: udpConn,
		Police: true,
	}
	// ENDPOINTS
	http.HandleFunc("/setup", centralServer.setupAPI)
	http.HandleFunc("/addCarCrash", centralServer.addCarCrashAPI)
	http.HandleFunc("/addParkingSpot", centralServer.addParkingSpotAPI)
	http.HandleFunc("/addCar", centralServer.addCarAPI)
	http.HandleFunc("/update", centralServer.updateAPI)
	// Web socket
	http.ListenAndServe(":" + strconv.Itoa(8086), nil)
}

/***
* Writing all the information for setup.
***/
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
		centralServer.carsMutex.Lock()
		centralServer.Cars = carsDict
		centralServer.carsMutex.Unlock()
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
	centralServer.printMap()
}

/***
* Disigning the map and launching the nodes
***/
func (centralServer *CentralServer) startProtocol(){
	centralServer.mapAddBuildings()
	centralServer.startNodes()
	centralServer.mapAddCarCrashes()
	centralServer.mapAddParkingSpots()
}

/***
* Add buildings
***/
func (centralServer *CentralServer) mapAddBuildings(){
	centralServer.mapMutex.Lock()
	for _, building := range centralServer.Buildings {
		centralServer.Map[building.X][building.Y] = "b"
	}
	centralServer.mapMutex.Unlock()
	centralServer.printMap()
}

/***
* Add car crashes
***/
func (centralServer *CentralServer) mapAddCarCrashes(){
	centralServer.mapMutex.Lock()
	for _, carCrash := range centralServer.CarCrashes {
		centralServer.Map[carCrash.X][carCrash.Y] = "cc"
	}
	centralServer.mapMutex.Unlock()
	centralServer.printMap()
}

/***
* Add parking spots
***/
func (centralServer *CentralServer) mapAddParkingSpots(){
	centralServer.mapMutex.Lock()
	for _, parkingSpot := range centralServer.ParkingSpots {
		centralServer.Map[parkingSpot.X][parkingSpot.Y] = "p"
	}
	centralServer.mapMutex.Unlock()
	centralServer.printMap()
}

/***
* Launching nodes
***/
func (centralServer *CentralServer) startNodes(){
	
}

/*** 
* Socket for listening to nodes
***/
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
			centralServer.carsMutex.RLock()
			fmt.Println("IT IS A POSITION FROM", addrString, "NAME", centralServer.Cars[addrString].Id, ": X", packet.Position.X, "Y", packet.Position.Y)
			c, ok := centralServer.Cars[addrString]
			centralServer.carsMutex.RUnlock()
			if ok {
				centralServer.carsMutex.Lock()
				c.X = int(packet.Position.X)
				c.Y = int(packet.Position.Y)
				centralServer.Cars[addrString] = c
				if c.Id == "police" {
					if c.X != -1 || c.Y != -1 {
						if centralServer.Police {
							fmt.Println("Police out of the station!")
						}
						centralServer.Police = false
					}else{
						centralServer.Police = true
						fmt.Println("Police available again!")
					}
				}
				centralServer.carsMutex.Unlock()
				centralServer.mapMutex.RLock()
				if centralServer.Map[c.X][c.Y] == "p" {
					centralServer.sendNode(Parking, addrString)
				} else if centralServer.Map[c.X][c.Y] == "cc" {
					if centralServer.Police {
						centralServer.sendNode(Crash, addrString)
					}
				}
				centralServer.mapMutex.RUnlock()
			}
		}else if(packet.Trace != nil){
			centralServer.carsMutex.Lock()
			fmt.Println("IT IS A TRACE FROM", addrString, "NAME", centralServer.Cars[addrString].Id, "TYPE", packet.Trace.Type, "TEXT", packet.Trace.Text)
			c, ok := centralServer.Cars[addrString]
			if ok {
				c.Messages = append(c.Messages, *packet.Trace)
			}
			centralServer.carsMutex.Unlock()
		}
	}
	
}

/***
* Sending an event to a nodes
***/
func (centralServer *CentralServer) sendNode(text string, addr string){
	fmt.Println("Infroming of a", text, "to", addr)
	// Sending the packet
	addrs_receive, _ := net.ResolveUDPAddr("udp", addr)
	packet := &ServerMessage{Type: text}
	packetBytes, _ := protobuf.Encode(packet)
	centralServer.conn.WriteToUDP(packetBytes, addrs_receive)
}

/***
* Printing map
***/
func (centralServer *CentralServer) printMap(){
	centralServer.mapMutex.RLock()
	fmt.Println("MAP", centralServer.Map)
	centralServer.mapMutex.RUnlock()
}