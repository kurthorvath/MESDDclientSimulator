package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/paulmach/orb"
	"gopkg.in/ini.v1"
)

type ClientConfig []struct {
	ServiceInterval   int    `json:"service_interval"`
	DiscoveryInterval int    `json:"discovery_interval"`
	Startlat          string `json:"startlat"`
	Startlon          string `json:"startlon"`
	Speed             int    `json:"speed"`
	Bearing           int    `json:"bearing"`
}

type Location struct {
	Lat             float64
	Lon             float64
	LocationDesc    string
	L1              string
	L2              string
	L3              string
	L4              string
	Speed           int
	Bearing         int
	ServiceInterval int
}

type client struct {
	Id                int      `json:"Id"`
	Status            string   `json:"Status"`
	Terminate         bool     `json:"Terminate"`
	DoneInit          bool     `json:"DoneInit"`
	UpdateInterval    int      `json:"UpdateInterval"`
	DiscoveryInterval int      `json:"DiscoveryInterval"`
	Loc               Location `json:"Location"`
	BaseURL           string   `json:"BaseURL"`
}

type configClient struct {
	id                int
	startLat          float64
	startLon          float64
	speed             int
	bearing           int
	ServiceInterval   int
	DiscoveryInterval int
}

type mainConfig struct {
	pathClients        string
	Clients            []configClient
	service_interval   int
	discovery_interval int
}

var CONF mainConfig

func TurnOnKthBit(n, k int) int {
	return n | (1 << (k))
}

func (c *client) inWhichZonesIsUserLocated() bool {
	//send location to API to retrieve zones
	resp, err := http.Get("http://neverssl.com")

	if err != nil {
		log.Printf("Request Failed: %s", err)
		return false
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Reading body failed: %s", err)
		return false
	}
	// workaround until final consul integration
	body = []byte(`["uni.waidmannsdorf.klagenfurt.austria","waidmannsdorf.klagenfurt.austria","klagenfurt.austria", "austria"]`)
	var arr []string
	_ = json.Unmarshal(body, &arr)
	log.Printf("Unmarshaled: %v", arr)

	for _, element := range arr {
		switch strings.Count(element, ".") {
		case 1:
			log.Println(c.Id, "L1")
			c.Loc.L1 = element
		case 2:
			log.Println(c.Id, "L2")
			c.Loc.L2 = element
		case 3:
			log.Println(c.Id, "L3")
			c.Loc.L3 = element
		case 4:
			log.Println(c.Id, "L4")
			c.Loc.L4 = element
		}
	}

	if c.Loc.L1 != "" {
		c.Loc.LocationDesc = c.Loc.L1
	}
	if c.Loc.L2 != "" {
		c.Loc.LocationDesc = c.Loc.L2
	}
	if c.Loc.L3 != "" {
		c.Loc.LocationDesc = c.Loc.L3
	}
	if c.Loc.L4 != "" {
		c.Loc.LocationDesc = c.Loc.L4
	}
	log.Println(c.Id, "LD", c.Loc)
	return true
}

func (c *client) areLocationDescriptorsValid() bool {
	//todo
	return true
}

func (c *client) downloadTargetApplication() bool {
	var URL string
	URL = c.Loc.LocationDesc + "." + c.BaseURL
	log.Println(c.Id, "download ..."+URL)
	return true
}

func (c *client) discoveryProcess() bool {
	//lookup geofence based on location
	var ret bool
	ret = c.inWhichZonesIsUserLocated()

	//validate location descriptors
	ret = c.areLocationDescriptorsValid()

	//download target app from edge server
	ret = c.downloadTargetApplication()

	if ret == false {
		return false
	}
	return true

}
func updateClientbyId(c *client) {
	for ind, elem := range arrClients {
		if elem.Id == c.Id {
			arrClients[ind] = *c
		}
	}
}

func (c *client) process() {

	for c.DoneInit != true {
		log.Println("starting discovery for ", c.Id)
		c.DoneInit = c.discoveryProcess()
	}

	log.Println("time-config for ", c.Id, " updateInterval:", c.UpdateInterval, " discoveryInterval:", c.DiscoveryInterval, c.Loc.Bearing)

	// main timer causing full discovery
	mainticker := time.NewTicker(time.Duration(c.DiscoveryInterval) * time.Second)
	doneMain := make(chan bool)
	go func() {
		for {
			select {
			case <-doneMain:
				return
			case t := <-mainticker.C:
				log.Println(c.Id, "FULL Discovery Tick at", t, "", c.Loc.Lat, c.Loc.Lon)
				c.DoneInit = false
				c.DoneInit = c.discoveryProcess()
			}
		}
	}()

	// update timer, while remaining on same service instance
	ticker := time.NewTicker(time.Duration(c.UpdateInterval) * time.Second)
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-done:
				return
			case t := <-ticker.C:
				c.Loc.Lat, c.Loc.Lon = newPosition(c.Loc.Lat, c.Loc.Lon, c.Loc.Bearing, c.Loc.Speed, c.Loc.ServiceInterval)
				updateClientbyId(c)

				log.Println(c.Id, "Tick at", t, "", c.Loc.Lat, c.Loc.Lon)
				c.downloadTargetApplication()
			}
		}
	}()

	for c.Terminate != true {
		time.Sleep(200 * time.Millisecond)
	}

	ticker.Stop()
	done <- true

}

func (c *client) Start() {
	c.Terminate = false
	go c.process()
}

func (c *client) Stop() {
	c.Terminate = true
}

var menuItems = []string{"List clients", "Add client", "Delete client", "Status", "Add Client from list"}
var arrClients = []client{}

