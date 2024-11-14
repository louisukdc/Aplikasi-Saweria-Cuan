package main

import (
    "log"
    "github.com/gorilla/websocket"
)

var wsURL = "ws://localhost:8080/ws" // URL WebSocket server

func main() {
    // Mencoba untuk menghubungkan ke WebSocket server
    conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
    if err != nil {
        // Jika koneksi gagal, tampilkan pesan error dan hentikan program
        log.Fatal("Koneksi WebSocket gagal:", err)
    }
    defer conn.Close() // Menutup koneksi WebSocket saat program selesai

    for {
        // Mendeklarasikan variabel untuk menyimpan data donasi
        var donation Donation
        // Membaca data JSON dari WebSocket dan memparsingnya ke dalam variabel donation
        err := conn.ReadJSON(&donation)
        if err != nil {
            // Jika terjadi error saat membaca JSON, tampilkan pesan error dan hentikan program
            log.Println("Error membaca JSON:", err)
            return
        }
        // Menampilkan informasi donasi yang diterima
        log.Printf("Donasi diterima dari %s sejumlah %.2f dengan pesan: %s", donation.From, donation.Amount, donation.Message)
    }
}

// Struktur data untuk menyimpan informasi donasi
type Donation struct {
    From    string  `json:"from"`    // Pengirim donasi
    Amount  float64 `json:"amount"`  // Jumlah donasi
    Message string  `json:"message"` // Pesan dari pengirim
}
