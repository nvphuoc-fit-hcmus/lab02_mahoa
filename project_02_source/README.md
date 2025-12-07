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
project_02_source/
├── client/                      # Mã nguồn Client - Desktop GUI App
│   ├── main.go                  # Entry point - Khởi động Fyne GUI
│   ├── ui/                      # Module giao diện người dùng
│   │   ├── gui.go               # GUI coordinator
│   │   ├── login/               # Module màn hình đăng nhập/đăng ký
│   │   │   └── login_screen.go
│   │   └── notes/               # Module màn hình notes
│   │       └── notes_screen.go
│   ├── api/                     # Module HTTP client
│   │   └── client.go            # API client gọi backend
│   ├── crypto/                  # Module mã hóa
│   │   └── encryption.go        # AES-256-GCM encryption
│   ├── cli/                     # Module CLI (command-line interface)
│   │   └── cli.go               # CLI commands handler
│   └── secure-notes.exe         # Compiled client executable (sau khi build)
├── server/                      # Mã nguồn Backend - RESTful API
│   ├── main.go                  # API server entry point
│   ├── auth/                    # Module xác thực
│   │   ├── jwt.go               # JWT token generation & validation
│   │   └── password.go          # Bcrypt password hashing
│   ├── database/                # Module database
│   │   └── database.go          # SQLite connection & migration
│   ├── handlers/                # Module xử lý HTTP requests
│   │   ├── auth_handler.go      # Login/Register handlers
│   │   ├── note_handler.go      # CRUD operations cho notes
│   │   └── utils.go             # JSON response helpers
│   ├── models/                  # Module data models
│   │   ├── user.go              # User model
│   │   ├── note.go              # Note & SharedLink models
│   │   └── requests.go          # Request/Response structs
│   ├── storage/                 # Database của server (auto-generated)
│   │   └── app.db               # SQLite database file
│   └── server.exe               # Compiled server executable (sau khi build)
├── storage/                     # Thư mục database chung (auto-generated)
│   └── app.db                   # SQLite database file
├── go.mod                       # Quản lý thư viện Go
├── go.sum                       # Checksum các thư viện
├── start.bat                    # Script tự động khởi động (Windows)
├── start.sh                     # Script tự động khởi động (Linux/Mac/Git Bash)
├── build.bat                    # Script build executable (Windows)
├── SRS.md                       # Software Requirements Specification
└── README.md                    # Tài liệu hướng dẫn này
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

### 3. Khởi chạy Server và Client

#### Cách 1: Sử dụng script tự động (Đơn giản nhất - Khuyến nghị)

**Git Bash:**
```bash
./start.sh
```

Script sẽ tự động:
- Khởi động Server trước (port 8080)
- Đợi 2 giây
- Khởi động Client GUI

#### Cách 2: Chạy thủ công từng thành phần

**Terminal 1 - Chạy Server:**
```bash
# Từ thư mục project_02_source
go run server/main.go
```

**Kết quả:** Server sẽ chạy trên `http://localhost:8080`
```
🚀 RESTful API Server is running on http://localhost:8080
```

**Terminal 2 - Chạy Client GUI:**
```bash
# Từ thư mục project_02_source
go run client/main.go
```

Ứng dụng desktop sẽ mở ra với màn hình đăng nhập.

#### Cách 3: Build thành file exe rồi chạy

**Build cả 2 components:**
```cmd
# Windows
build.bat

# Hoặc thủ công
cd server
go build -o server.exe
cd ..

cd client
go build -o secure-notes.exe
cd ..
```

**Chạy file exe:**
```cmd
# Terminal 1 - Chạy Server
cd server
server.exe

# Terminal 2 - Chạy Client
cd client
secure-notes.exe
```

**Lưu ý:** Sau khi build, các file exe sẽ được tạo:
- `server/server.exe` - Backend API server
- `client/secure-notes.exe` - Desktop GUI application

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

