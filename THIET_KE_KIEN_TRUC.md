# Thiết Kế và Kiến Trúc Hệ Thống

## 1. Tổng Quan Kiến Trúc

### 1.1 Mô Hình Kiến Trúc Tổng Quát

Hệ thống được thiết kế theo **mô hình Client-Server** với kiến trúc **Zero-Knowledge Privacy**, đảm bảo:

- **Client:** Xử lý mã hóa/giải mã dữ liệu, quản lý khóa mã hóa
- **Server:** Chỉ lưu trữ dữ liệu đã mã hóa, không có khả năng truy cập nội dung gốc
- **Zero-Knowledge:** Server không cần biết hoặc quản lý bất kỳ thông tin nhạy cảm nào

```
┌─────────────────────────────────────┐
│         CLIENT APPLICATION          │
├──────────┬──────────────┬───────────┤
│   GUI    │    CLI       │   API     │
│ (Fyne)   │   (Command)  │  Client   │
└─────┬────┴──────┬───────┴─────┬─────┘
      │           │             │
      ├─Encryption (AES-256-GCM)┤
      ├─Key Management (PBKDF2)─┤
      ├─E2EE (ECDH X25519)──────┤
      │                         │
      └───── HTTPS/REST API ────┘
            (Port: 8080)
             │
             ▼
┌──────────────────────────────────────┐
│      SERVER APPLICATION (Go)         │
├──────────┬──────────────┬────────────┤
│   Auth   │   Note Mgmt  │ Share Mgmt │
│ Handler  │   Handler    │  Handler   │
└────┬─────┴──────┬───────┴──────┬─────┘
     │            │              │
     └─SQLite Database────────────┘
        (Encrypted Data)
        ├─ Users Table
        ├─ Notes Table (Ciphertext)
        ├─ Shared Links Table
        └─ E2EE Shares Table
```

### 1.2 Mục Đích Thiết Kế

| Mục Tiêu | Mô Tả |
|----------|-------|
| **Bảo Mật Dữ Liệu** | Mã hóa toàn bộ ghi chú trước khi tải lên server |
| **Zero-Knowledge** | Server không thể đọc nội dung dữ liệu người dùng |
| **Chia Sẻ An Toàn** | Chia sẻ qua URL với kiểm soát thời gian, số lần truy cập |
| **E2EE Riêng Tư** | Chia sẻ dữ liệu giữa hai người dùng qua key exchange |
| **Kiểm Soát Truy Cập** | Tự động xóa dữ liệu hết hạn, yêu cầu xác thực |
| **Nhân Chứng Dữ Liệu** | Xác thực tính toàn vẹn dữ liệu bằng GCM mode |

---

## 2. Các Thành Phần Chính

### 2.1 Client Component

#### a) **UI/GUI (Giao diện Đồ họa)**
- **Framework:** Fyne v2.7.1 (Go GUI)
- **Tính năng:**
  - Login/Register screen
  - Notes management screen
  - Share creation & viewing

#### b) **API Client**
- Giao tiếp REST với server
- Quản lý token JWT
- Xử lý HTTP requests/responses

#### c) **Crypto Module**
Xử lý tất cả các phép toán mã hóa:
- AES-256-GCM encryption/decryption
- PBKDF2 key derivation
- ECDH X25519 key exchange
- SHA-256 hashing
- Random key generation

#### d) **CLI Interface**
- Command-line interface để kiểm thử
- Hỗ trợ các câu lệnh note, share, auth

### 2.2 Server Component

#### a) **Authentication Handler**
- Register: Tạo tài khoản mới
- Login: Xác thực người dùng, cấp JWT token
- Logout: Hủy phiên làm việc
- Password hashing: Bcrypt + Salt

#### b) **Note Handler**
- Create: Lưu note đã mã hóa
- Read: Trả về dữ liệu mã hóa cho chủ sở hữu
- Update: Cập nhật note (verify owner)
- Delete: Xóa note
- List: Liệt kê notes của user

#### c) **Share Handler**
- CreateShare: Tạo URL chia sẻ tạm thời
- GetSharedNote: Trả về note đã mã hóa qua share token
- RevokeShare: Hủy chia sẻ
- Cleanup: Xóa share hết hạn (background job)

