# Thiết Kế và Kiến Trúc Hệ Thống

---

### Sơ đồ kiến trúc tổng quát

```
┌─────────────┐        HTTPS/REST API        ┌─────────────┐
│   CLIENT    │ <-------------------------> │   SERVER    │
│ (Fyne GUI,  │                            │ (Go, SQLite │
│  CLI, API)  │                            │  GORM, JWT) │
└─────┬───────┘                            └─────┬───────┘
      │                                         │
      │  Mã hóa AES-256-GCM, PBKDF2, ECDH       │
      │                                         │
      ▼                                         ▼
┌─────────────┐                        ┌─────────────────┐
│  Người dùng │                        │  Lưu trữ dữ liệu│
└─────────────┘                        │  đã mã hóa      │
                                       └─────────────────┘
```

---

### Sơ đồ kiến trúc hệ thống chi tiết (đầy đủ các lớp)

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                                 CLIENT                                     │
├─────────────────────────────────────────────────────────────────────────────┤
│ ┌──────────────────────────────────────────────────────────────────────────┐ │
│ │ 1. PRESENTATION LAYER (UI)                                              │ │
│ │ ┌──────────────┐   ┌──────────────┐                                     │ │
│ │ │ Login Screen │   │ Notes Screen │  (Fyne GUI)                         │ │
│ │ └──────┬───────┘   └──────┬───────┘                                     │ │
│ │        └────────────┬─────────────────┘                                 │ │
│ └─────────────────────┼──────────────────────────────────────────────────┘ │
│ ┌─────────────────────▼──────────────────────────────────────────────────┐ │
│ │ 2. BUSINESS LOGIC LAYER (CLI)                                          │ │
│ │ ┌──────────────┐   ┌──────────────┐   ┌──────────────┐                 │ │
│ │ │ User Input   │   │ Command Parse│   │ Note Manager │                 │ │
│ │ └──────┬───────┘   └──────┬───────┘   └──────┬───────┘                 │ │
│ │        └────────────┬─────────────────────────┘                         │ │
│ └─────────────────────┼──────────────────────────────────────────────────┘ │
│ ┌─────────────────────▼──────────────────────────────────────────────────┐ │
│ │ 3. CRYPTO LAYER (encryption.go, diffie_hellman.go, keystore.go)        │ │
│ │ ┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐            │ │
│ │ │ AES-256-GCM     │ │ PBKDF2          │ │ ECDH X25519     │            │ │
│ │ │ Encryption      │ │ Key Derivation  │ │ Key Exchange    │            │ │
│ │ └─────────────────┘ └─────────────────┘ └─────────────────┘            │ │
│ │ ┌─────────────────────────────────────────────────────────┐             │ │
│ │ │ Keystore: Lưu trữ local key, password, JWT token        │             │ │
│ │ └─────────────────────────────────────────────────────────┘             │ │
│ └─────────────────────┼──────────────────────────────────────────────────┘ │
│ ┌─────────────────────▼──────────────────────────────────────────────────┐ │
│ │ 4. API CLIENT LAYER (client.go, api/)                                  │ │
│ │ ┌──────────────────────────────────────────────────────────┐            │ │
│ │ │ HTTP Client: net/http                                   │            │ │
│ │ │ Endpoints:                                              │            │ │
│ │ │   POST /register    - Đăng ký tài khoản                │            │ │
│ │ │   POST /login       - Đăng nhập, nhận JWT              │            │ │
│ │ │   POST /note        - Tạo ghi chú (ciphertext + JWT)   │            │ │
│ │ │   GET  /note/{id}   - Lấy ghi chú (JWT)               │            │ │
│ │ │   POST /share       - Tạo link chia sẻ (JWT)           │            │ │
│ │ │   GET  /share/{token} - Lấy ghi chú chia sẻ (token)   │            │ │
│ │ │   POST /e2ee        - Chia sẻ E2EE (public key + JWT)   │            │ │
│ │ │   GET  /e2ee/{id}   - Lấy E2EE (public key)            │            │ │
│ │ └──────────────────────────────────────────────────────────┘            │ │
│ └─────────────────────┬──────────────────────────────────────────────────┘ │
│                       │                                                    │
│ ┌─────────────────────▼──────────────────────────────────────────────────┐ │
│ │ 5. STORAGE LAYER (Local)                                               │ │
│ │ ┌─────────────────────────────────────────────────────────┐            │ │
│ │ │ File: JWT token, user credentials, decrypted notes      │            │ │
│ │ └─────────────────────────────────────────────────────────┘            │ │
│ └─────────────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────────┘
                              │
                              ▼
                     ┌────────────────┐
                     │   HTTPS/TLS    │
                     │   (Port 8080)  │
                     └────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                                 SERVER                                     │
