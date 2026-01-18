# Envie Website

Marketing website for Envie built with Astro and Vue.

## Tech Stack

- **[Astro](https://astro.build/)** - Static site generator with island architecture
- **[Vue 3](https://vuejs.org/)** - UI components (via `@astrojs/vue`)
- **[Tailwind CSS v4](https://tailwindcss.com/)** - Styling
- **[shadcn/ui](https://ui.shadcn.com/)** - UI component library (via reka-ui)
- **[Lucide](https://lucide.dev/)** - Icons

## Project Structure

```
website/
├── public/              # Static assets
│   └── favicon.svg
├── src/
│   ├── components/
│   │   ├── landing/     # Page-specific components
│   │   │   ├── Navigation.vue
│   │   │   ├── Hero.vue
│   │   │   ├── Features.vue
│   │   │   ├── Security.vue
│   │   │   ├── Footer.vue
│   │   │   ├── FeaturesPage.vue
│   │   │   └── SecurityPage.vue
│   │   └── ui/          # Reusable UI components (shadcn)
│   ├── layouts/
│   │   └── Layout.astro
│   ├── pages/
│   │   ├── index.astro      # Home page
│   │   ├── features.astro   # Features page
│   │   └── security.astro   # Security page
│   ├── styles/
│   │   └── global.css   # Tailwind + CSS variables
│   └── lib/
│       └── utils.ts     # Utility functions (cn)
├── astro.config.mjs
├── package.json
└── tsconfig.json
```

## Development

```bash
# Install dependencies
npm install

# Start dev server
npm run dev

# Build for production
npm run build

# Preview production build
npm run preview
```

## Pages

- `/` - Home page with hero, features overview, and security section
- `/features` - Detailed features with HTML illustrations
- `/security` - Deep dive into encryption and security architecture

## Styling

Uses Tailwind CSS v4 with shadcn/ui theming via CSS variables. The site is dark-mode only (`class="dark"` on `<html>`).

Theme colors are defined in `src/styles/global.css` using oklch color space.
