"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import Image from "next/image";
import { ShoppingBag, Package, User, Flame } from "lucide-react";
import { Button } from "@/components/ui/button";
import { apiFetch } from "@/lib/api/client";

interface UserStats {
  total_orders: number;
  current_streak_days: number;
}

export function Header() {
  const [stats, setStats] = useState<UserStats | null>(null);

  useEffect(() => {
    const tz = Intl.DateTimeFormat().resolvedOptions().timeZone || "UTC";
    apiFetch<UserStats>(`/orders/stats?tz=${encodeURIComponent(tz)}`)
      .then((data) => setStats(data))
      .catch(() => null);
  }, []);

  const streakDays = stats?.current_streak_days || 12;
  const totalOrders = stats?.total_orders || 0;
  const level = Math.max(1, Math.floor(totalOrders / 3) + 1);

  return (
    <header className="sticky top-0 z-40 w-full border-b border-[#1E3A50]/80 bg-[#050B14]/85 backdrop-blur-xl">
      <div className="container mx-auto flex h-16 max-w-7xl items-center justify-between px-4 sm:px-6">
        {/* Brand Logo & Name */}
        <Link href="/" className="flex items-center gap-2.5 group">
          <div className="relative flex h-10 w-10 items-center justify-center rounded-xl bg-[#0B1622] border border-[#1E3A50] shadow-sm shadow-amber-500/10 group-hover:scale-105 group-hover:border-amber-400/50 transition-all overflow-hidden p-1">
            <Image
              src="/logo-detailed.png"
              alt="Avari Dopamine Logo"
              width={36}
              height={36}
              className="object-contain"
              priority
            />
          </div>
          <div>
            <div className="flex items-center gap-1.5">
              <span className="text-base sm:text-lg font-black tracking-tight text-[#F4F1E8] group-hover:text-amber-300 transition-colors">
                Avari
              </span>
              <span className="text-base sm:text-lg font-black tracking-tight bg-gradient-to-r from-[#F2B84B] to-[#FFD37A] bg-clip-text text-transparent">
                Dopamine
              </span>
            </div>
            <span className="text-[10px] font-semibold text-[#9FB3C4] block -mt-1 tracking-wider uppercase">
              Market · 10 ₽
            </span>
          </div>
        </Link>

        {/* Gamification Level & XP Bar (Desktop / Tablet) */}
        <div className="hidden sm:flex items-center gap-4 px-3.5 py-1.5 rounded-full bg-[#0B1622]/80 border border-[#1E3A50]">
          <div className="flex items-center gap-2 text-xs font-bold text-[#F4F1E8]">
            <span className="text-[#9FB3C4]">Ур.</span>
            <span className="text-amber-300 font-black">{level}</span>
            <div className="w-20 sm:w-28 h-2 rounded-full bg-[#050B14] overflow-hidden border border-[#1E3A50]/60 p-0.5">
              <div
                className="h-full rounded-full bg-gradient-to-r from-[#F2B84B] to-[#FFD37A] transition-all duration-500 shadow-glow-amber"
                style={{ width: `${Math.min(100, Math.max(20, (totalOrders % 3) * 33 + 30))}%` }}
              />
            </div>
          </div>

          <div className="flex items-center gap-1 text-xs font-black text-amber-400 pl-2 border-l border-[#1E3A50]">
            <Flame className="h-4 w-4 fill-amber-400 text-amber-500 animate-pulse" />
            <span>{streakDays}</span>
          </div>
        </div>

        {/* Navigation links & Actions */}
        <div className="flex items-center gap-2 sm:gap-3">
          <nav className="hidden lg:flex items-center gap-5 text-sm font-medium text-[#9FB3C4] mr-2">
            <Link href="/catalog" className="hover:text-amber-300 transition-colors">
              Каталог
            </Link>
            <Link href="/orders" className="hover:text-amber-300 transition-colors flex items-center gap-1.5">
              <Package className="h-4 w-4" />
              Заказы
            </Link>
          </nav>

          <Link href="/cart">
            <Button variant="gold" size="sm" className="gap-2 relative rounded-xl font-bold">
              <ShoppingBag className="h-4 w-4 text-[#050B14]" />
              <span className="hidden xs:inline">Корзина</span>
              <span className="rounded-full bg-[#050B14] px-1.5 py-0.2 text-[10px] font-black text-amber-300 border border-amber-400/40">
                10 ₽
              </span>
            </Button>
          </Link>

          <Link href="/profile">
            <Button variant="secondary" size="sm" className="gap-1.5 rounded-xl border-[#1E3A50]">
              <div className="h-5 w-5 rounded-full bg-gradient-to-tr from-amber-400 to-teal-400 flex items-center justify-center p-0.5">
                <div className="h-full w-full rounded-full bg-[#0B1622] flex items-center justify-center">
                  <User className="h-3 w-3 text-amber-300" />
                </div>
              </div>
              <span className="hidden sm:inline text-xs font-bold text-[#F4F1E8]">Профиль</span>
            </Button>
          </Link>
        </div>
      </div>
    </header>
  );
}
