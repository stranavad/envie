<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
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
  { name: 'CI/CD', href: '/cicd' },
  { name: 'Security', href: '/security' },
  { name: 'Download', href: '/download' },
]

const githubUrl = 'https://github.com/stranavad/envie'

const isScrolled = ref(false)

function handleScroll() {
  isScrolled.value = window.scrollY > 20
}

onMounted(() => {
  window.addEventListener('scroll', handleScroll, { passive: true })
  handleScroll() // Check initial state
})

onUnmounted(() => {
  window.removeEventListener('scroll', handleScroll)
})
</script>

<template>
  <nav
    class="sticky top-0 z-50 w-full transition-all duration-300"
    :class="isScrolled
      ? 'bg-background/80 backdrop-blur-xl border-b border-border/40'
      : 'bg-transparent border-b border-transparent'"
  >
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
          <SheetContent class="w-[280px] p-0">
            <SheetHeader class="sr-only">
                <SheetTitle>Menu</SheetTitle>
                <SheetDescription>Mobile navigation menu</SheetDescription>
            </SheetHeader>

            <!-- Logo -->
            <div class="p-6 border-b border-border/50">
              <a href="/" class="flex items-center gap-2.5 font-bold text-xl">
                <div class="size-9 bg-primary rounded-lg flex items-center justify-center text-primary-foreground font-semibold shadow-lg shadow-primary/20">
                  E
                </div>
                <span class="tracking-tight">Envie</span>
              </a>
            </div>

            <!-- Navigation Links -->
            <nav class="p-4">
              <ul class="space-y-1">
                <li v-for="link in links" :key="link.name">
                  <a
                    :href="link.href"
                    class="flex items-center px-4 py-3 rounded-lg text-base font-medium text-muted-foreground hover:text-foreground hover:bg-secondary/80 transition-colors"
                  >
                    {{ link.name }}
                  </a>
                </li>
              </ul>

              <!-- Separator -->
              <div class="my-4 border-t border-border/50"></div>

              <!-- GitHub Link -->
              <a
                :href="githubUrl"
                target="_blank"
                rel="noopener noreferrer"
                class="flex items-center gap-3 px-4 py-3 rounded-lg text-base font-medium text-muted-foreground hover:text-foreground hover:bg-secondary/80 transition-colors"
              >
                <Github class="size-5" />
                GitHub
              </a>
            </nav>

            <!-- CTA Button -->
            <div class="absolute bottom-0 left-0 right-0 p-6 border-t border-border/50 bg-background">
              <Button class="w-full" as="a" :href="githubUrl" target="_blank" rel="noopener noreferrer">
                <Github class="mr-2 size-4" />
                View on GitHub
              </Button>
            </div>
          </SheetContent>
        </Sheet>
      </div>
    </div>
  </nav>
</template>