#### d) **E2EE Handler**
- CreateE2EEShare: Chia sẻ với ECDH key exchange
- GetE2EEShare: Lấy shared note từ recipient
- ListE2EEShares: Liệt kê shares nhận được
- DeleteE2EEShare: Xóa share E2EE

#### e) **Public Key Handler**
- UpdatePublicKey: Upload DH public key
- GetPublicKey: Lấy DH public key của user khác

#### f) **Cleanup Job**
- Background job chạy định kỳ
- Xóa shared links hết hạn
- Xóa E2EE shares hết hạn
- Xóa shares hết lượt truy cập

### 2.3 Database Schema

```
Users
├─ ID (PK)
├─ Username (UNIQUE)
├─ PasswordHash (Bcrypt)
└─ DHPublicKey (X25519 công khai)

Notes
├─ ID (PK)
├─ UserID (FK)
├─ Title
├─ EncryptedContent (AES-256-GCM ciphertext)
├─ IV (Initialization Vector)
├─ EncryptedKey (DEK mã hóa bằng KEK)
├─ EncryptedKeyIV
└─ CreatedAt

SharedLinks
├─ ID (PK)
├─ NoteID (FK)
├─ UserID (FK)
├─ ShareToken (Unique token)
├─ ExpiresAt (TTL)
├─ MaxAccessCount (0 = unlimited)
├─ AccessCount
├─ RequirePassword
├─ PasswordHash
└─ CreatedAt

E2EEShares
├─ ID (PK)
├─ NoteID (FK)
├─ SenderID (FK)
├─ RecipientID (FK)
├─ SenderPublicKey (DH public key)
├─ EncryptedContent (AES-256-GCM)
├─ ContentIV
├─ ExpiresAt
└─ CreatedAt
```

---

## 3. Công Nghệ Sử Dụng

### 3.1 Backend (Server)

| Thành Phần | Tên | Phiên Bản | Mục Đích |
|-----------|-----|----------|---------|
| **Ngôn Ngữ** | Go | 1.25.4 | Xây dựng server |
| **Web Framework** | Standard Library (net/http) | 1.25.4 | REST API |
| **Database** | SQLite | 3 | Lưu trữ dữ liệu |
| **ORM** | GORM | v1.25.5 | Query builder, migrations |
| **SQLite Driver** | gorm.io/driver/sqlite | v1.5.4 | SQLite integration |
| **Hashing Mật khẩu** | golang.org/x/crypto/bcrypt | v0.33.0 | Bcrypt password hashing |
| **JWT** | github.com/golang-jwt/jwt | v5.2.0 | JWT token management |
| **Crypto** | golang.org/x/crypto | v0.33.0 | AES, PBKDF2, SHA-256, ECDH |

### 3.2 Frontend (Client)

| Thành Phần | Tên | Phiên Bản | Mục Đích |
|-----------|-----|----------|---------|
| **Ngôn Ngữ** | Go | 1.25.4 | Xây dựng client |
| **GUI Framework** | Fyne | v2.7.1 | Giao diện đồ họa |
| **HTTP Client** | Standard Library (net/http) | 1.25.4 | API calls |
| **Crypto** | golang.org/x/crypto | v0.33.0 | Mã hóa/giải mã |
| **JSON** | Standard Library (encoding/json) | 1.25.4 | JSON parsing |

### 3.3 Testing & Development

| Thành Phần | Tên | Phiên Bản | Mục Đích |
|-----------|-----|----------|---------|
| **Testing** | stretchr/testify | v1.11.1 | Unit test framework |
| **Build Tool** | Go Build | 1.25.4 | Compile & build |

### 3.4 Thuật Toán Mã Hóa

| Thuật Toán | Mode/Type | Key Size | Mục Đích |
|-----------|-----------|----------|---------|
| **AES** | GCM (Galois/Counter Mode) | 256-bit | Mã hóa dữ liệu, xác thực tính toàn vẹn |
| **PBKDF2** | HMAC-SHA-256 | 32-byte | Derive khóa từ mật khẩu |
| **Bcrypt** | Adaptive Hash | Cost: 10 | Hashing mật khẩu an toàn |
| **ECDH** | X25519 | 256-bit | Key exchange cho E2EE |
| **SHA-256** | Hash | 256-bit | Derive session key từ DH secret |

