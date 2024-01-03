package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/manifoldco/promptui"
)

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
	updateInterval time.Duration
	loc            Location
	baseURL        string
}

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
	body = []byte(`["uni.waidmannsdorf.klagenfurt.austria","waidmannsdorf.klagenfurt.austria","klagenfurt.austria", "austria"]`)
	var arr []string
	_ = json.Unmarshal(body, &arr)
	log.Printf("Unmarshaled: %v", arr)
	fmt.Println(arr)

	for _, element := range arr {
		switch strings.Count(element, ".") {
		case 1:
			fmt.Println("L1")
			c.loc.L1 = element
		case 2:
			fmt.Println("L2")
			c.loc.L2 = element
		case 3:
			fmt.Println("L3")
			c.loc.L3 = element
		case 4:
			fmt.Println("L4")
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
	fmt.Println("LD", c.loc)
	return true
}

func (c *client) areLocationDescriptorsValid() bool {
	//todo
	return true
}

func (c *client) downloadTargetApplication() bool {
	var URL string
	URL = c.loc.LocationDesc + "." + c.baseURL
	fmt.Println("download ..." + URL)
	return true
}

func (c *client) discoveryProcess() bool {
	//lookup geofence based on location
	var ret bool
	ret = c.inWhichZonesIsUserLocated()
	//if ret == false
	//return ret

	//validate location descriptors
	ret = c.areLocationDescriptorsValid()
	//if ret == false
	//	return ret

	//download target app from edge server
	ret = c.downloadTargetApplication()

	if ret == false {
		return false
	}
	return true

}

func (c *client) process() {

	for c.DoneInit != true {
		fmt.Println("starting discovery", c.id)
		c.DoneInit = c.discoveryProcess()
	}

	fmt.Println("update", c.id)
	ticker := time.NewTicker(c.updateInterval * time.Second)
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-done:
				return
			case t := <-ticker.C:
				fmt.Println("Tick at", t)
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

var menuItems = []string{"List clients", "Add client", "Delete client", "Status"}
var arrClients = []client{}

func delete_at_index(slice []client, index int) []client {
	return append(slice[:index], slice[index+1:]...)
}

func listClientHandler() {
	fmt.Println(len(arrClients), arrClients)
}

func addClientHandler() {
	var INDEX int
	INDEX = len(arrClients)
	LOC := Location{0, 0, "", "", "", "", ""}
	arrClients = append(arrClients, client{INDEX, "test", false, false, 3, LOC, "app.service.consul"})
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

func eval(selected string) {
	fmt.Println(selected, menuItems[1])

	switch selected {
	case menuItems[0]: //list
		fmt.Println(menuItems[0])
		listClientHandler()
	case menuItems[1]: //add
		fmt.Println(menuItems[1])
		addClientHandler()
	case menuItems[2]: //delete
		fmt.Println(menuItems[2])
		deleteClientHandler()
	case menuItems[3]: //status
		fmt.Println(menuItems[3])
		statusClientHandler()
	default:
		fmt.Println("invalid command")

	}
}

func main() {
	prompt := promptui.Select{
		Label: "Select Action",
		Items: menuItems,
	}

	keepRunning := true

	for keepRunning {
		_, result, err := prompt.Run()

		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return
		}

		eval(result)
	}
}
