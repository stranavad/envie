<script setup lang="ts">
import {
  Card,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { KeyRound, FolderLock, Users, RefreshCw, Cloud, ArrowRight, HardDrive } from 'lucide-vue-next'

const features = [
  {
    title: 'Environment Variables',
    description: 'Store and manage environment variables with end-to-end encryption. Your secrets are encrypted before they leave your device.',
    icon: KeyRound,
  },
  {
    title: 'Local .env File Sync',
    description: 'Link projects to local .env files on each device. Push local changes or pull remote updates with one click.',
    icon: HardDrive,
  },
  {
    title: 'Encrypted File Sharing',
    description: 'Share sensitive files like .env, credentials, and certificates securely within your team.',
    icon: FolderLock,
  },
  {
    title: 'Teams & Collaboration',
    description: 'Structure secrets with organizations, teams, and projects. Each member has their own encryption keys with fine-grained access control.',
    icon: Users,
  },
  {
    title: 'Key Rotation',
    description: 'Rotate encryption keys with double-admin approval for organizations. Security without single points of failure.',
    icon: RefreshCw,
  },
  {
    title: 'Google Secret Manager',
    description: 'Optional integration with Google Cloud. Pull secrets directly - your GCP credentials never touch our servers.',
    icon: Cloud,
  },
]
</script>

<template>
  <section id="features" class="py-24 bg-secondary/20">
    <div class="container mx-auto px-4">
      <div class="text-center mb-16">
        <p class="text-primary font-medium mb-3">Features</p>
        <h2 class="text-3xl md:text-5xl font-bold mb-4">Everything you need for secret management</h2>
        <p class="text-muted-foreground text-lg max-w-2xl mx-auto">
          Built for teams who refuse to compromise on security.
        </p>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        <Card v-for="feature in features" :key="feature.title" class="border-border/50 bg-background/80 backdrop-blur-sm hover:border-primary/30 hover:bg-background transition-all group">
          <CardHeader>
            <div class="size-11 rounded-lg bg-primary/10 flex items-center justify-center mb-4 text-primary group-hover:bg-primary/20 transition-colors">
              <component :is="feature.icon" class="size-5" />
            </div>
            <CardTitle class="text-lg mb-2">{{ feature.title }}</CardTitle>
            <CardDescription class="text-sm leading-relaxed">{{ feature.description }}</CardDescription>
          </CardHeader>
        </Card>
      </div>

      <div class="text-center mt-12">
        <Button variant="outline" size="lg" as="a" href="/features">
          Learn more about features
          <ArrowRight class="ml-2 size-4" />
        </Button>
      </div>
    </div>
  </section>
</template>
