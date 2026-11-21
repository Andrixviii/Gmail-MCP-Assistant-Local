# Gmail MCP Server

MCP Server yang menghubungkan AI Assistant dengan akun Gmail menggunakan Gmail API dan Model Context Protocol (MCP). Project ini memungkinkan AI Assistant untuk mengambil email terbaru, mencari email berdasarkan kriteria tertentu, dan membaca isi email secara langsung melalui tools yang disediakan.

## Features

Project ini menyediakan tiga tools utama:

| Tool | Description |
|---|---|
| `get_recent_emails` | Mengambil daftar email terbaru dari Gmail. |
| `search_emails` | Mencari email berdasarkan query Gmail. |
| `read_email` | Membaca detail dan isi email berdasarkan Message ID. |

### Technology Stack

- **Go** – Backend programming language.
- **Model Context Protocol (MCP)** – Communication protocol antara AI Assistant dan MCP Server.
- **Gmail API** – Mengakses data email dari akun Gmail.
- **OAuth 2.0** – Authentication dan authorization.
- **OpenCode** – AI coding agent yang digunakan sebagai MCP client.

## Project Structure

```text
gmail-mcp/
├── auth.go
├── gmail.go
├── main.go
├── credentials.json
├── token.json
├── go.mod
├── go.sum
├── .gitignore
└── README.md
```

| File | Description |
|---|---|
| `main.go` | Inisialisasi MCP Server dan registrasi tools. |
| `auth.go` | Mengelola OAuth 2.0 dan autentikasi Gmail API. |
| `gmail.go` | Mengelola pengambilan, pencarian, dan pembacaan email. |
| `credentials.json` | Kredensial OAuth dari Google Cloud. |
| `token.json` | Token autentikasi yang dihasilkan setelah login pertama. |

## Requirements

Pastikan perangkat sudah memiliki:

- Go (sesuai versi pada `go.mod`).
- Akun Google.
- Google Cloud Project.
- OpenCode atau MCP client lain.
- Koneksi internet.

## Installation

Clone repository:

```bash
git clone https://github.com/Andrixviii/Gmail-MCP-Assistant-test.git
```

Masuk ke direktori project:

```bash
cd gmail-mcp
```

Download dependencies:

```bash
go mod tidy
```

Build project:

```bash
go build -o gmail-mcp .
```

Setelah proses build berhasil, binary `gmail-mcp` akan tersedia di direktori project.

## Google Cloud Configuration

Sebelum menggunakan Gmail MCP, aplikasi perlu mendapatkan akses ke Gmail API melalui Google Cloud.

### 1. Create Google Cloud Project