├─────────────────────────────────────────────────────────────────────────────┤
│ ┌──────────────────────────────────────────────────────────────────────────┐ │
│ │ 1. ROUTER & MIDDLEWARE LAYER (main.go)                                  │ │
│ │ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐    │ │
│ │ │ CORS         │ │ Auth JWT     │ │ Error        │ │ Logging      │    │ │
│ │ │ Middleware   │ │ Middleware   │ │ Handler      │ │ Middleware   │    │ │
│ │ └──────────────┘ └──────────────┘ └──────────────┘ └──────────────┘    │ │
│ └──────────────────────────────────────────────────────────────────────────┘ │
│ ┌──────────────────────────────────────────────────────────────────────────┐ │
│ │ 2. HANDLER LAYER (handlers/)                                            │ │
│ │ ┌──────────────────┐ ┌──────────────────┐ ┌──────────────────┐          │ │
│ │ │ AuthHandler      │ │ NoteHandler      │ │ ShareHandler     │          │ │
│ │ │ - Register       │ │ - CreateNote     │ │ - CreateShare    │          │ │
│ │ │ - Login          │ │ - GetNote        │ │ - GetSharedNote  │          │ │
│ │ │ - ValidateJWT    │ │ - DeleteNote     │ │ - CheckToken     │          │ │
│ │ └──────────────────┘ └──────────────────┘ └──────────────────┘          │ │
│ │ ┌──────────────────┐ ┌──────────────────┐                               │ │
│ │ │ E2EEHandler      │ │ PublicKeyHandler │                               │ │
│ │ │ - CreateE2EE     │ │ - GetPublicKey   │                               │ │
│ │ │ - GetE2EE        │ │ - StorePublicKey │                               │ │
│ │ │ - ShareE2EE      │ └──────────────────┘                               │ │
│ │ └──────────────────┘                                                    │ │
│ └──────────────────────────────────────────────────────────────────────────┘ │
│ ┌──────────────────────────────────────────────────────────────────────────┐ │
│ │ 3. MODEL & BUSINESS LOGIC LAYER (models/)                               │ │
│ │ ┌──────────────────┐ ┌──────────────────┐ ┌──────────────────┐          │ │
│ │ │ User Model       │ │ Note Model       │ │ Share Model      │          │ │
│ │ │ - ID             │ │ - ID             │ │ - ID             │          │ │
│ │ │ - Username       │ │ - UserID         │ │ - NoteID         │          │ │
│ │ │ - Password Hash  │ │ - Ciphertext     │ │ - Token          │          │ │
│ │ │ - CreatedAt      │ │ - EncryptedKey   │ │ - ExpireAt       │          │ │
│ │ │                  │ │ - IV             │ │ - AccessCount    │          │ │
│ │ └──────────────────┘ │ - CreatedAt      │ │ - Password       │          │ │
│ │ ┌──────────────────┐ └──────────────────┘ └──────────────────┘          │ │
│ │ │ E2EEShare Model  │ ┌──────────────────┐                               │ │
│ │ │ - ID             │ │ Request Models   │                               │ │
│ │ │ - NoteID         │ │ - RegisterReq    │                               │ │
│ │ │ - RecipientID    │ │ - LoginReq       │                               │ │
│ │ │ - Ciphertext     │ │ - CreateNoteReq  │                               │ │
│ │ │ - PublicKeyA     │ │ - ShareReq       │                               │ │
│ │ │ - PublicKeyB     │ │ - E2EEReq        │                               │ │
│ │ └──────────────────┘ └──────────────────┘                               │ │
│ └──────────────────────────────────────────────────────────────────────────┘ │
│ ┌──────────────────────────────────────────────────────────────────────────┐ │
│ │ 4. AUTH LAYER (auth/)                                                   │ │
│ │ ┌──────────────────┐ ┌──────────────────┐                               │ │
│ │ │ JWT (v5.2.0)     │ │ Password Hash    │                               │ │
│ │ │ - Sign Token     │ │ - Bcrypt         │                               │ │
│ │ │ - Verify Token   │ │ - HashPassword   │                               │ │
│ │ │ - GetUserFromJWT │ │ - VerifyPassword │                               │ │
│ │ └──────────────────┘ └──────────────────┘                               │ │
│ └──────────────────────────────────────────────────────────────────────────┘ │
│ ┌──────────────────────────────────────────────────────────────────────────┐ │
│ │ 5. DATABASE LAYER (database/, GORM v1.25.5)                            │ │
│ │ ┌──────────────────────────────────────────────────────────┐            │ │
│ │ │ SQLite 3 Database Connection                             │            │ │
│ │ │ AutoMigrate: User, Note, Share, E2EEShare tables        │            │ │
│ │ │ ORM Operations: Create, Read, Update, Delete             │            │ │
│ │ └──────────────────────────────────────────────────────────┘            │ │
│ └──────────────────────────────────────────────────────────────────────────┘ │
│ ┌──────────────────────────────────────────────────────────────────────────┐ │
│ │ 6. BACKGROUND JOBS LAYER (jobs/)                                       │ │
│ │ ┌──────────────────────────────────────────────────────────┐            │ │
│ │ │ CleanupJob (cleanup.go)                                  │            │ │
│ │ │ - Xóa note hết hạn (expireAt < now)                      │            │ │
│ │ │ - Xóa share hết lượt truy cập (accessCount >= maxCount)  │            │ │
│ │ │ - Xóa E2EE share hết hạn                                 │            │ │
│ │ │ - Chạy định kỳ: 1 lần/giờ hoặc 1 lần/ngày               │            │ │
│ │ └──────────────────────────────────────────────────────────┘            │ │
│ └──────────────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────────┘
                              │
                              ▼
                ┌────────────────────────────┐
                │ SQLite 3 Database File     │
                │ (project_02_source/storage│
                │  /notes.db)                │
                └────────────────────────────┘
