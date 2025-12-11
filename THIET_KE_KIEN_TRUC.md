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

### Sơ đồ luồng hoạt động chính

**Đăng ký/Đăng nhập:**
```
Người dùng → Client → [Gửi thông tin] → Server → [Kiểm tra, tạo tài khoản, trả về JWT]
```

**Tạo ghi chú:**
```
Người dùng → Client (mã hóa ghi chú) → [Gửi dữ liệu đã mã hóa] → Server (lưu ciphertext)
```

**Chia sẻ ghi chú:**
```
Người dùng → Client (tạo URL chia sẻ, mã hóa key) → [Gửi yêu cầu] → Server (tạo token, lưu metadata)
→ Người nhận truy cập URL → Client (giải mã bằng key từ URL) → Server (trả về ciphertext nếu hợp lệ)
```

**Chia sẻ đầu-cuối (E2EE):**
```
Người dùng A → Client (ECDH key exchange, mã hóa) → [Gửi dữ liệu] → Server (lưu ciphertext, public key)
Người dùng B → Client (lấy public key, giải mã bằng session key)
```

**Cleanup tự động:**
```
Server chạy job định kỳ → Xóa dữ liệu hết hạn, share hết lượt truy cập
```

---

### Sơ đồ thiết kế chi tiết hệ thống

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                                 CLIENT                                     │
├─────────────────────────────────────────────────────────────────────────────┤
│ ┌─────────────┐   ┌─────────────┐   ┌─────────────┐   ┌─────────────┐       │
│ │  UI/GUI     │   │  CLI        │   │  API Client │   │  Crypto     │       │
│ │ (Fyne)      │   │ (Command)   │   │ (net/http)  │   │ (AES, ECDH) │       │
│ └─────┬───────┘   └─────┬───────┘   └─────┬───────┘   └─────┬───────┘       │
│       │                │                │                │                  │
│       └─────┬──────────┴────────────────┴────────────────┴─────┬───────────┘
│             │                                              │                  │
│   ┌─────────▼──────────────────────────────────────────────▼─────────────┐    │
│   │                 Mã hóa dữ liệu (AES-256-GCM, PBKDF2)                │    │
│   │                 Key exchange (ECDH X25519)                          │    │
│   └──────────────────────────────────────────────────────────────────────┘    │
│             │                                              │                  │
│             ▼                                              ▼                  │
│   ┌──────────────────────────────────────────────────────────────────────┐    │
│   │        Gửi dữ liệu đã mã hóa, JWT token qua HTTPS/REST API           │    │
│   └──────────────────────────────────────────────────────────────────────┘    │
└─────────────┬─────────────────────────────────────────────────────────────────┘
              │
              ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                                 SERVER                                     │
├─────────────────────────────────────────────────────────────────────────────┤
│ ┌─────────────┐   ┌─────────────┐   ┌─────────────┐   ┌─────────────┐       │
│ │ AuthHandler │   │ NoteHandler │   │ ShareHandler│   │ E2EEHandler │       │
│ └─────┬───────┘   └─────┬───────┘   └─────┬───────┘   └─────┬───────┘       │
│       │                │                │                │                  │
│       └─────┬──────────┴────────────────┴────────────────┴─────┬───────────┘
│             │                                              │                  │
│   ┌─────────▼──────────────────────────────────────────────▼─────────────┐    │
│   │                 Xác thực JWT, kiểm tra quyền, kiểm tra token         │    │
│   │                 Lưu dữ liệu đã mã hóa vào SQLite (GORM ORM)          │    │
│   │                 Quản lý metadata, share token, E2EE key              │    │
│   └──────────────────────────────────────────────────────────────────────┘    │
│             │                                              │                  │
│             ▼                                              ▼                  │
│   ┌──────────────────────────────────────────────────────────────────────┐    │
│   │        Cleanup job tự động xóa dữ liệu hết hạn, share exhausted      │    │
│   └──────────────────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────────────┘
```

**Luồng dữ liệu:**
- Client mã hóa dữ liệu, gửi lên server qua API.
- Server xác thực JWT, lưu ciphertext, không thể giải mã.
- Chia sẻ ghi chú qua URL, E2EE dùng ECDH để tạo session key riêng cho từng người nhận.
- Cleanup job server tự động xóa dữ liệu hết hạn, share hết lượt truy cập.

---

## 1. Kiến trúc hệ thống
- Mô hình Client-Server, client mã hóa dữ liệu trước khi gửi lên server.
- Server chỉ lưu trữ dữ liệu đã mã hóa, không thể truy cập nội dung gốc.
- Sử dụng Zero-Knowledge Privacy, bảo vệ tối đa quyền riêng tư người dùng.

## 2. Mục đích thiết kế
- Đảm bảo bảo mật dữ liệu, xác thực mạnh, chia sẻ an toàn qua URL tạm thời.
- Hỗ trợ chia sẻ đầu-cuối (E2EE) giữa hai người dùng, server không thể giải mã.
- Tự động xóa dữ liệu hết hạn, kiểm soát truy cập theo thời gian và số lần.

## 3. Các thành phần chính
- Client: giao diện Fyne, module mã hóa AES-256-GCM, PBKDF2, ECDH, quản lý JWT.
- Server: các handler xác thực, ghi chú, chia sẻ, E2EE, cleanup job tự động.
- Database: SQLite lưu user, note, share, E2EE share, chỉ lưu ciphertext và metadata.

## 4. Công nghệ sử dụng
- Ngôn ngữ Go 1.25.4, Fyne v2.7.1 (GUI), GORM v1.25.5, SQLite 3, JWT v5.2.0.
- Thư viện mã hóa: golang.org/x/crypto v0.33.0 (AES, PBKDF2, Bcrypt, ECDH).
- Build bằng Go compiler, test với testify v1.11.1.
