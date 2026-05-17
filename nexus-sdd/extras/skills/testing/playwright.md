---
name: playwright
description: Playwright para E2E testing cross-browser + playwright-cli agent skills + mobile emulation
category: testing
stack: [playwright, typescript, javascript, e2e, browser, chromium, firefox, webkit, mobile]
triggers: [playwright, e2e, browser, page, locator, chromium, firefox, webkit, playwright-cli]
---

# Playwright Skill

## Agent Attitude
Eres un QA Automation Engineer especializado en E2E.
Tests deterministas. NUNCA `waitForTimeout` sin razon.
Locators basados en rol/texto, no en CSS fragil.
Cada flujo critico del spec tiene su test E2E.

## Playwright Agent CLI (playwright-cli)

Playwright ofrece su propio CLI para agentes con 9 skills oficiales:

```bash
# Instalar el CLI y sus skills
npm install -g playwright-cli
playwright-cli install --skills

# Las 9 skills que se instalan:
```

| # | Skill | Descripcion |
|---|-------|-------------|
| 1 | Running and Debugging | Ejecutar, debugear y gestionar suites de test |
| 2 | Request mocking | Interceptar y mockear peticiones de red |
| 3 | Running Playwright code | Ejecutar scripts Playwright arbitrarios |
| 4 | Browser session management | Gestionar multiples sesiones de navegador |
| 5 | Storage state | Persistir y restaurar cookies/localStorage |
| 6 | Test generation | Generar tests desde interacciones grabadas |
| 7 | Tracing | Grabar e inspeccionar trazas de ejecucion |
| 8 | Video recording | Capturar videos de sesiones de navegador |
| 9 | Inspecting element attributes | Atributos de elementos no visibles en snapshots |

### Integracion con Nexus-SDD

```bash
# Nexus-SDD instala playwright-cli + sus skills automaticamente
nexus-sdd skill install playwright   # instala esta skill + playwright-cli --skills

# Usar playwright-cli dentro de los agentes
playwright-cli --help                # descubrir comandos sin skills (modo discovery)
```

### Mobile Emulation con Playwright

```typescript
// Emular dispositivos Android/iOS en Playwright
import { devices } from '@playwright/test';

// Android
test.use({ ...devices['Pixel 7'] });
test.use({ ...devices['Galaxy Tab S4'] });

// iOS
test.use({ ...devices['iPhone 15'] });
test.use({ ...devices['iPad Pro'] });

// O configurar viewport + user agent manual
test.use({
  viewport: { width: 412, height: 915 },
  userAgent: 'Mozilla/5.0 ... Android ...',
  deviceScaleFactor: 2.625,
  isMobile: true,
  hasTouch: true,
});
```

### Conexion con ADB (Android)

```bash
# Playwright puede testear en dispositivos Android reales via ADB
adb devices                          # listar dispositivos
adb forward tcp:9222 tcp:9222       # forward del puerto de debug

# En Playwright, conectar al Chrome del dispositivo
const browser = await chromium.connectOverCDP('http://localhost:9222');
```

## Rules
- `page.getByRole()` y `page.getByLabel()` sobre selectores CSS.
- `page.waitForResponse()` para esperar APIs. NO `waitForTimeout`.
- Fixtures para datos de prueba (login, seeds).
- `test.describe` para agrupar flujos relacionados.
- `test.beforeEach` para estado inicial limpio.
- `webServer` en config para levantar el app automaticamente.
- `trace: 'on-first-retry'` para debugging.
- Preferir `playwright-cli` para tareas rapidas de navegador (no requiere escribir tests).

## Do's
- Tests independientes (cada test su propio contexto).
- `expect(locator).toBeVisible()` antes de interactuar.
- `page.route()` para mockear APIs externas en tests.
- `storageState` para reutilizar sesiones autenticadas.
- `test.step` para legibilidad en flujos largos.
- Screenshots en fallos (`screenshot: 'only-on-failure'`).
- `devices` para emular mobile en tests E2E.
- `playwright-cli` para exploracion rapida y debugging interactivo.

## Don'ts
- NO `waitForTimeout(5000)` — usa `waitForSelector` o `waitForResponse`.
- NO selectores CSS anidados complejos.
- NO tests que dependen de orden de ejecucion.
- NO `.click()` sin antes verificar que el elemento esta habilitado.
- NO tests de +30 lineas sin `test.step`.
- NO `test.only` commiteado.

## Example E2E for OpenSpec HDU
```typescript
import { test, expect } from '@playwright/test';

test.describe('HDU-01: User Login', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login');
  });

  test('successful login redirects to dashboard', async ({ page }) => {
    await page.getByLabel('Email').fill('test@example.com');
    await page.getByLabel('Password').fill('ValidPass1!');
    await page.getByRole('button', { name: 'Sign in' }).click();
    await expect(page).toHaveURL('/dashboard');
    await expect(page.getByText('Welcome back')).toBeVisible();
  });

  test('shows error on invalid credentials', async ({ page }) => {
    await page.getByLabel('Email').fill('test@example.com');
    await page.getByLabel('Password').fill('Wrong!');
    await page.getByRole('button', { name: 'Sign in' }).click();
    await expect(page.getByText('Invalid credentials')).toBeVisible();
  });
});
```

## Recommended Commands
- `npx playwright test` — Run all E2E
- `npx playwright test --ui` — Interactive mode
- `npx playwright show-trace test-results/.../trace.zip` — Debug
- `npx playwright codegen localhost:3000` — Record tests
