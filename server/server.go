package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
)

type client struct {
	username string
	room     string
}

var (
	clients = make(map[net.Conn]client)
	// variable usernameMap adalah map / set untuk melihat apakah username sudah dipakai atau tidak
	usernameMap = make(map[string]bool)
	// variable room adalah map yang memetakan string nama room ke int jumlah client didalamnya
	rooms = make(map[string]int)
	mtx   sync.Mutex
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

	defer func() {
		mtx.Lock()
		defer mtx.Unlock()

		c := clients[conn]

		if c.room != "" {
			rooms[c.room]--
		}

		delete(clients, conn)
		delete(usernameMap, strings.ToLower(c.username))
	}()

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
			lowerName := strings.ToLower(username)

			mtx.Lock()

			isDupe := usernameMap[lowerName]

			if isDupe {
				mtx.Unlock()
				fmt.Fprint(conn, "REJECTED\n")
			} else {
				// keep username not lowerName
				tempC := client{
					username: username,
					room:     "",
				}

				clients[conn] = tempC
				usernameMap[lowerName] = true
				mtx.Unlock()

				fmt.Fprint(conn, "APPROVED\n")
				fmt.Println("Username has been recieved!")

				break
			}
		}
	}

	commandInput := ""
	roomInput := ""
	message := ""
	for {
		commandInput, err = reader.ReadString('\n')

		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read room input!")
			break

		} else {
			commandInput = strings.TrimSpace(commandInput)

			if commandInput == "/create" {

				for {
					roomInput, err = reader.ReadString('\n')
					if err != nil {
						fmt.Fprintf(os.Stderr, "Failed to read room input name!")
						break
					}

					if createRoom(roomInput) {
						fmt.Fprint(conn, "APPROVED\n")
						break
					} else {
						fmt.Fprint(conn, "REJECTED\n")
					}
				}

			} else if commandInput == "/join" {

				for {
					roomInput, err = reader.ReadString('\n')

					if err != nil {
						fmt.Fprintf(os.Stderr, "Failed to read room input name!")
					}

					if joinRoom(conn, roomInput) {

						fmt.Fprint(conn, "APPROVED\n")
						joinMsg := fmt.Sprintf("%s has joined\n", username)
						broadcast(joinMsg, conn)

						for {
							message, err = reader.ReadString('\n')
							if err != nil {
								exitMsg := fmt.Sprintf("%s has left the room unexpectedly\n", username)
								broadcast(exitMsg, conn)
								leaveRoom(conn)
								break
							}

							if strings.TrimSpace(message) == "/exit" {
								fmt.Printf("%s left room", username)
								exitMsg := fmt.Sprintf("%s has left the room\n", username)
								broadcast(exitMsg, conn)
								leaveRoom(conn)
								break
							}

							formatM := fmt.Sprintf("%s: %s", username, message)
							broadcast(formatM, conn)
						}
						break
					} else {
						fmt.Fprint(conn, "REJECTED\n")
					}
				}

			} else if commandInput == "/list" {
				listRooms(conn)
			} else if commandInput == "/exit" {
				break
			} else {
				fmt.Fprintf(conn, "CREATE OR JOIN A ROOM FIRST")
			}
		}

	}
}

func broadcast(message string, sender net.Conn) {
	mtx.Lock()
	defer mtx.Unlock()

	senderRoom := clients[sender].room

	for conn := range clients {
		if conn != sender {
			room := clients[conn].room

			if room == senderRoom {
				fmt.Fprint(conn, message)
			}
		}
	}
}

func joinRoom(conn net.Conn, roomName string) bool {
	mtx.Lock()
	defer mtx.Unlock()

	roomName = strings.ToLower(strings.TrimSpace(roomName))

	// exist kalau key udah ada
	_, exists := rooms[roomName]

	if !exists {
		return false
	}
	c := clients[conn]
	c.room = roomName
	clients[conn] = c

	rooms[roomName]++

	return true
}

func createRoom(roomName string) bool {
	mtx.Lock()
	defer mtx.Unlock()

	roomName = strings.ToLower(strings.TrimSpace(roomName))
	_, exists := rooms[roomName]

	if exists {
		return false
	}

	rooms[roomName] = 0
	return true
}

func listRooms(conn net.Conn) {
	mtx.Lock()
	defer mtx.Unlock()

	if len(rooms) == 0 {
		fmt.Fprint(conn, "No rooms exists\n")
		return
	}

	for roomName, count := range rooms {
		fmt.Fprintf(conn, "- %s (%d users)\n", roomName, count)
	}

	fmt.Fprint(conn, "END\n")
}

func leaveRoom(conn net.Conn) {
	mtx.Lock()
	defer mtx.Unlock()

	c := clients[conn]

	if c.room == "" {
		return
	}

	rooms[c.room]--

	if rooms[c.room] <= 0 {
		delete(rooms, c.room)
	}

	c.room = ""
	clients[conn] = c
}
