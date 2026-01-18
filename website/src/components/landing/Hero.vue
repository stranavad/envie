<script setup lang="ts">
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Github, Shield, Lock, Monitor } from 'lucide-vue-next'

const githubUrl = 'https://github.com/stranavad/envie'
</script>

<template>
  <section class="relative pt-24 pb-32 overflow-hidden">
    <div class="container mx-auto px-4 relative z-10 flex flex-col items-center text-center">

      <Badge variant="secondary" class="mb-8 px-4 py-1.5 text-sm rounded-full animate-fade-in-up border border-border/50">
        <Shield class="size-3.5 mr-1.5" />
        Open Source &amp; End-to-End Encrypted
      </Badge>

      <h1 class="text-5xl md:text-7xl font-bold tracking-tight mb-6 animate-fade-in-up delay-100 max-w-5xl leading-[1.1]">
        Your secrets.
        <span class="bg-gradient-to-r from-primary via-primary/80 to-primary/60 bg-clip-text text-transparent">
          Your control.
        </span>
      </h1>

      <p class="text-xl md:text-2xl text-muted-foreground max-w-3xl mb-6 animate-fade-in-up delay-200 leading-relaxed">
        The zero-trust secret manager where your data never leaves your device unencrypted.
        Not even we can see your secrets.
      </p>

      <div class="flex flex-wrap items-center justify-center gap-4 text-sm text-muted-foreground mb-10 animate-fade-in-up delay-250">
        <div class="flex items-center gap-1.5 bg-secondary/50 px-3 py-1.5 rounded-full">
          <Lock class="size-3.5" />
          <span>Client-side encryption</span>
        </div>
        <div class="flex items-center gap-1.5 bg-secondary/50 px-3 py-1.5 rounded-full">
          <Monitor class="size-3.5" />
          <span>Desktop app with secure storage</span>
        </div>
      </div>

      <div class="flex flex-col sm:flex-row gap-4 animate-fade-in-up delay-300">
        <Button size="lg" class="h-12 px-8 text-base" as="a" :href="githubUrl" target="_blank" rel="noopener noreferrer">
          <Github class="mr-2 size-5" />
          View on GitHub
        </Button>
        <Button variant="outline" size="lg" class="h-12 px-8 text-base" as="a" href="/features">
          Explore Features
        </Button>
      </div>

      <!-- Abstract Background Elements -->
      <div class="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[900px] h-[900px] bg-gradient-to-br from-primary/15 via-primary/5 to-transparent rounded-full blur-3xl -z-10 pointer-events-none"></div>
      <div class="absolute top-1/3 left-1/4 w-[300px] h-[300px] bg-primary/10 rounded-full blur-3xl -z-10 pointer-events-none"></div>
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
