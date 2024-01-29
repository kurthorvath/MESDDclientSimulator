package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/paulmach/orb"
	"gopkg.in/ini.v1"
)

type ClientConfig []struct {
	ServiceInterval   string `json:"service_interval"`
	DiscoveryInterval string `json:"discovery_interval"`
	Startlat          string `json:"startlat"`
	Startlon          string `json:"startlon"`
	Speed             string `json:"speed"`
	Bearing           string `json:"bearing"`
}

type Location struct {
	Lat          float64
	Lon          float64
	LocationDesc string
	L1           string
	L2           string
	L3           string
	L4           string
}

type client struct {
	id             int
	status         string
	terminate      bool
	DoneInit       bool
	updateInterval int
	loc            Location
	baseURL        string
}

type configClient struct {
	id       int
	startLat float64
	startLon float64
	speed    int
	bearing  int
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
			log.Println(c.id, "L1")
			c.loc.L1 = element
		case 2:
			log.Println(c.id, "L2")
			c.loc.L2 = element
		case 3:
			log.Println(c.id, "L3")
			c.loc.L3 = element
		case 4:
			log.Println(c.id, "L4")
			c.loc.L4 = element
		}
	}

	if c.loc.L1 != "" {
		c.loc.LocationDesc = c.loc.L1
	}
	if c.loc.L2 != "" {
		c.loc.LocationDesc = c.loc.L2
	}
	if c.loc.L3 != "" {
		c.loc.LocationDesc = c.loc.L3
	}
	if c.loc.L4 != "" {
		c.loc.LocationDesc = c.loc.L4
	}
	log.Println(c.id, "LD", c.loc)
	return true
}

func (c *client) areLocationDescriptorsValid() bool {
	//todo
	return true
}

func (c *client) downloadTargetApplication() bool {
	var URL string
	URL = c.loc.LocationDesc + "." + c.baseURL
	log.Println(c.id, "download ..."+URL)
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

func (c *client) process() {

	for c.DoneInit != true {
		log.Println("starting discovery for ", c.id)
		c.DoneInit = c.discoveryProcess()
	}

	log.Println("update", c.id, c.updateInterval)

	// main timer causing full discovery
	mainticker := time.NewTicker(time.Duration(CONF.discovery_interval) * time.Second)
	doneMain := make(chan bool)
	go func() {
		for {
			select {
			case <-doneMain:
				return
			case t := <-mainticker.C:
				log.Println(c.id, "FULL Discovery Tick at", t, "", c.loc.Lat, c.loc.Lon)
				c.DoneInit = false
				c.DoneInit = c.discoveryProcess()
			}
		}
	}()

	// update timer, while remaining on same service instance
	ticker := time.NewTicker(time.Duration(c.updateInterval) * time.Second)
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-done:
				return
			case t := <-ticker.C:
				c.loc.Lat, c.loc.Lon = newPosition(c.loc.Lat, c.loc.Lon, 90, 5, 200)
				log.Println(c.id, "Tick at", t, "", c.loc.Lat, c.loc.Lon)
				c.downloadTargetApplication()
			}
		}
	}()

	for c.terminate != true {
		time.Sleep(200 * time.Millisecond)
	}

	ticker.Stop()
	done <- true

}

func (c *client) Start() {
	c.terminate = false
	go c.process()
}

func (c *client) Stop() {
	c.terminate = true
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
	LOC := Location{CONF.Clients[IND].startLat, CONF.Clients[IND].startLon, "", "", "", "", ""}
	arrClients = append(arrClients, client{INDEX, "test", false, false, CONF.service_interval, LOC, "app.service.consul"})
	arrClients[len(arrClients)-1].Start()
}

func deleteClientHandler() {
	var INDEX int
	fmt.Scanf("%d", &INDEX)
	c := arrClients[INDEX]
	c.terminate = true
	arrClients = delete_at_index(arrClients, INDEX)
}

func statusClientHandler() {
	var INDEX int
	fmt.Scanf("%d", &INDEX)
	fmt.Println("Client is", arrClients[INDEX])

}

func addfromListHandler() {

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

func newPosition(slat float64, slon float64, bearing float64, speed float64, interval int) (lat float64, lon float64) {
	dist := speed * float64(interval)
	p1 := PointAtBearingAndDistance(orb.Point{slat, slon}, bearing, dist)
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

	CONF.service_interval, _ = section.Key("service_interval").Int()
	CONF.discovery_interval, _ = section.Key("discovery_interval").Int()

	//currently only one item
	slat, _ := section.Key("startlat").Float64()
	slon, _ := section.Key("startlon").Float64()
	speed, _ := section.Key("speed").Int()
	bearing, _ := section.Key("bearing").Int()

	CONF.Clients = append(CONF.Clients, configClient{1, slat, slon, speed, bearing})
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

func main() {
	readConfig()
	initLogger()

	prompt := promptui.Select{
		Label: "Select Action",
		Items: menuItems,
	}

	keepRunning := true

	for keepRunning {
		_, result, err := prompt.Run()
		if err != nil {
			log.Printf("Prompt failed %v\n", err)
			return
		}

		eval(result)
	}

}
