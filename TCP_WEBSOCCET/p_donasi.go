package main

import (
	"log"

	"github.com/gorilla/websocket"
)

// URL WebSocket server
var wsURL = "ws://localhost:8084/ws"

// Struktur data untuk menyimpan informasi donasi
type Donation struct {
	From    string  `json:"from"`    // Pengirim donasi
	Amount  float64 `json:"amount"`  // Jumlah donasi
	Message string  `json:"message"` // Pesan dari pengirim
}

func main() {
	// Mencoba untuk menghubungkan ke WebSocket server
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		log.Fatal("Koneksi WebSocket gagal:", err)
	}
	defer conn.Close() // Menutup koneksi WebSocket saat program selesai

	log.Println("Terhubung ke WebSocket server di", wsURL)

	// Loop untuk mendengarkan pesan donasi yang diterima dari server
	for {
		// Mendeklarasikan variabel untuk menyimpan data donasi yang diterima
		var donation Donation
		// Membaca data JSON dari WebSocket dan memparsingnya ke dalam variabel donation
		err := conn.ReadJSON(&donation)
		if err != nil {
			// Jika terjadi error saat membaca JSON, tampilkan pesan error dan hentikan program
			log.Println("Error membaca JSON:", err)
			break
		}
		// Menampilkan informasi donasi yang diterima
		log.Printf("Donasi diterima dari %s sejumlah %.2f dengan pesan: %s", donation.From, donation.Amount, donation.Message)
	}
}
