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
  Shield,
  Lock,
  Key,
  Monitor,
  Users,
  RefreshCw,
  Fingerprint,
  Server,
  Github,
  CheckCircle2
} from 'lucide-vue-next'

const githubUrl = 'https://github.com/stranavad/envie'

const securityPrinciples = [
  {
    title: 'Zero-Knowledge Architecture',
    description: 'Envie servers never have access to your unencrypted data. All encryption and decryption happens locally on your device.',
    icon: Shield,
  },
  {
    title: 'Client-Side Encryption',
    description: 'Every piece of data is encrypted using XChaCha20-Poly1305 before leaving your device. The server only stores encrypted blobs.',
    icon: Lock,
  },
  {
    title: 'Key Ownership',
    description: 'You control all encryption keys. Project keys, team keys, and your master identity key never leave your device unencrypted.',
    icon: Key,
  },
  {
    title: 'Secure Local Storage',
    description: 'Encryption keys are stored in Stronghold, a secure storage library designed for cryptocurrency wallets.',
    icon: Monitor,
  },
]

const keyHierarchy = [
  {
    name: 'Master Identity Key',
    description: 'Your personal key that unlocks access to all your projects and teams. Stored securely in Stronghold on your device.',
    color: 'primary',
  },
  {
    name: 'Team Keys',
    description: 'Shared keys for team collaboration. Encrypted with each team member\'s public key so only members can decrypt.',
    color: 'blue-500',
  },
  {
    name: 'Project Keys',
    description: 'Individual keys for each project. Can be rotated independently when team members leave.',
    color: 'green-500',
  },
  {
    name: 'File Encryption Keys',
    description: 'Each file has its own encryption key, further wrapped by the project key for defense in depth.',
    color: 'yellow-500',
  },
]

const securityFeatures = [
  {
    title: 'Double-Admin Key Rotation',
    description: 'For organizations, key rotation requires approval from two admins. This prevents a single compromised account from affecting all secrets.',
    icon: RefreshCw,
  },
  {
    title: 'Device Identities',
    description: 'Each device generates its own Ed25519 key pair. New devices must be approved by existing devices before accessing secrets.',
    icon: Fingerprint,
  },
  {
    title: 'Asymmetric Key Exchange',
    description: 'Team keys are securely shared using X25519 key exchange. Only recipients with the correct private key can decrypt.',
    icon: Users,
  },
  {
    title: 'Server Blindness',
    description: 'The server facilitates sync but cannot read your data. Even a compromised server reveals nothing useful to attackers.',
    icon: Server,
  },
]
</script>

