---
name: swiftui
description: SwiftUI moderno con Swift 6, MVVM, SwiftData, y async/await
category: mobile
stack: [swift, swiftui, ios, xcode, swiftdata, combine]
triggers: [swift, swiftui, ios, xcode, view, observable, combine, swiftdata, uikit]
---

# SwiftUI Skill

## Agent Attitude
Desarrollador iOS con Swift 6 + SwiftUI. `@Observable` sobre `@Published`. SwiftData para persistencia. `async/await` sobre completion handlers. NUNCA UIKit en features nuevos.

## Rules
- Swift 6 con `@Observable` macro. NO `@Published` ni `ObservableObject`.
- SwiftData para persistencia local. NO Core Data directo.
- `async/await` para networking. NUNCA `completionHandler`.
- `NavigationStack` con `NavigationDestination`. NO `NavigationView` (deprecated).
- PreviewProvider para cada View. Desarrollo iterativo rápido.
- SF Symbols para iconografía. NO PNGs para iconos de sistema.

## Do's
- `View` compuestas y pequeñas (max 150 líneas).
- `@Environment` para dependencias inyectables.
- `@AppStorage` para preferencias simples.
- `.task { }` modifier para async on-appear.
- `UIViewRepresentable` solo para UIKit legacy (y documentar por qué).

## Don'ts
- NO UIKit en features nuevos sin justificación documentada.
- NO `DispatchQueue.main.async` — usa `@MainActor`.
- NO `print()` en producción. Usa `os_log` o `Logger`.
- NO force unwrap (`!`). Usa `guard let` o `if let`.
- NO `Color.white` hardcodeado. Usa `.primary`, `.secondary` del sistema.

## Recommended Commands
- `xcodebuild -scheme MyApp -destination 'platform=iOS Simulator' test` — Tests
- `swiftlint --strict` — Lint
- `swiftformat .` — Formateo
