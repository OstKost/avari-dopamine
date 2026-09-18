import Link from "next/link";
import { Sparkles, ArrowRight, Zap, ShieldCheck, HeartHandshake } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";

export default function HomePage() {
  return (
    <div className="flex flex-col items-center">
      {/* Hero Section */}
      <section className="relative w-full overflow-hidden py-20 md:py-32 bg-gradient-to-b from-rose-50/50 via-white to-transparent dark:from-zinc-900/50 dark:via-zinc-950 dark:to-transparent">
        <div className="container mx-auto max-w-7xl px-4 sm:px-6 text-center">
          <Badge variant="dopamine" className="mb-6 px-4 py-1.5 text-sm gap-2">
            <Sparkles className="h-4 w-4 animate-spin text-rose-500" />
            Инвариант INV-01: Любой заказ ровно за 10 ₽
          </Badge>

          <h1 className="text-4xl font-extrabold tracking-tight sm:text-6xl md:text-7xl text-zinc-900 dark:text-zinc-50 max-w-4xl mx-auto leading-tight">
            Мгновенный выброс{" "}
            <span className="bg-gradient-to-r from-rose-500 via-pink-500 to-purple-600 bg-clip-text text-transparent">
              дофамина
            </span>{" "}
            в каждом заказе
          </h1>

          <p className="mt-6 text-lg sm:text-xl text-zinc-600 dark:text-zinc-300 max-w-2xl mx-auto">
            Синтетический маркетплейс с реальной стейт-машиной доставки, геодезической генерацией ПВЗ и transactional outbox.
          </p>

          <div className="mt-10 flex flex-wrap items-center justify-center gap-4">
            <Link href="/catalog">
              <Button size="lg" variant="glow" className="gap-2 text-base">
                Открыть каталог
                <ArrowRight className="h-5 w-5" />
              </Button>
            </Link>
            <Link href="/login">
              <Button size="lg" variant="outline" className="text-base">
                Войти в профиль
              </Button>
            </Link>
          </div>
        </div>
      </section>

      {/* Features Cards */}
      <section className="container mx-auto max-w-7xl px-4 sm:px-6 py-16">
        <div className="grid grid-cols-1 gap-8 md:grid-cols-3">
          <Card className="hover:shadow-lg transition-shadow border-rose-100 dark:border-zinc-800">
            <CardContent className="p-8 space-y-4">
              <div className="flex h-12 w-12 items-center justify-center rounded-2xl bg-rose-500/10 text-rose-500">
                <Zap className="h-6 w-6" />
              </div>
              <h3 className="text-xl font-bold text-zinc-900 dark:text-zinc-100">
                10 ₽ Фиксированная цена
              </h3>
              <p className="text-sm text-zinc-600 dark:text-zinc-400">
                Сколько бы товаров ни было в корзине, итоговая сумма к оплате строго фиксирована инвариантом INV-01.
              </p>
            </CardContent>
          </Card>

          <Card className="hover:shadow-lg transition-shadow border-purple-100 dark:border-zinc-800">
            <CardContent className="p-8 space-y-4">
              <div className="flex h-12 w-12 items-center justify-center rounded-2xl bg-purple-500/10 text-purple-600">
                <ShieldCheck className="h-6 w-6" />
              </div>
              <h3 className="text-xl font-bold text-zinc-900 dark:text-zinc-100">
                Синтетические ПВЗ
              </h3>
              <p className="text-sm text-zinc-600 dark:text-zinc-400">
                Геодезический алгоритм вычисляет ближайшие пункты выдачи на расстоянии от 100 до 500 метров.
              </p>
            </CardContent>
          </Card>

          <Card className="hover:shadow-lg transition-shadow border-pink-100 dark:border-zinc-800">
            <CardContent className="p-8 space-y-4">
              <div className="flex h-12 w-12 items-center justify-center rounded-2xl bg-pink-500/10 text-pink-500">
                <HeartHandshake className="h-6 w-6" />
              </div>
              <h3 className="text-xl font-bold text-zinc-900 dark:text-zinc-100">
                Живой трекинг
              </h3>
              <p className="text-sm text-zinc-600 dark:text-zinc-400">
                Асинхронная стейт-машина заказов с публикацией Kafka-событий через Transactional Outbox.
              </p>
            </CardContent>
          </Card>
        </div>
      </section>
    </div>
  );
}