---

## 4. Luồng Hoạt Động Chính

### 4.1 Luồng Đăng Ký (Registration)

```
┌─────────────────────────────────────────────────────────┐
│                    CLIENT (GUI)                         │
│  User nhập: Username, Password                          │
└────────────────────┬────────────────────────────────────┘
                     │ POST /api/auth/register
                     │ {username, password}
                     ▼
┌─────────────────────────────────────────────────────────┐
│               SERVER (Register Handler)                 │
│  1. Validate input (length, format)                     │
│  2. Hash password: Bcrypt(password)                     │
│  3. Create User record                                  │
│  4. Save to Database                                    │
└────────────────────┬────────────────────────────────────┘
                     │ JSON Response
                     │ {success: true, message: "..."}
                     ▼
┌─────────────────────────────────────────────────────────┐
│                    CLIENT (GUI)                         │
│  Display: "User registered successfully"               │
└─────────────────────────────────────────────────────────┘
```

### 4.2 Luồng Đăng Nhập (Login)

```
┌─────────────────────────────────────────────────────────┐
│                    CLIENT (GUI)                         │
│  User nhập: Username, Password                          │
└────────────────────┬────────────────────────────────────┘
                     │ POST /api/auth/login
                     │ {username, password}
                     ▼
┌─────────────────────────────────────────────────────────┐
│                 SERVER (Login Handler)                  │
│  1. Find User by username                              │
│  2. Verify password: Bcrypt.Check(password, hash)      │
│  3. Generate JWT token: token = JWT.Sign(userID)       │
│  4. Return token                                        │
└────────────────────┬────────────────────────────────────┘
                     │ JSON Response
                     │ {token: "JWT_TOKEN", username: "..."}
                     ▼
┌─────────────────────────────────────────────────────────┐
│                    CLIENT (GUI)                         │
│  1. Store token locally (in memory)                    │
│  2. Use token for subsequent API calls                 │
│  Display: "Logged in successfully"                     │
└─────────────────────────────────────────────────────────┘
```

### 4.3 Luồng Tạo & Tải Ghi Chú Mã Hóa

```
┌──────────────────────────────────────────────────────────┐
│                      CLIENT (GUI)                        │
│  1. User chọn note content                              │
│  2. Generate DEK (Data Encryption Key): 32 bytes random  │
└────────────────────┬─────────────────────────────────────┘
                     │
     ┌───────────────┴────────────────┐
     │  Encryption Process (Client)    │
     ├──────────────────────────────────┤
     │ 1. Generate IV: 12 bytes random │
     │ 2. Encrypt content:              │
     │    ciphertext = AES-256-GCM      │
     │      (content, DEK, IV)          │
     │ 3. Derive KEK from password:     │
     │    KEK = PBKDF2(password, salt) │
     │ 4. Encrypt DEK:                  │
     │    enc_DEK = AES-256-GCM         │
     │      (DEK, KEK, DEK_IV)          │
     └────────────────┬─────────────────┘
                      │
                      ▼
            ┌──────────────────────┐
            │  ciphertext (32KB+)  │
            │  IV (12 bytes)       │
            │  enc_DEK (32 bytes)  │
            │  DEK_IV (12 bytes)   │
            └──────────────────────┘
                      │
                      │ POST /api/notes
                      │ {title, encryptedContent, IV, 
                      │  encryptedKey, encryptedKeyIV}
                      │ Authorization: Bearer JWT_TOKEN
                      ▼
┌──────────────────────────────────────────────────────────┐
│              SERVER (Note Handler)                       │
│  1. Verify JWT token                                     │
│  2. Validate user ownership                              │
│  3. Save to database (ciphertext only)                   │
│  4. Return success response                              │
└────────────────────┬─────────────────────────────────────┘
                     │
                     ▼
┌──────────────────────────────────────────────────────────┐
│                    CLIENT (GUI)                          │
│  Display: "Note uploaded successfully"                  │
│  Clear sensitive data from memory (DEK, password)       │
└──────────────────────────────────────────────────────────┘
```

### 4.4 Luồng Giải Mã & Đọc Ghi Chú

