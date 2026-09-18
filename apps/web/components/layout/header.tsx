"use client";

import Link from "next/link";
import { Sparkles, ShoppingBag, Package, User } from "lucide-react";
import { Button } from "@/components/ui/button";

export function Header() {
  return (
    <header className="sticky top-0 z-40 w-full border-b border-zinc-200/80 bg-white/80 backdrop-blur-md dark:border-zinc-800 dark:bg-zinc-950/80">
      <div className="container mx-auto flex h-16 max-w-7xl items-center justify-between px-4 sm:px-6">
        <Link href="/" className="flex items-center gap-2 group">
          <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-gradient-to-tr from-rose-500 to-purple-600 text-white shadow-md shadow-rose-500/20 group-hover:scale-105 transition-transform">
            <Sparkles className="h-5 w-5 animate-pulse" />
          </div>
          <div>
            <span className="text-xl font-black tracking-tight bg-gradient-to-r from-rose-600 via-pink-600 to-purple-600 bg-clip-text text-transparent">
              Dopamine
            </span>
            <span className="text-xs font-semibold text-zinc-500 block -mt-1">
              Market
            </span>
          </div>
        </Link>

        <nav className="hidden md:flex items-center gap-6 text-sm font-medium text-zinc-600 dark:text-zinc-400">
          <Link href="/catalog" className="hover:text-rose-600 transition-colors">
            Каталог
          </Link>
          <Link href="/orders" className="hover:text-rose-600 transition-colors flex items-center gap-1.5">
            <Package className="h-4 w-4" />
            Заказы
          </Link>
          <Link href="/profile" className="hover:text-rose-600 transition-colors flex items-center gap-1.5">
            <User className="h-4 w-4" />
            Профиль
          </Link>
        </nav>

        <div className="flex items-center gap-3">
          <Link href="/cart">
            <Button variant="outline" size="sm" className="gap-2 relative">
              <ShoppingBag className="h-4 w-4 text-rose-500" />
              <span>Корзина</span>
              <span className="rounded-full bg-rose-500 px-1.5 py-0.2 text-[10px] font-bold text-white">
                10 ₽
              </span>
            </Button>
          </Link>

          <Link href="/profile">
            <Button variant="ghost" size="sm" className="gap-2">
              <User className="h-4 w-4" />
              <span className="hidden sm:inline">Профиль</span>
            </Button>
          </Link>
        </div>
      </div>
    </header>
  );
}
