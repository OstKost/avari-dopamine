import Link from "next/link";
import { Sparkles, ArrowRight, ShieldCheck, Zap, CheckCircle2, Package, Gift, HelpCircle } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

export const metadata = {
  title: "О проекте и правила | Avari Dopamine",
  description: "Концепция сервиса мгновенного удовольствия и предвкушения. Как работает Dopamine Market и правила покупок.",
};

export default function AboutPage() {
  return (
    <div className="container mx-auto max-w-5xl px-4 sm:px-6 py-10 sm:py-14 space-y-12">
      {/* Hero Banner */}
      <div className="text-center space-y-4 relative">
        <div className="inline-flex items-center gap-2 px-4 py-1 rounded-full bg-amber-400/15 border border-amber-400/30 text-amber-300 text-xs font-bold shadow-glow-amber">
          <Sparkles className="h-3.5 w-3.5 animate-spin" />
          <span>Магия предвкушения и быстрой доставки</span>
        </div>

        <h1 className="text-3xl sm:text-5xl font-black text-[#F4F1E8] tracking-tight max-w-3xl mx-auto leading-tight">
          О проекте{" "}
          <span className="bg-gradient-to-r from-[#F2B84B] via-[#FFD37A] to-[#F2B84B] bg-clip-text text-transparent">
            Avari Dopamine
          </span>
        </h1>

        <p className="text-base sm:text-lg text-[#9FB3C4] max-w-2xl mx-auto leading-relaxed">
          Мы создали маркетплейс, который дарит чистые эмоции онлайн-шопинга без стресса для кошелька. 
          Выбирайте любые классные вещи, применяйте промокод и получайте микро-дозу радости всего за 10 рублей.
        </p>
      </div>

      {/* Concept Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <Card className="border-[#1E3A50] bg-[#0B1622]/90 rounded-3xl p-2 shadow-sm hover:border-amber-400/40 transition-all">
          <CardHeader className="space-y-3 pb-2">
            <div className="h-12 w-12 rounded-2xl bg-amber-400/15 border border-amber-400/30 flex items-center justify-center text-amber-300">
              <Gift className="h-6 w-6" />
            </div>
            <CardTitle className="text-lg font-bold text-[#F4F1E8]">
              Эмоции без лишних трат
            </CardTitle>
          </CardHeader>
          <CardContent className="text-xs sm:text-sm text-[#9FB3C4] leading-relaxed">
            Самая приятная часть шопинга — это выбор, ожидание курьера и открытие коробки. Мы сделали весь этот опыт доступным ровно за 10 рублей.
          </CardContent>
        </Card>

        <Card className="border-[#1E3A50] bg-[#0B1622]/90 rounded-3xl p-2 shadow-sm hover:border-teal-400/40 transition-all">
          <CardHeader className="space-y-3 pb-2">
            <div className="h-12 w-12 rounded-2xl bg-teal-400/15 border border-teal-400/30 flex items-center justify-center text-teal-300">
              <Zap className="h-6 w-6" />
            </div>
            <CardTitle className="text-lg font-bold text-[#F4F1E8]">
              Живой трекинг доставки
            </CardTitle>
          </CardHeader>
          <CardContent className="text-xs sm:text-sm text-[#9FB3C4] leading-relaxed">
            Следите за каждым этапом пути: сборка на складе, назначение курьера, движение по карте и вручение в пункте выдачи в 2–5 минутах от вас.
          </CardContent>
        </Card>

        <Card className="border-[#1E3A50] bg-[#0B1622]/90 rounded-3xl p-2 shadow-sm hover:border-amber-400/40 transition-all">
          <CardHeader className="space-y-3 pb-2">
            <div className="h-12 w-12 rounded-2xl bg-amber-400/15 border border-amber-400/30 flex items-center justify-center text-amber-300">
              <Sparkles className="h-6 w-6" />
            </div>
            <CardTitle className="text-lg font-bold text-[#F4F1E8]">
              Стрики и достижения
            </CardTitle>
          </CardHeader>
          <CardContent className="text-xs sm:text-sm text-[#9FB3C4] leading-relaxed">
            Совершайте покупки каждый день, поддерживайте огонь своего стрика, зарабатывайте опыт (XP) и открывайте редкие коллекционные награды.
          </CardContent>
        </Card>
      </div>

      {/* How It Works Steps */}
      <div className="space-y-6">
        <div className="text-center space-y-2">
          <h2 className="text-2xl sm:text-3xl font-black text-[#F4F1E8]">
            Как работает сервис
          </h2>
          <p className="text-sm text-[#9FB3C4]">
            Всего 4 простых шага от выбора до получения дофаминового заказа
          </p>
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
          <div className="p-5 rounded-2xl bg-[#0B1622] border border-[#1E3A50] space-y-3 relative">
            <span className="text-3xl font-black text-amber-400/30">01</span>
            <h4 className="font-bold text-base text-[#F4F1E8]">Соберите корзину</h4>
            <p className="text-xs text-[#9FB3C4] leading-relaxed">
              Выбирайте любые товары из каталога антистрессов, полезных гаджетов и вдохновляющих мелочей.
            </p>
          </div>

          <div className="p-5 rounded-2xl bg-[#0B1622] border border-[#1E3A50] space-y-3 relative">
            <span className="text-3xl font-black text-amber-400/30">02</span>
            <h4 className="font-bold text-base text-[#F4F1E8]">Промокод «DOPAMINE»</h4>
            <p className="text-xs text-[#9FB3C4] leading-relaxed">
              Спецпредложение автоматически пересчитывает общую стоимость любой корзины до фиксированных 10.00 ₽.
            </p>
          </div>

          <div className="p-5 rounded-2xl bg-[#0B1622] border border-[#1E3A50] space-y-3 relative">
            <span className="text-3xl font-black text-amber-400/30">03</span>
            <h4 className="font-bold text-base text-[#F4F1E8]">Удобный ПВЗ</h4>
            <p className="text-xs text-[#9FB3C4] leading-relaxed">
              Выберите пункт выдачи, расположенный в 100–500 метрах от вашей точки на карте.
            </p>
          </div>

          <div className="p-5 rounded-2xl bg-[#0B1622] border border-[#1E3A50] space-y-3 relative">
            <span className="text-3xl font-black text-amber-400/30">04</span>
            <h4 className="font-bold text-base text-[#F4F1E8]">Живой трекинг</h4>
            <p className="text-xs text-[#9FB3C4] leading-relaxed">
              Смотрите статус доставки онлайн, получайте XP и празднуйте успешное вручение салютом из конфетти!
            </p>
          </div>
        </div>
      </div>

      {/* Rules & FAQ Section */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-8 pt-4">
        {/* Rules */}
        <div className="space-y-4">
          <div className="flex items-center gap-2">
            <ShieldCheck className="h-5 w-5 text-teal-400" />
            <h3 className="text-xl font-bold text-[#F4F1E8]">Правила сервиса</h3>
          </div>
          <div className="space-y-3">
            <div className="p-4 rounded-2xl bg-[#0B1622] border border-[#1E3A50] flex items-start gap-3">
              <CheckCircle2 className="h-5 w-5 text-amber-400 flex-shrink-0 mt-0.5" />
              <div>
                <h5 className="text-sm font-bold text-[#F4F1E8]">Фиксированная стоимость заказа</h5>
                <p className="text-xs text-[#9FB3C4] mt-0.5">
                  При оформлении заказа с промокодом итоговая сумма к оплате всегда составляет ровно 10.00 рублей.
                </p>
              </div>
            </div>

            <div className="p-4 rounded-2xl bg-[#0B1622] border border-[#1E3A50] flex items-start gap-3">
              <CheckCircle2 className="h-5 w-5 text-amber-400 flex-shrink-0 mt-0.5" />
              <div>
                <h5 className="text-sm font-bold text-[#F4F1E8]">Один платеж в процессе</h5>
                <p className="text-xs text-[#9FB3C4] mt-0.5">
                  Для защиты от случайных списаний пользователь может иметь только один заказ на этапе ожидания оплаты.
                </p>
              </div>
            </div>

            <div className="p-4 rounded-2xl bg-[#0B1622] border border-[#1E3A50] flex items-start gap-3">
              <CheckCircle2 className="h-5 w-5 text-amber-400 flex-shrink-0 mt-0.5" />
              <div>
                <h5 className="text-sm font-bold text-[#F4F1E8]">Синтетический опыт</h5>
                <p className="text-xs text-[#9FB3C4] mt-0.5">
                  Каталог, курьеры и ПВЗ являются частью интерактивного симулятора радости покупок и быстрой доставки.
                </p>
              </div>
            </div>
          </div>
        </div>

        {/* FAQ */}
        <div className="space-y-4">
          <div className="flex items-center gap-2">
            <HelpCircle className="h-5 w-5 text-amber-400" />
            <h3 className="text-xl font-bold text-[#F4F1E8]">Частые вопросы</h3>
          </div>
          <div className="space-y-3 text-xs sm:text-sm">
            <details className="group p-4 rounded-2xl bg-[#0B1622] border border-[#1E3A50] [&_summary::-webkit-details-marker]:hidden">
              <summary className="font-bold text-[#F4F1E8] cursor-pointer flex items-center justify-between">
                <span>Почему заказ стоит 10 рублей?</span>
                <span className="text-amber-400 group-open:rotate-180 transition-transform">▼</span>
              </summary>
              <p className="text-[#9FB3C4] mt-2 leading-relaxed">
                Мы берем плату исключительно за запуск интерактивной симуляции доставки и генерацию вашего уникального дофаминового опыта.
              </p>
            </details>

            <details className="group p-4 rounded-2xl bg-[#0B1622] border border-[#1E3A50] [&_summary::-webkit-details-marker]:hidden">
              <summary className="font-bold text-[#F4F1E8] cursor-pointer flex items-center justify-between">
                <span>Как работает стрик ежедневных заказов?</span>
                <span className="text-amber-400 group-open:rotate-180 transition-transform">▼</span>
              </summary>
              <p className="text-[#9FB3C4] mt-2 leading-relaxed">
                Делая хотя бы один заказ в сутки (с учетом вашего часового пояса), вы увеличиваете счетчик огонька. Стрик открывает эксклюзивные бейджи и поднимает уровень аккаунта.
              </p>
            </details>

            <details className="group p-4 rounded-2xl bg-[#0B1622] border border-[#1E3A50] [&_summary::-webkit-details-marker]:hidden">
              <summary className="font-bold text-[#F4F1E8] cursor-pointer flex items-center justify-between">
                <span>Где посмотреть исходный код и архитектуру?</span>
                <span className="text-amber-400 group-open:rotate-180 transition-transform">▼</span>
              </summary>
              <p className="text-[#9FB3C4] mt-2 leading-relaxed">
                Для разработчиков, архитекторов и рекрутеров мы подготовили отдельную страницу{" "}
                <Link href="/tech" className="text-amber-300 font-bold underline hover:text-amber-200">
                  Архитектура и стек проекта
                </Link>
                .
              </p>
            </details>
          </div>
        </div>
      </div>

      {/* CTA Footer */}
      <div className="rounded-3xl bg-gradient-to-r from-amber-500/20 via-teal-500/20 to-amber-500/20 p-0.5 border border-amber-400/30 shadow-glow-amber">
        <div className="rounded-[22px] bg-[#0B1622]/95 p-6 sm:p-8 text-center space-y-5">
          <h3 className="text-2xl font-black text-[#F4F1E8]">
            Готовы получить свою дозу дофамина?
          </h3>
          <p className="text-sm text-[#9FB3C4] max-w-md mx-auto">
            Переходите в каталог, выбирайте понравившиеся товары и оформляйте доставку всего за 10 рублей!
          </p>
          <div className="flex flex-wrap items-center justify-center gap-4 pt-2">
            <Link href="/catalog">
              <Button size="lg" variant="glow" className="rounded-2xl gap-2">
                <Package className="h-5 w-5" />
                <span>Открыть каталог товаров</span>
                <ArrowRight className="h-4 w-4" />
              </Button>
            </Link>
            <Link href="/tech">
              <Button size="lg" variant="outline" className="rounded-2xl border-[#1E3A50] text-[#9FB3C4] hover:text-[#F4F1E8]">
                <span>Для разработчиков (Tech Stack)</span>
              </Button>
            </Link>
          </div>
        </div>
      </div>
    </div>
  );
}