```
┌──────────────────────────────────────────────────────────┐
│                    CLIENT (GUI)                          │
│  User click: "View Note"                                 │
│  Send: GET /api/notes/{noteID}                           │
│        Authorization: Bearer JWT_TOKEN                   │
└────────────────────┬─────────────────────────────────────┘
                     │
                     ▼
┌──────────────────────────────────────────────────────────┐
│              SERVER (Note Handler)                       │
│  1. Verify JWT token                                     │
│  2. Fetch note (ciphertext + encrypted DEK)              │
│  3. Return encrypted data                                │
└────────────────────┬─────────────────────────────────────┘
                     │ {encryptedContent, IV, 
                     │  encryptedKey, encryptedKeyIV}
                     ▼
┌──────────────────────────────────────────────────────────┐
│            CLIENT (Decryption Process)                   │
│  1. Prompt user for password                             │
│  2. Derive KEK: KEK = PBKDF2(password, salt)             │
│  3. Decrypt DEK:                                         │
│     DEK = AES-256-GCM.Decrypt(enc_DEK, KEK, DEK_IV)     │
│  4. Decrypt content:                                     │
│     content = AES-256-GCM.Decrypt(ciphertext, DEK, IV)   │
│  5. Verify authenticity (GCM tag)                        │
│  6. Display content                                      │
│  7. Zeroinize DEK, KEK from memory                       │
└──────────────────────────────────────────────────────────┘
```

### 4.5 Luồng Chia Sẻ Qua URL Tạm Thời

```
┌──────────────────────────────────────────────────────────┐
│                    CLIENT (GUI)                          │
│  User click: "Share Note"                                │
│  Input: Expiration time, max access count, password     │
└────────────────────┬─────────────────────────────────────┘
                     │ POST /api/notes/{noteID}/share
                     │ {expiresIn: "1h", maxAccessCount: 5,
                     │  password: "..."}
                     ▼
┌──────────────────────────────────────────────────────────┐
│            SERVER (Share Handler)                        │
│  1. Verify JWT token & user ownership                    │
│  2. Generate unique token: RandomBytes(32)               │
│  3. Hash password: Bcrypt(password) if provided          │
│  4. Calculate expiration: now() + duration               │
│  5. Create SharedLink record:                            │
│     {noteID, userID, shareToken, expiresAt,              │
│      maxAccessCount, passwordHash, accessCount=0}        │
│  6. Save to database                                     │
│  7. Generate share URL:                                  │
│     https://app.com/share/{shareToken}#key=<DEK_encoded> │
└────────────────────┬─────────────────────────────────────┘
                     │
                     ▼
┌──────────────────────────────────────────────────────────┐
│                    CLIENT (GUI)                          │
│  Display share URL (with encryption key in fragment)    │
│  User copies & shares URL                               │
└──────────────────────────────────────────────────────────┘

            ┌─────────────────────────────┐
            │  RECIPIENT (Any Device)     │
            │  1. Click share URL         │
            │  2. Browser reads fragment  │
            │     (never sent to server)  │
            │  3. App extracts DEK        │
            └─────────────┬───────────────┘
                          │ GET /api/shares/{shareToken}
                          │ (no Authorization header)
                          ▼
        ┌────────────────────────────────┐
        │  SERVER (Share Handler)         │
        │  1. Find SharedLink by token    │
        │  2. Verify not expired:         │
        │     now() < expiresAt? ✓        │
        │  3. Verify access count:        │
        │     accessCount < maxCount? ✓   │
        │  4. If password protected:      │
        │     Verify password hash        │
        │  5. Increment accessCount++     │
        │  6. Return encrypted note       │
        │     (ciphertext only, no DEK)   │
        │  7. Check if now at max:        │
        │     If accessCount>=maxCount,   │
        │     mark as exhausted (cleanup) │
        └────────────────┬────────────────┘
                         │ {encryptedContent, IV, owner}
                         │
                         ▼
        ┌────────────────────────────────┐
        │  RECIPIENT CLIENT               │
        │  1. Decrypt using DEK from URL  │
        │     content = AES-GCM.Decrypt   │
        │       (ciphertext, DEK, IV)     │
        │  2. Display note                │
        │  3. Clear sensitive data        │
        └────────────────────────────────┘

Background Cleanup Job (Server):
┌────────────────────────────────────────┐
│  Every 5 minutes:                      │
│  1. Find expired shares:               │
│     now() > expiresAt                  │
│  2. Find exhausted shares:             │
│     accessCount >= maxAccessCount      │
│  3. Delete matched records             │
│  4. Log cleanup count                  │
└────────────────────────────────────────┘
```

