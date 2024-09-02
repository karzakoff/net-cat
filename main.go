package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

// structure qui contients tous les clients
type client struct {
	conn net.Conn    // adresse
	name string      // nom
	ch   chan string // channel
}

// représente chaque client
var clients = make(map[net.Conn]*client)

// channel pour diffuser les messages à tous les clients connectés
var hub = make(chan string)

// historique des messages
var messageHistory []string

var allClients []client

var numberClients []string

var currentConnections int = 9

func handleConnections(conn net.Conn) {
	// on envoie ici dans la structure du clients toutes ses infos
	client := &client{conn: conn, name: "", ch: make(chan string, 1024)}
	clients[conn] = client

	date := time.Now()
	currentTime := date.Format("[2006-01-02 15:04:05]")

	fmt.Println(currentConnections)
	if len(numberClients) >= currentConnections {
		fmt.Println(currentConnections)
		for _, kick := range clients {
			if kick.conn == conn {
				fmt.Fprint(conn, "error")
				//delete(clients, conn)
				conn.Close()
				return
			}
		}
	} else {
		numberClients = append(numberClients, client.name)
		//currentConnections++
	}

	header, err := os.ReadFile("pinguin.txt")
	if err != nil {
		log.Fatal(err)
	}
	conn.Write(header)

	// et dans la boucle infinie on lit tous les messages
	go func() {
		for {

			msg, err := bufio.NewReader(conn).ReadString('\n')

			// si jamais y'a une erreur dans la lecture ça veut dire que le client est parti
			if err != nil {
				delete(clients, conn)
				currentConnections--
				fmt.Println(client.name, "has left the chat")
				hub <- client.name + " has left the chat" + "\n"
				break
			}

			// ici on rentre le nom du client
			fmt.Println(client.name)
			if client.name == "" {

				client.name = strings.TrimSpace(string(msg))
				fmt.Println(client.name, "has joined the chat")

				for _, actualClient := range allClients {
					actualClient.ch <- client.name + " has joined the chat\n"
				}

				allClients = append(allClients, *client)

				//fmt.Fprint(conn, " ====>> If you want to change your name, write: NEWNAME\n")

				if len(numberClients) == len(allClients) {
					for _, msg := range messageHistory {
						client.ch <- msg
					}
					continue
				}
			}

			if strings.TrimSpace(string(msg)) == "" {
				fmt.Fprint(conn, "Error: cannot send an empty message\n")
				continue
			}

			/*newName := ""

			for _, letter := range msg {
				newName += string(letter)
				if newName == "NEWNAME" {
					fmt.Fprint(conn, "\033[1A\033[K")
					fmt.Fprint(conn, "CHOOSE A NEW NAME: ")
					//hub <- "====== somebody change his name ======\n"
					break
				}
			}*/

			if conn == client.conn {
				fmt.Fprint(conn, "\033[1A\033[K")
				broadcast(currentTime + "[" + client.name + "]" + ":" + string(msg))
			}

			messageHistory = append(messageHistory, currentTime+"["+client.name+"]:"+string(msg))
		}
		numberClients = numberClients[:len(numberClients)-1]

	}()

	// go routine qui gère deux cas à la fois
	// 1 qui attend la réception d'un message sur le channel hub
	// 1 autre qui attend la réception d'un message sur le channel client.ch
	go func() {
		for {
			select {
			case msg := <-hub:
				if conn != client.conn {
					fmt.Fprint(conn, msg)
				}
			case msg := <-client.ch:
				if conn == client.conn {
					fmt.Fprint(conn, msg)
				}
			}
		}
	}()
}

// cette fonction sert à envoyer les messages à tous les clients
func broadcast(msg string) {
	for _, client := range clients {
		client.ch <- msg
	}
}

func main() {
	var port string
	if len(os.Args) > 1 && os.Args[1] != "" {
		port = os.Args[1]
	} else {
		port = "8989"
	}
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Server started on port : " + port)

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}

		go handleConnections(conn)
	}
}
