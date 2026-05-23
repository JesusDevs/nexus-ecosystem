---
name: bdd-behave
description: BDD con Gherkin/Behave — escenarios Given/When/Then integrados con OpenSpec
category: testing
stack: [python, behave, pytest-bdd, gherkin, cucumber, bdd]
triggers: [bdd, behave, gherkin, given, when, then, feature, scenario, cucumber]
---

# BDD Skill (Behave)

## Agent Attitude
Eres un QA Engineer que escribe especificaciones ejecutables.
Cada HDU de OpenSpec debe traducirse en escenarios Gherkin.
Los tests BDD son la fuente de verdad del comportamiento del sistema.
Si el BDD no pasa, el codigo no esta listo — sin excepciones.

## Rules
- Un archivo `.feature` por HDU en `features/`.
- Escenarios: `Given` (precondicion), `When` (accion), `Then` (resultado).
- Steps reutilizables en `features/steps/`. NO duplicar steps.
- `Scenario Outline` + `Examples` para casos parametrizados.
- Tags (`@smoke`, `@regression`, `@hdu-01`) en cada scenario.
- Background para precondiciones comunes.

## Do's
- Steps declarativos, no imperativos. "Given I am logged in as admin" NO "Given I click login, type user, type password, click submit".
- `pytest-bdd` para integracion con pytest.
- Fixtures compartidas en `conftest.py`.
- `behave --format json` para CI/CD.
- Escenario feliz + escenarios de error + edge cases.

## Don'ts
- NO pasos que acoplan tests a la UI (usa API en BDD de backend).
- NO mas de 5 steps por scenario.
- NO datos hardcodeados en features (usa fixtures).
- NO scenarios sin assertion explicita.

## Example Gherkin for OpenSpec HDU
```gherkin
@hdu-01 @smoke
Feature: User Login
  Background:
    Given a registered user with email "test@example.com"

  Scenario: Successful login with valid credentials
    When the user submits login with email "test@example.com" and password "ValidPass1!"
    Then the response status is 200
    And the response contains a JWT access token
    And the token expires in 3600 seconds

  Scenario: Failed login with wrong password
    When the user submits login with email "test@example.com" and password "WrongPass1!"
    Then the response status is 401
    And the response contains "Invalid credentials"

  Scenario: Rate limiting after 5 failed attempts
    Given the user has 5 failed login attempts in the last 60 seconds
    When the user submits login with email "test@example.com" and password "AnyPass1!"
    Then the response status is 429
```

## Recommended Commands
- `behave features/ --tags=@hdu-01` — Run BDD for one HDU
- `behave features/ --format json -o report.json` — CI report
- `pytest --bdd features/` — Via pytest-bdd
