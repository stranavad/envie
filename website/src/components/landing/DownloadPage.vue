<script setup lang="ts">
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  Download,
  Github,
  Terminal,
} from 'lucide-vue-next'

const githubUrl = 'https://github.com/stranavad/envie'

const linuxDownloads = [
  { name: 'AppImage', file: 'envie.AppImage', recommended: true, description: 'Universal Linux package' },
  { name: '.deb', file: 'envie.deb', recommended: false, description: 'Debian/Ubuntu' },
  { name: '.rpm', file: 'envie.rpm', recommended: false, description: 'Fedora/RHEL' },
]

const macDownload = { name: '.dmg', file: 'envie.dmg', description: 'macOS installer' }
</script>

<template>
  <div class="py-16 md:py-24">
    <div class="container mx-auto px-4">
      <!-- Hero Section -->
      <div class="text-center mb-16 max-w-3xl mx-auto">
        <Badge variant="secondary" class="mb-8 px-4 py-1.5 text-sm rounded-full animate-fade-in-up border border-border/50">
          <Download class="size-3.5 mr-1.5" />
          Desktop Application
        </Badge>

        <h1 class="text-4xl md:text-6xl font-bold tracking-tight mb-6 animate-fade-in-up delay-100">
          Download
          <span class="bg-gradient-to-r from-primary via-primary/80 to-primary/60 bg-clip-text text-transparent">
            Envie
          </span>
        </h1>

        <p class="text-xl text-muted-foreground leading-relaxed animate-fade-in-up delay-200">
          Get Envie for your platform. Available for Linux and macOS with secure local storage.
        </p>
      </div>

      <!-- Platform Cards -->
      <div class="grid md:grid-cols-2 gap-8 max-w-4xl mx-auto mb-16 animate-fade-in-up delay-300">
        <!-- Linux Card -->
        <div class="relative">
          <div class="absolute inset-0 bg-gradient-to-br from-orange-500/20 to-transparent blur-3xl rounded-full -z-10"></div>
          <div class="bg-card border border-border/50 rounded-xl shadow-2xl overflow-hidden h-full">
            <!-- Header -->
            <div class="bg-secondary/50 border-b border-border/50 px-6 py-4 flex items-center gap-3">
              <div class="size-10 rounded-lg bg-orange-500/10 flex items-center justify-center text-orange-500">
                <Terminal class="size-6" />
              </div>
              <div>
                <h3 class="font-semibold text-lg">Linux</h3>
                <p class="text-xs text-muted-foreground">Choose your preferred package format</p>
              </div>
            </div>

            <!-- Download Options -->
            <div class="p-6 space-y-4">
              <!-- Recommended: AppImage -->
              <a
                :href="`/downloads/${linuxDownloads[0].file}`"
                class="flex items-center justify-between p-4 rounded-lg bg-orange-500/10 border border-orange-500/30 hover:bg-orange-500/20 transition-colors group"
              >
                <div class="flex items-center gap-3">
                  <Download class="size-5 text-orange-500" />
                  <div>
                    <div class="font-medium flex items-center gap-2">
                      {{ linuxDownloads[0].name }}
                      <Badge class="bg-orange-500/20 text-orange-400 border-orange-500/30 text-xs">Recommended</Badge>
                    </div>
                    <p class="text-xs text-muted-foreground">{{ linuxDownloads[0].description }}</p>
                  </div>
                </div>
                <Download class="size-4 text-muted-foreground group-hover:text-orange-500 transition-colors" />
              </a>

              <!-- Alternative formats -->
              <div class="grid grid-cols-2 gap-3">
                <a
                  v-for="download in linuxDownloads.slice(1)"
                  :key="download.name"
                  :href="`/downloads/${download.file}`"
                  class="flex items-center gap-2 p-3 rounded-lg border border-border/50 hover:border-orange-500/30 hover:bg-secondary/50 transition-colors group"
                >
                  <Download class="size-4 text-muted-foreground group-hover:text-orange-500 transition-colors" />
                  <div>
                    <div class="font-medium text-sm">{{ download.name }}</div>
                    <p class="text-xs text-muted-foreground">{{ download.description }}</p>
                  </div>
                </a>
              </div>
            </div>
          </div>
        </div>

        <!-- macOS Card -->
        <div class="relative">
          <div class="absolute inset-0 bg-gradient-to-br from-blue-500/20 to-transparent blur-3xl rounded-full -z-10"></div>
          <div class="bg-card border border-border/50 rounded-xl shadow-2xl overflow-hidden h-full">
            <!-- Header -->
            <div class="bg-secondary/50 border-b border-border/50 px-6 py-4 flex items-center gap-3">
              <div class="size-10 rounded-lg bg-blue-500/10 flex items-center justify-center text-blue-500">
                <svg class="size-6" viewBox="0 0 24 24" fill="currentColor">
                  <path d="M18.71 19.5c-.83 1.24-1.71 2.45-3.05 2.47-1.34.03-1.77-.79-3.29-.79-1.53 0-2 .77-3.27.82-1.31.05-2.3-1.32-3.14-2.53C4.25 17 2.94 12.45 4.7 9.39c.87-1.52 2.43-2.48 4.12-2.51 1.28-.02 2.5.87 3.29.87.78 0 2.26-1.07 3.81-.91.65.03 2.47.26 3.64 1.98-.09.06-2.17 1.28-2.15 3.81.03 3.02 2.65 4.03 2.68 4.04-.03.07-.42 1.44-1.38 2.83M13 3.5c.73-.83 1.94-1.46 2.94-1.5.13 1.17-.34 2.35-1.04 3.19-.69.85-1.83 1.51-2.95 1.42-.15-1.15.41-2.35 1.05-3.11z"/>
                </svg>
              </div>
              <div>
                <h3 class="font-semibold text-lg">macOS</h3>
                <p class="text-xs text-muted-foreground">Universal binary for Apple Silicon & Intel</p>
              </div>
            </div>

            <!-- Download Option -->
            <div class="p-6">
              <a
                :href="`/downloads/${macDownload.file}`"
                class="flex items-center justify-between p-4 rounded-lg bg-blue-500/10 border border-blue-500/30 hover:bg-blue-500/20 transition-colors group"
              >
                <div class="flex items-center gap-3">
                  <Download class="size-5 text-blue-500" />
                  <div>
                    <div class="font-medium">{{ macDownload.name }}</div>
                    <p class="text-xs text-muted-foreground">{{ macDownload.description }}</p>
                  </div>
                </div>
                <Download class="size-4 text-muted-foreground group-hover:text-blue-500 transition-colors" />
              </a>

              <!-- Note -->
              <div class="mt-4 p-3 rounded-lg bg-secondary/50 border border-border/50">
                <p class="text-xs text-muted-foreground">
                  Supports both Apple Silicon (M1/M2/M3) and Intel-based Macs.
                </p>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Bottom CTA -->
      <div class="text-center bg-secondary/30 rounded-2xl p-12 border border-border/50 max-w-4xl mx-auto animate-fade-in-up delay-400">
        <h2 class="text-2xl md:text-3xl font-bold mb-4">Prefer to build from source?</h2>
        <p class="text-muted-foreground mb-8 max-w-xl mx-auto">
          Envie is open source. Clone the repository and build it yourself for maximum transparency.
        </p>
        <div class="flex flex-col sm:flex-row gap-4 justify-center">
          <Button size="lg" as="a" :href="githubUrl" target="_blank" rel="noopener noreferrer">
            <Github class="mr-2 size-5" />
            View on GitHub
          </Button>
          <Button variant="outline" size="lg" as="a" :href="`${githubUrl}#building-from-source`" target="_blank" rel="noopener noreferrer">
            <Terminal class="mr-2 size-5" />
            Build Instructions
          </Button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.animate-fade-in-up {
  animation: fadeInUp 0.8s cubic-bezier(0.16, 1, 0.3, 1) forwards;
  opacity: 0;
  transform: translateY(20px);
}

.delay-100 { animation-delay: 0.1s; }
.delay-200 { animation-delay: 0.2s; }
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
