<script setup lang="ts">
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import {
  KeyRound,
  FolderLock,
  Users,
  Building2,
  RefreshCw,
  Fingerprint,
  FolderTree,
  Shield,
  Github,
  Layers,
  Copy,
  Lock,
  EyeOff,
  FileText,
  FileKey,
  Upload,
  Download,
  CheckCircle2,
  User,
  Folder,
  Cloud,
  ArrowRight,
  Server,
  HardDrive,
  Link2,
  RefreshCcw
} from 'lucide-vue-next'

const githubUrl = 'https://github.com/stranavad/envie'

const additionalFeatures = [
  {
    title: 'Team Collaboration',
    description: 'Work together with your team on shared secrets. Each member has their own encryption keys for maximum security.',
    icon: Users,
  },
  {
    title: 'Key Rotation',
    description: 'Rotate encryption keys when team members leave. For organizations, double-admin approval ensures no single point of failure.',
    icon: RefreshCw,
  },
  {
    title: 'Device Identities',
    description: 'Each device has its own cryptographic identity. Approve new devices or revoke access instantly.',
    icon: Fingerprint,
  },
  {
    title: 'Categories',
    description: 'Organize your environment variables with categories like Development, Staging, and Production.',
    icon: FolderTree,
  },
  {
    title: 'Zero-Trust Security',
    description: 'Built on zero-trust principles. The server never sees your unencrypted data.',
    icon: Shield,
  },
  {
    title: 'Project Hierarchy',
    description: 'Organize projects within teams. Share secrets across projects or keep them isolated.',
    icon: Layers,
  },
]

// Mock data for illustrations
const mockSecrets = [
  { key: 'DATABASE_URL', value: 'postgres://user:***@db.example.com', category: 'Production', visible: false },
  { key: 'API_KEY', value: 'sk_live_***************', category: 'Production', visible: false },
  { key: 'JWT_SECRET', value: '*********************', category: 'All', visible: false },
]

const mockFiles = [
  { name: '.env.production', size: '2.4 KB', encrypted: true, shared: 3 },
  { name: 'service-account.json', size: '1.8 KB', encrypted: true, shared: 2 },
  { name: 'ssl-certificate.pem', size: '4.1 KB', encrypted: true, shared: 1 },
]
</script>

