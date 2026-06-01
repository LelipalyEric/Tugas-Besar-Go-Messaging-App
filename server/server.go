package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
    ln, err := net.Listen("tcp", ":9090")
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to listen!")
        os.Exit(1)
    } else {
        fmt.Println("Listening...")
    }
    conn, err := ln.Accept()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to accept connection!")
        os.Exit(1)
    } else {
        fmt.Println("New connection accepted!")
    }

    reader := bufio.NewReader(conn)

    fmt.Println("Waiting for message...")
    message, err := reader.ReadString('\n')
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to read message!")
        os.Exit(1)
    } else {
        fmt.Println("The message has been received!")
    }

    fmt.Fprintf(conn, "Echo: %s", message)
    fmt.Println("The message has been echoed back!")
}

