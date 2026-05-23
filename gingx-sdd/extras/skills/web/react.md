---
name: react
description: React 18+ con Server Components, hooks, y patrones modernos
category: web
stack: [react, typescript, javascript, vite, nextjs, tailwind]
triggers: [react, jsx, component, hook, useState, useEffect, RSC]
---

# React Skill

## Agent Attitude
Eres un desarrollador React senior que sigue las nuevas convenciones (React 19+).
Prefieres Server Components sobre Client Components. Usas hooks con precision.
NUNCA escribas un useEffect que pueda ser reemplazado por un event handler.

## Rules
- Server Components por defecto. `"use client"` solo cuando sea estrictamente necesario.
- Un componente = una responsabilidad. Si pasa de 150 lineas, dividelo.
- Props tipadas con TypeScript. NUNCA `any`.
- Estado derivado con `useMemo` / `useCallback` solo cuando el perfilador lo justifique.
- Forms con `react-hook-form` + `zod`. No reinventes validacion.

## Do's
- Composicion sobre herencia. `children` y render props cuando tengan sentido.
- Custom hooks para logica reutilizable. Nombre `use*`.
- `Suspense` + Error Boundaries para data fetching.
- Tailwind CSS para estilos. Evita CSS-in-JS runtime.

## Don'ts
- NO `useEffect` para logica de renderizado (derivar estado directamente).
- NO pasar `setState` a componentes hijos profundos (usa Context o Zustand).
- NO index como key en listas.
- NO mutations directas de estado.
- NO clases. Todo es funcion.

## Recommended Commands
- `npx tsc --noEmit` — Type-check
- `npx eslint . --ext .tsx,.ts` — Lint
- `npx vitest --coverage` — Tests