Hệ thống có bộ test tự động hoàn chỉnh cho 2 component chính: **Authentication** (Xác thực) và **Access Control** (Giới hạn truy cập).

### ⚠️ Vị trí Test Files

**Lưu ý quan trọng:** Test files được tách riêng ra thư mục `../project_02_test/` để dễ quản lý và không ảnh hưởng đến source code chính.

```
project_02_test/                   # Thư mục test riêng biệt
├── go.mod                         # Module config (link đến source)
├── auth/                          # Test xác thực người dùng (44 tests)
│   ├── register_test.go           # Test đăng ký người dùng
│   ├── login_test.go              # Test đăng nhập
│   ├── password_test.go           # Test hash và verify mật khẩu
│   └── jwt_test.go                # Test JWT token
└── access/                        # Test giới hạn truy cập (20 tests)
    ├── share_access_test.go       # Test share link access control
    └── expired_links_test.go      # Test expired link handling
```

**Tổng cộng:** 64 test cases với coverage đầy đủ cho các chức năng quan trọng.

---

### 🚀 Hướng dẫn Chạy Test

**Lưu ý:** Test files nằm trong thư mục `../project_02_test/`, không phải trong source code. 
Để chạy test, bạn cần di chuyển đến thư mục test:

#### 1. Chạy TẤT CẢ Tests

```bash
# Di chuyển đến thư mục test
cd ../project_02_test

# Chạy toàn bộ test suite (Auth + Access Control)
go test ./... -v

# Chạy với coverage report
go test ./... -cover

# Xuất coverage ra file HTML
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

#### 2. Chạy Test Theo Component

**Authentication Tests (44 tests):**
```bash
# Di chuyển đến thư mục test (nếu chưa)
cd ../project_02_test

# Chạy tất cả auth tests
go test ./auth -v

# Chạy test cụ thể
go test ./auth -run TestRegisterSuccess -v
go test ./auth -run TestLoginSuccess -v
go test ./auth -run TestHashPassword -v
go test ./auth -run TestGenerateJWT -v

# Với coverage
go test ./auth -cover
```

**Access Control Tests (20 tests):**
```bash
# Chạy tất cả access tests
go test ./access -v

# Chạy test cụ thể - Kiểm tra hết hạn
go test ./access -run TestAccessExpiredShareLink -v
go test ./access -run TestShareLinkExpirationBoundary -v

# Chạy test bảo mật
go test ./access -run TestUnauthorizedAccess -v
go test ./access -run TestExpiredShareNoLeakage -v

# Chạy performance test
go test ./access -run TestShareListNotesPerformance -v

# Skip slow tests (time-based tests)
go test ./access -short
```

#### 3. Chạy Test với Options Nâng cao

```bash
# Di chuyển đến thư mục test
cd ../project_02_test

# Chạy với race detector (phát hiện race conditions)
go test ./... -race

# Chạy với verbose output chi tiết
go test ./... -v -json | tee test-results.json

# Chạy với timeout
go test ./... -timeout 30s

# Chạy song song với nhiều CPUs
go test ./... -parallel 4

