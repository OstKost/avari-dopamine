"use client";

import { CheckCircle2, Clock, Package, Truck, UserCheck, Star } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { ProgressBar } from "@/components/ui/progress-bar";
import { CourierInfo } from "@/hooks/useOrderStatus";

interface DeliveryProgressProps {
  status: string;
  courier?: CourierInfo;
  estimatedCompletionAt?: string;
  isLive?: boolean;
}

export function DeliveryProgress({
  status,
  courier,
  estimatedCompletionAt,
  isLive,
}: DeliveryProgressProps) {
  const getProgress = (st: string) => {
    switch (st) {
      case "created":
      case "payment_pending":
        return 15;
      case "paid":
        return 30;
      case "assembling":
        return 45;
      case "courier_assigned":
        return 65;
      case "in_transit":
      case "delivery_delayed":
        return 85;
      case "delivered":
        return 100;
      default:
        return 0;
    }
  };

  const getStatusText = (st: string) => {
    switch (st) {
      case "created":
      case "payment_pending":
        return "Ожидает оплаты";
      case "paid":
        return "Оплачен — отправлен в сборку";
      case "assembling":
        return "Собираем заказ на складе";
      case "courier_assigned":
        return "Курьер назначен и спешит на склад";
      case "in_transit":
        return "Курьер везёт заказ в пункт выдачи";
      case "delivery_delayed":
        return "Курьер немного задерживается";
      case "delivered":
        return "Заказ прибыл в ПВЗ и готов к выдаче";
      case "cancelled":
        return "Заказ отменен";
      default:
        return st;
    }
  };

  const progress = getProgress(status);

  return (
    <Card className="border-zinc-200/80 dark:border-zinc-800 shadow-md">
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <CardTitle className="text-base">Статус доставки</CardTitle>
            {isLive && (
              <span className="flex items-center gap-1.5 rounded-full bg-emerald-50 px-2 py-0.5 text-[11px] font-semibold text-emerald-600 dark:bg-emerald-950/50 dark:text-emerald-400 border border-emerald-200 dark:border-emerald-800">
                <span className="h-1.5 w-1.5 rounded-full bg-emerald-500 animate-pulse" />
                Live SSE
              </span>
            )}
          </div>
          <span className="text-xs text-rose-500 font-bold">{progress}%</span>
        </div>
        <p className="text-sm font-semibold text-zinc-700 dark:text-zinc-300 mt-1">
          {getStatusText(status)}
        </p>
      </CardHeader>
      <CardContent className="space-y-6">
        <ProgressBar progress={progress} />

        {/* Timeline Stages */}
        <div className="grid grid-cols-4 gap-2 text-center text-xs">
          <div className={`space-y-1 ${progress >= 30 ? "text-rose-500 font-bold" : "text-zinc-400"}`}>
            <Clock className="h-4 w-4 mx-auto" />
            <span>Оплачен</span>
          </div>
          <div className={`space-y-1 ${progress >= 45 ? "text-rose-500 font-bold" : "text-zinc-400"}`}>
            <Package className="h-4 w-4 mx-auto" />
            <span>Сборка</span>
          </div>
          <div className={`space-y-1 ${progress >= 85 ? "text-rose-500 font-bold" : "text-zinc-400"}`}>
            <Truck className="h-4 w-4 mx-auto" />
            <span>В пути</span>
          </div>
          <div className={`space-y-1 ${progress >= 100 ? "text-emerald-500 font-bold" : "text-zinc-400"}`}>
            <CheckCircle2 className="h-4 w-4 mx-auto" />
            <span>В ПВЗ</span>
          </div>
        </div>

        {/* Courier Info Card */}
        {courier && (
          <div className="rounded-2xl border border-zinc-200/80 dark:border-zinc-800 bg-zinc-50/50 dark:bg-zinc-900/50 p-4 flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="flex h-11 w-11 items-center justify-center rounded-xl bg-purple-50 text-purple-600 dark:bg-purple-950/40 dark:text-purple-300">
                <UserCheck className="h-6 w-6" />
              </div>
              <div>
                <h5 className="font-bold text-sm text-zinc-900 dark:text-zinc-100">
                  {courier.name}
                </h5>
                <div className="flex items-center gap-1 text-xs text-amber-500 mt-0.5">
                  <Star className="h-3.5 w-3.5 fill-amber-400 text-amber-400" />
                  <span className="font-bold">{courier.rating.toFixed(2)}</span>
                  <span className="text-zinc-400 text-[11px] ml-1">Курьер Dopamine</span>
                </div>
              </div>
            </div>

            {estimatedCompletionAt && status !== "delivered" && (
              <div className="text-right">
                <span className="text-[11px] text-zinc-400 block">Ожидается к</span>
                <span className="text-xs font-bold text-zinc-800 dark:text-zinc-200">
                  {new Date(estimatedCompletionAt).toLocaleTimeString("ru-RU", { hour: "2-digit", minute: "2-digit" })}
                </span>
              </div>
            )}
          </div>
        )}
      </CardContent>
    </Card>
  );
}