### 4.6 Luồng E2EE (End-to-End Encryption)

```
┌──────────────────────────────────────────────────────────┐
│                  SENDER (Alice)                          │
│  1. Wants to share note with Bob                         │
│  2. Click "E2EE Share" & select recipient               │
└────────────────────┬─────────────────────────────────────┘
                     │
        ┌────────────┴──────────┐
        │  Client Side:         │
        │  1. Fetch Bob's DH    │
        │     public key        │
        │  2. Alice's DH        │
        │     secret = DH(      │
        │       alicePrivate,   │
        │       bobPublic)      │
        │  3. Derive session    │
        │     key = SHA256      │
        │       (secret)        │
        │  4. Encrypt content   │
        │     with session key  │
        │  5. Clear DEK from    │
        │     memory (SECURE!)  │
        └────────────┬──────────┘
                     │
                     │ POST /api/notes/{noteID}/e2ee
                     │ {recipientID, senderPublicKey,
                     │  encryptedContent, contentIV}
                     ▼
┌──────────────────────────────────────────────────────────┐
│        SERVER (E2EE Handler)                             │
│  1. Verify sender JWT token                              │
│  2. Validate recipient exists                            │
│  3. Create E2EEShare record:                             │
│     {noteID, senderID, recipientID,                      │
│      senderPublicKey, encryptedContent,                  │
│      contentIV, expiresAt}                               │
│  4. Return success                                       │
│  5. Server stores but CANNOT decrypt                     │
│     (doesn't have session key!)                          │
└────────────────────┬─────────────────────────────────────┘
                     │
                     ▼
┌──────────────────────────────────────────────────────────┐
│              RECIPIENT (Bob)                             │
│  1. Query: GET /api/e2ee (list shared notes)            │
│     Authorization: Bearer BOB_TOKEN                      │
└────────────────────┬─────────────────────────────────────┘
                     │
                     ▼
┌──────────────────────────────────────────────────────────┐
│        SERVER (List E2EE Shares)                         │
│  1. Find E2EEShares where recipientID = Bob's ID        │
│  2. Return list with senderPublicKey & encrypted data   │
└────────────────────┬─────────────────────────────────────┘
                     │ [{noteID, senderPublicKey,
                     │   encryptedContent, ...}]
                     ▼
┌──────────────────────────────────────────────────────────┐
│              RECIPIENT (Bob) Decryption:                 │
│  1. Bob has his DH private key                           │
│  2. Compute shared secret:                               │
│     secret = DH(bobPrivate, alicePublic)                │
│  3. Derive session key:                                  │
│     sessionKey = SHA256(secret)                          │
│  4. Decrypt content:                                     │
│     content = AES-GCM.Decrypt                            │
│       (encryptedContent, sessionKey, IV)                 │
│  5. Verify GCM tag (authenticity)                        │
│  6. Display note                                         │
│  7. Zeroinize secret & sessionKey                        │
└──────────────────────────────────────────────────────────┘

Security Properties:
├─ Forward Secrecy: Session key không được reuse
├─ Mutual Authentication: Only Alice & Bob can communicate
├─ Perfect Secrecy: Server cannot decrypt even if hacked
└─ No Key Escrow: Only peers hold encryption keys
```

---

## 5. Sơ Đồ Tương Tác Thành Phần

