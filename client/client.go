package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {

	// Ngehubungin si client ke server
	conn, err := net.Dial("tcp", ":9090")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot connect to server!")
	} else {
		fmt.Println("Connected to server!")
	}

	// Kalau koneksi sudah selesai tutup koneksi ke server
	defer conn.Close()

	// Menginstansasi reader untuk dari server dan reader untuk dari input pengguna via terminal
	connReader := bufio.NewReader(conn)
	localReader := bufio.NewReader(os.Stdin)

	// Variable untuk input nama dari client
	var username string = ""

	// Looping untuk pengecekan input nama dari client apakah sudah disetujui atau belum oleh server
	// Pertama pengguna mengisi username mereka lalu bila ada error checking bila tidak bisa
	// Setelah itu akan dikirimkan ke server dan melihat via error checking apakah berhasil atau tidak
	// Setelah itu menerima dari server apakah status pengisian username disetujui atau tidak
	// Bila sudah diterima maka akan membuka thread untuk menghandle message-message lain yang diberikan client lain
	for {
		fmt.Print("Type your username> ")
		username, err = localReader.ReadString('\n')

		if err != nil {
			fmt.Fprintf(os.Stderr, "Cannot read the username!")
			continue
		}

		conn.Write([]byte(username))
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

	// Membuat thread pengecekan incoming message client lain agar dapat di tampilkan di layar pengguna
	go handleIncomingMessage(conn)

	// Looping untuk menangkap input message dari client untuk dikirimkan ke client lain
	// Pertama akan dibaca dari terminal client dan dicek apakah terjadi error
	// bila message yang dikirimkan berupa exit maka akan dikirimkan ke server untuk mengatakan client ingin disconnect dan looping selesai
	// dan bila hanya message biasa maka akan dikirimkan ke server saja
	for {
		message, err := localReader.ReadString('\n')

		if err != nil {
			fmt.Fprintf(os.Stderr, "Cannot read the message!")
			continue
		}

		if message == "/exit" {
			fmt.Fprint(conn, "/exit\n")
			break
		}

		conn.Write([]byte(message))
	}
}

// Function untuk mengecek input yang datang dari sever dan menampilkannya ke layar client
func handleIncomingMessage(conn net.Conn) {
	// Instansiasi reader untuk membaca input dari server
	connReader := bufio.NewReader(conn)

	// Melakukan looping untuk mengecek apakah ada message dari server yang datang terus menerus dan bila ada maka akan di tampilkan di terminal client
	// bila terjadi error membaca maka akan ditutup
	for {
		incoming, err := connReader.ReadString('\n')
		if err != nil {
			os.Exit(0)
		} else {
			fmt.Print(incoming)
		}
	}
}
