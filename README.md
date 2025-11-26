# Lab02: Ứng dụng Chia sẻ Ghi chú Bảo mật (Secure Note Sharing)

Đây là đồ án xây dựng hệ thống chia sẻ ghi chú bảo mật theo mô hình **Client-Server**. Hệ thống được thiết kế theo kiến trúc **Zero-Knowledge Privacy** (Kiến thức bằng không), đảm bảo rằng Server chỉ đóng vai trò lưu trữ dữ liệu đã mã hóa và không bao giờ có khả năng truy cập hoặc đọc được nội dung gốc của người dùng.

---

## 📋 Tính năng của Hệ thống

Dựa trên yêu cầu của bài tập Lab02, ứng dụng bao gồm các tính năng cốt lõi sau:

### 1. Xác thực & Quản lý phiên (Authentication)

- **Đăng ký & Đăng nhập:** Người dùng cần tạo tài khoản để sử dụng hệ thống
- **Bảo mật mật khẩu:** Mật khẩu được băm (Hashing) kết hợp với Salt trước khi lưu vào cơ sở dữ liệu. Server tuyệt đối không lưu mật khẩu dạng văn bản rõ
- **Quản lý phiên:** Sử dụng **JWT (JSON Web Token)** để xác thực và duy trì phiên làm việc an toàn cho người dùng mà không cần gửi lại mật khẩu nhiều lần

### 2. Mã hóa phía Client (Client-side Encryption)

- **Mã hóa dữ liệu:** Sử dụng thuật toán **AES-GCM** để mã hóa toàn bộ ghi chú ngay tại máy người dùng trước khi tải lên Server
- **Quản lý khóa:** Mỗi ghi chú được mã hóa bằng một khóa riêng biệt để tăng cường bảo mật. Khóa này sau đó được bảo vệ bằng mật khẩu của người dùng
- **Bảo mật dữ liệu:** Server chỉ nhận được chuỗi mã hóa (ciphertext), ngăn chặn rủi ro rò rỉ dữ liệu từ phía máy chủ

### 3. Chia sẻ qua URL có giới hạn (Time-sensitive Access)

- Cho phép người dùng tạo đường dẫn chia sẻ (URL) tạm thời cho ghi chú
- **Cơ chế bảo mật URL:** Khóa giải mã được đặt trong phần **Fragment** của URL (phần sau dấu `#`). Trình duyệt hoặc Client sẽ đọc phần này để giải mã, nhưng phần này **không bao giờ được gửi lên Server** qua HTTP Request
- **Kiểm soát thời gian:** Server thực thi quy tắc metadata, tự động chặn truy cập nếu liên kết đã quá thời gian hết hạn

### 4. Chia sẻ Mã hóa đầu cuối (End-to-End Encryption - E2EE)

- Hỗ trợ chia sẻ dữ liệu riêng tư giữa hai người dùng cụ thể
- Sử dụng thuật toán trao đổi khóa **Diffie-Hellman** để tạo ra một **Khóa phiên (Session Key)** duy nhất giữa người gửi và người nhận
- Khóa này được sinh ra tại máy người dùng và sẽ bị hủy sau khi phiên làm việc kết thúc

---

## 📂 Cấu trúc Thư mục

Dự án được tổ chức theo cấu trúc phân tách rõ ràng giữa Client và Server:

```
lab02_mahoa/
├── client/              # Mã nguồn Client - Desktop GUI App
│   ├── main.go          # Entry point - Khởi động Fyne GUI
│   ├── ui/              # Module giao diện người dùng
│   │   ├── gui.go       # GUI coordinator
│   │   ├── login/       # Module màn hình đăng nhập/đăng ký
│   │   │   └── login_screen.go
│   │   └── notes/       # Module màn hình notes
│   │       └── notes_screen.go
│   ├── api/             # Module HTTP client
│   │   └── client.go    # API client gọi backend
│   └── crypto/          # Module mã hóa
│       └── encryption.go # AES-256-GCM encryption
├── server/              # Mã nguồn Backend - RESTful API
│   ├── main.go          # API server entry point
│   ├── auth/            # Module xác thực
│   │   ├── jwt.go       # JWT token generation & validation
│   │   └── password.go  # Bcrypt password hashing
│   ├── database/        # Module database
│   │   └── database.go  # SQLite connection & migration
│   ├── handlers/        # Module xử lý HTTP requests
│   │   ├── auth_handler.go # Login/Register handlers
│   │   ├── note_handler.go # CRUD operations cho notes
│   │   └── utils.go        # JSON response helpers
│   └── models/          # Module data models
│       ├── user.go      # User model
│       ├── note.go      # Note & SharedLink models
│       └── requests.go  # Request/Response structs
├── storage/             # Thư mục chứa Database (auto-generated)
│   └── app.db           # SQLite database file
├── go.mod               # Quản lý thư viện Go
├── go.sum               # Checksum các thư viện
├── start.bat            # Script tự động khởi động (Windows)
├── start.sh             # Script tự động khởi động (Linux/Mac/Git Bash)
├── build.bat            # Script build executable
└── README.md            # Tài liệu hướng dẫn này
```