<template>
  <div class="py-16 md:py-24">
    <div class="container mx-auto px-4">
      <!-- Header -->
      <div class="text-center mb-16 max-w-3xl mx-auto">
        <div class="inline-flex items-center rounded-full bg-primary/10 px-4 py-1.5 text-sm font-medium text-primary border border-primary/20 mb-6">
          <Shield class="mr-2 size-4" />
          Security
        </div>
        <h1 class="text-4xl md:text-6xl font-bold mb-6 tracking-tight">
          Security you can verify
        </h1>
        <p class="text-xl text-muted-foreground leading-relaxed">
          Envie is built on zero-trust principles. We don't ask you to trust us &mdash;
          the code is open source and the architecture ensures we can't access your secrets even if we wanted to.
        </p>
      </div>

      <!-- Security Principles -->
      <div class="grid md:grid-cols-2 gap-6 mb-20">
        <Card v-for="principle in securityPrinciples" :key="principle.title"
          class="border-border/50 bg-secondary/20 backdrop-blur-sm"
        >
          <CardHeader>
            <div class="size-12 rounded-lg bg-primary/10 flex items-center justify-center mb-4 text-primary">
              <component :is="principle.icon" class="size-6" />
            </div>
            <CardTitle class="text-xl mb-2">{{ principle.title }}</CardTitle>
            <CardDescription class="text-base leading-relaxed">{{ principle.description }}</CardDescription>
          </CardHeader>
        </Card>
      </div>

      <!-- Encryption Flow Diagram -->
      <div class="mb-20">
        <h2 class="text-2xl md:text-3xl font-bold text-center mb-4">How encryption works</h2>
        <p class="text-muted-foreground text-center mb-12 max-w-2xl mx-auto">
          All encryption happens on your device. The server only sees encrypted data.
        </p>

        <div class="max-w-4xl mx-auto">
          <div class="grid md:grid-cols-3 gap-6 items-start">
            <!-- Your Device -->
            <Card class="border-green-500/30 bg-green-500/5">
              <CardHeader class="pb-2">
                <div class="flex items-center gap-2 text-green-400 mb-2">
                  <Monitor class="size-5" />
                  <span class="font-semibold">Your Device</span>
                </div>
              </CardHeader>
              <CardContent class="space-y-3 text-sm">
                <div class="p-3 bg-green-500/10 border border-green-500/20 rounded font-mono text-xs text-green-400">
                  DATABASE_URL=postgres://...<br/>
                  API_KEY=sk_live_...
                </div>
                <div class="flex items-center gap-2 text-muted-foreground">
                  <CheckCircle2 class="size-4 text-green-500" />
                  <span>Plaintext data</span>
                </div>
                <div class="flex items-center gap-2 text-muted-foreground">
                  <CheckCircle2 class="size-4 text-green-500" />
                  <span>Encryption keys</span>
                </div>
              </CardContent>
            </Card>

            <!-- Arrow / Encryption -->
            <div class="flex flex-col items-center justify-center py-8 md:py-0">
              <div class="flex items-center gap-3 text-primary mb-2">
                <Lock class="size-5" />
                <span class="text-sm font-medium">XChaCha20-Poly1305</span>
              </div>
              <div class="w-full h-px bg-gradient-to-r from-transparent via-primary to-transparent"></div>
              <p class="text-xs text-muted-foreground mt-2 text-center">
                Encrypted before leaving device
              </p>
            </div>

            <!-- Server -->
            <Card class="border-border/50 bg-secondary/20">
              <CardHeader class="pb-2">
                <div class="flex items-center gap-2 text-muted-foreground mb-2">
                  <Server class="size-5" />
                  <span class="font-semibold">Envie Server</span>
                </div>
              </CardHeader>
              <CardContent class="space-y-3 text-sm">
                <div class="p-3 bg-muted/50 rounded font-mono text-xs text-muted-foreground break-all">
                  7A93F4B2C1D8E9F0...<br/>
                  3B2A1C4D5E6F7890...
                </div>
                <div class="flex items-center gap-2 text-muted-foreground">
                  <CheckCircle2 class="size-4" />
                  <span>Only encrypted blobs</span>
                </div>
                <div class="flex items-center gap-2 text-muted-foreground">
                  <CheckCircle2 class="size-4" />
                  <span>Cannot decrypt</span>
                </div>
              </CardContent>
            </Card>
          </div>
        </div>
      </div>

      <!-- Key Hierarchy -->
      <div class="mb-20">
        <h2 class="text-2xl md:text-3xl font-bold text-center mb-4">Key hierarchy</h2>
        <p class="text-muted-foreground text-center mb-12 max-w-2xl mx-auto">
          Envie uses a hierarchical key structure for maximum security and flexibility.
        </p>

        <div class="max-w-3xl mx-auto space-y-4">
          <div v-for="(key, index) in keyHierarchy" :key="key.name"
            class="flex items-start gap-4 p-5 rounded-lg border border-border/50 bg-secondary/10"
          >
            <div class="flex items-center justify-center size-8 rounded-full text-sm font-bold"
              :class="`bg-${key.color}/20 text-${key.color}`"
              :style="{ backgroundColor: `hsl(var(--${key.color === 'primary' ? 'primary' : key.color.replace('-500', '')}) / 0.2)` }"
            >
              {{ index + 1 }}
            </div>
            <div>
              <h3 class="font-semibold mb-1">{{ key.name }}</h3>
              <p class="text-sm text-muted-foreground">{{ key.description }}</p>
            </div>
          </div>
        </div>
      </div>

      <!-- Additional Security Features -->
      <div class="mb-20">
        <h2 class="text-2xl md:text-3xl font-bold text-center mb-12">Additional security measures</h2>
        <div class="grid md:grid-cols-2 gap-6">
          <div v-for="feature in securityFeatures" :key="feature.title"
            class="flex gap-4 p-5 rounded-lg border border-border/50 bg-background/80"
          >
            <div class="size-10 rounded-lg bg-primary/10 flex items-center justify-center text-primary shrink-0">
              <component :is="feature.icon" class="size-5" />
            </div>
            <div>
              <h3 class="font-semibold mb-1">{{ feature.title }}</h3>
              <p class="text-sm text-muted-foreground leading-relaxed">{{ feature.description }}</p>
            </div>
          </div>
        </div>
      </div>

      <!-- Open Source CTA -->
      <div class="text-center bg-secondary/30 rounded-2xl p-12 border border-border/50">
        <h2 class="text-2xl md:text-3xl font-bold mb-4">Verify it yourself</h2>
        <p class="text-muted-foreground mb-8 max-w-xl mx-auto">
          Envie is fully open source. Review the code, audit the encryption, and verify our security claims yourself.
        </p>
        <Button size="lg" as="a" :href="githubUrl" target="_blank" rel="noopener noreferrer">
          <Github class="mr-2 size-5" />
          View Source on GitHub
        </Button>
      </div>
    </div>
  </div>
</template>