1. Buka [Google Cloud Console](https://console.cloud.google.com/).
2. Buat project baru atau pilih project yang sudah tersedia.
3. Buka bagian API Library.
4. Cari **Gmail API**.
5. Aktifkan Gmail API.

### 2. Configure OAuth Consent Screen

1. Buka menu Google Auth Platform.
2. Lengkapi informasi aplikasi.
3. Atur Audience sesuai kebutuhan.
4. Jika menggunakan External dengan status Testing, tambahkan akun Gmail yang akan digunakan ke bagian Test Users.

### 3. Create OAuth Client ID

1. Buka menu Clients.
2. Pilih Create Client.
3. Pilih Application Type: **Desktop app**.
4. Berikan nama aplikasi.
5. Buat OAuth Client ID.
6. Download file JSON yang dihasilkan.

Simpan file tersebut dengan nama:

```text
credentials.json
```

Letakkan file di direktori project.

### 4. Configure OAuth Scope

Project ini menggunakan scope:

```text
https://www.googleapis.com/auth/gmail.readonly
```

Scope tersebut memberikan izin untuk membaca email tanpa memberikan akses untuk mengirim, menghapus, atau mengubah email.

## Authentication

Autentikasi menggunakan OAuth 2.0.

Pada saat pertama kali MCP Server dijalankan:

1. Aplikasi membaca `credentials.json`.
2. Aplikasi membuat URL autentikasi Google.
3. Browser akan terbuka untuk proses login.
4. Pengguna memberikan izin akses Gmail.
5. Google mengarahkan kembali ke OAuth callback lokal.
6. Aplikasi menerima authorization code dan menukarnya dengan token.
7. Token disimpan dalam file `token.json`.

Pada penggunaan berikutnya, aplikasi akan menggunakan token yang sudah tersimpan sehingga pengguna tidak perlu melakukan login ulang selama token masih valid.

## Configuration

Project mendukung konfigurasi direktori menggunakan environment variable:

```text
GMAIL_MCP_CONFIG_DIR
```

Environment variable ini digunakan untuk menentukan lokasi penyimpanan `credentials.json` dan `token.json`.

Jika environment variable tidak diatur, aplikasi menggunakan working directory saat dijalankan.

Contoh direktori:

```text
/home/user/Project/gmail-mcp
```

### OpenCode Configuration

Tambahkan konfigurasi berikut ke file konfigurasi OpenCode:

```json
{
  "mcp": {
    "gmail": {
      "type": "local",
      "command": [
        "/path/to/gmail-mcp/gmail-mcp"
      ],
      "environment": {
        "GMAIL_MCP_CONFIG_DIR": "/path/to/gmail-mcp"
      },
      "enabled": true
    }
  }
}
```

Sesuaikan `/path/to/gmail-mcp` dengan lokasi project pada perangkat.

Setelah konfigurasi disimpan, restart OpenCode agar MCP Server dapat dimuat.

## Available Tools

### 1. get_recent_emails

Mengambil email terbaru berdasarkan query Gmail.

**Arguments:**

```json
{
  "query": "in:inbox",
  "limit": 5
}
```

**Parameters:**

| Parameter | Type | Description |
|---|---|---|
| `query` | string | Query pencarian Gmail. |
| `limit` | integer | Jumlah maksimal email yang diambil. |

**Example queries:**

```text
in:inbox
is:unread
newer_than:7d
from:example@gmail.com
```

### 2. search_emails

Mencari email berdasarkan kriteria tertentu menggunakan Gmail Search Operators.

**Arguments:**

```json
{
  "query": "subject:bimbingan",
  "limit": 5
}
```

**Example queries:**

```text
from:example@gmail.com
subject:bimbingan
has:attachment
newer_than:7d
is:unread
```

Dokumentasi operator pencarian Gmail:

[ Gmail Search Operators ](https://support.google.com/mail/answer/7190)

### 3. read_email

Membaca detail dan isi email berdasarkan Message ID.

**Arguments:**

```json
{
  "message_id": "YOUR_MESSAGE_ID"
}
```

Message ID diperoleh dari hasil `get_recent_emails` atau `search_emails`.

Tool ini mengembalikan informasi seperti:

- Sender
- Recipient
- Subject
- Date
- Email snippet
- Email body

## Testing with MCP Inspector

Project dapat diuji menggunakan MCP Inspector sebelum diintegrasikan dengan AI Assistant.

Jalankan MCP Inspector:

```bash
npx @modelcontextprotocol/inspector
```

Kemudian hubungkan Inspector dengan binary MCP Server dan uji setiap tool melalui bagian Tools.

Contoh pengujian:

**get_recent_emails**

```json
{
  "query": "in:inbox",
  "limit": 5
}
```

**search_emails**

```json
{
  "query": "subject:bimbingan",
  "limit": 5
}
```

**read_email**

```json
{
  "message_id": "YOUR_MESSAGE_ID"
}
```

## Security

Beberapa hal yang perlu diperhatikan:

- Jangan membagikan `credentials.json` kepada pihak lain.
- Jangan membagikan `token.json`.
- Jangan mengunggah file kredensial atau token ke repository publik.
- Gunakan OAuth scope sesuai kebutuhan aplikasi.
- Perlakukan isi email sebagai data eksternal yang tidak tepercaya.

Pastikan `.gitignore` berisi:

```gitignore
credentials.json
token.json
gmail-mcp
.env
```

## Limitations

- Project hanya mendukung operasi membaca email.
- Tidak menyediakan fitur mengirim email atau membuat draft (untuk menghindari ai menulis tanpa izin).
- Tidak mendukung pengelolaan label maupun penghapusan email.
- Membutuhkan koneksi internet untuk berkomunikasi dengan Gmail API.
- Isi email yang diproses AI dapat menggunakan context/token dari model AI yang digunakan.



---

**Gmail MCP Server** — A personal Gmail integration powered by Go, Gmail API, and Model Context Protocol.