"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import Image from "next/image";
import { User, Flame, Trophy, Sparkles, ArrowRight, Package, Lock, CheckCircle2 } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { StreakBadge } from "@/components/features/gamification/StreakBadge";
import { AchievementModal, Achievement } from "@/components/features/gamification/AchievementModal";
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

const INITIAL_ACHIEVEMENTS: Achievement[] = [
  {
    id: "first-order",
    title: "Первый шаг",
    description: "Оформили ваш самый первый синтетический заказ за 10 ₽",
    xp_reward: 50,
    is_unlocked: true,
    unlocked_at: "2026-09-15",
    progress: 1,
    max_progress: 1,
  },
  {
    id: "night-hunter",
    title: "Ночной охотник",
    description: "Заказали дофамин 3 раза в ночное время (после 23:00)",
    xp_reward: 50,
    is_unlocked: true,
    unlocked_at: "2026-09-17",
    progress: 3,
    max_progress: 3,
  },
  {
    id: "streak-master-7",
    title: "Огненный ритуал",
    description: "Удерживали стрик заказов 7 дней подряд без перерыва",
    xp_reward: 100,
    is_unlocked: true,
    unlocked_at: "2026-09-18",
    progress: 7,
    max_progress: 7,
  },
  {
    id: "snack-connoisseur",
    title: "Гурман снеков",
    description: "Попробовали все виды чипсов и закусок из каталога",
    xp_reward: 75,
    is_unlocked: true,
    unlocked_at: "2026-09-18",
    progress: 5,
    max_progress: 5,
  },
  {
    id: "streak-legend-30",
    title: "Легенда стрика",
    description: "Не прерывайте ежедневные заказы на протяжении 30 дней",
    xp_reward: 250,
    is_unlocked: false,
    progress: 12,
    max_progress: 30,
  },
  {
    id: "fast-delivery-collector",
    title: "Сверхзвуковой",
    description: "Получите 10 заказов с синтетической доставкой менее 30 секунд",
    xp_reward: 150,
    is_unlocked: false,
    progress: 6,
    max_progress: 10,
  },
  {
    id: "big-basket",
    title: "Полная корзина",
    description: "Соберите корзину из 10+ товаров и примените промокод INV-01",
    xp_reward: 100,
    is_unlocked: false,
    progress: 4,
    max_progress: 10,
  },
  {
    id: "dopamine-overload",
    title: "Дофаминовый взрыв",
    description: "Достигните 10-го уровня профиля и накопите 2500 XP",
    xp_reward: 500,
    is_unlocked: false,
    progress: 740,
    max_progress: 2500,
  },
];

