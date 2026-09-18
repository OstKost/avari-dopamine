"use client";

import { AlertTriangle, Clock } from "lucide-react";

interface DeliveryDelayedBannerProps {
  reason?: string;
}

export function DeliveryDelayedBanner({ reason }: DeliveryDelayedBannerProps) {
  return (
    <div className="rounded-2xl border border-amber-500/30 bg-[#122234] p-4 text-[#F4F1E8] flex items-start gap-3.5 shadow-lg shadow-amber-500/5 transition-all animate-fadeIn">
      <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-amber-500/20 text-amber-400 border border-amber-500/30 flex-shrink-0 mt-0.5">
        <AlertTriangle className="h-5 w-5" />
      </div>
      <div className="space-y-1 text-sm">
        <div className="flex items-center gap-2">
          <h4 className="font-bold text-base text-amber-400">
            Курьер немного задерживается
          </h4>
          <span className="inline-flex items-center gap-1 rounded-full bg-amber-400/20 border border-amber-400/30 px-2 py-0.5 text-xs font-semibold text-amber-300">
            <Clock className="h-3 w-3" />
            +30-60с
          </span>
        </div>
        <p className="text-xs text-[#9FB3C4]">
          {reason || "На синтетическом маршруте возникла небольшая задержка, но заказ уже скоро будет в пункте выдачи!"}
        </p>
      </div>
    </div>
  );
}
