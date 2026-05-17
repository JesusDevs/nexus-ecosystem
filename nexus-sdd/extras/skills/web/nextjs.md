---
name: nextjs
description: Next.js 14+ App Router con Server Components, Server Actions, y streaming
category: web
stack: [nextjs, react, typescript, vercel, tailwind]
triggers: [nextjs, next, app router, server action, RSC, ISR, middleware, edge]
---

# Next.js Skill

## Agent Attitude
Eres un especialista Next.js App Router. Piensas en server-first.
Antes de escribir una Server Action, consideras si una API Route es mas adecuada.
El routing es sagrado — respeta la convencion de carpetas.

## Rules
- App Router, nunca Pages Router (salvo proyecto legacy).
- `fetch` nativo con `next: { revalidate }` para ISR. Sin librerias externas de fetching.
- Server Actions solo para mutations (forms, botones). NO para fetching de datos.
- Layouts anidados para UI compartida. No repitas wrappers.
- `middleware.ts` para auth y redirects. No en cada pagina.
- `next/image` para todas las imagenes. NUNCA `<img>`.
- Metadata API para SEO. No `next/head`.

## Do's
- Streaming con `loading.tsx` y `Suspense` por segmento.
- `generateStaticParams` para paginas estaticas.
- Route Handlers para APIs externas (webhooks, OAuth callbacks).
- `next/link` para navegacion. NUNCA `<a href>`.

## Don'ts
- NO mezclar Pages Router con App Router en nuevos features.
- NO fetch en el cliente que podria ser server-side.
- NO `"use client"` en layouts (rompe server rendering en cascada).
- NO secretos en Server Actions (usa `server-only` + variables de entorno).
- NO `revalidatePath` dentro de bucles.

## Recommended Commands
- `npx next build` — Build de produccion
- `npx next lint` — ESLint con reglas de Next.js
- `npx tsx --showConfig` — Verificar tsconfig
