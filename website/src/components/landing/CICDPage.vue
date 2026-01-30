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
  Terminal,
  Key,
  Shield,
  Lock,
  Server,
  Github,
  CheckCircle2,
  Copy,
  ArrowRight,
  Container,
  Cloud,
  Clock,
  Eye,
  EyeOff,
  Fingerprint,
  Settings,
  Layers
} from 'lucide-vue-next'

const githubUrl = 'https://github.com/stranavad/envie'

const benefits = [
  {
    title: 'Zero Exposure to Hosting Providers',
    description: 'Your secrets are fetched at runtime and never stored in your CI/CD platform. Vercel, AWS, or any provider never sees your actual values.',
    icon: EyeOff,
  },
  {
    title: 'Instant Secret Rotation',
    description: 'Rotate secrets in Envie and all deployments automatically use the new values on next run. No manual updates across platforms.',
    icon: Clock,
  },
  {
    title: 'Audit Trail',
    description: 'Track which tokens accessed which secrets and when. Full visibility into your CI/CD secret usage.',
    icon: Eye,
  },
  {
    title: 'Scoped Access',
    description: 'Create tokens with read-only access to specific projects. Each pipeline only sees what it needs.',
    icon: Shield,
  },
]

const useCases = [
  {
    title: 'Docker Builds',
    description: 'Inject secrets during container builds without baking them into images.',
  },
  {
    title: 'GitHub Actions',
    description: 'Fetch secrets at workflow runtime instead of storing in repository secrets.',
  },
  {
    title: 'Kubernetes Deployments',
    description: 'Generate ConfigMaps and Secrets on the fly during deployment.',
  },
  {
    title: 'Local Development',
    description: 'New team members can bootstrap their environment in seconds.',
  },
]
</script>

