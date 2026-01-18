<script setup lang="ts">
import { Button } from '@/components/ui/button'
import { ShieldCheck, Lock, Key, Monitor, ArrowRight } from 'lucide-vue-next'
</script>

<template>
  <section class="py-24 bg-background">
    <div class="container mx-auto px-4">
      <div class="flex flex-col lg:flex-row items-center gap-16">
        <div class="flex-1 space-y-6">
          <div class="inline-flex items-center rounded-full bg-primary/10 px-4 py-1.5 text-sm font-medium text-primary border border-primary/20">
            <ShieldCheck class="mr-2 size-4" />
            Zero Trust Architecture
          </div>
          <h2 class="text-3xl md:text-5xl font-bold tracking-tight leading-tight">
            Your secrets never leave your device unencrypted.
          </h2>
          <p class="text-lg text-muted-foreground leading-relaxed">
            Envie is built on zero-trust principles. All encryption and decryption happens locally on your device.
            The server stores only encrypted blobs that are meaningless without your keys.
          </p>

          <div class="grid gap-5 pt-4">
            <div class="flex gap-4 p-4 rounded-lg bg-secondary/30 border border-border/50">
              <div class="bg-primary/10 p-2.5 rounded-lg h-fit text-primary">
                <Lock class="size-5" />
              </div>
              <div>
                <h3 class="font-semibold mb-1">Client-Side Encryption</h3>
                <p class="text-sm text-muted-foreground">
                   All data is encrypted locally using XChaCha20-Poly1305. The server only stores ciphertext.
                </p>
              </div>
            </div>
            <div class="flex gap-4 p-4 rounded-lg bg-secondary/30 border border-border/50">
              <div class="bg-primary/10 p-2.5 rounded-lg h-fit text-primary">
                <Key class="size-5" />
              </div>
              <div>
                <h3 class="font-semibold mb-1">You Hold All the Keys</h3>
                <p class="text-sm text-muted-foreground">
                  Decryption keys never leave your app. Not even Envie can access your secrets.
                </p>
              </div>
            </div>
            <div class="flex gap-4 p-4 rounded-lg bg-secondary/30 border border-border/50">
              <div class="bg-primary/10 p-2.5 rounded-lg h-fit text-primary">
                <Monitor class="size-5" />
              </div>
              <div>
                <h3 class="font-semibold mb-1">Secure Local Storage</h3>
                <p class="text-sm text-muted-foreground">
                  Keys are stored in Stronghold, the same secure storage used by cryptocurrency wallets.
                </p>
              </div>
            </div>
          </div>

          <div class="pt-4">
            <Button variant="outline" as="a" href="/security">
              Learn more about security
              <ArrowRight class="ml-2 size-4" />
            </Button>
          </div>
        </div>

        <div class="flex-1 relative max-w-md w-full">
            <div class="absolute inset-0 bg-gradient-to-tr from-primary/20 via-primary/5 to-transparent blur-3xl rounded-full"></div>
            <div class="relative bg-card border border-border/50 rounded-xl p-8 shadow-2xl">
                <div class="space-y-5 font-mono text-sm">
                    <div class="flex justify-between items-center">
                        <span class="text-muted-foreground text-xs uppercase tracking-wider">Your Device</span>
                        <span class="text-xs bg-green-500/20 text-green-400 px-2 py-0.5 rounded">Decrypted</span>
                    </div>
                    <div class="p-4 bg-green-500/10 border border-green-500/20 rounded text-xs">
                        <div class="text-green-400">DATABASE_URL=postgres://...</div>
                        <div class="text-green-400">API_KEY=sk_live_...</div>
                    </div>

                    <div class="flex justify-center py-2">
                        <div class="flex items-center gap-2 text-muted-foreground">
                            <Lock class="size-4" />
                            <span class="text-xs">Encrypted in transit</span>
                        </div>
                    </div>

                    <div class="flex justify-between items-center">
                        <span class="text-muted-foreground text-xs uppercase tracking-wider">Envie Server</span>
                        <span class="text-xs bg-primary/20 text-primary px-2 py-0.5 rounded">Encrypted</span>
                    </div>
                    <div class="p-4 bg-muted/50 rounded break-all text-xs text-muted-foreground">
                        7A93F4B2...8B2C1D9E<br/>
                        F3E8A7C1...2D4F6B8A
                    </div>
                </div>
            </div>
        </div>
      </div>
    </div>
  </section>
</template>
