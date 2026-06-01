package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
)

var (
	clients = make(map[net.Conn]string)
	mtx     sync.Mutex
)

func main() {

	//Setup nyalain server suruh connect
	ln, err := net.Listen("tcp", ":9090")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to listen!")
	} else {
		fmt.Println("Listening...")
	}

	//Looping buat nge accept multiple client
	for {
		conn, err := ln.Accept()

		// Error handling kalau kena error dia ga masuk lagi kebawah
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to accept connection!")
			continue
		}

		// Ga kena error dia ngomong connectionya dapet
		fmt.Println("New connection accepted!")

		// Manggil function buat handle si client

		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	usernameInput := ""
	username := ""

	err := fmt.Errorf("")
	for {
		usernameInput, err = reader.ReadString('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read username!")
		} else {
			username = strings.TrimSpace(usernameInput)

			mtx.Lock()
			isDupe := false
			for _, names := range clients {
				if strings.EqualFold(names, username) {
					isDupe = true
					break
				}
			}

			if isDupe {
				mtx.Unlock()
				fmt.Fprint(conn, "REJECTED\n")
			} else {
				clients[conn] = username
				mtx.Unlock()
				fmt.Fprint(conn, "APPROVED\n")
				fmt.Println("Username has been recieved!")
				joinMsg := fmt.Sprintf("%s has joined\n", username)
				broadcast(joinMsg, conn)
				break
			}
		}
	}

	message := ""

	for {
		message, err = reader.ReadString('\n')
		if err != nil {
			exitMsg := fmt.Sprintf("%s has left the groupchat unexpectedly\n", username)
			broadcast(exitMsg, conn)
			break
		} else {
			fmt.Println("The message has been received!")
		}

		if strings.TrimSpace(message) == "/exit" {
			fmt.Printf("%s left groupchat", username)
			exitMsg := fmt.Sprintf("%s has left the groupchat\n", username)
			broadcast(exitMsg, conn)
			break
		}

		formatM := fmt.Sprintf("%s: %s", username, message)
		broadcast(formatM, conn)
	}

	mtx.Lock()
	delete(clients, conn)
	mtx.Unlock()
}

func broadcast(message string, sender net.Conn) {
	mtx.Lock()
	defer mtx.Unlock()
	for conn := range clients {
		if conn != sender {
			fmt.Fprint(conn, message)
		}
	}
}
