# Envie Frontend

Desktop application built with Tauri 2, Vue 3, and TypeScript.

## Tech Stack

### Core
- **[Tauri 2](https://v2.tauri.app/)** - Desktop app framework (Rust backend, web frontend)
- **[Vue 3](https://vuejs.org/)** - UI framework with Composition API
- **[TypeScript](https://www.typescriptlang.org/)** - Type safety
- **[Vite](https://vitejs.dev/)** - Build tool

### UI
- **[Tailwind CSS v4](https://tailwindcss.com/)** - Styling
- **[shadcn/ui](https://ui.shadcn.com/)** - UI component library (via reka-ui)
- **[Lucide](https://lucide.dev/)** - Icons

### State & Data
- **[Pinia](https://pinia.vuejs.org/)** - State management
- **[Vue Router](https://router.vuejs.org/)** - Routing
- **[Zod](https://zod.dev/)** - Schema validation

### Security
- **[Tauri Stronghold](https://v2.tauri.app/plugin/stronghold/)** - Secure key storage
- **[@noble/curves](https://github.com/paulmillr/noble-curves)** - Ed25519, X25519 cryptography
- **[@noble/hashes](https://github.com/paulmillr/noble-hashes)** - XChaCha20-Poly1305, hashing

## Project Structure

```
frontend/
├── src/
│   ├── assets/          # CSS and static assets
│   ├── components/
│   │   ├── ui/          # Reusable UI components (shadcn)
│   │   ├── layout/      # App layout (sidebar, etc.)
│   │   ├── project/     # Project-related components
│   │   ├── organization/# Organization management
│   │   └── identity/    # Device identity management
│   ├── composables/     # Vue composables
│   ├── lib/             # Utilities (cn, toast)
│   ├── router/          # Vue Router config
│   ├── services/        # API service layer
│   │   ├── api.ts           # Base API client
│   │   ├── auth.service.ts
│   │   ├── project.service.ts
│   │   ├── team.service.ts
│   │   ├── organization.service.ts
│   │   ├── device.service.ts
│   │   ├── identity.service.ts
│   │   ├── encryption.service.ts
│   │   └── secret-manager.service.ts
│   ├── stores/          # Pinia stores
│   │   ├── auth.ts          # Authentication state
│   │   ├── vault.ts         # Encryption keys
│   │   ├── organization.ts  # Current organization
│   │   └── secret-manager.store.ts
│   ├── views/           # Page components
│   │   ├── Home.vue
│   │   ├── Settings.vue
│   │   ├── Identities.vue
│   │   ├── ProjectDetail.vue
│   │   └── OrganizationDetail.vue
│   ├── App.vue          # Root component
│   ├── main.ts          # Entry point
│   └── config.ts        # App configuration
├── src-tauri/           # Tauri/Rust backend
│   ├── src/
│   │   └── lib.rs       # Rust commands
│   ├── Cargo.toml
│   └── tauri.conf.json
├── package.json
├── vite.config.ts
└── tsconfig.json
```

## Key Concepts

### Vault & Encryption

The app uses a local vault (Stronghold) to store encryption keys:

1. **Master Identity Key** - Generated on first setup, stored in Stronghold
2. **Device Key Pair** - Ed25519 for signatures, X25519 for key exchange
3. **Project Keys** - Symmetric keys for encrypting project data

### Services Layer

All API calls go through typed service classes:
- Centralized error handling with toast notifications
- Typed request/response interfaces
- Automatic token refresh

## Development

```bash
# Install dependencies
npm install

# Start development (opens Tauri app)
npm run tauri dev

# Build for production
npm run tauri build

# Run Vite dev server only (no Tauri)
npm run dev
```

## Configuration

Create `src/config.ts`:
```typescript
export const config = {
  backendUrl: 'http://localhost:8080'
}
```
