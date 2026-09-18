---
id: EPIC-10
title: Frontend Catalog & Cart UX
module: [web]
priority: P0
status: done
depends_on: [EPIC-09, EPIC-02, EPIC-03, EPIC-04]
covers_requirements: [FR-CATALOG-01, FR-CATALOG-02, FR-PICKUP-01, FR-CART-01]
related_adrs: [ADR-009, ADR-008]
scope:
  paths:
    - apps/web/app/(main)/catalog/**
    - apps/web/app/(main)/product/**
    - apps/web/app/(main)/cart/**
    - apps/web/app/(main)/onboarding/**
    - apps/web/components/features/catalog/**
    - apps/web/components/features/cart/**
    - apps/web/components/features/pickup/**
---

# EPIC-10: Frontend Catalog & Cart UX

## Цель

Онбординг выбора ПВЗ, каталог товаров (SSR + поиск), страница товара, корзина с оптимистичными обновлениями.

## Задачи

1. `app/(main)/onboarding/page.tsx`: запрос геолокации браузера (`navigator.geolocation`), fallback на ручной ввод города, вызов `POST /pickup/generate`, отображение сгенерированных точек для выбора (список/карточки, БЕЗ реальной карты согласно продуктовому решению — ADR-005 контекст также применим к пункту выдачи: не показываем на карте).
2. `app/(main)/catalog/page.tsx`: Server Component, SSR-фетч `GET /catalog/products` с параметрами категории/страницы из searchParams, рендер сетки товаров.
3. `components/features/catalog/SearchBar.tsx`: Client Component с debounced input, обновляет URL searchParam `q` (используя `useRouter`/`usePathname` — SSR перерендеривает результаты без полного client-side fetch waterfall благодаря Next.js App Router навигации).
4. `components/features/catalog/ProductCard.tsx`: карточка товара с изображением (placeholder-сервис из ADR-012), ценой, кнопкой "В корзину" (оптимистичное обновление счётчика через `useOptimistic` + Server Action).
5. `app/(main)/product/[id]/page.tsx`: SSR страница товара, `generateMetadata` для SEO (даже для портфолио — демонстрация полноты SSR-владения).
6. `app/(main)/cart/page.tsx`: список товаров в корзине, изменение количества (inline +/- с debounced PATCH-запросом), удаление, отображение выбранного ПВЗ с возможностью сменить (модалка/отдельная страница выбора из уже сгенерированных точек), явный акцент в UI на "к оплате: 10₽" НЕЗАВИСИМО от суммы товаров (продуктовая фишка — обязательно визуально выделить это как "фичу", не как баг).
7. `hooks/useCart.ts`: TanStack Query для серверного состояния корзины, мутации с optimistic updates и rollback при ошибке.
8. Обработка пустой корзины — явный empty-state с CTA "Перейти в каталог".

## Acceptance Criteria

- [x] Онбординг работает с геолокацией и без неё (fallback на город) (`FR-PICKUP-01`)
- [x] Поиск и фильтр по категории работают, обновляют URL (шарибельные ссылки на результаты поиска) (`FR-CATALOG-02`)
- [x] Добавление в корзину — оптимистичное, без полной перезагрузки страницы (`FR-CART-01`)
- [x] Корзина отображает фиксированную стоимость заказа 10₽ явно и заметно, независимо от состава
- [x] Смена ПВЗ не сбрасывает товары в корзине (совпадает с backend acceptance criteria EPIC-04)
- [x] Изображения товаров используют `next/image` с корректной конфигурацией доменов

## Definition of Done

Стандартный DoD (frontend: `pnpm lint`, `pnpm typecheck`, `pnpm test` — компонентные тесты через Vitest/Testing Library для ProductCard, SearchBar, CartItem). Ручная проверка на мобильной ширине экрана (responsive) — обязательна перед мержем, приложить скриншоты в PR.
