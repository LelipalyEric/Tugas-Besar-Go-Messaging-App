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
	// Variable tempat nyimpen koneksi client mapping koneksi ke string nama
	clients = make(map[net.Conn]string)
	mtx     sync.Mutex
)

func main() {
	//Setup nyalain server buat bisa connect
	ln, err := net.Listen("tcp", ":9090")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to listen!")
	} else {
		fmt.Println("Listening...")
	}

	// Looping buat nge accept multiple client
	// Pertama Terima koneksi kalau ada yang mau masuk
	// Kedua ada Error handling kalau kena error saat di accept masuk
	// Kalau berhasil koneksi kasih tahu di terminal server dan panggil thread buat nge handle koneksi client itu
	for {
		conn, err := ln.Accept()

		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to accept connection!")
			continue
		}

		fmt.Println("New connection accepted!")
		go handleClient(conn)
	}
}

// Fungsi utama yang bakal di run jadi thread buat masing-masing koneksi pengguna
func handleClient(conn net.Conn) {

	// Tutup koneksi kalau udah selesai execute
	defer conn.Close()

	// Instansiasi variable reader dan error
	reader := bufio.NewReader(conn)
	err := fmt.Errorf("")

	// Input username yang diberikan client dan username yang username yang berhasil di set
	var usernameInput string = ""
	var username string = ""

	// Looping pengecekan input username sampai di terima usernamenya tidak sama dengan username client lain
	// Pertama, melakukan pengecekan error input lalu mengecek dari map apakah ada username yang sama dengan input baru
	// Kedua, bila ada dengan ditandai flag boolean isdupe maka akan dikembalikan Rejected atau Approved dari server
	// Bila hasilnya approved maka melakukan input ke map dan lakukan broadcast bahwa pengguna telah masuk
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

	// Input message adalah input dari client berupa message yang ingin dikirim dan message adalah final bentuk message setelah formatting untuk dibroadcast
	var message string = ""
	var messageInput string = ""

	// Looping dilakukan agar client dapat melakukan pengiriman message berkali-kali
	// Pertama server membaca input dari client dan disimpan di variable
	// Lalu dilakukan pengecekan bila ada kegagalan dalam menangkap input dari client
	// Bila berhasil maka akan dilakukan pengecekan kedua kali bila message yang diberikan berupa command exit maka server akan memutus hubungan dengan client dan memberi tahu client lain dengan broadcast
	// Bila message hanya berisi message biasa maka akan di format message tersebut lalu di broadcast agar memiliki nama pengirimnya siapa
	for {
		messageInput, err = reader.ReadString('\n')
		if err != nil {
			message := fmt.Sprintf("%s has left the groupchat unexpectedly\n", username)
			broadcast(message, conn)
			break
		} else {
			fmt.Println("The message has been received!")
		}

		if strings.TrimSpace(message) == "/exit" {
			fmt.Printf("%s left groupchat", username)
			message := fmt.Sprintf("%s has left the groupchat\n", username)
			broadcast(message, conn)
			break
		}

		message := fmt.Sprintf("%s: %s", username, messageInput)
		broadcast(message, conn)
	}

	// Bila keluar dari looping message maka melakukan delete koneksi client dari map agar usernamenya dapat diguanakan kembali
	mtx.Lock()
	delete(clients, conn)
	mtx.Unlock()

}

// Fungsi broadcast untuk menyebar message ke client lain selain pengirim client itu sendiri dengan input berupa message yang akan dikirimkan
// Dilakukan looping untuk isi map dan bila bukan client pengirim maka akan dilakukan penulisan ke client tersebut berisi input message
func broadcast(message string, sender net.Conn) {
	mtx.Lock()
	for conn := range clients {
		if conn != sender {
			fmt.Fprint(conn, message)
		}
	}
	mtx.Unlock()
}