# Chạy benchmark tests
go test ./... -bench=.
```

---

### 📊 Test Components Chi tiết

#### 1. Test Tự động (Unit Tests)

**✅ Test Đăng ký (register_test.go):**
- ✓ Đăng ký thành công với thông tin hợp lệ
- ✓ Từ chối username đã tồn tại
- ✓ Từ chối username quá ngắn (< 3 ký tự)
- ✓ Từ chối password quá ngắn (< 6 ký tự)
- ✓ Từ chối request body không hợp lệ
- ✓ Từ chối HTTP method sai
- ✓ Xử lý các trường rỗng

**✅ Test Đăng nhập (login_test.go):**
- ✓ Đăng nhập thành công và nhận JWT token
- ✓ Từ chối mật khẩu sai
- ✓ Từ chối username không tồn tại
- ✓ Kiểm tra phân biệt chữ hoa/thường
- ✓ Cho phép đăng nhập nhiều lần
- ✓ Xử lý thông tin đăng nhập rỗng

**✅ Test Mật khẩu (password_test.go):**
- ✓ Hash password với bcrypt
- ✓ Mỗi lần hash tạo salt khác nhau
- ✓ Verify password đúng/sai
- ✓ Phân biệt chữ hoa/thường
- ✓ Hỗ trợ ký tự đặc biệt và Unicode
- ✓ Giới hạn password dài (> 72 bytes)

**✅ Test JWT Token (jwt_test.go):**
- ✓ Tạo JWT token hợp lệ
- ✓ Validate token thành công
- ✓ Từ chối token không hợp lệ/bị sửa đổi
- ✓ Kiểm tra token hết hạn
- ✓ Extract token từ Authorization header
- ✓ Kiểm tra claims (UserID, Username, ExpiresAt)

#### Kết quả Test

```bash
# Kết quả mẫu khi chạy: cd ../project_02_test && go test ./auth -v
=== RUN   TestRegisterSuccess
--- PASS: TestRegisterSuccess (0.21s)
=== RUN   TestLoginSuccess
--- PASS: TestLoginSuccess (0.42s)
=== RUN   TestHashPassword
--- PASS: TestHashPassword (0.21s)
=== RUN   TestGenerateJWT
--- PASS: TestGenerateJWT (0.00s)
...
PASS
ok      project_02_test/auth   17.452s
```

**Tổng cộng:** 44 test cases covering authentication system

---

#### 2. Test Giới hạn Truy cập (Access Control Tests)

Test suite này đảm bảo rằng **các liên kết chia sẻ hết hạn không thể truy cập**, bảo vệ dữ liệu người dùng khỏi truy cập trái phép.

**📁 Vị trí:** `test/access/`

**🎯 Mục đích:**
Kiểm tra tính năng giới hạn truy cập theo thời gian của share links, đảm bảo:
- Liên kết hết hạn **KHÔNG thể truy cập**
- Chỉ liên kết còn hạn mới có thể sử dụng
- Bảo mật dữ liệu người dùng được đảm bảo

**✅ Test Cases (20 tests):**

**Core Access Control Tests (share_access_test.go):**
- ✓ `TestAccessActiveShareLink` - Truy cập liên kết còn hạn
- ✓ `TestAccessExpiredShareLink` ⭐ - Liên kết hết hạn KHÔNG truy cập được
- ✓ `TestMultipleExpiredShareLinks` - Lọc nhiều liên kết hết hạn
- ✓ `TestListNotesWithExpiredShares` - Hiển thị trạng thái IsShared đúng
- ✓ `TestRevokeExpiredShare` - Thu hồi liên kết đã hết hạn
- ✓ `TestCreateShareWithCustomExpiration` - Tạo liên kết với thời gian tùy chỉnh
- ✓ `TestShareLinkExpirationBoundary` ⭐ - Kiểm tra điều kiện `expires_at > NOW()`
- ✓ `TestCleanupExpiredShares` - Dọn dẹp hàng loạt liên kết hết hạn
- ✓ `TestUnauthorizedAccessToExpiredShare` 🔒 - Bảo mật unauthorized access
- ✓ `TestShareLinkTokenUniqueness` - UNIQUE constraint hoạt động đúng

**Advanced Expiration Tests (expired_links_test.go):**
- ✓ `TestExpiredShareLinkAccessViaAPI` - Truy cập qua API endpoint
- ✓ `TestMultipleUsersExpiredShares` - Nhiều users với liên kết hết hạn
- ✓ `TestShareLinkExpirationTransition` ⏱️ - Chuyển đổi active → expired
- ✓ `TestConcurrentShareAccess` 🔀 - 5 goroutines truy cập đồng thời
- ✓ `TestExpiredSharesDoNotAffectActiveNotes` - Owner vẫn truy cập được note
- ✓ `TestExpiredShareDeletion` - Xóa chọn lọc liên kết hết hạn
- ✓ `TestShareExpirationWithDifferentTimezones` 🌍 - Xử lý timezone
- ✓ `TestShareListNotesPerformance` 🚀 - Hiệu năng với 100 notes, 400 shares
- ✓ `TestExpiredShareNoLeakage` 🔒 - Không leak thông tin
- ✓ `TestRevokeAllSharesIncludingExpired` - Thu hồi tất cả shares

**🔑 Logic Kiểm tra Hết hạn:**
```sql
WHERE expires_at > NOW()
```

Điều kiện truy cập:
- `expires_at > NOW()` → ✅ CÒN HẠN (có thể truy cập)
- `expires_at = NOW()` → ❌ HẾT HẠN (không thể truy cập)
- `expires_at < NOW()` → ❌ HẾT HẠN (không thể truy cập)

**🏃 Chạy Access Tests:**

```bash
# Di chuyển đến thư mục test
cd ../project_02_test

