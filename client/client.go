package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

var (
	// ini variable untuk menghandle bila client belum ada di dalam room jangan membuka buffered reader dari server untuk menerima message-message
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

	//block for loop input username dimana minta dari reader standard input (terminal) minta username clientnya mau apa
	for {
		fmt.Print("Type your username> ")
		username, err := localReader.ReadString('\n')

		if err != nil {
			fmt.Fprintf(os.Stderr, "Cannot read the username!")
			continue
		}

		//dia ngirim ke server hasil input dari standard input yang dimasukin ke variable username
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

	// block of information mau clientnya dan diberikan list of command yang bisa diterima oleh server
	fmt.Println("What do you wanna do?")
	fmt.Println("* /create to create room")
	fmt.Println("* /join to join a room")
	fmt.Println("* /list to list all available room")
	fmt.Println("* /exit to close the program")

	// ngeset variable untuk ngecek apakah client sudah didalam room di set jadi false karena masih di luar
	isInRoom = false

	// looping block untuk input command client
	for {
		commandInput, err := localReader.ReadString('\n')

		if err != nil {
			fmt.Fprintf(os.Stderr, "Cannot read the command!")
			continue
		}

		commandInput = strings.TrimSpace(commandInput)

		if commandInput == "/create" {
			fmt.Fprint(conn, "/create\n")
			fmt.Println("Choose room name")

			// looping untuk pengecekan oleh server apakah input room name valid atau tidak
			for {
				commandInput, err = localReader.ReadString('\n')

				if err != nil {
					fmt.Fprintf(os.Stderr, "disconnected during room creating session.")
					return
				}

				conn.Write([]byte(commandInput))

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

			commandInput, err = localReader.ReadString('\n')

			if err != nil {
				fmt.Fprintf(os.Stderr, "Server disconnected during room joining session.")
				return
			}

			conn.Write([]byte(commandInput))

			status, err := connReader.ReadString('\n')
			if err != nil {
				fmt.Fprintf(os.Stderr, "disconnected during room creating session.")
				return
			}

			status = strings.TrimSpace(status)

			if status == "APPROVED" {
				// Kalau di approved dia akan masuk ke looping message untuk handle client sekarang ada di room
				fmt.Println("Successfully Joined The Room")
				// variable pengecekan apakah sudah masuk room menjadi true karena dia masuk ke room menandakan client ini sudah dalam posisi menerima message-message pengguna lain yang di handle di incoming message go routine
				isInRoom = true

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
						// kalau exit varible ini jadi false kembali sampai di join
						isInRoom = false
						break
					}

					conn.Write([]byte(message))
				}

			} else {
				fmt.Println("ERROR INVALID ROOM NAME")
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

				// looping nerima message ini hanya akan end kalau dari server diberikan special message END
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

// function yang bilang client sudah siap menerima message-message dari client lain dan bukan server acc/status
// untuk mengetahui kapan client ini siap untuk menerima message-message dari client lain kita cek variable isInRoom
// karena client hanya akan menerima message-message dari client lain kalau dia sudah masuk room
func handleIncomingMessage(conn net.Conn) {
	connReader := bufio.NewReader(conn)

	for isInRoom {
		// kalau is in roomnya jadi false/ client keluar dari room maka function ini selesai
		if !isInRoom {
			return
		}

		incoming, err := connReader.ReadString('\n')
		if err != nil {
			os.Exit(0)
			return
		} else {
			// \r <-- cursor pindahin ke paling kiri
			// \033 <-- escape character stop nge print ke layar untuk sementara
			// [K <-- hapus semua isi line sampe end of line secara visual
			fmt.Print("\r\033[K") // <-- ANSI escape codes untuk formatting message yang masuk
			// nge print message yang didapat dari server
			fmt.Print(incoming)
		}
	}
}
