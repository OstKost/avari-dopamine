import Link from "next/link";
import { Sparkles, ShieldCheck, Zap, Heart } from "lucide-react";

export function Footer() {
  return (
    <footer className="mt-auto border-t border-zinc-200 bg-zinc-50/50 py-12 dark:border-zinc-800 dark:bg-zinc-950/50">
      <div className="container mx-auto max-w-7xl px-4 sm:px-6">
        <div className="grid grid-cols-1 gap-8 md:grid-cols-4">
          <div className="space-y-3 md:col-span-2">
            <div className="flex items-center gap-2">
              <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-rose-500 text-white">
                <Sparkles className="h-4 w-4" />
              </div>
              <span className="font-bold text-lg text-zinc-900 dark:text-zinc-100">
                Dopamine Market
              </span>
            </div>
            <p className="text-sm text-zinc-600 dark:text-zinc-400 max-w-sm">
              Синтетический маркетплейс мгновенной радости. Фиксированная цена любого заказа — ровно 10 рублей (INV-01).
            </p>
          </div>

          <div>
            <h4 className="text-sm font-semibold text-zinc-900 dark:text-zinc-100 mb-3">
              Разделы
            </h4>
            <ul className="space-y-2 text-sm text-zinc-600 dark:text-zinc-400">
              <li>
                <Link href="/catalog" className="hover:text-rose-600 transition-colors">
                  Каталог
                </Link>
              </li>
              <li>
                <Link href="/cart" className="hover:text-rose-600 transition-colors">
                  Корзина
                </Link>
              </li>
              <li>
                <Link href="/orders" className="hover:text-rose-600 transition-colors">
                  Мои заказы
                </Link>
              </li>
            </ul>
          </div>

          <div>
            <h4 className="text-sm font-semibold text-zinc-900 dark:text-zinc-100 mb-3">
              Особенности
            </h4>
            <ul className="space-y-2 text-sm text-zinc-600 dark:text-zinc-400">
              <li className="flex items-center gap-1.5">
                <ShieldCheck className="h-4 w-4 text-emerald-500" />
                <span>Все заказы по 10 ₽</span>
              </li>
              <li className="flex items-center gap-1.5">
                <Zap className="h-4 w-4 text-amber-500" />
                <span>Синтетическая доставка</span>
              </li>
              <li className="flex items-center gap-1.5">
                <Heart className="h-4 w-4 text-rose-500" />
                <span>Портфолио проект</span>
              </li>
            </ul>
          </div>
        </div>

        <div className="mt-8 border-t border-zinc-200 dark:border-zinc-800 pt-8 text-center text-xs text-zinc-500">
          © {new Date().getFullYear()} Dopamine Market. Все товары и курьеры синтетические.
        </div>
      </div>
    </footer>
  );
}