func delete_at_index(slice []client, index int) []client {
	return append(slice[:index], slice[index+1:]...)
}

func listClientHandler() {
	fmt.Println(len(arrClients), arrClients)
}

func addClientHandlers() {
	for ind, _ := range CONF.Clients {
		addClientHandler(ind)
	}
}

func addClientHandler(IND int) {
	var INDEX int
	INDEX = len(arrClients)
	LOC := Location{CONF.Clients[IND].startLat, CONF.Clients[IND].startLon, "", "", "", "", "", CONF.Clients[IND].speed, CONF.Clients[IND].bearing, CONF.Clients[IND].ServiceInterval}
	fmt.Println(LOC)
	arrClients = append(arrClients, client{INDEX, "test", false, false, CONF.Clients[IND].ServiceInterval, CONF.Clients[IND].DiscoveryInterval, LOC, "app.service.consul"})
	arrClients[len(arrClients)-1].Start()
}

func deleteClientHandler() {
	var INDEX int
	fmt.Scanf("%d", &INDEX)
	c := arrClients[INDEX]
	c.Terminate = true
	arrClients = delete_at_index(arrClients, INDEX)
}

func statusClientHandler() {
	var INDEX int
	fmt.Scanf("%d", &INDEX)
	fmt.Println("Client is", arrClients[INDEX])

}

func eval(selected string) {
	fmt.Println(selected, menuItems[1])

	switch selected {
	case menuItems[0]: //list
		fmt.Println(menuItems[0])
		listClientHandler()
	case menuItems[1]: //add
		fmt.Println(menuItems[1])
		addClientHandler(0)
	case menuItems[2]: //delete
		fmt.Println(menuItems[2])
		deleteClientHandler()
	case menuItems[3]: //status
		fmt.Println(menuItems[3])
		statusClientHandler()
	case menuItems[4]: //add from list
		fmt.Println(menuItems[4])
		addClientHandlers()
	default:
		fmt.Println("invalid command")

	}
}

func deg2rad(d float64) float64 {
	return d * math.Pi / 180.0
}

func rad2deg(r float64) float64 {
	return 180.0 * r / math.Pi
}

func PointAtBearingAndDistance(p orb.Point, bearing, distance float64) orb.Point {
	aLat := deg2rad(p[1])
	aLon := deg2rad(p[0])

	bearingRadians := deg2rad(bearing)

	distanceRatio := distance / orb.EarthRadius
	bLat := math.Asin(math.Sin(aLat)*math.Cos(distanceRatio) + math.Cos(aLat)*math.Sin(distanceRatio)*math.Cos(bearingRadians))
	bLon := aLon +
		math.Atan2(
			math.Sin(bearingRadians)*math.Sin(distanceRatio)*math.Cos(aLat),
			math.Cos(distanceRatio)-math.Sin(aLat)*math.Sin(bLat),
		)

	return orb.Point{rad2deg(bLon), rad2deg(bLat)}
}

func newPosition(slat float64, slon float64, bearing int, speed int, interval int) (lat float64, lon float64) {
	dist := float64(speed) * float64(interval)
	p1 := PointAtBearingAndDistance(orb.Point{slat, slon}, float64(bearing), dist)
	return p1.X(), p1.Y()
}

func readConfig() {
	inidata, err := ini.Load("config.ini")
	if err != nil {
		fmt.Printf("Fail to read file: %v", err)
		os.Exit(1)
	}
	section := inidata.Section("defaults")

	CONF.pathClients = section.Key("listclients").String()

	jsonFile, _ := os.Open(CONF.pathClients)
	byteValue, _ := ioutil.ReadAll(jsonFile)
	var cc ClientConfig
	json.Unmarshal(byteValue, &cc)

	CONF.service_interval, _ = section.Key("service_interval").Int()
	CONF.discovery_interval, _ = section.Key("discovery_interval").Int()

	//currently only one item
	slat, _ := section.Key("startlat").Float64()
	slon, _ := section.Key("startlon").Float64()
	speed, _ := section.Key("speed").Int()
	bearing, _ := section.Key("bearing").Int()

	if len(cc) > 0 {
		for ind, client := range cc {
			lat, _ := strconv.ParseFloat(client.Startlat, 64)
			lon, _ := strconv.ParseFloat(client.Startlon, 64)
			CONF.Clients = append(CONF.Clients, configClient{ind, lat, lon, client.Speed, client.Bearing, client.ServiceInterval, client.DiscoveryInterval})
		}
	} else {
		CONF.Clients = append(CONF.Clients, configClient{1, slat, slon, speed, bearing, CONF.service_interval, CONF.discovery_interval})
	}
}

func initLogger() {
	fileName := "simulator.log"
	// open log file
	logFile, err := os.OpenFile(fileName, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Panic(err)
	}

	log.SetOutput(logFile)
	log.SetFlags(log.Lshortfile | log.LstdFlags)

}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "map.html")
}

func getpos(w http.ResponseWriter, req *http.Request) {
	enableCors(&w)
	buffer, err := json.Marshal(arrClients)
	if err != nil {
		fmt.Printf("error marshaling JSON: %v\n", err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(buffer)
}

func httpServer() {
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/getpos", getpos)
	http.ListenAndServe(":8080", nil)
}

func main() {
	readConfig()
	initLogger()

	prompt := promptui.Select{
		Label: "Select Action",
		Items: menuItems,
	}

	keepRunning := true

	go httpServer()

	for keepRunning {
		_, result, err := prompt.Run()
		if err != nil {
			log.Printf("Prompt failed %v\n", err)
			return
		}

		eval(result)
	}

}
