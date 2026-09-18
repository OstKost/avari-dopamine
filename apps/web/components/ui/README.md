# Dopamine Design System Primitives

Набор переиспользуемых UI-компонентов на основе Tailwind CSS и React 19 / Next.js 15.

## Токены и цвета
- **Primary / Brand:** Rose (`#f43f5e`), Electric Violet (`#8b5cf6`), Pink Glow gradients.
- **Fixed Order Price Invariant (INV-01):** 10.00 RUB.

## Примитивы
- **`Button`**: Варианты: `default`, `secondary`, `outline`, `ghost`, `destructive`, `glow`. Поддерживает размеры `sm`, `md`, `lg`, `icon` и состояние `isLoading`.
- **`Input`**: Поле ввода с поддержкой `label` и отображением ошибок валидации `error`.
- **`Card`**, `CardHeader`, `CardTitle`, `CardDescription`, `CardContent`, `CardFooter`: Контейнеры с мягкими закруглениями (`rounded-2xl`).
- **`Badge`**: Метки категорий и статусов заказа (`default`, `secondary`, `success`, `warning`, `destructive`, `dopamine`).
- **`ProgressBar`**: Линейный индикатор прогресса доставки и сборки.
- **`Skeleton`**: Анимированные плейсхолдеры для стриминга SSR и Suspense-границ.
- **`Spinner`**: Круговой индикатор загрузки.