<template>
  <div class="py-16 md:py-24">
    <div class="container mx-auto px-4">
      <!-- Header -->
      <div class="text-center mb-16 max-w-3xl mx-auto">
        <p class="text-primary font-medium mb-3">Features</p>
        <h1 class="text-4xl md:text-6xl font-bold mb-6 tracking-tight">
          Everything you need for secure secret management
        </h1>
        <p class="text-xl text-muted-foreground leading-relaxed">
          Envie provides all the tools your team needs to manage secrets securely,
          without compromising on usability.
        </p>
      </div>

      <!-- Environment Variables Feature -->
      <div class="grid lg:grid-cols-2 gap-12 items-center mb-24">
        <div>
          <div class="inline-flex items-center rounded-full bg-primary/10 px-4 py-1.5 text-sm font-medium text-primary border border-primary/20 mb-4">
            <KeyRound class="mr-2 size-4" />
            Environment Variables
          </div>
          <h2 class="text-3xl font-bold mb-4">Environment Variables</h2>
          <p class="text-lg text-muted-foreground mb-6 leading-relaxed">
            Store and manage your environment variables with military-grade encryption. Every secret is encrypted on your device before being synced to your team.
          </p>
          <ul class="space-y-3">
            <li class="flex items-center gap-3 text-muted-foreground">
              <div class="size-1.5 rounded-full bg-primary"></div>
              End-to-end encrypted storage
            </li>
            <li class="flex items-center gap-3 text-muted-foreground">
              <div class="size-1.5 rounded-full bg-primary"></div>
              Easy copy and export to .env files
            </li>
            <li class="flex items-center gap-3 text-muted-foreground">
              <div class="size-1.5 rounded-full bg-primary"></div>
              Organize with categories (Dev, Staging, Prod)
            </li>
            <li class="flex items-center gap-3 text-muted-foreground">
              <div class="size-1.5 rounded-full bg-primary"></div>
              Share securely across your team
            </li>
          </ul>
        </div>
        <div>
          <!-- Environment Variables Illustration -->
          <div class="relative">
            <div class="absolute inset-0 bg-gradient-to-br from-primary/20 to-transparent blur-3xl rounded-full -z-10"></div>
            <div class="bg-card border border-border/50 rounded-xl shadow-2xl overflow-hidden">
              <!-- Header -->
              <div class="bg-secondary/50 border-b border-border/50 px-4 py-3 flex items-center justify-between">
                <div class="flex items-center gap-2">
                  <Folder class="size-4 text-primary" />
                  <span class="font-medium text-sm">api-backend</span>
                  <span class="text-xs text-muted-foreground">/</span>
                  <span class="text-xs bg-green-500/20 text-green-400 px-2 py-0.5 rounded">Production</span>
                </div>
                <div class="flex items-center gap-1">
                  <Lock class="size-3.5 text-green-500" />
                  <span class="text-xs text-green-500">Encrypted</span>
                </div>
              </div>
              <!-- Secrets List -->
              <div class="divide-y divide-border/50">
                <div v-for="secret in mockSecrets" :key="secret.key" class="px-4 py-3 flex items-center justify-between hover:bg-secondary/30 transition-colors">
                  <div class="flex-1 min-w-0">
                    <div class="flex items-center gap-2 mb-1">
                      <span class="font-mono text-sm font-medium">{{ secret.key }}</span>
                      <span class="text-xs bg-secondary px-1.5 py-0.5 rounded text-muted-foreground">{{ secret.category }}</span>
                    </div>
                    <div class="font-mono text-xs text-muted-foreground truncate">{{ secret.value }}</div>
                  </div>
                  <div class="flex items-center gap-2 ml-4">
                    <button class="p-1.5 hover:bg-secondary rounded transition-colors text-muted-foreground hover:text-foreground">
                      <EyeOff class="size-4" />
                    </button>
                    <button class="p-1.5 hover:bg-secondary rounded transition-colors text-muted-foreground hover:text-foreground">
                      <Copy class="size-4" />
                    </button>
                  </div>
                </div>
              </div>
              <!-- Footer -->
              <div class="bg-secondary/30 border-t border-border/50 px-4 py-2 flex items-center justify-between text-xs text-muted-foreground">
                <span>3 variables</span>
                <span>Last synced: just now</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Local .env File Sync Feature -->
      <div class="grid lg:grid-cols-2 gap-12 items-center mb-24">
        <div class="lg:order-2">
          <div class="inline-flex items-center rounded-full bg-emerald-500/10 px-4 py-1.5 text-sm font-medium text-emerald-400 border border-emerald-500/20 mb-4">
            <HardDrive class="mr-2 size-4" />
            Local File Sync
          </div>
          <h2 class="text-3xl font-bold mb-4">Local .env File Sync</h2>
          <p class="text-lg text-muted-foreground mb-6 leading-relaxed">
            Link any Envie project to a local .env file on your computer. Each device can have its own linked file, and you control when to sync - push your local changes or pull the latest from your team.
          </p>
          <ul class="space-y-3">
            <li class="flex items-center gap-3 text-muted-foreground">
              <div class="size-1.5 rounded-full bg-emerald-500"></div>
              Link different .env files per device
            </li>
            <li class="flex items-center gap-3 text-muted-foreground">
              <div class="size-1.5 rounded-full bg-emerald-500"></div>
              Automatic change detection
            </li>
            <li class="flex items-center gap-3 text-muted-foreground">
              <div class="size-1.5 rounded-full bg-emerald-500"></div>
              Push local changes to share with team
            </li>
            <li class="flex items-center gap-3 text-muted-foreground">
              <div class="size-1.5 rounded-full bg-emerald-500"></div>
              Pull remote updates to local file
            </li>
          </ul>
        </div>
        <div class="lg:order-1">
          <!-- Local File Sync Illustration -->
          <div class="relative">
            <div class="absolute inset-0 bg-gradient-to-br from-emerald-500/20 to-transparent blur-3xl rounded-full -z-10"></div>
            <div class="bg-card border border-border/50 rounded-xl shadow-2xl overflow-hidden">
              <!-- Header -->
              <div class="bg-secondary/50 border-b border-border/50 px-4 py-3 flex items-center justify-between">
                <div class="flex items-center gap-2">
                  <Folder class="size-4 text-primary" />
                  <span class="font-medium text-sm">my-app</span>
                </div>
                <div class="flex items-center gap-1.5 text-xs">
                  <CheckCircle2 class="size-3.5 text-emerald-500" />
                  <span class="text-emerald-400">Synced</span>
                </div>
              </div>

              <!-- Linked File Info -->
              <div class="p-4 border-b border-border/50">
                <div class="flex items-center gap-3 p-3 bg-emerald-500/5 border border-emerald-500/20 rounded-lg">
                  <div class="size-10 rounded-lg bg-emerald-500/10 flex items-center justify-center text-emerald-500">
                    <FileText class="size-5" />
                  </div>
                  <div class="flex-1 min-w-0">
                    <div class="font-mono text-sm truncate">/Users/dev/my-app/.env</div>
                    <div class="text-xs text-muted-foreground flex items-center gap-1.5 mt-0.5">
                      <Link2 class="size-3" />
                      Linked on this device
                    </div>
                  </div>
                </div>
              </div>

              <!-- Sync Actions -->
              <div class="p-4 space-y-3">
                <div class="flex gap-2">
                  <button class="flex-1 flex items-center justify-center gap-2 px-3 py-2.5 bg-secondary/50 hover:bg-secondary rounded-lg text-sm transition-colors">
                    <Download class="size-4 text-blue-400" />
                    <span>Pull</span>
                  </button>
                  <button class="flex-1 flex items-center justify-center gap-2 px-3 py-2.5 bg-secondary/50 hover:bg-secondary rounded-lg text-sm transition-colors">
                    <Upload class="size-4 text-orange-400" />
                    <span>Push</span>
                  </button>
                </div>
                <p class="text-xs text-muted-foreground text-center">
                  Pull overwrites local file • Push uploads to Envie
                </p>
              </div>

              <!-- Sync Status Examples -->
              <div class="border-t border-border/50 divide-y divide-border/50">
                <div class="px-4 py-2.5 flex items-center justify-between text-xs">
                  <div class="flex items-center gap-2">
                    <RefreshCcw class="size-3.5 text-orange-400" />
                    <span class="text-orange-400">Local file changed</span>
                  </div>
                  <span class="text-muted-foreground">Push to sync</span>
                </div>
                <div class="px-4 py-2.5 flex items-center justify-between text-xs">
                  <div class="flex items-center gap-2">
                    <Download class="size-3.5 text-blue-400" />
                    <span class="text-blue-400">Remote changed</span>
                  </div>
                  <span class="text-muted-foreground">Pull to sync</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Encrypted File Sharing Feature -->
      <div class="grid lg:grid-cols-2 gap-12 items-center mb-24">
        <div class="lg:order-2">
          <div class="inline-flex items-center rounded-full bg-primary/10 px-4 py-1.5 text-sm font-medium text-primary border border-primary/20 mb-4">
            <FolderLock class="mr-2 size-4" />
            Encrypted File Sharing
          </div>
          <h2 class="text-3xl font-bold mb-4">Encrypted File Sharing</h2>
          <p class="text-lg text-muted-foreground mb-6 leading-relaxed">
            Share sensitive files like .env files, certificates, and credentials with your team. Files are encrypted with the same zero-trust model.
          </p>
          <ul class="space-y-3">
            <li class="flex items-center gap-3 text-muted-foreground">
              <div class="size-1.5 rounded-full bg-primary"></div>
              Share any file type securely
            </li>
            <li class="flex items-center gap-3 text-muted-foreground">
              <div class="size-1.5 rounded-full bg-primary"></div>
              Automatic encryption/decryption
            </li>
            <li class="flex items-center gap-3 text-muted-foreground">
              <div class="size-1.5 rounded-full bg-primary"></div>
              Download history tracking
            </li>
            <li class="flex items-center gap-3 text-muted-foreground">
              <div class="size-1.5 rounded-full bg-primary"></div>
              Per-project file storage
            </li>
          </ul>
        </div>
        <div class="lg:order-1">
          <!-- File Sharing Illustration -->
          <div class="relative">
            <div class="absolute inset-0 bg-gradient-to-bl from-blue-500/20 to-transparent blur-3xl rounded-full -z-10"></div>
            <div class="bg-card border border-border/50 rounded-xl shadow-2xl overflow-hidden">
              <!-- Header -->
              <div class="bg-secondary/50 border-b border-border/50 px-4 py-3 flex items-center justify-between">
                <div class="flex items-center gap-2">
                  <FolderLock class="size-4 text-blue-500" />
                  <span class="font-medium text-sm">Project Files</span>
                </div>
                <button class="flex items-center gap-1.5 text-xs bg-primary/10 text-primary px-2.5 py-1.5 rounded-md hover:bg-primary/20 transition-colors">
                  <Upload class="size-3.5" />
                  Upload
                </button>
              </div>
              <!-- Files List -->
              <div class="divide-y divide-border/50">
                <div v-for="file in mockFiles" :key="file.name" class="px-4 py-3 flex items-center justify-between hover:bg-secondary/30 transition-colors">
                  <div class="flex items-center gap-3">
                    <div class="size-10 rounded-lg bg-blue-500/10 flex items-center justify-center text-blue-500">
                      <FileKey class="size-5" />
                    </div>
                    <div>
                      <div class="font-medium text-sm flex items-center gap-2">
                        {{ file.name }}
                        <Lock class="size-3 text-green-500" />
                      </div>
                      <div class="text-xs text-muted-foreground">{{ file.size }}</div>
                    </div>
                  </div>
                  <div class="flex items-center gap-3">
                    <div class="flex -space-x-2">
                      <div v-for="i in file.shared" :key="i" class="size-6 rounded-full bg-secondary border-2 border-card flex items-center justify-center">
                        <User class="size-3 text-muted-foreground" />
                      </div>
                    </div>
                    <button class="p-1.5 hover:bg-secondary rounded transition-colors text-muted-foreground hover:text-foreground">
                      <Download class="size-4" />
                    </button>
                  </div>
                </div>
              </div>
              <!-- Encryption Status -->
              <div class="bg-green-500/5 border-t border-green-500/20 px-4 py-3 flex items-center gap-2">
                <CheckCircle2 class="size-4 text-green-500" />
                <span class="text-xs text-green-400">All files encrypted with XChaCha20-Poly1305</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Organizations & Teams Feature -->
      <div class="grid lg:grid-cols-2 gap-12 items-center mb-20">
        <div>
          <div class="inline-flex items-center rounded-full bg-primary/10 px-4 py-1.5 text-sm font-medium text-primary border border-primary/20 mb-4">
            <Building2 class="mr-2 size-4" />
            Organizations & Teams
          </div>
          <h2 class="text-3xl font-bold mb-4">Organizations & Teams</h2>
          <p class="text-lg text-muted-foreground mb-6 leading-relaxed">
            Structure your secrets with a hierarchical model. Create organizations, organize teams, and manage projects with fine-grained access control.
          </p>
          <ul class="space-y-3">
            <li class="flex items-center gap-3 text-muted-foreground">
              <div class="size-1.5 rounded-full bg-primary"></div>
              Multi-organization support
            </li>
            <li class="flex items-center gap-3 text-muted-foreground">
              <div class="size-1.5 rounded-full bg-primary"></div>
              Flexible team structure
            </li>
            <li class="flex items-center gap-3 text-muted-foreground">
              <div class="size-1.5 rounded-full bg-primary"></div>
              Project-level permissions
            </li>
            <li class="flex items-center gap-3 text-muted-foreground">
              <div class="size-1.5 rounded-full bg-primary"></div>
              Admin and member roles
            </li>
          </ul>
        </div>
        <div>
          <!-- Organization Hierarchy Illustration -->
          <div class="relative">
            <div class="absolute inset-0 bg-gradient-to-br from-purple-500/20 to-transparent blur-3xl rounded-full -z-10"></div>
            <div class="bg-card border border-border/50 rounded-xl shadow-2xl p-6">
              <!-- Organization -->
              <div class="flex flex-col items-center">
                <div class="bg-purple-500/10 border border-purple-500/30 rounded-xl px-6 py-4 flex items-center gap-3 mb-6">
                  <Building2 class="size-6 text-purple-500" />
                  <div>
                    <div class="font-semibold">Acme Corp</div>
                    <div class="text-xs text-muted-foreground">Organization</div>
                  </div>
                </div>

                <!-- Connection Lines -->
                <div class="w-px h-6 bg-border"></div>
                <div class="flex items-center gap-20">
                  <div class="w-20 h-px bg-border"></div>
                  <div class="w-20 h-px bg-border"></div>
                </div>

                <!-- Teams -->
                <div class="flex gap-6 mt-2">
                  <div class="flex flex-col items-center">
                    <div class="w-px h-4 bg-border"></div>
                    <div class="bg-blue-500/10 border border-blue-500/30 rounded-lg px-4 py-3 flex items-center gap-2">
                      <Users class="size-4 text-blue-500" />
                      <div>
                        <div class="font-medium text-sm">Backend</div>
                        <div class="text-xs text-muted-foreground">Team</div>
                      </div>
                    </div>
                    <div class="w-px h-4 bg-border mt-2"></div>
                    <div class="flex gap-2 mt-2">
                      <div class="bg-green-500/10 border border-green-500/30 rounded px-2 py-1.5 text-xs">
                        <div class="text-green-400 font-medium">API</div>
                      </div>
                      <div class="bg-green-500/10 border border-green-500/30 rounded px-2 py-1.5 text-xs">
                        <div class="text-green-400 font-medium">Auth</div>
                      </div>
                    </div>
                  </div>

                  <div class="flex flex-col items-center">
                    <div class="w-px h-4 bg-border"></div>
                    <div class="bg-orange-500/10 border border-orange-500/30 rounded-lg px-4 py-3 flex items-center gap-2">
                      <Users class="size-4 text-orange-500" />
                      <div>
                        <div class="font-medium text-sm">Frontend</div>
                        <div class="text-xs text-muted-foreground">Team</div>
                      </div>
                    </div>
                    <div class="w-px h-4 bg-border mt-2"></div>
                    <div class="flex gap-2 mt-2">
                      <div class="bg-green-500/10 border border-green-500/30 rounded px-2 py-1.5 text-xs">
                        <div class="text-green-400 font-medium">Web</div>
                      </div>
                      <div class="bg-green-500/10 border border-green-500/30 rounded px-2 py-1.5 text-xs">
                        <div class="text-green-400 font-medium">Mobile</div>
                      </div>
                    </div>
                  </div>
                </div>

                <!-- Legend -->
                <div class="flex items-center gap-4 mt-8 text-xs text-muted-foreground">
                  <div class="flex items-center gap-1.5">
                    <div class="size-2 rounded-full bg-purple-500"></div>
                    Organization
                  </div>
                  <div class="flex items-center gap-1.5">
                    <div class="size-2 rounded-full bg-blue-500"></div>
                    Teams
                  </div>
                  <div class="flex items-center gap-1.5">
                    <div class="size-2 rounded-full bg-green-500"></div>
                    Projects
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Google Secret Manager Integration Feature -->
      <div class="grid lg:grid-cols-2 gap-12 items-center mb-20">
        <div class="lg:order-2">
          <div class="inline-flex items-center rounded-full bg-blue-500/10 px-4 py-1.5 text-sm font-medium text-blue-400 border border-blue-500/20 mb-4">
            <Cloud class="mr-2 size-4" />
            Optional Integration
          </div>
          <h2 class="text-3xl font-bold mb-4">Google Secret Manager Sync</h2>
          <p class="text-lg text-muted-foreground mb-6 leading-relaxed">
            Already using Google Secret Manager? Connect directly and pull secrets into Envie. Your service account credentials are encrypted locally - we never see them.
          </p>
          <ul class="space-y-3">
            <li class="flex items-center gap-3 text-muted-foreground">
              <div class="size-1.5 rounded-full bg-blue-500"></div>
              Direct connection to Google Cloud API
            </li>
            <li class="flex items-center gap-3 text-muted-foreground">
              <div class="size-1.5 rounded-full bg-blue-500"></div>
              Service account key encrypted locally
            </li>
            <li class="flex items-center gap-3 text-muted-foreground">
              <div class="size-1.5 rounded-full bg-blue-500"></div>
              No intermediate server - secrets go directly to your app
            </li>
            <li class="flex items-center gap-3 text-muted-foreground">
              <div class="size-1.5 rounded-full bg-blue-500"></div>
              Seamless workflow with existing infrastructure
            </li>
          </ul>
        </div>
        <div class="lg:order-1">
          <!-- Google Secret Manager Illustration -->
          <div class="relative">
            <div class="absolute inset-0 bg-gradient-to-br from-blue-500/20 to-transparent blur-3xl rounded-full -z-10"></div>
            <div class="bg-card border border-border/50 rounded-xl shadow-2xl p-6">
              <div class="space-y-6">
                <!-- Step 1: Your Device -->
                <div class="flex items-center gap-4">
                  <div class="size-12 rounded-xl bg-green-500/10 border border-green-500/30 flex items-center justify-center text-green-500 shrink-0">
                    <Server class="size-6" />
                  </div>
                  <div class="flex-1">
                    <div class="font-medium text-sm mb-1">Your Device</div>
                    <div class="text-xs text-muted-foreground">Decrypts service account key using your Project Key</div>
                  </div>
                </div>

                <!-- Arrow -->
                <div class="flex items-center gap-3 pl-5">
                  <div class="w-px h-8 bg-border"></div>
                  <div class="flex items-center gap-2 text-xs text-muted-foreground">
                    <Lock class="size-3.5 text-green-500" />
                    <span>Decrypted locally</span>
                  </div>
                </div>

                <!-- Step 2: Direct API Call -->
                <div class="flex items-center gap-4">
                  <div class="size-12 rounded-xl bg-blue-500/10 border border-blue-500/30 flex items-center justify-center text-blue-500 shrink-0">
                    <Cloud class="size-6" />
                  </div>
                  <div class="flex-1">
                    <div class="font-medium text-sm mb-1">Google Cloud API</div>
                    <div class="text-xs text-muted-foreground">Direct connection - no Envie servers involved</div>
                  </div>
                </div>

                <!-- Arrow -->
                <div class="flex items-center gap-3 pl-5">
                  <div class="w-px h-8 bg-border"></div>
                  <div class="flex items-center gap-2 text-xs text-muted-foreground">
                    <ArrowRight class="size-3.5" />
                    <span>Secrets fetched directly</span>
                  </div>
                </div>

                <!-- Step 3: Secrets in App -->
                <div class="flex items-center gap-4">
                  <div class="size-12 rounded-xl bg-primary/10 border border-primary/30 flex items-center justify-center text-primary shrink-0">
                    <KeyRound class="size-6" />
                  </div>
                  <div class="flex-1">
                    <div class="font-medium text-sm mb-1">Secrets in Envie</div>
                    <div class="text-xs text-muted-foreground">Available in app state, never stored on our servers</div>
                  </div>
                </div>

                <!-- Note -->
                <div class="bg-blue-500/5 border border-blue-500/20 rounded-lg p-3 mt-4">
                  <p class="text-xs text-blue-400 text-center">
                    Your GCP credentials never touch Envie servers
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Additional Features Grid -->
      <div class="mb-16">
        <h2 class="text-2xl md:text-3xl font-bold text-center mb-12">And much more...</h2>
        <div class="grid md:grid-cols-2 lg:grid-cols-3 gap-6">
          <Card v-for="feature in additionalFeatures" :key="feature.title"
            class="border-border/50 bg-background/80 backdrop-blur-sm hover:border-primary/30 transition-all group"
          >
            <CardHeader>
              <div class="size-11 rounded-lg bg-primary/10 flex items-center justify-center mb-4 text-primary group-hover:bg-primary/20 transition-colors">
                <component :is="feature.icon" class="size-5" />
              </div>
              <CardTitle class="text-lg">{{ feature.title }}</CardTitle>
              <CardDescription class="text-sm leading-relaxed">{{ feature.description }}</CardDescription>
            </CardHeader>
          </Card>
        </div>
      </div>

      <!-- CTA -->
      <div class="text-center bg-secondary/30 rounded-2xl p-12 border border-border/50">
        <h2 class="text-2xl md:text-3xl font-bold mb-4">Ready to secure your secrets?</h2>
        <p class="text-muted-foreground mb-8 max-w-xl mx-auto">
          Envie is open source and free to use. Check out the code on GitHub and start securing your team's secrets today.
        </p>
        <div class="flex flex-col sm:flex-row gap-4 justify-center">
          <Button size="lg" as="a" :href="githubUrl" target="_blank" rel="noopener noreferrer">
            <Github class="mr-2 size-5" />
            View on GitHub
          </Button>
          <Button variant="outline" size="lg" as="a" href="/security">
            Learn about Security
          </Button>
        </div>
      </div>
    </div>
  </div>
</template>
