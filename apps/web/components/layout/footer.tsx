import Link from "next/link";
import Image from "next/image";
import { ShieldCheck, Zap, Heart, Sparkles, ExternalLink } from "lucide-react";

export function Footer() {
  return (
    <footer className="mt-auto border-t border-[#1E3A50]/80 bg-[#050B14]/90 py-12 pb-24 md:pb-12 text-[#9FB3C4]">
      <div className="container mx-auto max-w-7xl px-4 sm:px-6">
        <div className="grid grid-cols-1 gap-8 md:grid-cols-4">
          <div className="space-y-4 md:col-span-2">
            <div className="flex items-center gap-3">
              <div className="relative flex h-10 w-10 items-center justify-center rounded-xl bg-[#0B1622] border border-[#1E3A50] p-1 shadow-sm">
                <Image
                  src="/logo-detailed.png"
                  alt="Avari Dopamine Logo"
                  width={36}
                  height={36}
                  className="object-contain"
                />
              </div>
              <div>
                <span className="font-black text-lg text-[#F4F1E8]">
                  Avari <span className="bg-gradient-to-r from-amber-400 to-amber-300 bg-clip-text text-transparent">Dopamine</span>
                </span>
                <span className="text-xs text-[#5E7488] block -mt-1 font-semibold">
                  Dopamine Market Engine
                </span>
              </div>
            </div>
            <p className="text-sm text-[#9FB3C4] max-w-sm leading-relaxed">
              Маркетплейс мгновенной радости и предвкушения. Любой заказ всего за 10 рублей по промокоду DOPAMINE.
            </p>
          </div>

          <div>
            <h4 className="text-sm font-bold text-[#F4F1E8] mb-3 uppercase tracking-wider text-xs">
              Разделы
            </h4>
            <ul className="space-y-2 text-sm text-[#9FB3C4]">
              <li>
                <Link href="/catalog" className="hover:text-amber-300 transition-colors">
                  🛍️ Каталог товаров
                </Link>
              </li>
              <li>
                <Link href="/cart" className="hover:text-amber-300 transition-colors">
                  🛒 Корзина
                </Link>
              </li>
              <li>
                <Link href="/orders" className="hover:text-amber-300 transition-colors">
                  📦 Мои заказы
                </Link>
              </li>
              <li>
                <Link href="/profile" className="hover:text-amber-300 transition-colors">
                  🏆 Достижения и стрик
                </Link>
              </li>
              <li>
                <Link href="/about" className="hover:text-amber-300 transition-colors font-medium text-amber-400">
                  ✨ О проекте и правила
                </Link>
              </li>
              <li>
                <Link href="/tech" className="hover:text-teal-300 transition-colors font-medium text-teal-400">
                  ⚙️ Архитектура (For Tech Leads)
                </Link>
              </li>
            </ul>
          </div>

          <div>
            <h4 className="text-sm font-bold text-[#F4F1E8] mb-3 uppercase tracking-wider text-xs">
              Автор и поддержка
            </h4>
            <ul className="space-y-3 text-sm">
              <li>
                <a
                  href="https://boosty.to/ostkost"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="inline-flex items-center gap-2 px-3.5 py-2 rounded-xl bg-gradient-to-r from-[#F2B84B]/15 via-[#FFD37A]/10 to-transparent border border-amber-400/40 text-amber-300 font-bold hover:brightness-125 hover:border-amber-400 transition-all shadow-glow-amber group"
                >
                  <Sparkles className="h-4 w-4 text-amber-400 animate-pulse" />
                  <span>Boosty: ostkost</span>
                  <ExternalLink className="h-3.5 w-3.5 opacity-70 group-hover:translate-x-0.5 group-hover:-translate-y-0.5 transition-transform" />
                </a>
              </li>
              <li className="flex items-center gap-2 text-xs text-[#5E7488]">
                <ShieldCheck className="h-4 w-4 text-teal-400" />
                <span>Фикс-прайс: 10 ₽ по промокоду</span>
              </li>
              <li className="flex items-center gap-2 text-xs text-[#5E7488]">
                <Zap className="h-4 w-4 text-amber-400" />
                <span>Быстрая интерактивная доставка</span>
              </li>
            </ul>
          </div>
        </div>

        <div className="mt-10 border-t border-[#1E3A50]/60 pt-8 flex flex-col sm:flex-row items-center justify-between gap-4 text-xs text-[#5E7488]">
          <p>© {new Date().getFullYear()} Avari Dopamine. Все товары, курьеры и ПВЗ синтетические.</p>
          <div className="flex items-center gap-4">
            <a href="https://boosty.to/ostkost" target="_blank" rel="noopener noreferrer" className="hover:text-amber-300 transition-colors flex items-center gap-1">
              <span>Поддержать автора (ostkost)</span>
              <Heart className="h-3 w-3 text-amber-400 fill-amber-400" />
            </a>
          </div>
        </div>
      </div>
    </footer>
  );
}