```
┌─────────────────────────────────────────────────────────────────┐
│                     CLIENT ARCHITECTURE                         │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐  │
│  │              UI Layer (Fyne)                            │  │
│  │  ├─ Login Screen                                        │  │
│  │  ├─ Notes Management Screen                            │  │
│  │  ├─ Share Creation Dialog                              │  │
│  │  └─ Shared Note Viewer                                 │  │
│  └────────────────────┬────────────────────────────────────┘  │
│                       │                                        │
│  ┌────────────────────▼────────────────────────────────────┐  │
│  │          API Client Layer                              │  │
│  │  ├─ HTTP requests to server                            │  │
│  │  ├─ JWT token management                               │  │
│  │  ├─ Response parsing                                   │  │
│  │  └─ Error handling                                     │  │
│  └────────────────────┬────────────────────────────────────┘  │
│                       │                                        │
│  ┌────────────────────▼────────────────────────────────────┐  │
│  │         Crypto Layer                                   │  │
│  │  ├─ AES-256-GCM encrypt/decrypt                        │  │
│  │  ├─ PBKDF2 key derivation                              │  │
│  │  ├─ ECDH X25519 key exchange                           │  │
│  │  ├─ SHA-256 hashing                                    │  │
│  │  ├─ Random key generation                              │  │
│  │  └─ Memory zeroinization                               │  │
│  └─────────────────────────────────────────────────────────┘  │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                    SERVER ARCHITECTURE                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐  │
│  │        HTTP Router (net/http)                          │  │
│  │  ├─ CORS Middleware                                    │  │
│  │  ├─ Route handlers                                     │  │
│  │  └─ Method validation                                  │  │
│  └────────────────────┬────────────────────────────────────┘  │
│                       │                                        │
│  ┌────────────────────▼────────────────────────────────────┐  │
│  │            Handler Layer                               │  │
│  │  ├─ AuthHandler (register, login, logout)              │  │
│  │  ├─ NoteHandler (CRUD operations)                      │  │
│  │  ├─ ShareHandler (create, access, revoke)              │  │
│  │  ├─ E2EEHandler (E2EE operations)                      │  │
│  │  └─ PublicKeyHandler (key management)                  │  │
│  └────────────────────┬────────────────────────────────────┘  │
│                       │                                        │
│  ┌────────────────────▼────────────────────────────────────┐  │
│  │          Database Layer (GORM + SQLite)                │  │
│  │  ├─ User operations                                    │  │
│  │  ├─ Note operations                                    │  │
│  │  ├─ SharedLink operations                              │  │
│  │  └─ E2EEShare operations                               │  │
│  └────────────────────┬────────────────────────────────────┘  │
│                       │                                        │
│  ┌────────────────────▼────────────────────────────────────┐  │
│  │     Database File (app.db - SQLite)                    │  │
│  │  ├─ Encrypted note data                                │  │
│  │  ├─ User hashes                                        │  │
│  │  ├─ Share metadata                                     │  │
│  │  └─ E2EE share data                                    │  │
│  └─────────────────────────────────────────────────────────┘  │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐  │
│  │       Background Job (Cleanup)                         │  │
│  │  ├─ Check expired shares every 5 minutes               │  │
│  │  ├─ Delete exhausted shares                            │  │
│  │  └─ Log cleanup statistics                             │  │
│  └─────────────────────────────────────────────────────────┘  │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

---

## 6. Luồng Xác Thực & Ủy Quyền

```
┌─────────────────────────────────────────────────────────────┐
│              Authentication Flow                            │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  1. Login                                                   │
│     ├─ POST /api/auth/login {username, password}           │
│     └─ Response: {token, username}                         │
│                                                             │
│  2. Client stores JWT token in memory                       │
│     └─ Token format: eyJhbGciOiJIUzI1NiIsInR5cCI6...      │
│                                                             │
│  3. For authenticated requests:                            │
│     ├─ Add header: Authorization: Bearer <JWT_TOKEN>       │
│     └─ Server verifies token signature & expiry            │
│                                                             │
│  4. Token expiry:                                          │
│     ├─ Duration: 24 hours (configurable)                   │
│     └─ Expired token → 401 Unauthorized                    │
│                                                             │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│         Authorization Rules                                │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  /api/notes (List & Create)                                │
│  ├─ POST   → Requires valid JWT                            │
│  └─ GET    → Requires valid JWT                            │
│                                                             │
│  /api/notes/{id}                                           │
│  ├─ GET    → Requires JWT + Owner verification             │
│  ├─ PUT    → Requires JWT + Owner verification             │
│  └─ DELETE → Requires JWT + Owner verification             │
│                                                             │
│  /api/shares/{token}                                       │
│  ├─ GET    → No authentication (public access)             │
│  │         But must pass validation rules                  │
│  │         (expiry, access count, password)                │
│  └─ Validation is done server-side                         │
│                                                             │
│  /api/e2ee                                                 │
│  ├─ GET    → Requires JWT (list received shares)           │
│  └─ POST   → Requires JWT (create new share)               │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## 7. Tóm Tắt Công Nghệ & Công Cụ