# Chạy tất cả access tests
go test ./access -v

# Chạy một test cụ thể
go test ./access -run TestAccessExpiredShareLink -v

# Chạy với coverage
go test ./access -cover

# Skip slow tests (time-based tests)
go test ./access -short
```

**📊 Kết quả Test:**
```bash
=== RUN   TestAccessExpiredShareLink
--- PASS: TestAccessExpiredShareLink (0.05s)
=== RUN   TestShareLinkExpirationBoundary
--- PASS: TestShareLinkExpirationBoundary (0.06s)
=== RUN   TestUnauthorizedAccessToExpiredShare
--- PASS: TestUnauthorizedAccessToExpiredShare (0.11s)
...
PASS
ok      project_02_test/access 4.547s
```

**✅ Kết quả:** Tất cả 20 tests PASS - Giới hạn truy cập hoạt động đúng!

**🔍 Test Coverage:**
- ✅ Security: Unauthorized access, information leakage
- ✅ Performance: Concurrent access, bulk operations (1.2ms cho 100 notes)
- ✅ Edge Cases: Boundary times, timezone handling
- ✅ Database: Constraints, cleanup, transactions

**Tổng cộng:** 20 test cases covering access control system

---

### Test Tất cả Components

Để chạy toàn bộ test suite (Authentication + Access Control):

```bash
# Di chuyển đến thư mục test
cd ../project_02_test

# Chạy tất cả tests
go test ./... -v

# Chạy với coverage report
go test ./... -cover -coverprofile=coverage.out

# Xem coverage chi tiết
go tool cover -html=coverage.out

# Chạy theo thư mục
go test ./auth -v    # Chỉ auth tests
go test ./access -v  # Chỉ access tests
```

**📊 Tổng kết Test Suite:**
- **Authentication Tests:** 44 test cases
- **Access Control Tests:** 20 test cases
- **Tổng cộng:** 64 test cases
- **Status:** ✅ ALL TESTS PASSING

---

### Test Thủ công (Manual Testing)

Để kiểm tra các tính năng thủ công, bạn có thể:

1. **Test Authentication:**
   ```bash
   # Khởi động server
   go run server/main.go
   
   # Khởi động client GUI
   go run client/main.go
   
   # Thử đăng ký và đăng nhập
   ```

2. **Test Encryption:**
   - Tạo note mới trong GUI
   - Kiểm tra dữ liệu trong database (storage/app.db) đã được mã hóa

3. **Test API với curl:**
   ```bash
   # Đăng ký
   curl -X POST http://localhost:8080/api/auth/register \
     -H "Content-Type: application/json" \
     -d '{"username":"testuser","password":"password123"}'
   
   # Đăng nhập
   curl -X POST http://localhost:8080/api/auth/login \
     -H "Content-Type: application/json" \
     -d '{"username":"testuser","password":"password123"}'
   ```