export default function ProfilePage() {
  const [user, setUser] = useState<UserProfile | null>(null);
  const [stats, setStats] = useState<UserStats | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [achievements] = useState<Achievement[]>(INITIAL_ACHIEVEMENTS);
  const [selectedAchievement, setSelectedAchievement] = useState<Achievement | null>(null);

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
      <div className="container mx-auto max-w-2xl px-4 py-20 text-center space-y-4">
        <div className="animate-spin rounded-full h-10 w-10 border-2 border-amber-400 border-t-transparent mx-auto" />
        <p className="text-sm text-[#9FB3C4]">Загрузка профиля...</p>
      </div>
    );
  }

  if (!user) {
    return (
      <div className="container mx-auto max-w-md px-4 py-24 text-center space-y-6">
        <div className="flex h-20 w-20 items-center justify-center rounded-3xl bg-[#0B1622] border border-[#1E3A50] text-amber-400 mx-auto shadow-xl">
          <User className="h-10 w-10" />
        </div>
        <div className="space-y-2">
          <h2 className="text-2xl font-black text-[#F4F1E8]">Вы не вошли в аккаунт</h2>
          <p className="text-sm text-[#9FB3C4]">
            Войдите или зарегистрируйтесь, чтобы просматривать профиль, награды и стрик.
          </p>
        </div>
        <Link href="/login">
          <Button variant="reward" className="rounded-2xl px-8">
            Войти в аккаунт
          </Button>
        </Link>
      </div>
    );
  }

  const unlockedCount = achievements.filter((a) => a.is_unlocked).length;
  const currentStreak = stats?.current_streak_days || 12;
  const totalOrders = stats?.total_orders || 43;

  return (
    <div className="container mx-auto max-w-4xl px-4 sm:px-6 py-8 space-y-8">
      {/* Profile Header matching screen-05 */}
      <div className="relative overflow-hidden rounded-[28px] bg-gradient-to-br from-[#0B1622] via-[#122234] to-[#0B1622] border border-[#1E3A50] p-6 sm:p-8 shadow-xl">
        <div className="pointer-events-none absolute -right-16 -top-16 h-64 w-64 rounded-full bg-amber-500/10 blur-3xl" />
        <div className="pointer-events-none absolute -left-16 -bottom-16 h-64 w-64 rounded-full bg-teal-500/10 blur-3xl" />

        <div className="relative z-10 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-6">
          <div className="flex items-center gap-5">
            {/* Avatar with Level Ring */}
            <div className="relative flex-shrink-0">
              <div className="relative flex h-20 w-20 items-center justify-center rounded-2xl bg-gradient-to-tr from-[#1E3A50] to-[#0E1B29] border-2 border-amber-400 text-amber-400 font-black text-2xl shadow-lg shadow-amber-500/15">
                <Image
                  src="/logo-minimal-star.png"
                  alt="Level Icon"
                  width={40}
                  height={40}
                  className="object-contain"
                />
              </div>
              <span className="absolute -bottom-2 -right-2 px-2 py-0.5 rounded-full bg-amber-400 text-[#050B14] font-black text-[11px] shadow-sm">
                Ур. 7
              </span>
            </div>

            <div className="space-y-1">
              <h1 className="text-xl sm:text-2xl font-black text-[#F4F1E8] tracking-tight">
                {user.email}
              </h1>
              <p className="text-xs text-[#9FB3C4]">
                В Avari Dopamine с {new Date(user.created_at).toLocaleDateString("ru-RU")}
              </p>
              <div className="flex items-center gap-2 pt-1">
                <div className="h-2 w-32 rounded-full bg-[#0E1B29] border border-[#1E3A50] overflow-hidden">
                  <div className="h-full w-[74%] rounded-full bg-gradient-to-r from-amber-400 to-amber-300" />
                </div>
                <span className="text-[11px] font-bold text-amber-400">740 / 1000 XP</span>
              </div>
            </div>
          </div>

          <div className="flex items-center gap-3">
            {currentStreak > 0 && (
              <StreakBadge days={currentStreak} className="text-sm px-4 py-2" />
            )}
          </div>
        </div>
      </div>

      {/* 3 Gamification Stat Cards matching screen-05 */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        {/* Streak Card */}
        <Card className="border-[#1E3A50] bg-[#0B1622] hover:border-amber-400/40 transition-colors shadow-lg">
          <CardHeader className="p-5 pb-2">
            <CardTitle className="text-xs text-[#9FB3C4] font-semibold uppercase tracking-wider flex items-center justify-between">
              <span>Текущий Стрик</span>
              <Flame className="h-4 w-4 text-amber-400 fill-amber-400 animate-pulse" />
            </CardTitle>
          </CardHeader>
          <CardContent className="p-5 pt-0">
            <div className="text-3xl font-black text-amber-400">
              {currentStreak} <span className="text-sm font-bold text-[#9FB3C4]">дней</span>
            </div>
            <p className="text-[11px] text-[#5E7488] mt-1">
              Рекорд: {stats?.longest_streak_days || 18} дней подряд
            </p>
          </CardContent>
        </Card>

        {/* Rewards / Achievements Card */}
        <Card className="border-[#1E3A50] bg-[#0B1622] hover:border-teal-400/40 transition-colors shadow-lg">
          <CardHeader className="p-5 pb-2">
            <CardTitle className="text-xs text-[#9FB3C4] font-semibold uppercase tracking-wider flex items-center justify-between">
              <span>Достижения</span>
              <Trophy className="h-4 w-4 text-teal-400" />
            </CardTitle>
          </CardHeader>
          <CardContent className="p-5 pt-0">
            <div className="text-3xl font-black text-teal-400">
              {unlockedCount} <span className="text-sm font-bold text-[#9FB3C4]">/ {achievements.length}</span>
            </div>
            <p className="text-[11px] text-[#5E7488] mt-1">
              Разблокировано наград
            </p>
          </CardContent>
        </Card>

        {/* Orders / Benefit Card */}
        <Card className="border-[#1E3A50] bg-[#0B1622] hover:border-emerald-400/40 transition-colors shadow-lg">
          <CardHeader className="p-5 pb-2">
            <CardTitle className="text-xs text-[#9FB3C4] font-semibold uppercase tracking-wider flex items-center justify-between">
              <span>Всего заказов</span>
              <Package className="h-4 w-4 text-emerald-400" />
            </CardTitle>
          </CardHeader>
          <CardContent className="p-5 pt-0">
            <div className="text-3xl font-black text-[#5FD98A]">
              {totalOrders}
            </div>
            <p className="text-[11px] text-[#5E7488] mt-1">
              Сэкономлено: {formatPrice(stats?.total_saved_rub || "18450.00")}
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Achievements Section matching screen-05 */}
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <div className="space-y-1">
            <div className="flex items-center gap-2">
              <Trophy className="h-5 w-5 text-amber-400" />
              <h2 className="text-xl sm:text-2xl font-black text-[#F4F1E8] tracking-tight">
                Коллекция наград
              </h2>
            </div>
            <p className="text-xs text-[#9FB3C4]">
              Нажмите на достижение, чтобы просмотреть награду и условия
            </p>
          </div>

          <span className="text-xs font-bold text-amber-400 px-3 py-1 rounded-full bg-amber-500/10 border border-amber-500/20">
            {unlockedCount} из {achievements.length} получено
          </span>
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
          {achievements.map((item) => (
            <div
              key={item.id}
              onClick={() => setSelectedAchievement(item)}
              className={`relative flex flex-col justify-between p-5 rounded-2xl border transition-all cursor-pointer group ${
                item.is_unlocked
                  ? "bg-[#0B1622] border-[#1E3A50] hover:border-amber-400/60 hover:shadow-lg hover:shadow-amber-500/10"
                  : "bg-[#0B1622]/50 border-[#152B3D] opacity-75 hover:opacity-100 hover:border-[#1E3A50]"
              }`}
            >
              <div className="space-y-3">
                <div className="flex items-center justify-between">
                  <div
                    className={`flex h-12 w-12 items-center justify-center rounded-2xl border ${
                      item.is_unlocked
                        ? "bg-gradient-to-tr from-[#1E3A50] to-[#0E1B29] border-amber-400 text-amber-400 shadow-md shadow-amber-500/20"
                        : "bg-[#0E1B29] border-[#1E3A50] text-[#5E7488]"
                    }`}
                  >
                    {item.is_unlocked ? (
                      <Sparkles className="h-6 w-6 text-amber-400 animate-pulse" />
                    ) : (
                      <Lock className="h-5 w-5 text-[#5E7488]" />
                    )}
                  </div>

                  <span
                    className={`text-xs font-bold px-2 py-0.5 rounded-full ${
                      item.is_unlocked
                        ? "bg-amber-500/15 text-amber-400 border border-amber-500/30"
                        : "bg-[#0E1B29] text-[#5E7488] border border-[#152B3D]"
                    }`}
                  >
                    +{item.xp_reward} XP
                  </span>
                </div>

                <div className="space-y-1">
                  <h4 className="font-bold text-sm text-[#F4F1E8] group-hover:text-amber-400 transition-colors">
                    {item.title}
                  </h4>
                  <p className="text-xs text-[#9FB3C4] line-clamp-2 leading-relaxed">
                    {item.description}
                  </p>
                </div>
              </div>

              <div className="pt-4 mt-4 border-t border-[#152B3D]">
                {item.is_unlocked ? (
                  <div className="flex items-center gap-1.5 text-xs text-[#5FD98A] font-semibold">
                    <CheckCircle2 className="h-3.5 w-3.5" />
                    <span>Получено</span>
                  </div>
                ) : (
                  <div className="space-y-1.5">
                    <div className="flex justify-between text-[11px] text-[#5E7488]">
                      <span>Прогресс</span>
                      <span>
                        {item.progress || 0} / {item.max_progress || 1}
                      </span>
                    </div>
                    <div className="h-1.5 w-full rounded-full bg-[#0E1B29] overflow-hidden">
                      <div
                        className="h-full rounded-full bg-teal-400/80 transition-all"
                        style={{
                          width: `${Math.min(
                            100,
                            (((item.progress || 0) / (item.max_progress || 1)) * 100)
                          )}%`,
                        }}
                      />
                    </div>
                  </div>
                )}
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Interactive Achievement Modal */}
      <AchievementModal
        achievement={selectedAchievement}
        onClose={() => setSelectedAchievement(null)}
      />

      {/* Orders link banner */}
      <div className="rounded-3xl bg-gradient-to-r from-[#122234] via-[#0B1622] to-[#122234] border border-[#1E3A50] p-6 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
        <div className="space-y-1">
          <div className="flex items-center gap-2 text-amber-400 text-xs font-bold uppercase tracking-wider">
            <Sparkles className="h-4 w-4" />
            <span>История синтетических покупок</span>
          </div>
          <h3 className="text-lg font-bold text-[#F4F1E8]">
            Просмотрите статус и архив ваших заказов
          </h3>
          <p className="text-xs text-[#9FB3C4] max-w-md">
            Все заказы выполняются по правилу INV-01 ровно за 10 ₽ с применением скидки Avari Dopamine.
          </p>
        </div>

        <Link href="/orders" className="flex-shrink-0">
          <Button variant="outline" className="rounded-2xl gap-2 text-[#F4F1E8] border-[#1E3A50] hover:bg-[#1E3A50]/50">
            <span>Мои заказы</span>
            <ArrowRight className="h-4 w-4" />
          </Button>
        </Link>
      </div>
    </div>
  );
}