<template>
  <div class="py-16 md:py-24">
    <div class="container mx-auto px-4">
      <!-- Header -->
      <div class="text-center mb-16 max-w-3xl mx-auto">
        <div class="inline-flex items-center rounded-full bg-cyan-500/10 px-4 py-1.5 text-sm font-medium text-cyan-400 border border-cyan-500/20 mb-6">
          <Terminal class="mr-2 size-4" />
          CI/CD Integration
        </div>
        <h1 class="text-4xl md:text-6xl font-bold mb-6 tracking-tight">
          Secrets for your pipelines
        </h1>
        <p class="text-xl text-muted-foreground leading-relaxed">
          Inject encrypted secrets into your CI/CD pipelines with access tokens.
          Your hosting provider never sees your actual secret values.
        </p>
      </div>

      <!-- How Access Tokens Work -->
      <div class="mb-24">
        <h2 class="text-2xl md:text-3xl font-bold text-center mb-4">How access tokens work</h2>
        <p class="text-muted-foreground text-center mb-12 max-w-2xl mx-auto">
          Access tokens provide secure, scoped access to your project's secrets without exposing your encryption keys.
        </p>

        <div class="max-w-4xl mx-auto">
          <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-8 lg:gap-4">
            <!-- Step 1 -->
            <div class="flex flex-col items-center text-center relative">
              <div class="size-14 rounded-xl bg-cyan-500/10 border border-cyan-500/30 flex items-center justify-center text-cyan-500 mb-4">
                <Key class="size-7" />
              </div>
              <h3 class="font-semibold mb-2">1. Create Token</h3>
              <p class="text-sm text-muted-foreground">
                Generate an access token in Envie with an optional expiration date
              </p>
              <!-- Arrow (desktop only) -->
              <div class="hidden lg:block absolute -right-2 top-7">
                <ArrowRight class="size-5 text-muted-foreground" />
              </div>
            </div>

            <!-- Step 2 -->
            <div class="flex flex-col items-center text-center relative">
              <div class="size-14 rounded-xl bg-green-500/10 border border-green-500/30 flex items-center justify-center text-green-500 mb-4">
                <Lock class="size-7" />
              </div>
              <h3 class="font-semibold mb-2">2. Token Contains Key</h3>
              <p class="text-sm text-muted-foreground">
                The token embeds an encrypted project key that only the CLI can decrypt
              </p>
              <!-- Arrow (desktop only) -->
              <div class="hidden lg:block absolute -right-2 top-7">
                <ArrowRight class="size-5 text-muted-foreground" />
              </div>
            </div>

            <!-- Step 3 -->
            <div class="flex flex-col items-center text-center relative">
              <div class="size-14 rounded-xl bg-blue-500/10 border border-blue-500/30 flex items-center justify-center text-blue-500 mb-4">
                <Terminal class="size-7" />
              </div>
              <h3 class="font-semibold mb-2">3. CLI Fetches Config</h3>
              <p class="text-sm text-muted-foreground">
                The CLI uses the token to fetch encrypted config from Envie servers
              </p>
              <!-- Arrow (desktop only) -->
              <div class="hidden lg:block absolute -right-2 top-7">
                <ArrowRight class="size-5 text-muted-foreground" />
              </div>
            </div>

            <!-- Step 4 -->
            <div class="flex flex-col items-center text-center">
              <div class="size-14 rounded-xl bg-primary/10 border border-primary/30 flex items-center justify-center text-primary mb-4">
                <CheckCircle2 class="size-7" />
              </div>
              <h3 class="font-semibold mb-2">4. Decrypts Locally</h3>
              <p class="text-sm text-muted-foreground">
                Secrets are decrypted in your pipeline and exported as environment variables
              </p>
            </div>
          </div>
        </div>
      </div>

      <!-- Security Benefits -->
      <div class="grid md:grid-cols-2 gap-6 mb-24">
        <Card v-for="benefit in benefits" :key="benefit.title"
          class="border-border/50 bg-secondary/20 backdrop-blur-sm"
        >
          <CardHeader>
            <div class="size-12 rounded-lg bg-cyan-500/10 flex items-center justify-center mb-4 text-cyan-500">
              <component :is="benefit.icon" class="size-6" />
            </div>
            <CardTitle class="text-xl mb-2">{{ benefit.title }}</CardTitle>
            <CardDescription class="text-base leading-relaxed">{{ benefit.description }}</CardDescription>
          </CardHeader>
        </Card>
      </div>

      <!-- Docker Example -->
      <div class="mb-24">
        <h2 class="text-2xl md:text-3xl font-bold text-center mb-4">Using with Docker</h2>
        <p class="text-muted-foreground text-center mb-12 max-w-2xl mx-auto">
          Inject secrets at build time or runtime without storing them in your image layers.
        </p>

        <div class="max-w-4xl mx-auto grid lg:grid-cols-2 gap-8">
          <!-- Dockerfile Example -->
          <div>
            <div class="flex items-center gap-2 mb-4">
              <Container class="size-5 text-blue-500" />
              <h3 class="font-semibold">Multi-stage Dockerfile</h3>
            </div>
            <div class="bg-card border border-border/50 rounded-xl overflow-hidden">
              <div class="bg-secondary/50 border-b border-border/50 px-4 py-2 flex items-center justify-between">
                <span class="text-xs text-muted-foreground font-mono">Dockerfile</span>
                <button class="p-1 hover:bg-secondary rounded transition-colors text-muted-foreground">
                  <Copy class="size-3.5" />
                </button>
              </div>
              <pre class="p-4 text-sm font-mono overflow-x-auto"><code class="text-muted-foreground"><span class="text-blue-400">FROM</span> node:20-alpine <span class="text-blue-400">AS</span> builder

<span class="text-gray-500"># Install Envie CLI</span>
<span class="text-blue-400">RUN</span> npm install -g @envie/cli

<span class="text-gray-500"># Build argument for the token</span>
<span class="text-blue-400">ARG</span> ENVIE_TOKEN

<span class="text-gray-500"># Fetch and export secrets</span>
<span class="text-blue-400">RUN</span> envie export --token $ENVIE_TOKEN > .env

<span class="text-gray-500"># Your build steps here</span>
<span class="text-blue-400">RUN</span> npm ci && npm run build

