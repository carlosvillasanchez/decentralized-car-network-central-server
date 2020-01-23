package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"encoding/json"
)

type CentralServer struct {
	Cars 			map[string]Car
	Buildings 		[]Building
	ParkingSpots 	[]ParkingSpot
	CarCrashes 		[]CarCrash
	Map				[9][9]string
}

type Car struct {
	Id 				string
	IP 				string
	Port 			string
	X 				int
	Y 				int
	DestinationX 	int
	DestinationY  	int
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

func (centralServer *CentralServer) updateAPI(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		toSend := make(map[string][]int)
		for k, v := range centralServer.Cars {
			v.X = (v.X +1)%9
			centralServer.Cars[k] = v
			toSend[v.Id] = []int{v.X, v.Y, v.DestinationX, v.DestinationY}
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
	var centralServer CentralServer 
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