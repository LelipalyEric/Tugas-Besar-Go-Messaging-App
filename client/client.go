package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
    conn, err := net.Dial("tcp", ":9090")
    if err != nil {
        fmt.Fprintf(os.Stderr, "Cannot connect to server!")
    } else {
        fmt.Println("Connected to server!")
    }
    
    connReader := bufio.NewReader(conn)
    localReader := bufio.NewReader(os.Stdin)
    fmt.Print("Type your message> ")
    message, err := localReader.ReadString('\n')
    if err != nil {
        fmt.Fprintf(os.Stderr, "Cannot read the message!")
    } else {
        fmt.Println("The message has been read!")
    }

    conn.Write([]byte(message)) // fmt.Fprint(conn, message)
    fmt.Println("The message has been sent!")
    fmt.Println("Waiting for reply...")
    echo, err := connReader.ReadString('\n')
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to read the echo!")
        os.Exit(1)
    } else {
        fmt.Println("The echo has been received!")
    }

    fmt.Println(echo)
}