```

---

### Database Schema Chi Tiết

```
TABLE: users
├── id (INT, PK, AI)
├── username (VARCHAR, UNIQUE)
├── password_hash (VARCHAR)
├── created_at (DATETIME)

TABLE: notes
├── id (INT, PK, AI)
├── user_id (INT, FK → users.id)
├── ciphertext (BLOB)
├── encrypted_key (VARCHAR)
├── iv (VARCHAR)
├── expire_at (DATETIME, NULL)
├── created_at (DATETIME)

TABLE: shares
├── id (INT, PK, AI)
├── note_id (INT, FK → notes.id)
├── token (VARCHAR, UNIQUE)
├── password_hash (VARCHAR, NULL)
├── expire_at (DATETIME)
├── access_count (INT, DEFAULT: 0)
├── max_access_count (INT)
├── created_at (DATETIME)

TABLE: e2ee_shares
├── id (INT, PK, AI)
├── note_id (INT, FK → notes.id)
├── sender_id (INT, FK → users.id)
├── recipient_id (INT, FK → users.id)
├── ciphertext (BLOB)
├── public_key_a (VARCHAR)
├── public_key_b (VARCHAR)
├── expire_at (DATETIME)
├── created_at (DATETIME)

TABLE: public_keys
├── id (INT, PK, AI)
├── user_id (INT, FK → users.id)
├── public_key (VARCHAR)
├── created_at (DATETIME)
```

---

### Sơ đồ luồng hoạt động chi tiết

**1. Đăng ký/Đăng nhập:**
```
Người dùng nhập thông tin → Client gửi /register hoặc /login →
  Server (AuthHandler): kiểm tra, tạo tài khoản, sinh JWT → trả về JWT cho client
```

**2. Tạo ghi chú:**
```
Người dùng nhập ghi chú → Client mã hóa ghi chú (AES-256-GCM, PBKDF2) →
  Gửi /note kèm JWT → Server (NoteHandler): xác thực JWT, lưu ciphertext, metadata
