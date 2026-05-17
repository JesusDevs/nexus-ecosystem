---
name: vue
description: Vue 3 Composition API con TypeScript, Pinia, y Vue Router
category: web
stack: [vue, typescript, vite, pinia, tailwind, nuxt]
triggers: [vue, composition api, ref, reactive, pinia, vue-router, SFC]
---

# Vue Skill

## Agent Attitude
Eres un desarrollador Vue 3 con Composition API. NUNCA uses Options API en nuevo codigo.
`<script setup>` es tu default. TypeScript siempre. Pinia para estado global.

## Rules
- `<script setup lang="ts">` siempre. Sin excepciones.
- Pinia para estado global. NO Vuex (deprecated).
- `ref` para primitivos, `reactive` para objetos.
- `defineProps<T>()` con tipos. NUNCA props sin tipo.
- `defineEmits<T>()` tipado.
- Composables para logica reutilizable. Archivos `use*.ts`.

## Do's
- `watchEffect` para side effects auto-limpieza.
- `Suspense` + `defineAsyncComponent` para lazy loading.
- `Teleport` para modales y overlays.
- Tailwind o CSS Modules para estilos. NO estilos globales.

## Don'ts
- NO Options API (`data`, `methods`, `computed` objeto).
- NO mutar props directamente.
- NO `v-if` con `v-for` en el mismo elemento.
- NO `this` en Composition API.
- NO `$ref`魔法 (magia de reactividad). Usa `ref()` de Vue.

## Recommended Commands
- `npx vue-tsc --noEmit` — Type-check
- `npx vitest --coverage` — Tests
- `npx eslint . --ext .vue,.ts` — Lint
