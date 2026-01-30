<script setup lang="ts">
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Github, Shield, Lock, Monitor, Key, Folder, EyeOff, Copy, CheckCircle2 } from 'lucide-vue-next'

const githubUrl = 'https://github.com/stranavad/envie'

const mockSecrets = [
  { key: 'DATABASE_URL', value: 'postgres://prod:***@db.example.com', category: 'Production' },
  { key: 'STRIPE_SECRET', value: 'sk_live_***************', category: 'Production' },
  { key: 'JWT_SECRET', value: '************************', category: 'All' },
]
</script>

<template>
  <section class="relative pt-20 pb-24">
    <!-- Gradient background that fades out -->
    <div class="absolute inset-x-0 -top-20 h-[800px] overflow-hidden pointer-events-none">
      <div class="absolute top-20 left-1/2 -translate-x-1/2 w-[800px] h-[600px] bg-primary/15 rounded-full blur-[120px]"></div>
      <div class="absolute top-40 left-1/4 w-[400px] h-[400px] bg-cyan-500/10 rounded-full blur-[100px]"></div>
      <div class="absolute top-32 right-1/4 w-[300px] h-[300px] bg-violet-500/10 rounded-full blur-[80px]"></div>
      <!-- Fade out at bottom -->
      <div class="absolute inset-x-0 bottom-0 h-48 bg-gradient-to-t from-background to-transparent"></div>
    </div>

    <!-- Subtle grid background -->
    <div class="absolute inset-0 bg-[linear-gradient(to_right,hsl(var(--border)/0.3)_1px,transparent_1px),linear-gradient(to_bottom,hsl(var(--border)/0.3)_1px,transparent_1px)] bg-[size:4rem_4rem] [mask-image:radial-gradient(ellipse_60%_40%_at_50%_20%,#000_40%,transparent_100%)] pointer-events-none"></div>

    <div class="container mx-auto px-4 relative z-10">
      <div class="grid lg:grid-cols-2 gap-12 lg:gap-16 items-center">
        <!-- Left side - Text content -->
        <div class="flex flex-col items-center lg:items-start text-center lg:text-left">
          <Badge variant="secondary" class="mb-6 px-4 py-1.5 text-sm rounded-full animate-fade-in-up border border-primary/20 bg-primary/5">
            <Shield class="size-3.5 mr-1.5 text-primary" />
            Open Source & End-to-End Encrypted
          </Badge>

          <h1 class="text-4xl sm:text-5xl lg:text-6xl font-bold tracking-tight mb-6 animate-fade-in-up delay-100 leading-[1.1]">
            Your secrets.
            <span class="block mt-2">
              Your control.
            </span>
          </h1>

          <p class="text-lg lg:text-xl text-muted-foreground max-w-xl mb-8 animate-fade-in-up delay-200 leading-relaxed">
            The zero-trust secret manager where your data never leaves your device unencrypted.
            Not even we can see your secrets.
          </p>

          <div class="flex flex-wrap items-center justify-center lg:justify-start gap-3 text-sm text-muted-foreground mb-8 animate-fade-in-up delay-250">
            <div class="flex items-center gap-1.5 bg-secondary/80 px-3 py-1.5 rounded-full border border-border/50">
              <Lock class="size-3.5 text-primary" />
              <span>Client-side encryption</span>
            </div>
            <div class="flex items-center gap-1.5 bg-secondary/80 px-3 py-1.5 rounded-full border border-border/50">
              <Monitor class="size-3.5 text-primary" />
              <span>Desktop app</span>
            </div>
          </div>

          <div class="flex flex-col sm:flex-row gap-3 animate-fade-in-up delay-300">
            <Button size="lg" class="h-12 px-8 text-base shadow-lg shadow-primary/20" as="a" :href="githubUrl" target="_blank" rel="noopener noreferrer">
              <Github class="mr-2 size-5" />
              View on GitHub
            </Button>
            <Button variant="outline" size="lg" class="h-12 px-8 text-base" as="a" href="/features">
              Explore Features
            </Button>
          </div>
        </div>

        <!-- Right side - App mockup -->
        <div class="relative animate-fade-in-up delay-400">
          <div class="absolute inset-0 bg-gradient-to-br from-primary/20 via-transparent to-violet-500/10 blur-3xl rounded-full -z-10"></div>

          <!-- Main card -->
          <div class="bg-card/80 backdrop-blur-xl border border-border/50 rounded-2xl shadow-2xl overflow-hidden">
            <!-- Window header -->
            <div class="bg-secondary/50 border-b border-border/50 px-3 sm:px-4 py-2.5 sm:py-3 flex items-center justify-between">
              <div class="flex items-center gap-2">
                <div class="hidden sm:flex gap-1.5">
                  <div class="size-3 rounded-full bg-red-500/80"></div>
                  <div class="size-3 rounded-full bg-yellow-500/80"></div>
                  <div class="size-3 rounded-full bg-green-500/80"></div>
                </div>
                <span class="text-xs text-muted-foreground sm:ml-2 font-medium">Envie</span>
              </div>
              <div class="flex items-center gap-1.5">
                <Lock class="size-3.5 text-green-500" />
                <span class="text-xs text-green-500">Encrypted</span>
              </div>
            </div>

            <!-- Project header -->
            <div class="px-3 sm:px-5 py-3 sm:py-4 border-b border-border/50 flex items-center justify-between">
              <div class="flex items-center gap-2 sm:gap-3">
                <div class="size-8 sm:size-10 rounded-lg bg-primary/10 flex items-center justify-center text-primary">
                  <Folder class="size-4 sm:size-5" />
                </div>
                <div>
                  <div class="font-semibold text-xs sm:text-sm">api-backend</div>
                  <div class="text-xs text-muted-foreground hidden sm:block">Production environment</div>
                </div>
              </div>
              <div class="flex items-center gap-1.5 text-xs bg-green-500/10 text-green-400 px-2 sm:px-2.5 py-1 rounded-full border border-green-500/20">
                <CheckCircle2 class="size-3" />
                <span class="hidden sm:inline">Synced</span>
              </div>
            </div>

            <!-- Secrets list -->
            <div class="divide-y divide-border/50">
              <div v-for="(secret, index) in mockSecrets" :key="secret.key"
                class="px-3 sm:px-5 py-2.5 sm:py-3.5 flex items-center justify-between hover:bg-secondary/30 transition-colors"
                :class="{ 'bg-secondary/20': index === 1 }"
              >
                <div class="flex-1 min-w-0">
                  <div class="flex items-center gap-1.5 sm:gap-2 mb-0.5 sm:mb-1">
                    <Key class="size-3 sm:size-3.5 text-primary shrink-0" />
                    <span class="font-mono text-xs sm:text-sm font-medium truncate">{{ secret.key }}</span>
                    <span class="text-xs bg-secondary px-1 sm:px-1.5 py-0.5 rounded text-muted-foreground hidden sm:inline">{{ secret.category }}</span>
                  </div>
                  <div class="font-mono text-xs text-muted-foreground truncate pl-4 sm:pl-5 hidden sm:block">{{ secret.value }}</div>
                </div>
                <div class="flex items-center gap-1 sm:gap-1.5 ml-2 sm:ml-4">
                  <button class="p-1 sm:p-1.5 hover:bg-secondary rounded transition-colors text-muted-foreground hover:text-foreground">
                    <EyeOff class="size-3 sm:size-3.5" />
                  </button>
                  <button class="p-1 sm:p-1.5 hover:bg-secondary rounded transition-colors text-muted-foreground hover:text-foreground">
                    <Copy class="size-3 sm:size-3.5" />
                  </button>
                </div>
              </div>
            </div>

            <!-- Footer -->
            <div class="bg-secondary/30 border-t border-border/50 px-3 sm:px-5 py-2 sm:py-3 flex items-center justify-between text-xs text-muted-foreground">
              <span>3 secrets</span>
              <span class="flex items-center gap-1.5">
                <div class="size-1.5 rounded-full bg-green-500 animate-pulse"></div>
                <span class="hidden sm:inline">Last synced:</span> just now
              </span>
            </div>
          </div>

          <!-- Floating encryption badge - hidden on mobile -->
          <div class="absolute -bottom-4 -left-4 bg-card border border-border/50 rounded-xl px-3 sm:px-4 py-2 sm:py-3 shadow-xl hidden sm:flex items-center gap-2 sm:gap-3">
            <div class="size-8 sm:size-10 rounded-lg bg-green-500/10 flex items-center justify-center text-green-500">
              <Shield class="size-4 sm:size-5" />
            </div>
            <div>
              <div class="text-xs font-medium">XChaCha20-Poly1305</div>
              <div class="text-xs text-muted-foreground hidden sm:block">Military-grade encryption</div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>

<style scoped>
.animate-fade-in-up {
  animation: fadeInUp 0.8s cubic-bezier(0.16, 1, 0.3, 1) forwards;
  opacity: 0;
  transform: translateY(20px);
}

.delay-100 { animation-delay: 0.1s; }
.delay-200 { animation-delay: 0.2s; }
.delay-250 { animation-delay: 0.25s; }
.delay-300 { animation-delay: 0.3s; }
.delay-400 { animation-delay: 0.4s; }

@keyframes fadeInUp {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style>