<span class="text-gray-500"># Production image - no secrets!</span>
<span class="text-blue-400">FROM</span> node:20-alpine
<span class="text-blue-400">COPY</span> --from=builder /app/dist ./dist
<span class="text-blue-400">CMD</span> ["node", "dist/index.js"]</code></pre>
            </div>
          </div>

          <!-- Build Command -->
          <div>
            <div class="flex items-center gap-2 mb-4">
              <Terminal class="size-5 text-green-500" />
              <h3 class="font-semibold">Build Command</h3>
            </div>
            <div class="bg-card border border-border/50 rounded-xl overflow-hidden mb-6">
              <div class="bg-secondary/50 border-b border-border/50 px-4 py-2 flex items-center justify-between">
                <span class="text-xs text-muted-foreground font-mono">terminal</span>
                <button class="p-1 hover:bg-secondary rounded transition-colors text-muted-foreground">
                  <Copy class="size-3.5" />
                </button>
              </div>
              <pre class="p-4 text-sm font-mono overflow-x-auto"><code class="text-muted-foreground">docker build \
  --build-arg ENVIE_TOKEN=$ENVIE_TOKEN \
  -t my-app .</code></pre>
            </div>

            <div class="p-4 bg-green-500/5 border border-green-500/20 rounded-lg">
              <div class="flex items-start gap-3">
                <CheckCircle2 class="size-5 text-green-500 shrink-0 mt-0.5" />
                <div class="text-sm">
                  <p class="font-medium text-green-400 mb-1">Secrets never in image</p>
                  <p class="text-muted-foreground">
                    The .env file is only present during the build stage.
                    The final production image contains no secrets.
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- GitHub Actions Example -->
      <div class="mb-24">
        <h2 class="text-2xl md:text-3xl font-bold text-center mb-4">GitHub Actions Example</h2>
        <p class="text-muted-foreground text-center mb-12 max-w-2xl mx-auto">
          Fetch secrets at workflow runtime. Store only the access token in GitHub Secrets.
        </p>

        <div class="max-w-3xl mx-auto">
          <div class="bg-card border border-border/50 rounded-xl overflow-hidden">
            <div class="bg-secondary/50 border-b border-border/50 px-4 py-2 flex items-center justify-between">
              <div class="flex items-center gap-2">
                <Github class="size-4" />
                <span class="text-xs text-muted-foreground font-mono">.github/workflows/deploy.yml</span>
              </div>
              <button class="p-1 hover:bg-secondary rounded transition-colors text-muted-foreground">
                <Copy class="size-3.5" />
              </button>
            </div>
            <pre class="p-4 text-sm font-mono overflow-x-auto"><code class="text-muted-foreground"><span class="text-purple-400">name:</span> Deploy

<span class="text-purple-400">on:</span>
  <span class="text-purple-400">push:</span>
    <span class="text-purple-400">branches:</span> [main]