---

## 🛠️ Công nghệ sử dụng

### Backend (Server)
- **Go (Golang)** 1.20+
- **SQLite** với GORM ORM
- **JWT** authentication (`github.com/golang-jwt/jwt/v5`)
- **Bcrypt** password hashing (`golang.org/x/crypto/bcrypt`)
- **RESTful API** với CORS middleware

### Frontend (Client)
- **Fyne v2.7** - Modern cross-platform GUI framework
- **AES-256-GCM** encryption (`crypto/aes`, `crypto/cipher`)
- **HTTP Client** - Gọi API backend
- **Desktop App** - Native Windows/Linux/macOS

---

## 🚀 Hướng dẫn Cài đặt & Sử dụng

### 1. Yêu cầu Môi trường (Prerequisites)

Trước khi bắt đầu, hãy đảm bảo máy tính của bạn đã cài đặt:

- **Go (Golang):** Phiên bản 1.20 trở lên
- **Git Bash:** Để chạy script `start.sh` trên Windows (tùy chọn - có thể dùng `start.bat` thay thế)

#### Cách cài đặt Go trên Windows

**Nếu chưa có Go, hãy làm theo các bước sau:**

1. **Tải Go từ trang chính thức:**
   - Truy cập: https://golang.org/dl/
   - Chọn phiên bản Windows (tìm file có tên `go1.x.x.windows-amd64.msi`)

2. **Cài đặt:**
   - Nhấp đôi vào file `.msi` vừa tải
   - Làm theo hướng dẫn cài đặt (thường cài vào `C:\Program Files\Go`)
   - Nhấn "Finish" để hoàn thành

3. **Khởi động lại Terminal/CMD:**
   - Đóng cửa sổ cmd/PowerShell hiện tại
   - Mở cmd/PowerShell mới để Go có sẵn trong `PATH`

4. **Kiểm tra cài đặt:**
   ```cmd
   go version
   ```
   
   Nếu thành công, bạn sẽ thấy phiên bản Go được cài đặt

**Cách cài đặt Go trên macOS/Linux:**

   ```bash
   # macOS (sử dụng Homebrew)
   brew install go
   
   # Linux
   wget https://golang.org/dl/go1.21.0.linux-amd64.tar.gz
   sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
   export PATH=$PATH:/usr/local/go/bin
   ```

Kiểm tra cài đặt bằng lệnh: `go version`

### 2. Thiết lập Dự án

Mở terminal tại thư mục gốc của dự án và chạy lệnh sau để tải các thư viện cần thiết:

```bash
go mod tidy
```

Lệnh này sẽ tự động đọc file `go.mod` và tải các dependencies về máy.

### 3. Khởi chạy Server và CLient

**Cách 1: Sử dụng script tự động (Đơn giản nhất)**

-Chạy `./start.sh` trong Git Bash 

**Cách 2: Chạy thủ công**

Mở Terminal đầu tiên và chạy Server:

```bash
cd c:\Users\Admin\lab02_mahoa
go run server/main.go server/auth.go server/db.go server/handlers.go server/models.go
```

**Cách 3: Build thành exe rồi chạy**

```bash
# Build
cd server
go build -o server.exe

# Chạy
./server.exe
```

**Kết quả:** Bạn sẽ thấy thông báo:
```
🚀 RESTful API Server is running on http://localhost:8080
```

Giữ Terminal này mở để Server tiếp tục chạy.

---

## 📝 Lưu ý Bảo mật

- **Không bao giờ chia sẻ mật khẩu** hoặc private key
- **URL chia sẻ có thời hạn** - hãy chuẩn bị sẵn trước khi người nhận lấy dữ liệu
- **Xóa dữ liệu nhạy cảm** sau khi không cần sử dụng
- **Kiểm tra chứng chỉ SSL/TLS** khi triển khai trên production
- **Giữ bí mật JWT Token** - Không chia sẻ token với người khác

---

## 🔗 API Endpoints

Dưới đây là các endpoint REST API mà Server cần implement:

### Authentication (Xác thực)
| Method | Endpoint | Mô tả |
|--------|----------|-------|
| POST | `/auth/register` | Đăng ký tài khoản mới |
| POST | `/auth/login` | Đăng nhập và lấy JWT Token |
| POST | `/auth/logout` | Đăng xuất |

### Notes Management (Quản lý ghi chú)
| Method | Endpoint | Mô tả |
|--------|----------|-------|
| POST | `/notes/upload` | Tải lên ghi chú mã hóa |
| GET | `/notes/list` | Lấy danh sách ghi chú của người dùng |
| GET | `/notes/:id` | Lấy ghi chú theo ID |
| DELETE | `/notes/:id` | Xóa ghi chú |

### Sharing (Chia sẻ)
| Method | Endpoint | Mô tả |
|--------|----------|-------|
| POST | `/share/public` | Tạo link chia sẻ công khai có thời hạn |
| GET | `/share/:shareId` | Lấy dữ liệu từ link chia sẻ |
| POST | `/share/e2ee` | Tạo chia sẻ E2EE với người dùng khác |

---

## 💾 Cấu trúc Database

