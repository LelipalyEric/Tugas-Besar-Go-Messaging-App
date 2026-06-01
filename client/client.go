package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {

	// Ngehubungin si client ke seerver
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

	fmt.Println("Start typing your messages")

	go handleIncomingMessage(conn)

	for {
		//Input messages
		message, err := localReader.ReadString('\n')

		if err != nil {
			fmt.Fprintf(os.Stderr, "Cannot read the message!")
			continue
		}

		if message == "/exit" {
			fmt.Fprint(conn, "/exit\n")
			break
		}

		conn.Write([]byte(message)) // fmt.Fprint(conn, message)

	}
}

func handleIncomingMessage(conn net.Conn) {
	connReader := bufio.NewReader(conn)

	for {
		incoming, err := connReader.ReadString('\n')
		if err != nil {
			os.Exit(0)
		} else {
			fmt.Print(incoming)
		}
	}

}