```

**3. Chia sẻ ghi chú:**
```
Client tạo URL chia sẻ, mã hóa key → Gửi /share kèm JWT →
  Server (ShareHandler): tạo token, lưu metadata (expire, access_count)
Người nhận truy cập URL → Client lấy key từ URL, gửi /share/{token} →
  Server kiểm tra token, trả về ciphertext nếu hợp lệ
```

**4. Chia sẻ đầu-cuối (E2EE):**
```
Người dùng A nhập ghi chú → Client thực hiện ECDH, mã hóa ghi chú →
  Gửi /e2ee kèm public key, JWT → Server (E2EEHandler): lưu ciphertext, public key
Người dùng B lấy public key, thực hiện ECDH, gửi /e2ee/{id} →
  Server trả về ciphertext, client giải mã bằng session key
```

**5. Cleanup tự động:**
```
Server chạy job định kỳ (jobs/cleanup) → Xóa dữ liệu hết hạn, share exhausted
```

---

### Security & Validation Flow

```
┌──────────────────┐
│ Client Request   │
└────────┬─────────┘
         │
         ▼
┌──────────────────────────────────────────────────────┐
│ 1. HTTPS/TLS (Transport Security)                    │
│    - Mã hóa toàn bộ kết nối                          │
└────────┬─────────────────────────────────────────────┘
         │
         ▼
┌──────────────────────────────────────────────────────┐
│ 2. CORS Middleware                                   │
│    - Kiểm tra origin, allow origin cụ thể            │
└────────┬─────────────────────────────────────────────┘
         │
         ▼
┌──────────────────────────────────────────────────────┐
│ 3. JWT Validation (ngoại trừ /register, /login)     │
│    - Kiểm tra token hợp lệ, chưa hết hạn            │
│    - Lấy user_id từ JWT claim                        │
└────────┬─────────────────────────────────────────────┘
         │
         ▼
┌──────────────────────────────────────────────────────┐
│ 4. Handler Processing                                │
│    - Xác thực thông tin đầu vào (validation)         │
│    - Kiểm tra quyền truy cập (permission)            │
│    - Kiểm tra share token hợp lệ, hết hạn           │
│    - Kiểm tra access_count < max_access_count        │
│    - Xử lý dữ liệu, lưu DB                           │
└────────┬─────────────────────────────────────────────┘
         │
         ▼
┌──────────────────────────────────────────────────────┐
│ 5. Error Handling                                    │
│    - Trả về error message, status code               │
│    - Log error cho debugging                         │
└────────┬─────────────────────────────────────────────┘
         │
         ▼
┌──────────────────┐
│ Response Client  │
└──────────────────┘
```

---

**Các điểm kiểm tra bảo mật:**
- Xác thực JWT ở mọi request (ngoại trừ /register, /login).
- Kiểm tra quyền truy cập khi lấy/chia sẻ ghi chú.
- Kiểm tra token chia sẻ hợp lệ, hết hạn, số lượt truy cập.
- Kiểm tra session key E2EE khi giải mã ghi chú chia sẻ đầu-cuối.
- Mã hóa đầu-cuối (E2EE): Server không thể giải mã dữ liệu của người dùng.

---

## 1. Kiến trúc hệ thống
- Mô hình Client-Server, client mã hóa dữ liệu trước khi gửi lên server.
- Server chỉ lưu trữ dữ liệu đã mã hóa, không thể truy cập nội dung gốc.
- Hỗ trợ chia sẻ đầu-cuối (E2EE) giữa hai người dùng, server không thể giải mã.
- Tự động xóa dữ liệu hết hạn, kiểm soát truy cập theo thời gian và số lần.

## 3. Các thành phần chính
- Client: giao diện Fyne, module mã hóa AES-256-GCM, PBKDF2, ECDH, quản lý JWT.
- Server: các handler xác thực, ghi chú, chia sẻ, E2EE, cleanup job tự động.
- Database: SQLite lưu user, note, share, E2EE share, chỉ lưu ciphertext và metadata.

- Ngôn ngữ Go 1.25.4, Fyne v2.7.1 (GUI), GORM v1.25.5, SQLite 3, JWT v5.2.0.
- Thư viện mã hóa: golang.org/x/crypto v0.33.0 (AES, PBKDF2, Bcrypt, ECDH).
- Build bằng Go compiler, test với testify v1.11.1.
