package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

var (
	isInRoom bool
)

func main() {

	// Ngehubungin si client ke server
	conn, err := net.Dial("tcp", ":9090")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot connect to server!")
	} else {
		fmt.Println("Connected to server!")
	}
	defer conn.Close()
	connReader := bufio.NewReader(conn)
	localReader := bufio.NewReader(os.Stdin)

	//Input username
	username := ""
	for {
		fmt.Print("Type your username> ")
		username, err = localReader.ReadString('\n')

		if err != nil {
			fmt.Fprintf(os.Stderr, "Cannot read the username!")
			continue
		}

		conn.Write([]byte(username)) // fmt.Fprint(conn, message)
		status, err := connReader.ReadString('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, "Server disconnected during registration.")
			return
		}

		status = strings.TrimSpace(status)

		if status == "APPROVED" {
			fmt.Println("Username Approved :D")
			break
		} else {
			fmt.Println("ERROR DUPLICATE USERNAME, TRY AGAIN")
		}
	}

	//Input room
	fmt.Println("What do you wanna do?")
	fmt.Println("* /create to create room")
	fmt.Println("* /join to join a room")
	fmt.Println("* /list to list all available room")
	fmt.Println("* /exit to close the program")
	commandInput := ""

	isInRoom = false

	for {
		commandInput, err = localReader.ReadString('\n')

		if err != nil {
			fmt.Fprintf(os.Stderr, "Cannot read the command!")
			continue
		}

		commandInput = strings.TrimSpace(commandInput)

		if commandInput == "/create" {
			fmt.Fprint(conn, "/create\n")
			fmt.Println("Choose room name")

			for {
				commandInput, err = localReader.ReadString('\n')

				if err != nil {
					fmt.Fprintf(os.Stderr, "disconnected during room creating session.")
					return
				}

				conn.Write([]byte(commandInput)) // fmt.Fprint(conn, message)
				status, err := connReader.ReadString('\n')
				if err != nil {
					fmt.Fprintf(os.Stderr, "disconnected during room creating session.")
					return
				}

				status = strings.TrimSpace(status)

				if status == "APPROVED" {
					fmt.Println("Room is Created")
					break
				} else {
					fmt.Println("ERROR DUPLICATE ROOM NAME, TRY AGAIN")
				}
			}

		} else if commandInput == "/join" {
			fmt.Fprint(conn, "/join\n")
			fmt.Println("Choose room you wanna join")

			for {
				commandInput, err = localReader.ReadString('\n')

				if err != nil {
					fmt.Fprintf(os.Stderr, "Server disconnected during room joining session.")
					return
				}

				conn.Write([]byte(commandInput)) // fmt.Fprint(conn, message)
				status, err := connReader.ReadString('\n')
				if err != nil {
					fmt.Fprintf(os.Stderr, "disconnected during room creating session.")
					return
				}

				status = strings.TrimSpace(status)

				if status == "APPROVED" {
					fmt.Println("Successfully Joined The Room")
					isInRoom = true
					break
				} else {
					fmt.Println("ERROR INVALID ROOM NAME")
				}
			}

			fmt.Println("Start typing your messages")

			go handleIncomingMessage(conn)

			for {
				//Input messages
				message, err := localReader.ReadString('\n')

				if err != nil {
					fmt.Fprintf(os.Stderr, "Cannot read the message!")
					continue
				}

				if strings.TrimSpace(message) == "/exit" {
					fmt.Println("Exiting Room")
					fmt.Fprint(conn, "/exit\n")
					isInRoom = false
					break
				}

				conn.Write([]byte(message)) // fmt.Fprint(conn, message)

			}

		} else if commandInput == "/list" {
			fmt.Fprint(conn, "/list\n")
			res := ""

			for {
				listRoom, err := connReader.ReadString('\n')

				if err != nil {
					fmt.Fprintf(os.Stderr, "Server disconnected during room joining session.")
					return
				}

				listRoom = strings.TrimSpace(listRoom)

				if listRoom == "END" {
					break
				}

				res += listRoom + "\n"

			}

			fmt.Print(res)

		} else if commandInput == "/exit" {
			fmt.Fprint(conn, "/exit\n")
			break
		} else {
			fmt.Println("Wrong Command Input")
		}
	}
}

func handleIncomingMessage(conn net.Conn) {
	connReader := bufio.NewReader(conn)

	for isInRoom {
		if !isInRoom {
			return
		}

		incoming, err := connReader.ReadString('\n')
		if err != nil {
			os.Exit(0)
			return
		} else {
			//SAKTI ANJIR
			fmt.Print("\r\033[K")
			fmt.Print(incoming)
		}
	}
}