**SQLite Database: `storage/app.db`**

### Bảng Users
```sql
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### Bảng Notes
```sql
CREATE TABLE notes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    encrypted_content TEXT NOT NULL,
    iv TEXT NOT NULL,  -- Initialization Vector
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);
```

### Bảng SharedLinks
```sql
CREATE TABLE shared_links (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    note_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    share_token TEXT UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(note_id) REFERENCES notes(id) ON DELETE CASCADE,
    FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);
```

---

## 🔐 Quy trình Mã hóa & Giải mã

### Quy trình Mã hóa (Client → Server)
1. Người dùng nhập nội dung ghi chú
2. Client sinh khóa AES ngẫu nhiên
3. Client mã hóa nội dung bằng AES-GCM
4. Client tạo IV (Initialization Vector) ngẫu nhiên
5. Client gửi dữ liệu mã hóa + IV lên Server (nội dung gốc không gửi)
6. Server lưu trữ ciphertext + IV

### Quy trình Giải mã (Server → Client)
1. Server gửi ciphertext + IV cho Client
2. Client sử dụng khóa AES để giải mã
3. Client hiển thị nội dung gốc cho người dùng

---

## ❓ Troubleshooting (Giải quyết Sự cố)

### 1. Lỗi: "go: go.mod file not found"
**Giải pháp:**
```bash
go mod init lab02_mahoa
go mod tidy
```

### 2. Lỗi: "cannot find module"
**Giải pháp:**
```bash
go mod download
go mod verify
go mod tidy
```

### 3. Lỗi: "Server address already in use"
**Giải pháp:** Port 8080 đang được sử dụng
```bash
# Tìm process đang dùng port 8080
netstat -ano | findstr :8080

# Hoặc thay đổi port trong code Server
```

### 4. Lỗi: "database is locked"
**Giải pháp:** Đóng các instance khác của Server hoặc Client đang truy cập database

### 5. Lỗi: "invalid token"
**Giải pháp:** Token JWT hết hạn hoặc không hợp lệ
- Đăng nhập lại: `go run client/*.go login -u [username] -p [password]`

---

## 📊 Sơ đồ Kiến trúc

```
┌─────────────────────────────────────────────────────────────┐
│           CLIENT - Fyne Desktop GUI App                      │
├─────────────────────────────────────────────────────────────┤
│  ┌──────────────────────────────────────────────────────┐  │
│  │  • main.go           - Khởi động Fyne app           │  │
│  │  • ui/gui.go         - GUI coordinator              │  │
│  │  • ui/login/         - Login/Register screen        │  │
│  │  • ui/notes/         - Notes screen                 │  │
│  │  • api/client.go     - HTTP client gọi API backend  │  │
│  │  • crypto/encryption.go - AES-256-GCM encryption    │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                              │
│  Tính năng:                                                  │
│  ✓ Login/Register UI                                         │
│  ✓ Notes Manager với Create/View/Delete                     │
│  ✓ Client-side encryption (Zero-Knowledge)                   │
│  ✓ JWT token management                                      │
└───────────────────────────┬──────────────────────────────────┘
                            │
                            │ RESTful API (HTTP/JSON)
                            │ CORS enabled
                            │
┌───────────────────────────▼──────────────────────────────────┐
│              SERVER - RESTful API Backend                     │
├──────────────────────────────────────────────────────────────┤
│  ┌──────────────────────────────────────────────────────┐  │
│  │  • main.go                - API server với CORS      │  │
│  │  • auth/jwt.go            - JWT generation          │  │
│  │  • auth/password.go       - Bcrypt hashing          │  │
│  │  • database/database.go   - SQLite + GORM setup     │  │
│  │  • handlers/auth_handler.go - Auth endpoints        │  │
│  │  • handlers/note_handler.go - Notes endpoints       │  │
│  │  • handlers/utils.go      - JSON helpers            │  │
│  │  • models/*               - Data structures         │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                              │
│  API Endpoints:                                              │
│  • POST   /api/auth/register                                 │
│  • POST   /api/auth/login                                    │
│  • POST   /api/notes          (JWT required)                 │
│  • GET    /api/notes          (JWT required)                 │
│  • GET    /api/notes/:id      (JWT required)                 │
│  • DELETE /api/notes/:id      (JWT required)                 │
└───────────────────────────┬──────────────────────────────────┘
                            │
                            ↓
                  ┌──────────────────┐
                  │  SQLite Database │
                  │  (storage/app.db)│
                  │                  │
                  │  • users         │
                  │  • notes         │
                  │  • shared_links  │
                  └──────────────────┘
                    (Encrypted Data)
```

---

## 🧪 Testing

Để kiểm tra các tính năng, bạn có thể:

1. **Test Authentication:**
   ```bash
   go run client/*.go register -u testuser -p password123
   go run client/*.go login -u testuser -p password123
   ```

2. **Test Encryption:**
   ```bash
   echo "Đây là nội dung bí mật" > test.txt
   go run client/*.go upload -f test.txt
   ```

3. **Test Sharing:**
   ```bash
   go run client/*.go share -id 1 -time 60
   # Chia sẻ URL với người khác
   ```