---
id: EPIC-09
title: Frontend Shell & Design System
module: [web]
priority: P0
status: todo
depends_on: [EPIC-00]
covers_requirements: [NFR-MAINT-01]
related_adrs: [ADR-009, ADR-007]
scope:
  paths:
    - apps/web/**
---

# EPIC-09: Frontend Shell & Design System

## Цель

Каркас Next.js 15 App Router приложения: проектная структура, дизайн-система (Tailwind токены), авторизационный слой (middleware, работающий с cookie из EPIC-01), генерация типов из OpenAPI, базовые layout-компоненты. Это фундамент для EPIC-10/11 — приоритетная область портфолио (frontend SSR), стоит уделить архитектуре здесь особое внимание.

## Задачи

1. Инициализировать Next.js 15 проект (`create-next-app` с TypeScript, App Router, Tailwind, ESLint) в `apps/web`.
2. Структура: `app/` (роуты), `components/ui/` (примитивы дизайн-системы — Button, Card, Badge, ProgressBar и т.д.), `components/features/{feature}/` (композитные компоненты, специфичные для домена — catalog, cart, order), `lib/api/` (сгенерированные типы + fetch-обёртки), `lib/auth/` (server-side auth helpers), `hooks/`.
3. `lib/api/client.ts`: типобезопасный fetch-клиент, читающий `openapi-typescript`-сгенерированные типы из `docs/api/openapi.yaml` (пайплайн генерации — `pnpm generate:api-types`, выходной файл `lib/api/generated-types.ts`, коммитится в репозиторий для простоты CI, регенерируется при изменении контракта).
4. `middleware.ts`: проверка JWT access-cookie (без похода в сеть — self-contained проверка подписи, ADR-007), редирект неавторизованных с защищённых роутов (`/cart`, `/checkout`, `/orders`) на `/login`.
5. Дизайн-система: определить токены (цвета, типографика, spacing) в `tailwind.config.ts`, ориентируясь на "лёгкий, игровой, дофаминовый" визуальный тон (яркие акценты, но не перегруженно — референс на современные food-delivery приложения). Загрузить `load_skill(name="design-foundations")` при реализации для консистентных решений по цвету/типографике, если потребуется отдельная агентная сессия.
6. Базовые layout-компоненты: `app/layout.tsx` (корневой), навигация (header с логотипом, ссылка на корзину со счётчиком, профиль), `app/(auth)/layout.tsx` для страниц логина/регистрации (отдельный минималистичный layout).
7. Страницы аутентификации: `app/(auth)/login/page.tsx`, `app/(auth)/register/page.tsx` — Server Actions для отправки формы, `react-hook-form` + `zod` валидация на клиенте, отображение ошибок сервера.
8. Error/Loading boundaries: `app/error.tsx`, `app/loading.tsx`, а также per-route `loading.tsx` со skeleton-компонентами (важно для SSR UX — избегать резких "прыжков" контента).
9. Настроить `next.config.ts` (image domains для placeholder-изображений товаров, если используется внешний сервис типа picsum).

## Acceptance Criteria

- [ ] `pnpm dev` поднимает приложение, страницы логина/регистрации работают против реального backend (EPIC-01)
- [ ] Middleware корректно блокирует доступ к `/cart`, `/checkout`, `/orders` без валидного access-токена
- [ ] Типы API генерируются из `openapi.yaml` без ручного дублирования интерфейсов
- [ ] TypeScript strict mode включён, `pnpm lint`/`pnpm typecheck` без ошибок
- [ ] Дизайн-система документирована (Storybook опционален, но минимум — README в `components/ui/` с примерами использования)

## Definition of Done

Стандартный DoD. Дополнительно: skeleton loading states для минимум 2 типов страниц (каталог, история заказов) — демонстрация зрелого SSR/streaming UX (`Suspense` boundaries).
