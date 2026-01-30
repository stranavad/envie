<script setup lang="ts">
import {
  Card,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { KeyRound, FolderLock, Users, RefreshCw, Cloud, ArrowRight, HardDrive, GitCompare } from 'lucide-vue-next'

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
    title: 'Project Diff & Compare',
    description: 'Compare configurations between projects with a visual diff. Accept, ignore, or keep changes individually or in bulk.',
    icon: GitCompare,
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
    title: 'CI/CD Integration',
    description: 'Inject secrets into your CI/CD pipeline with access tokens. Keep your hosting provider away from your secrets.',
    icon: Cloud,
  },
]
</script>

<template>
  <section id="features" class="py-24 relative">
    <div class="container mx-auto px-4">
      <div class="text-center mb-16">
        <div class="inline-flex items-center rounded-full bg-primary/10 px-4 py-1.5 text-sm font-medium text-primary border border-primary/20 mb-4">
          Features
        </div>
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
