package main

import (
	"fmt"

	"github.com/manifoldco/promptui"
)

type client struct {
	id     int
	status string
}

var menuItems = []string{"List clients", "Add client", "Delete client", "Status"}
var arrClients = []client{}

func listClientHandler() {
	fmt.Println(len(arrClients), arrClients)
}

func addClientHandler() {
	arrClients = append(arrClients, client{1, "test"})
}

func eval(selected string) {
	fmt.Println(selected, menuItems[1])

	switch selected {
	case menuItems[0]:
		fmt.Println(menuItems[0])
		listClientHandler()
	case menuItems[1]:
		fmt.Println(menuItems[1])
		addClientHandler()
	case menuItems[2]:
		fmt.Println(menuItems[2])
	case menuItems[3]:
		fmt.Println(menuItems[3])
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
