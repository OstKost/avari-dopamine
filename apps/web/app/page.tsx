import Link from "next/link";
import Image from "next/image";
import { Sparkles, ArrowRight, Zap, ShieldCheck, Flame } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";

export default function HomePage() {
  return (
    <div className="flex flex-col items-center">
      {/* Hero Section with detailed logo and ambient glow */}
      <section className="relative w-full overflow-hidden py-16 sm:py-24 md:py-28">
        {/* Ambient radial glows */}
        <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[600px] h-[600px] bg-teal-500/10 rounded-full blur-[120px] pointer-events-none" />
        <div className="absolute top-1/3 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[350px] h-[350px] bg-amber-500/15 rounded-full blur-[90px] pointer-events-none" />

        <div className="container relative z-10 mx-auto max-w-5xl px-4 sm:px-6 text-center space-y-6">
          {/* Detailed Emblem in Hero */}
          <div className="inline-flex relative p-3 sm:p-4 rounded-3xl bg-[#0B1622]/90 border border-amber-400/40 shadow-glow-amber-lg group">
            <Image
              src="/logo-detailed.png"
              alt="Avari Dopamine Emblem"
              width={88}
              height={88}
              className="object-contain drop-shadow-[0_0_20px_rgba(242,184,75,0.4)] group-hover:scale-105 transition-transform duration-500"
              priority
            />
          </div>

          <div className="flex items-center justify-center gap-2">
            <Badge variant="gold" className="px-4 py-1 text-xs gap-1.5 shadow-glow-amber">
              <Sparkles className="h-3.5 w-3.5 text-amber-300 animate-spin" />
              <span>Спеццена доставки: 10.00 ₽ по промокоду DOPAMINE</span>
            </Badge>
          </div>

          <h1 className="text-3xl sm:text-5xl md:text-6xl font-black tracking-tight text-[#F4F1E8] max-w-3xl mx-auto leading-tight">
            Мгновенный выброс{" "}
            <span className="bg-gradient-to-r from-[#F2B84B] via-[#FFD37A] to-[#F2B84B] bg-clip-text text-transparent">
              дофамина
            </span>{" "}
            в каждом заказе
          </h1>

          <p className="mt-6 text-lg sm:text-xl text-[#9FB3C4] max-w-2xl mx-auto leading-relaxed">
            Маркетплейс мгновенной радости и предвкушения: выбирайте классные штуки для настроения, применяйте промокод и наблюдайте за быстрой доставкой курьером.
          </p>

          <div className="mt-10 flex flex-wrap items-center justify-center gap-4">
            <Link href="/catalog">
              <Button size="lg" variant="glow" className="gap-2 text-base">
                Открыть каталог
                <ArrowRight className="h-5 w-5" />
              </Button>
            </Link>
            <Link href="/about">
              <Button size="lg" variant="outline" className="text-base border-[#1E3A50] text-[#F4F1E8] hover:bg-[#1E3A50]">
                Как это работает
              </Button>
            </Link>
          </div>
        </div>
      </section>

      {/* Features Grid */}
      <section className="container mx-auto max-w-6xl px-4 sm:px-6 py-8 sm:py-12">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-5 sm:gap-6">
          <Card className="border-[#1E3A50] bg-[#0B1622]/90 shadow-sm rounded-2xl hover:border-amber-400/50 hover:shadow-glow-amber transition-all">
            <CardContent className="p-6 space-y-3">
              <div className="flex h-12 w-12 items-center justify-center rounded-2xl bg-amber-400/15 text-amber-300 border border-amber-400/30">
                <Zap className="h-6 w-6" />
              </div>
              <h3 className="text-lg font-bold text-[#F4F1E8]">
                Фикс-прайс 10 ₽ по промокоду
              </h3>
              <p className="text-xs sm:text-sm text-[#9FB3C4] leading-relaxed">
                Набирайте любые товары в корзину — секретный промокод снижает итоговую стоимость всего заказа до символических 10 рублей.
              </p>
            </CardContent>
          </Card>

          <Card className="border-[#1E3A50] bg-[#0B1622]/90 shadow-sm rounded-2xl hover:border-teal-400/50 hover:shadow-glow-teal transition-all">
            <CardContent className="p-6 space-y-3">
              <div className="flex h-12 w-12 items-center justify-center rounded-2xl bg-teal-400/15 text-teal-300 border border-teal-400/30">
                <ShieldCheck className="h-6 w-6" />
              </div>
              <h3 className="text-lg font-bold text-[#F4F1E8]">
                Быстрая доставка рядом с вами
              </h3>
              <p className="text-xs sm:text-sm text-[#9FB3C4] leading-relaxed">
                Удобные пункты выдачи в 2-5 минутах ходьбы и интерактивный живой трекинг движения курьера прямо на экране.
              </p>
            </CardContent>
          </Card>

          <Card className="border-[#1E3A50] bg-[#0B1622]/90 shadow-sm rounded-2xl hover:border-amber-400/50 hover:shadow-glow-amber transition-all">
            <CardContent className="p-6 space-y-3">
              <div className="flex h-12 w-12 items-center justify-center rounded-2xl bg-amber-400/15 text-amber-300 border border-amber-400/30">
                <Flame className="h-6 w-6" />
              </div>
              <h3 className="text-lg font-bold text-[#F4F1E8]">
                Стрики и награды
              </h3>
              <p className="text-xs sm:text-sm text-[#9FB3C4] leading-relaxed">
                Сохраняйте ежедневный стрик, повышайте уровень аккаунта и открывайте коллекцию достижений с золотым сиянием.
              </p>
            </CardContent>
          </Card>
        </div>
      </section>
    </div>
  );
}
