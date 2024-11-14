// server.go
package main

import (
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
	broadcast = make(chan Donasi)              // Channel untuk mengirim data donation ke semua klien
	upgrader  = websocket.Upgrader{            // WebSocket upgrader untuk mengupgrade koneksi HTTP menjadi WebSocket
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	clientMutex = sync.Mutex{}             // Mutex untuk mengamankan akses ke map 'clients' dari banyak goroutine
	balances    = make(map[string]float64) // Menyimpan saldo masing-masing pengguna
)

type Donasi struct {
	From    string  `json:"from"`    // Nama pengirim
	Amount  float64 `json:"amount"`  // Jumlah donasi
	Message string  `json:"message"` // Pesan dari pengirim
}

func main() {
	go handleUDP()                               // Menangani koneksi UDP
	go handleWebSocket()                         // Menangani koneksi WebSocket
	http.HandleFunc("/ws", wsHandler)            // Menangani request WebSocket di endpoint "/ws"
	log.Println("Server started on port :8080")  // Menampilkan pesan server sudah mulai berjalan
	log.Fatal(http.ListenAndServe(":8080", nil)) // Menjalankan server HTTP di port 8080
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

	// Mendengarkan dan menerima donation dari klien WebSocket
	for {
		var donasi Donasi
		if err := conn.ReadJSON(&donasi); err != nil {
			log.Printf("Error reading JSON: %v", err)
			break
		}
		broadcast <- donasi // Mengirim donasi yang diterima ke channel broadcast
	}
}

func handleUDP() {
	addr, err := net.ResolveUDPAddr("udp", ":8081") // Mengatur alamat UDP pada port 8081
	if err != nil {
		log.Fatalf("Failed to resolve UDP address: %v", err)
	}
	conn, err := net.ListenUDP("udp", addr) // Mendengarkan koneksi UDP
	if err != nil {
		log.Fatalf("Failed to start UDP server: %v", err)
	}
	defer conn.Close()

	buffer := make([]byte, 1024) // Buffer untuk menerima data

	for {
		n, _, err := conn.ReadFromUDP(buffer) // Membaca data dari klien
		if err != nil {
			log.Println("Error reading from UDP client:", err)
			continue
		}
		go handleUDPMessage(buffer[:n]) // Memproses pesan UDP
	}
}

func handleUDPMessage(data []byte) {
	line := string(data)

	// Memisahkan input berdasarkan spasi
	parts := strings.Fields(line)
	if len(parts) < 2 {
		log.Println("Invalid input format:", line)
		return
	}

	// Menangani perintah TOP_UP atau DONASI
	username := parts[0]
	var donasi Donasi
	donasi.From = username
	var err error
	donasi.Amount, err = strconv.ParseFloat(parts[1], 64) // Mengonversi jumlah donasi dari string ke float
	if err != nil {
		log.Println("Invalid amount format:", parts[1])
		return
	}

	// Menggabungkan sisa bagian sebagai pesan
	donasi.Message = strings.Join(parts[2:], " ")

	clientMutex.Lock()
	if donasi.Message == "TOP_UP" {
		balances[donasi.From] += donasi.Amount
	} else {
		balances[donasi.From] -= donasi.Amount
		broadcast <- donasi
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
