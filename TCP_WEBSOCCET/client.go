package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

var saldo float64   // Menyimpan saldo saat ini
var username string // Menyimpan username pengguna

func main() {
	reader := bufio.NewReader(os.Stdin)

	// Langsung panggil fungsi login
	login(reader)

	// Setelah login, masuk ke menu utama donasi
	mainMenu(reader)
}

func login(reader *bufio.Reader) {
	// Proses login untuk memasukkan username
	fmt.Print("Masukkan Username: ")
	uname, _ := reader.ReadString('\n')
	username = strings.TrimSpace(uname)
	fmt.Printf("Halo, %s! Selamat datang di aplikasi donasi.\n", username)
}

func mainMenu(reader *bufio.Reader) {
	for {
		// Menampilkan menu donasi setelah login
		fmt.Println("\n==== MENU DONASI ====")
		fmt.Println("1. Berikan Donasi")
		fmt.Println("2. Isi Saldo")
		fmt.Println("3. Keluar")
		fmt.Print("Pilih menu (1-3): ")
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			donate(reader) // Panggil fungsi donasi
		case "2":
			topUpSaldo(reader) // Panggil fungsi top-up saldo
		case "3":
			fmt.Println("Keluar dari menu donasi.") // Keluar dari menu donasi
			return
		default:
			fmt.Println("Pilihan tidak valid, silakan coba lagi.") // Pilihan tidak valid
		}
	}
}

func donate(reader *bufio.Reader) {
	// Cek jika saldo tidak cukup untuk donasi
	if saldo <= 0 {
		fmt.Println("Saldo anda habis. Harap isi saldo terlebih dahulu.")
		return
	}

	// Input nama penerima donasi
	fmt.Print("Masukkan Nama Penerima: ")
	to, _ := reader.ReadString('\n')
	to = strings.TrimSpace(to)

	// Input jumlah donasi
	fmt.Print("Jumlah Donasi: ")
	amountStr, _ := reader.ReadString('\n')
	amountStr = strings.TrimSpace(amountStr)
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil || amount <= 0 {
		fmt.Println("Jumlah donasi tidak valid.")
		return
	}

	// Cek jika jumlah donasi lebih besar dari saldo
	if amount > saldo {
		fmt.Println("Saldo tidak mencukupi untuk donasi ini.")
		return
	}

	// Input pesan donasi
	fmt.Print("Pesan Donasi: ")
	message, _ := reader.ReadString('\n')
	message = strings.TrimSpace(message)

	// Mengirimkan data donasi ke server menggunakan koneksi TCP
	conn, err := net.Dial("tcp", "localhost:8083")
	if err != nil {
		fmt.Println("Gagal terhubung ke server:", err)
		return
	}
	defer conn.Close()

	// Mengirimkan data donasi ke server
	fmt.Fprintf(conn, "%s %f %s\n", username, amount, message)
	saldo -= amount // Memperbarui saldo lokal
	fmt.Printf("Donasi sebesar %.2f berhasil dikirim ke %s.\n", amount, to)
}

func topUpSaldo(reader *bufio.Reader) {
	// Input jumlah saldo yang ingin ditambahkan
	fmt.Print("Masukkan jumlah saldo yang ingin ditambahkan: ")
	amountStr, _ := reader.ReadString('\n')
	amountStr = strings.TrimSpace(amountStr)
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil || amount <= 0 {
		fmt.Println("Jumlah saldo tidak valid.")
		return
	}

	// Mengirimkan informasi top-up saldo ke server menggunakan koneksi TCP
	conn, err := net.Dial("tcp", "localhost:8084")
	if err != nil {
		fmt.Println("Gagal terhubung ke server:", err)
		return
	}
	defer conn.Close()

	// Mengirimkan data top-up ke server
	fmt.Fprintf(conn, "%s %f TOP_UP\n", username, amount)

	saldo += amount // Memperbarui saldo lokal
	fmt.Printf("Saldo berhasil ditambahkan sebesar %.2f. Saldo saat ini: %.2f\n", amount, saldo)
}
