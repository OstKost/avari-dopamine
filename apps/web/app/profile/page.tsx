"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { User, Flame, Trophy, Sparkles, ArrowRight, Package } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { StreakBadge } from "@/components/features/gamification/StreakBadge";
import { apiFetch } from "@/lib/api/client";
import { formatPrice } from "@/lib/utils";

interface UserProfile {
  id: string;
  email: string;
  created_at: string;
}

interface UserStats {
  total_orders: number;
  current_streak_days: number;
  longest_streak_days: number;
  total_saved_rub: string;
}

export default function ProfilePage() {
  const [user, setUser] = useState<UserProfile | null>(null);
  const [stats, setStats] = useState<UserStats | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    async function loadData() {
      try {
        const tz = Intl.DateTimeFormat().resolvedOptions().timeZone || "UTC";
        const [userData, statsData] = await Promise.all([
          apiFetch<{ user: UserProfile }>("/auth/me").catch(() => null),
          apiFetch<UserStats>(`/orders/stats?tz=${encodeURIComponent(tz)}`).catch(() => null),
        ]);

        if (userData?.user) setUser(userData.user);
        if (statsData) setStats(statsData);
      } finally {
        setIsLoading(false);
      }
    }
    loadData();
  }, []);

  if (isLoading) {
    return (
      <div className="container mx-auto max-w-2xl px-4 py-16 text-center">
        <div className="animate-spin rounded-full h-10 w-10 border-b-2 border-rose-500 mx-auto mb-4" />
        <p className="text-sm text-zinc-500">Загрузка профиля...</p>
      </div>
    );
  }

  if (!user) {
    return (
      <div className="container mx-auto max-w-md px-4 py-20 text-center space-y-6">
        <h2 className="text-2xl font-black text-zinc-900 dark:text-zinc-50">Вы не вошли в аккаунт</h2>
        <p className="text-sm text-zinc-500">Войдите или зарегистрируйтесь, чтобы просматривать профиль и streak.</p>
        <Link href="/login">
          <Button variant="glow" className="rounded-2xl">Войти</Button>
        </Link>
      </div>
    );
  }

  return (
    <div className="container mx-auto max-w-3xl px-4 sm:px-6 py-8 space-y-8">
      {/* Profile Header */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 p-6 rounded-3xl bg-zinc-50 dark:bg-zinc-900 border border-zinc-200/80 dark:border-zinc-800">
        <div className="flex items-center gap-4">
          <div className="flex h-16 w-16 items-center justify-center rounded-2xl bg-gradient-to-tr from-rose-500 to-purple-600 text-white font-black text-xl shadow-md shadow-rose-500/20">
            <User className="h-8 w-8" />
          </div>
          <div className="space-y-1">
            <h1 className="text-xl sm:text-2xl font-black text-zinc-900 dark:text-zinc-50 tracking-tight">
              {user.email}
            </h1>
            <p className="text-xs text-zinc-400">
              В Dopamine Market с {new Date(user.created_at).toLocaleDateString("ru-RU")}
            </p>
          </div>
        </div>

        {stats && stats.current_streak_days > 0 && (
          <StreakBadge days={stats.current_streak_days} className="text-sm px-4 py-1.5" />
        )}
      </div>

      {/* Gamification Stats Grid */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <Card className="border-zinc-200/80 dark:border-zinc-800 shadow-sm">
          <CardHeader className="p-5 pb-2">
            <CardTitle className="text-xs text-zinc-400 font-semibold uppercase tracking-wider flex items-center justify-between">
              <span>Всего заказов</span>
              <Package className="h-4 w-4 text-rose-500" />
            </CardTitle>
          </CardHeader>
          <CardContent className="p-5 pt-0">
            <div className="text-3xl font-black text-zinc-900 dark:text-zinc-50">
              {stats?.total_orders || 0}
            </div>
            <p className="text-[11px] text-zinc-500 mt-1">Каждый заказ ровно 10 ₽</p>
          </CardContent>
        </Card>

        <Card className="border-zinc-200/80 dark:border-zinc-800 shadow-sm">
          <CardHeader className="p-5 pb-2">
            <CardTitle className="text-xs text-zinc-400 font-semibold uppercase tracking-wider flex items-center justify-between">
              <span>Текущий Streak</span>
              <Flame className="h-4 w-4 text-orange-500" />
            </CardTitle>
          </CardHeader>
          <CardContent className="p-5 pt-0">
            <div className="text-3xl font-black text-amber-500">
              {stats?.current_streak_days || 0} <span className="text-base font-bold text-zinc-400">дн.</span>
            </div>
            <p className="text-[11px] text-zinc-500 mt-1">
              Рекорд: {stats?.longest_streak_days || 0} дн. подряд
            </p>
          </CardContent>
        </Card>

        <Card className="border-zinc-200/80 dark:border-zinc-800 shadow-sm">
          <CardHeader className="p-5 pb-2">
            <CardTitle className="text-xs text-zinc-400 font-semibold uppercase tracking-wider flex items-center justify-between">
              <span>Dopamine выгода</span>
              <Trophy className="h-4 w-4 text-emerald-500" />
            </CardTitle>
          </CardHeader>
          <CardContent className="p-5 pt-0">
            <div className="text-3xl font-black text-emerald-500">
              {formatPrice(stats?.total_saved_rub || "0.00")}
            </div>
            <p className="text-[11px] text-zinc-500 mt-1">Сэкономлено по INV-01</p>
          </CardContent>
        </Card>
      </div>

      {/* Dopamine Banner */}
      <div className="rounded-3xl bg-gradient-to-r from-purple-900/40 via-rose-900/40 to-pink-900/40 border border-purple-500/30 p-6 flex items-center justify-between gap-4">
        <div className="space-y-1">
          <div className="flex items-center gap-2 text-rose-400 text-xs font-bold uppercase tracking-wider">
            <Sparkles className="h-4 w-4" />
            <span>Геймификация и ачивки</span>
          </div>
          <h3 className="text-lg font-bold text-zinc-100">
            Делайте заказы каждый день для сохранения Streak!
          </h3>
          <p className="text-xs text-zinc-400 max-w-md">
            Синтетическая доставка симулируется за секунды, а удовольствие от доставленного заказа остается реальным.
          </p>
        </div>

        <Link href="/orders" className="flex-shrink-0">
          <Button variant="outline" className="rounded-2xl gap-2">
            <span>Мои заказы</span>
            <ArrowRight className="h-4 w-4" />
          </Button>
        </Link>
      </div>
    </div>
  );
}
