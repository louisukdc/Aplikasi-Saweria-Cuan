// server.go
package main

import (
	"bufio"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

var (
	clients   = make(map[*websocket.Conn]bool) // Menyimpan koneksi WebSocket aktif
	broadcast = make(chan Donation)            // Channel untuk mengirim data donasi ke semua klien
	upgrader  = websocket.Upgrader{            // WebSocket upgrader untuk mengubah koneksi HTTP menjadi WebSocket
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	clientMutex = sync.Mutex{}             // Mutex untuk mengamankan akses ke map 'clients' dari banyak goroutine
	balances    = make(map[string]float64) // Menyimpan saldo masing-masing pengguna
)

type Donation struct {
	From    string  `json:"from"`    // Nama pengirim
	Amount  float64 `json:"amount"`  // Jumlah donasi
	Message string  `json:"message"` // Pesan dari pengirim
}

func main() {
	go handleTCP()                               // Menangani koneksi TCP
	go handleWebSocket()                         // Menangani koneksi WebSocket
	http.HandleFunc("/ws", wsHandler)            // Menangani request WebSocket di endpoint "/ws"
	log.Println("Server started on port :8084")  // Menampilkan pesan server sudah mulai berjalan
	log.Fatal(http.ListenAndServe(":8084", nil)) // Menjalankan server HTTP di port 8084
}

// Fungsi untuk menangani koneksi WebSocket
func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil) // Upgrade koneksi HTTP menjadi WebSocket
	if err != nil {
		log.Println("WebSocket upgrade failed:", err)
		return
	}
	defer func() {
		clientMutex.Lock()
		delete(clients, conn) // Hapus koneksi saat WebSocket ditutup
		clientMutex.Unlock()
		conn.Close()
		log.Println("Client disconnected")
	}()

	clientMutex.Lock()
	clients[conn] = true // Menambahkan koneksi WebSocket yang baru ke dalam map clients
	clientMutex.Unlock()

	// Mendengarkan dan menerima donasi dari klien WebSocket
	for {
		var donation Donation
		if err := conn.ReadJSON(&donation); err != nil {
			log.Printf("Error reading JSON: %v", err)
			break
		}
		broadcast <- donation // Mengirim donasi yang diterima ke channel broadcast
	}
}

func handleTCP() {
	listener, err := net.Listen("tcp", ":8083") // Ganti port 8082 ke 8083
	if err != nil {
		log.Fatalf("Failed to start TCP server: %v", err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept() // Menerima koneksi dari klien
		if err != nil {
			log.Println("Error accepting TCP connection:", err)
			continue
		}
		go handleTCPConnection(conn)
	}
}

func handleTCPConnection(c net.Conn) {
	defer c.Close() // Menutup koneksi TCP setelah selesai

	reader := bufio.NewReader(c)
	line, err := reader.ReadString('\n') // Membaca input baris per baris
	if err != nil {
		log.Println("Error reading from TCP client:", err)
		return
	}

	// Memisahkan input berdasarkan spasi
	parts := strings.Fields(line)
	if len(parts) < 2 {
		log.Println("Invalid input format:", line)
		return
	}

	// Menangani perintah TOP_UP atau DONASI
	username := parts[0]
	var donation Donation
	donation.From = username
	donation.Amount, err = strconv.ParseFloat(parts[1], 64) // Mengonversi jumlah donasi dari string ke float
	if err != nil {
		log.Println("Invalid amount format:", parts[1])
		return
	}

	// Menggabungkan sisa bagian sebagai pesan
	donation.Message = strings.Join(parts[2:], " ")

	clientMutex.Lock()
	if donation.Message == "TOP_UP" {
		balances[donation.From] += donation.Amount
	} else {
		balances[donation.From] -= donation.Amount
		broadcast <- donation
	}
	clientMutex.Unlock()
}

// Fungsi untuk menangani broadcast donasi ke semua klien WebSocket
func handleWebSocket() {
	for donation := range broadcast {
		clientMutex.Lock()
		for client := range clients {
			if err := client.WriteJSON(donation); err != nil {
				log.Printf("Error broadcasting to client: %v", err)
				client.Close()
				delete(clients, client) // Menghapus klien yang terputus
			}
		}
		clientMutex.Unlock()
	}
}
