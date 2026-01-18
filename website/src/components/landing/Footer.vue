<script setup lang="ts">
import { Separator } from '@/components/ui/separator'
import { Github } from 'lucide-vue-next'

const githubUrl = 'https://github.com/stranavad/envie'
</script>

<template>
  <footer class="py-12 bg-background border-t border-border/40">
    <div class="container mx-auto px-4">
      <div class="grid grid-cols-1 md:grid-cols-3 gap-8 mb-12">
        <div>
            <a href="/" class="flex items-center gap-2.5 font-bold text-xl mb-4">
                <div class="size-9 bg-primary rounded-lg flex items-center justify-center text-primary-foreground shadow-lg shadow-primary/20">
                  E
                </div>
                <span class="tracking-tight">Envie</span>
            </a>
            <p class="text-muted-foreground max-w-sm text-sm leading-relaxed">
                The zero-trust secret manager. Your secrets, your control.
                Open source and end-to-end encrypted.
            </p>
        </div>
        <div>
            <h4 class="font-semibold mb-4">Product</h4>
            <ul class="space-y-2.5 text-sm text-muted-foreground">
                <li><a href="/features" class="hover:text-foreground transition-colors">Features</a></li>
                <li><a href="/security" class="hover:text-foreground transition-colors">Security</a></li>
            </ul>
        </div>
        <div>
            <h4 class="font-semibold mb-4">Resources</h4>
            <ul class="space-y-2.5 text-sm text-muted-foreground">
                <li>
                  <a :href="githubUrl" target="_blank" rel="noopener noreferrer" class="flex items-center gap-2 hover:text-foreground transition-colors">
                    <Github class="size-4" />
                    GitHub
                  </a>
                </li>
            </ul>
        </div>
      </div>

      <Separator class="mb-8" />

      <div class="flex flex-col md:flex-row justify-between items-center text-sm text-muted-foreground gap-4">
        <p>&copy; {{ new Date().getFullYear() }} Envie. Open source under MIT license.</p>
        <a :href="githubUrl" target="_blank" rel="noopener noreferrer" class="flex items-center gap-2 hover:text-foreground transition-colors">
          <Github class="size-4" />
          Star us on GitHub
        </a>
      </div>
    </div>
  </footer>
</template>
