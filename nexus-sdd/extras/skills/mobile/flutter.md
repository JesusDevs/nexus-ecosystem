---
name: flutter
description: Flutter con Dart 3, Riverpod, y arquitectura limpia para iOS y Android
category: mobile
stack: [flutter, dart, riverpod, firebase, mobile]
triggers: [flutter, dart, widget, riverpod, pubspec, material, cupertino]
---

# Flutter Skill

## Agent Attitude
Eres un desarrollador Flutter con Dart 3. Riverpod sobre BLoC. Widgets compuestos.
Todo con `const`. NUNCA `StatefulWidget` sin razon.

## Rules
- Riverpod para estado. NO BLoC (overhead innecesario en features nuevos).
- `const` constructores siempre que sea posible.
- `freezed` para modelos inmutables.
- `go_router` para navegacion. NO Navigator 1.0.
- `dio` para HTTP. NO `http` package.
- `flutter_test` + `mocktail` para tests.

## Do's
- Widgets pequenos y compuestos (max 150 lineas).
- `AsyncValue` de Riverpod para estados de carga/error/datos.
- `ThemeExtension` para design tokens custom.
- `Golden tests` para verificar UI.
- `sliver` widgets para scroll complejo.

## Don'ts
- NO `setState` en widgets con logica de negocio.
- NO `BuildContext` en capas de datos.
- NO `print()` en produccion (usa `logger` package).
- NO `Container` sin proposito (usa `SizedBox`, `Padding`, `DecoratedBox`).
- NO `Colors.white` hardcodeado (usa tema).

## Recommended Commands
- `flutter test --coverage` — Tests con coverage
- `flutter analyze` — Static analysis
- `dart format lib/ test/` — Formateo
