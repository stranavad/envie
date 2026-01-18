# Envie

**Your secrets. Your control.**

Envie is a zero-trust, end-to-end encrypted secret manager for teams. All encryption and decryption happens locally on your device - not even Envie servers can access your secrets.

## Features

- **Environment Variables** - Store and manage environment variables with E2E encryption
- **Encrypted File Sharing** - Share sensitive files (.env, certificates, credentials) securely
- **Organizations & Teams** - Hierarchical structure with fine-grained access control
- **Key Rotation** - Rotate encryption keys with double-admin approval for organizations
- **Device Identities** - Each device has its own cryptographic identity
- **Google Secret Manager Integration** - Optional sync with Google Cloud (your GCP credentials never touch our servers)

## Architecture

Envie is built on zero-trust principles:

1. **Client-side encryption** - All data is encrypted using XChaCha20-Poly1305 before leaving your device
2. **Secure key storage** - Keys are stored in [Stronghold](https://github.com/iotaledger/stronghold.rs), the same secure storage used by cryptocurrency wallets
3. **Server blindness** - The backend only stores encrypted blobs that are meaningless without your keys

## Project Structure

```
envie/
├── frontend/     # Tauri desktop app (Vue 3 + Rust)
├── backend/      # Go API server
└── website/      # Marketing website (Astro)
```

## Getting Started

### Prerequisites

- Node.js 24+
- Rust (for Tauri)
- Go 1.25+
- PostgreSQL
- S3-compatible storage (e.g., Tigris, AWS S3)

### Development

**Backend:**
```bash
cd backend
cp .env.example .env  # Configure environment variables
go run cmd/api/main.go
```

**Frontend:**
```bash
cd frontend
npm install
npm run tauri dev
```

**Website:**
```bash
cd website
npm install
npm run dev
```

## Security Model

### Key Hierarchy

1. **Master Identity Key** - Your personal key stored in Stronghold
2. **Team Keys** - Shared keys encrypted with each member's public key
3. **Project Keys** - Individual keys for each project
4. **File Encryption Keys** - Per-file keys wrapped by project keys

### Encryption

- Symmetric encryption: XChaCha20-Poly1305
- Key exchange: X25519
- Signatures: Ed25519

## License

MIT License - see [LICENSE](LICENSE) for details.

## Links

- [GitHub](https://github.com/stranavad/envie)
- [Website](https://envie.gandalfthegray.dev)
