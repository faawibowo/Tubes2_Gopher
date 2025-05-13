# Tubes2 Gopher - DFS & BFS Element Finder

## 🧠 Penjelasan Singkat Algoritma DFS dan BFS

### Proyek ini mengimplementasikan dua algoritma pencarian graf:

- DFS (Depth-First Search): Menelusuri pohon resep secara mendalam untuk menemukan semua kombinasi bahan dasar yang membentuk elemen target. Cocok untuk eksplorasi menyeluruh.

- BFS (Breadth-First Search): Menelusuri pohon secara melebar untuk menemukan jalur tercepat atau terpendek dalam membentuk elemen target. Cocok untuk menemukan solusi optimal dengan langkah minimal.

## ⚙️ Persyaratan Sistem dan Instalasi

### 1. Menggunakan Docker (Direkomendasikan)

- Docker dan Docker Compose harus terinstal di sistem Anda.
- Untuk panduan instalasi Docker:
  - https://docs.docker.com/desktop/install/windows-install/
  - https://docs.docker.com/desktop/install/mac-install/
  - https://docs.docker.com/engine/install/

### 2. Tanpa Docker (Manual)

- Backend:
  - Go versi 1.21 atau lebih baru.
  - Untuk panduan instalasi Go: https://go.dev/doc/install
- Frontend:
  - Node.js dan npm.
  - Untuk panduan instalasi Node.js dan npm: https://nodejs.org/en/download/

Catatan: Persyaratan hanya diperlukan jika ingin dikompilasi secara lokal dan tidak diperlukan jika mnengakses melalui web.

## 🛠️ Langkah-Langkah Build dan Menjalankan Aplikasi

💡 Sebelum menjalankan aplikasi (baik menggunakan Docker maupun manual), pastikan Anda telah mengatur file `.env`:

- Copy file `.env.example` masing-masing folder menjadi `.env` di:
  - Root folder: /.env
  - Frontend folder: /web/.env

### 1. Menggunakan Docker (Direkomendasikan)

```
docker compose down --remove-orphans
docker compose --env-file .env up --build
```

Setelah proses selesai, buka aplikasi di browser melalui:
http://localhost:8080

### 2. Menjalankan Secara Manual (Tanpa Docker)

Backend (Go):

```
cd backend
go mod tidy
go run ./cmd
```

Frontend (npm):

```
cd web
npm install
npm run dev
```

Setelah kedua server berjalan, buka aplikasi di browser melalui:
http://localhost:3000

### 3. Akses Langsung

https://p01--app-frontend--7jx42hjyr69l.code.run/

Catatan: Pastikan Docker dan npm telah diinstal dan dikonfigurasi dengan benar sebelum menjalankan perintah di atas. Jika belum, silakan merujuk pada panduan instalasi yang telah disediakan di bagian Persyaratan Sistem dan Instalasi.

## 👤 Author

| Nama                   | NIM      |
| ---------------------- | -------- |
| Brian Ricardo Tamin    | 13523126 |
| Nathanael Rachmat      | 13523142 |
| Muhammad Farrel Wibowo | 13523153 |
