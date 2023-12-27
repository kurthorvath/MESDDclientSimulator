package main

import (
	"fmt"
	"time"

	"github.com/manifoldco/promptui"
)

type Location struct {
}

type client struct {
	id             int
	status         string
	terminate      bool
	DoneInit       bool
	updateInterval time.Duration
	loc            Location
}

func (c *client) discoveryProcess() {
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

	arrClients = append(arrClients, client{INDEX, "test", false, false, 3})
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
