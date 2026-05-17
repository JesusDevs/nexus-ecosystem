---
name: svelte
description: Svelte 5 con runes, SvelteKit, y Tailwind
category: web
stack: [svelte, sveltekit, typescript, tailwind, vite]
triggers: [svelte, sveltekit, rune, $state, $derived, $effect, .svelte]
---

# Svelte Skill

## Agent Attitude
Eres un desarrollador Svelte 5 con runes. `$state` sobre `let`.
SvelteKit para routing y data loading. NUNCA Svelte 4 `$:` reactivity.

## Rules
- Svelte 5 runes: `$state()`, `$derived()`, `$effect()`. NO `$:`.
- `$props()` para props tipadas. NO `export let`.
- `onclick` no `on:click`. Nuevo sistema de eventos.
- SvelteKit: `+page.svelte`, `+layout.svelte`, `+server.ts`.
- `load` functions para data fetching. NO fetch en componentes.
- `enhance` para forms progresivos.

## Do's
- Stores de Svelte para estado global.
- `$effect` para side effects con auto-limpieza.
- `{#snippet}` para templates reusables.
- Tailwind para estilos. NO `<style>` global.
- `+page.server.ts` para data sensible.

## Don'ts
- NO `$:` reactive statements (Svelte 4 legacy).
- NO `export let` para props.
- NO `on:click` — usa `onclick`.
- NO `store.subscribe()` manual — usa `$store`.
- NO `<svelte:head>` para SEO (usa SvelteKit SEO hooks).

## Recommended Commands
- `npx svelte-check --tsconfig ./tsconfig.json` — Type-check
- `npx vite build` — Build
- `npx vitest --coverage` — Tests
