package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
)

// Kelas / struct client buat nyimpen hubungan username client dan room yang dia join
type client struct {
	username string
	room     string
}

var (
	// Variable map untuk memetakan connection-connection ke clientnya (username sama roomnya)
	clients = make(map[net.Conn]client)

	// variable usernameMap adalah map / set untuk melihat apakah username sudah dipakai atau tidak
	usernameMap = make(map[string]bool)

	// variable room adalah map yang memetakan string nama room ke int jumlah client didalamnya
	rooms = make(map[string]int)

	mtx sync.Mutex
)

func main() {

	//Setup nyalain server suruh connect ke port 9090 over tcp
	ln, err := net.Listen("tcp", ":9090")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to listen!")
		return
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
	// defer selalu dipanggil di akhir function
	// yang dilakukan di akhir function adalah menutup connection dan menghapus client dari map dan usernamenya dari map/set username dan juga mengurangi int penanda banyak orang yang ada di room
	// defer bekerja seperti stack yang masuk pertama di lakukan di akhir oleh karena itu conn.Close() disimpan di paling atas menandakan hubungan sudah di putus dari server
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

	// block looping ini fungsinya buat ngecek username bentrok atau ga sama username client yang lain
	for {

		usernameInput, err := reader.ReadString('\n')

		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read username!")
			break
		} else {
			// ubah username input jadi lowercase hanya mengecek character ga ngecek capitalisasi
			username := strings.TrimSpace(usernameInput)
			lowerName := strings.ToLower(username)

			mtx.Lock()

			// ngecek apakah username duplikat dengan masukin keynya username tadi apakah true atau false duplikat
			isDupe := usernameMap[lowerName]

			if isDupe {
				mtx.Unlock()
				fmt.Fprint(conn, "REJECTED!!! TRY AGAIN\n")
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

	// Block for loop ini fungsinya ngurusin command setelah client masukin username
	roomInput := ""

	for {
		commandInput, err := reader.ReadString('\n')

		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read room input!")
			break
		} else {
			commandInput = strings.TrimSpace(commandInput)

			// command /create fungsinya supaya client bisa bikin room baru dan ngasih nama room tersebut apa sekalian di cek apakah nama room itu udah di pake atau belum
			if commandInput == "/create" {

				for {
					roomInput, err = reader.ReadString('\n')
					if err != nil {
						fmt.Fprintf(os.Stderr, "Failed to read room input name!")
						break
					}

					// dia manggil function buat ngecek input room returnya true false nandain berhasil atau ga
					if createRoom(roomInput) {
						fmt.Fprint(conn, "APPROVED\n")
						break
					} else {
						fmt.Fprint(conn, "REJECTED\n")
					}
				}

				// command /join fungsinya handle client mau join ke room apa nanti di cek input roomnya sesuai dengan nama room yang ada atau tidak kalau tidak sesuai dia harus masukin command /join lagi
			} else if commandInput == "/join" {

				roomInput, err = reader.ReadString('\n')
				username := clients[conn].username

				if err != nil {
					fmt.Fprintf(os.Stderr, "Failed to read room input name!")
					break
				}

				// function buat ngecek apakah room itu beneneran ada atau tidak dan juga bila ada memasukan client ke room tersebut
				if joinRoom(conn, roomInput) {

					fmt.Fprint(conn, "APPROVED\n")
					// saat berhasil join client tersebut akan di announce joinnya dengan cara melakukan broadcast dengan format berikut dengan namanya tertera dan dikirim ke penghuni room yang sama
					joinMsg := fmt.Sprintf("%s has joined\n", username)
					broadcast(joinMsg, conn)

					// block looping untuk client message pengguna yang ada di room yang sama
					for {
						message, err := reader.ReadString('\n')

						if err != nil {
							exitMsg := fmt.Sprintf("%s has left the room unexpectedly\n", username)
							broadcast(exitMsg, conn)
							// client akan dipanggilkan function leave room dimana akan ada penghapusan room yang dia sekarang di assign
							leaveRoom(conn)
							break
						}

						// client bisa keluar dari room yang sekarang dengan memanggil command exit dimana akan dilakukan broacast ke isi room kalau client tersebut sudah keluar
						if strings.TrimSpace(message) == "/exit" {
							fmt.Printf("%s left room", username)
							exitMsg := fmt.Sprintf("%s has left the room\n", username)
							broadcast(exitMsg, conn)
							// client akan dipanggilkan function leave room dimana akan ada penghapusan room yang dia sekarang di assign
							leaveRoom(conn)
							break
						}

						formatM := fmt.Sprintf("%s: %s", username, message)
						broadcast(formatM, conn)
					}
				} else {
					fmt.Fprint(conn, "REJECTED\n")
				}

				// command /list fungsinya untuk client memanggil nama-nama room yang ada beserta jumlah orang yang join di room teresebut
			} else if commandInput == "/list" {
				listRooms(conn)
				// command /exit diluar join adalah untuk client keluar dari aplikasi dan menutup koneksi yang nantinya akan di handle oleh function defer
			} else if commandInput == "/exit" {
				break
				// kalau command tidak dikenali maka akan diberikan message error
			} else {
				fmt.Fprintf(conn, "CREATE OR JOIN A ROOM FIRST")
			}
		}
	}
}

// function broadcast dia ngirim message ke semua client yang satu room dengan sender kecuali sender itu sendiri
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
	// karena gaboleh ngengati variable dari sebuah struct yang ada di map
	// kita buat objek client baru yang sekarang nama roomnya bukan kosong tapi sudah di assign sama roomname input
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

// function list room dia bakal looping sebanyak jumlah room yang ada di map dan akan mengeluarkan nama roomnya dan jumlah anggotanya
// di dalam looping akan terus melakukan write ke connection
// untuk mengetahui kapan selesai si looping roomnya maka dibutuhkan special ending message untuk memberitahu loop sudah selesai dan tidak mengirim list room name dan jumlah anggotanya
// hal ini dihandle dengan cara memberikan special messagenya yaitu fmt.Fprint(conn, "END\n")
func listRooms(conn net.Conn) {
	mtx.Lock()
	defer mtx.Unlock()

	if len(rooms) == 0 {
		fmt.Fprint(conn, "No rooms exists\n")
		fmt.Fprint(conn, "END\n")
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

	// kalau misalnya sebuah room di leave oleh orang terakhir dan alhasil sisa 0 client dalam room tersebut maka room akan di delete dari map
	if rooms[c.room] <= 0 {
		delete(rooms, c.room)
	}

	c.room = ""
	clients[conn] = c
}
