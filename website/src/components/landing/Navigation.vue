<script setup lang="ts">
import { Button } from '@/components/ui/button'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from '@/components/ui/sheet'
import { AlignRight, Github } from 'lucide-vue-next'

const links = [
  { name: 'Features', href: '/features' },
  { name: 'Security', href: '/security' },
]

const githubUrl = 'https://github.com/stranavad/envie'
</script>

<template>
  <nav class="sticky top-0 z-50 w-full border-b border-border/40 bg-background/80 backdrop-blur-xl">
    <div class="container mx-auto px-4 h-16 flex items-center justify-between">
      <a href="/" class="flex items-center gap-2.5 font-bold text-xl group">
        <div class="size-9 bg-primary rounded-lg flex items-center justify-center text-primary-foreground font-semibold shadow-lg shadow-primary/20 group-hover:shadow-primary/40 transition-shadow">
          E
        </div>
        <span class="tracking-tight">Envie</span>
      </a>

      <!-- Desktop Menu -->
      <div class="hidden md:flex items-center gap-8">
        <a v-for="link in links" :key="link.name" :href="link.href" class="text-sm font-medium text-muted-foreground hover:text-foreground transition-colors">
          {{ link.name }}
        </a>
        <div class="flex items-center gap-3">
          <Button variant="ghost" size="icon" as="a" :href="githubUrl" target="_blank" rel="noopener noreferrer">
            <Github class="size-5" />
          </Button>
          <Button as="a" :href="githubUrl" target="_blank" rel="noopener noreferrer">
            View on GitHub
          </Button>
        </div>
      </div>

      <!-- Mobile Menu -->
      <div class="md:hidden">
        <Sheet>
          <SheetTrigger as-child>
            <Button variant="ghost" size="icon">
              <AlignRight class="size-6" />
            </Button>
          </SheetTrigger>
          <SheetContent>
            <SheetHeader>
                <SheetTitle class="text-left">Menu</SheetTitle>
                <SheetDescription class="sr-only">Mobile navigation menu</SheetDescription>
            </SheetHeader>
            <div class="flex flex-col gap-4 mt-8">
              <a v-for="link in links" :key="link.name" :href="link.href" class="text-lg font-medium text-muted-foreground hover:text-foreground transition-colors">
                {{ link.name }}
              </a>
              <a :href="githubUrl" target="_blank" rel="noopener noreferrer" class="flex items-center gap-2 text-lg font-medium text-muted-foreground hover:text-foreground transition-colors">
                <Github class="size-5" />
                GitHub
              </a>
            </div>
          </SheetContent>
        </Sheet>
      </div>
    </div>
  </nav>
</template>
