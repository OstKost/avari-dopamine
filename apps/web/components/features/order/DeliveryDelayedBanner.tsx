"use client";

import { AlertTriangle, Clock } from "lucide-react";

interface DeliveryDelayedBannerProps {
  reason?: string;
}

export function DeliveryDelayedBanner({ reason }: DeliveryDelayedBannerProps) {
  return (
    <div className="rounded-2xl border border-amber-200 bg-amber-50/80 p-4 dark:border-amber-900/40 dark:bg-amber-950/20 text-amber-800 dark:text-amber-200 flex items-start gap-3.5 shadow-sm transition-all animate-fadeIn">
      <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-amber-100 dark:bg-amber-900/60 text-amber-600 dark:text-amber-400 flex-shrink-0 mt-0.5">
        <AlertTriangle className="h-5 w-5" />
      </div>
      <div className="space-y-1 text-sm">
        <div className="flex items-center gap-2">
          <h4 className="font-bold text-base text-amber-900 dark:text-amber-100">
            Курьер немного задерживается
          </h4>
          <span className="inline-flex items-center gap-1 rounded-full bg-amber-200/60 px-2 py-0.5 text-xs font-semibold text-amber-900 dark:bg-amber-800/60 dark:text-amber-100">
            <Clock className="h-3 w-3" />
            +30-60с
          </span>
        </div>
        <p className="text-xs text-amber-700 dark:text-amber-300">
          {reason || "На маршруте возникла небольшая задержка, но заказ уже скоро будет в пункте выдачи!"}
        </p>
      </div>
    </div>
  );
}