<span class="text-purple-400">jobs:</span>
  <span class="text-purple-400">deploy:</span>
    <span class="text-purple-400">runs-on:</span> ubuntu-latest
    <span class="text-purple-400">steps:</span>
      - <span class="text-purple-400">uses:</span> actions/checkout@v4

      - <span class="text-purple-400">name:</span> Install Envie CLI
        <span class="text-purple-400">run:</span> npm install -g @envie/cli

      - <span class="text-purple-400">name:</span> Load secrets
        <span class="text-purple-400">run:</span> |
          envie export --token $&#123;&#123; secrets.ENVIE_TOKEN &#125;&#125; > .env
          <span class="text-gray-500"># Or inject directly into environment</span>
          source &lt;(envie export --token $&#123;&#123; secrets.ENVIE_TOKEN &#125;&#125; --format shell)

      - <span class="text-purple-400">name:</span> Deploy
        <span class="text-purple-400">run:</span> npm run deploy</code></pre>
          </div>

          <div class="mt-6 grid sm:grid-cols-2 gap-4">
            <div class="p-4 bg-cyan-500/5 border border-cyan-500/20 rounded-lg">
              <div class="flex items-start gap-3">
                <Shield class="size-5 text-cyan-500 shrink-0 mt-0.5" />
                <div class="text-sm">
                  <p class="font-medium text-cyan-400 mb-1">Only one secret to manage</p>
                  <p class="text-muted-foreground">
                    Store only the ENVIE_TOKEN in GitHub. All other secrets are fetched at runtime.
                  </p>
                </div>
              </div>
            </div>
            <div class="p-4 bg-green-500/5 border border-green-500/20 rounded-lg">
              <div class="flex items-start gap-3">
                <CheckCircle2 class="size-5 text-green-500 shrink-0 mt-0.5" />
                <div class="text-sm">
                  <p class="font-medium text-green-400 mb-1">Always up to date</p>
                  <p class="text-muted-foreground">
                    Rotate secrets in Envie and deployments automatically use new values.
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Easy Deployment to New Environments -->
      <div class="mb-24">
        <div class="grid lg:grid-cols-2 gap-12 items-center">
          <div>
            <div class="inline-flex items-center rounded-full bg-green-500/10 px-4 py-1.5 text-sm font-medium text-green-400 border border-green-500/20 mb-4">
              <Layers class="mr-2 size-4" />
              Easy Onboarding
            </div>
            <h2 class="text-3xl font-bold mb-4">Deploy to new environments instantly</h2>
            <p class="text-lg text-muted-foreground mb-6 leading-relaxed">
              Spinning up a new staging environment or onboarding a new developer?
              One command pulls all the secrets they need. No more copying .env files
              or waiting for access to secret managers.
            </p>
            <ul class="space-y-3">
              <li class="flex items-center gap-3 text-muted-foreground">
                <div class="size-1.5 rounded-full bg-green-500"></div>
                New developer? Share a token, they're ready in seconds
              </li>
              <li class="flex items-center gap-3 text-muted-foreground">
                <div class="size-1.5 rounded-full bg-green-500"></div>
                New staging environment? Same config as production, one command
              </li>
              <li class="flex items-center gap-3 text-muted-foreground">
                <div class="size-1.5 rounded-full bg-green-500"></div>
                Disaster recovery? Rebuild with all secrets intact
              </li>
            </ul>
          </div>
          <div>
            <!-- Terminal Illustration -->
            <div class="relative">
              <div class="absolute inset-0 bg-gradient-to-br from-green-500/20 to-transparent blur-3xl rounded-full -z-10"></div>
              <div class="bg-card border border-border/50 rounded-xl shadow-2xl overflow-hidden">
                <div class="bg-secondary/50 border-b border-border/50 px-4 py-3 flex items-center gap-2">
                  <div class="flex gap-1.5">
                    <div class="size-3 rounded-full bg-red-500/80"></div>
                    <div class="size-3 rounded-full bg-yellow-500/80"></div>
                    <div class="size-3 rounded-full bg-green-500/80"></div>
                  </div>
                  <span class="text-xs text-muted-foreground font-mono ml-2">Terminal</span>
                </div>
                <div class="p-4 font-mono text-sm space-y-2">
                  <div class="flex items-center gap-2">
                    <span class="text-green-500">$</span>
                    <span class="text-muted-foreground">envie export --token $ENVIE_TOKEN > .env</span>
                  </div>
                  <div class="text-green-400 text-xs pl-4">
                    Fetching config from Envie...
                  </div>
                  <div class="text-green-400 text-xs pl-4">
                    Decrypting 12 variables...
                  </div>
                  <div class="text-green-400 text-xs pl-4 flex items-center gap-2">
                    <CheckCircle2 class="size-3.5" />
                    Written to .env
                  </div>
                  <div class="mt-4 flex items-center gap-2">
                    <span class="text-green-500">$</span>
                    <span class="text-muted-foreground">docker-compose up -d</span>
                  </div>
                  <div class="text-cyan-400 text-xs pl-4">
                    Starting services with secrets...
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Use Cases -->
      <div class="mb-20">
        <h2 class="text-2xl md:text-3xl font-bold text-center mb-12">Where to use it</h2>
        <div class="grid sm:grid-cols-2 lg:grid-cols-4 gap-4">
          <Card v-for="useCase in useCases" :key="useCase.title"
            class="border-border/50 bg-background/80 hover:border-cyan-500/30 transition-all"
          >
            <CardHeader class="pb-2">
              <CardTitle class="text-base">{{ useCase.title }}</CardTitle>
            </CardHeader>
            <CardContent>
              <p class="text-sm text-muted-foreground">{{ useCase.description }}</p>
            </CardContent>
          </Card>
        </div>
      </div>

      <!-- CTA -->
      <div class="text-center bg-secondary/30 rounded-2xl p-12 border border-border/50">
        <h2 class="text-2xl md:text-3xl font-bold mb-4">Ready to secure your CI/CD?</h2>
        <p class="text-muted-foreground mb-8 max-w-xl mx-auto">
          Get started with Envie CLI and stop exposing secrets to your hosting providers.
          Check out the documentation on GitHub.
        </p>
        <div class="flex flex-col sm:flex-row gap-4 justify-center">
          <Button size="lg" as="a" :href="githubUrl" target="_blank" rel="noopener noreferrer">
            <Github class="mr-2 size-5" />
            View CLI on GitHub
          </Button>
          <Button variant="outline" size="lg" as="a" href="/download">
            Download Desktop App
          </Button>
        </div>
      </div>
    </div>
  </div>
</template>