### 7.1 Lập Trình & Build
- **Compiler:** Go 1.25.4
- **Phương pháp build:** go build (tạo executable)

### 7.2 Các Thư Viện Mã Hóa
- **AES-256-GCM:** golang.org/x/crypto/aes (gcm mode)
- **PBKDF2:** golang.org/x/crypto/pbkdf2
- **Bcrypt:** golang.org/x/crypto/bcrypt
- **ECDH:** golang.org/x/crypto/ecdh (X25519)
- **SHA-256:** crypto/sha256 (standard library)

### 7.3 Database & ORM
- **Database:** SQLite 3 (file-based)
- **ORM:** GORM v1.25.5 (Go Object-Relational Mapping)
- **Driver:** gorm.io/driver/sqlite v1.5.4

### 7.4 API & Communication
- **Protocol:** HTTP/REST
- **Port:** 8080
- **JWT:** github.com/golang-jwt/jwt v5.2.0
- **CORS:** Implemented in middleware

### 7.5 GUI & CLI
- **GUI Framework:** Fyne v2.7.1 (Go native GUI)
- **CLI:** Custom command-line interface

### 7.6 Testing & Quality Assurance
- **Testing Framework:** stretchr/testify v1.11.1
- **Test Coverage:** Unit tests cho auth, crypto, handlers

---

## 8. Cơ Chế Bảo Mật Chi Tiết

### 8.1 Bảo Mật Mật Khẩu
- **Hashing:** Bcrypt (adaptive hashing)
- **Salt:** Tự động tạo bởi Bcrypt
- **Cost Factor:** 10 (2^10 iterations)
- **Lưu trữ:** Chỉ hash được lưu, mật khẩu không bao giờ lưu

### 8.2 Bảo Mật Dữ Liệu
- **Mã hóa:** AES-256-GCM
- **Chế độ:** GCM (Galois/Counter Mode)
  - Mã hóa (Confidentiality)
  - Xác thực (Authenticity)
  - Tính toàn vẹn (Integrity)
- **IV:** 12 bytes ngẫu nhiên cho mỗi lần mã hóa

### 8.3 Quản Lý Khóa
- **DEK (Data Encryption Key):** 32 bytes ngẫu nhiên cho mỗi note
- **KEK (Key Encryption Key):** Derived từ password qua PBKDF2
- **PBKDF2:** SHA-256, 600,000 iterations
- **Envelope Encryption:** DEK được mã hóa bằng KEK

### 8.4 E2EE Bảo Mật
- **Key Exchange:** ECDH X25519
- **Public Key Infrastructure:** Mỗi user có DH keypair
- **Session Key:** Derive từ shared secret bằng SHA-256
- **Forward Secrecy:** Session key bị hủy sau sử dụng
- **Perfect Secrecy:** Server không bao giờ có khóa

### 8.5 JWT Bảo Mật
- **Thuật toán:** HS256 (HMAC SHA-256)
- **Claims:** user_id, username, exp (hết hạn)
- **Expiration:** 24 hours
- **Verification:** Server verify signature trước xử lý

---

## 9. Kết Luận

Kiến trúc hệ thống được thiết kế theo nguyên tắc **Zero-Knowledge Privacy** với các tính năng bảo mật lớp nhiều:
- Mã hóa toàn bộ dữ liệu phía client
- Server không thể đọc nội dung dữ liệu
- E2EE cho chia sẻ riêng tư
- Kiểm soát truy cập thời gian và số lần
- Xác thực mạnh bằng JWT
- Cleanup tự động dữ liệu hết hạn

Hệ thống sử dụng các thư viện mã hóa chuẩn công nghiệp (golang.org/x/crypto) với các thuật toán hiện đại (AES-256-GCM, ECDH, Bcrypt) để đảm bảo tính bảo mật cao nhất.
