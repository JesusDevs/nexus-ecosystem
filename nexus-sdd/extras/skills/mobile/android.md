---
name: android
description: Android development with official CLI, Jetpack Compose, AGP 9, and Google skills
category: mobile
stack: [android, kotlin, java, jetpack-compose, gradle, agp]
triggers: [android, kotlin, java, compose, gradle, agp, apk, emulator, sdk, jetpack]
source: https://github.com/android/skills
---

# Android Skill

## Agent Attitude
Eres un desarrollador Android con Kotlin + Jetpack Compose. Siempre usas la CLI oficial (`android`) para tareas de configuración, build y emulación. Prefieres Compose sobre XML Views. Material 3 por defecto.

## Prerrequisitos

### Android CLI (obligatorio)
```bash
# Instalar
curl -fsSL https://developer.android.com/studio/cli -o android-cli.zip
unzip android-cli.zip && chmod +x android && sudo mv android /usr/local/bin/

# Verificar
android --version
android update

# Configurar para agentes
android init   # Instala la skill android-cli para Gemini/Claude/Codex
```

### Skills Oficiales de Google
```bash
# Listar skills disponibles
android skills list --long

# Instalar skills recomendadas para desarrollo general
android skills add --all

# Skills específicas según necesidad:
android skills add edge-to-edge          # Pantallas borde a borde
android skills add navigation-3          # Navegación con Navigation 3
android skills add migrate-xml-views-to-jetpack-compose  # Migración a Compose
android skills add agp-9-upgrade         # Actualización a AGP 9
android skills add camera1-to-camerax    # Migración Camera1 → CameraX
android skills add r8-analyzer           # Optimización R8
android skills add play-billing-library-version-upgrade  # Play Billing
android skills add perfetto-sql          # Profiling con Perfetto
```

## Reglas

### Proyecto
- `android create --name=<app> --output=<path>` para nuevos proyectos.
- `android describe` para metadata del proyecto actual.
- `android info` para verificar SDK path.

### Build y Deploy
- `android run --apks=<path>` para instalar en dispositivo/emulador.
- `android sdk list <pattern>` antes de instalar paquetes SDK.
- `android sdk install <package@version>` con versión específica.

### Emulador
- `android emulator list` para ver dispositivos disponibles.
- `android emulator create --profile=<profile>` para crear AVD.
- `android emulator start <name>` para iniciar.
- `android emulator stop <serial>` para detener.

### UI y Testing
- `android layout --output=ui.json` para árbol de layout.
- `android screen capture --annotate --output=screen.png` para screenshot anotada.
- `android screen resolve --screenshot=screen.png --string="input tap #5"` para coordenadas de UI.

### Documentación
- `android docs search "<query>"` para buscar en la knowledge base de Android.
- `android docs fetch kb://android/topic/...` para leer documentación oficial.

## Do's
- Jetpack Compose + Material 3 para UI nueva. NO XML Views.
- Kotlin 2.0+ con coroutines. NO AsyncTask ni callbacks.
- Navigation 3 para navegación type-safe.
- AGP 9+ y Gradle Kotlin DSL (`build.gradle.kts`).
- `android skills add` para mantener skills actualizadas (Google las actualiza).
- Edge-to-edge en todas las apps nuevas.
- Perfetto para profiling (reemplaza a systrace).

## Don'ts
- NO XML Views en features nuevos (usar Compose).
- NO `android:allowBackup="true"` sin saber qué estás exponiendo.
- NO hardcodear API keys en `AndroidManifest.xml` ni `build.gradle`.
- NO `AsyncTask`, `AsyncTaskLoader`, `LoaderManager` (deprecated).
- NO `support.v4` ni `support.v7` (usar AndroidX).
- NO `android run` sin antes verificar el device target si hay múltiples.

## Skills Oficiales (github.com/android/skills)

| Skill | Cuándo usar |
|-------|-------------|
| `android-cli` | Siempre. Base para cualquier proyecto. |
| `edge-to-edge` | Apps nuevas. Obligatorio para Android 15+. |
| `navigation-3` | Navegación type-safe entre pantallas. |
| `migrate-xml-views-to-jetpack-compose` | Proyecto legacy con XML Views. |
| `agp-9-upgrade` | Actualizar Android Gradle Plugin. |
| `camera1-to-camerax` | Migrar cámara antigua a CameraX. |
| `r8-analyzer` | Optimizar ofuscación y shrinking. |
| `play-billing-library-version-upgrade` | Actualizar billing de Google Play. |
| `perfetto-sql` | Profiling avanzado con Perfetto. |
| `display-ai-glasses-with-jetpack-compose-glimmer` | XR / AI glasses. |

## ADB (Android Debug Bridge)

El agente debe usar ADB para interactuar con dispositivos y emuladores:

```bash
# ── Dispositivos ──────────────────────────────
adb devices -l               # Listar dispositivos con detalles
adb wait-for-device           # Esperar a que el dispositivo este listo
adb reboot                    # Reiniciar dispositivo
adb reboot bootloader         # Entrar en fastboot
adb reboot recovery           # Entrar en recovery

# ── Instalacion y apps ────────────────────────
adb install app.apk           # Instalar APK
adb install -r app.apk        # Reinstalar manteniendo datos
adb uninstall <package>       # Desinstalar app
adb shell pm list packages    # Listar todos los paquetes
adb shell am start -n <package>/<activity>  # Lanzar actividad

# ── Logs y debug ──────────────────────────────
adb logcat                    # Logs en tiempo real
adb logcat -s TAG              # Filtrar por tag
adb logcat -v time             # Con timestamps
adb shell dumpsys activity    # Estado de actividades
adb shell dumpsys meminfo <package>  # Memoria de la app

# ── Archivos ──────────────────────────────────
adb push local.txt /sdcard/   # Subir archivo
adb pull /sdcard/remote.txt . # Bajar archivo
adb shell ls /sdcard/         # Explorar filesystem

# ── Input y UI ────────────────────────────────
adb shell input tap 500 1000  # Tap en coordenadas
adb shell input swipe 500 1000 500 500  # Swipe/drag
adb shell input text "hola"   # Escribir texto
adb shell input keyevent 4    # Keycode (4 = BACK, 3 = HOME)
adb shell screencap /sdcard/screen.png  # Screenshot

# ── Red y puertos ─────────────────────────────
adb forward tcp:9222 tcp:9222       # Forward para Chrome DevTools
adb reverse tcp:3000 tcp:3000       # Reverse (dispositivo → host)
adb shell netstat                   # Conexiones activas
```

## Emulador (Avanzado)

```bash
# ── Creacion y configuracion ──────────────────
avdmanager list device               # Dispositivos disponibles
avdmanager list target               # System images
sdkmanager "system-images;android-34;google_apis;x86_64"  # Descargar imagen
avdmanager create avd -n pixel_7 -k "system-images;android-34;google_apis;x86_64" -d pixel_7

# ── Inicio y control ──────────────────────────
emulator -avd pixel_7 -no-boot-anim -gpu host &  # Iniciar con GPU
emulator -avd pixel_7 -writable-system            # System writable
emulator -avd pixel_7 -http-proxy http://proxy:8080  # Con proxy

# ── Snapshot ──────────────────────────────────
emulator -avd pixel_7 -snapshot snap01   # Restaurar snapshot
emulator -avd pixel_7 -snapshot-save snap01  # Guardar snapshot

# ── Conexion multiple ─────────────────────────
adb -s emulator-5554 shell            # Dirigirse a un emulador especifico
adb -s emulator-5556 install app.apk  # Instalar en otro emulador
```

## Integracion con Playwright (Mobile Testing)

```bash
# Playwright puede testear webviews y PWAs en emuladores Android
# 1. Iniciar emulador
emulator -avd pixel_7 &

# 2. Forward del puerto de debug de Chrome
adb forward tcp:9222 tcp:9222

# 3. Conectar Playwright al Chrome del emulador via CDP
```

```typescript
// playwright.config.ts — emular dispositivos Android/iOS
import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
  projects: [
    { name: 'Pixel 7', use: { ...devices['Pixel 7'] } },
    { name: 'iPhone 15', use: { ...devices['iPhone 15'] } },
    {
      name: 'android-chrome',
      use: {
        ...devices['Pixel 7'],
        browserName: 'chromium',
        // Conectar a Chrome en emulador real via CDP
        browserWSEndpoint: 'http://localhost:9222',
      },
    },
  ],
});
```

```bash
# Flujo completo: emulador → ADB forward → Playwright test
nexus-sdd test --hdu-id HDU-01 --mobile android  # Corre tests E2E en emulador
```

## Seguridad

Las skills oficiales de Google (`github.com/android/skills`) son seguras por defecto.
Nexus-SDD las valida por hash SHA antes de instalarlas automaticamente.

```bash
# Nexus-SDD instala skills oficiales con verificación
nexus-sdd skill install @android/edge-to-edge   # Verifica firma de Google
nexus-sdd skill install android                 # Nuestra skill wrapper
```

## Comandos Recomendados
- `android create list` — Ver plantillas disponibles
- `android describe` — Metadata del proyecto
- `android sdk list platforms` — Plataformas instaladas
- `android sdk update` — Actualizar todo el SDK
- `android emulator list` — Dispositivos disponibles
- `android layout --pretty` — Árbol de UI legible
- `android run --apks=app/build/outputs/apk/debug/app-debug.apk` — Deploy
- `android skills list --long` — Skills oficiales disponibles
- `adb devices -l` — Dispositivos conectados
- `adb logcat -s <TAG>` — Logs filtrados
- `adb shell screencap /sdcard/screen.png` — Screenshot via ADB
- `adb forward tcp:9222 tcp:9222` — Forward para Playwright/Chrome DevTools
- `emulator -avd <name> -no-boot-anim -gpu host &` — Iniciar emulador rapido
