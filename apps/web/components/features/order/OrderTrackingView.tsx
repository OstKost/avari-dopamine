"use client";

import { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { ArrowLeft, MapPin, RefreshCw } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { useOrderStatus, CourierInfo } from "@/hooks/useOrderStatus";
import { DeliveryProgress } from "./DeliveryProgress";
import { DeliveryDelayedBanner } from "./DeliveryDelayedBanner";
import { CelebrationConfetti } from "./CelebrationConfetti";
import { formatPrice } from "@/lib/utils";
import { apiFetch } from "@/lib/api/client";

interface OrderItem {
  product_id: string;
  name: string;
  category_name?: string;
  price_rub: string;
  quantity: number;
  subtotal_rub: string;
}

interface OrderData {
  id: string;
  user_id: string;
  status: string;
  total_amount_rub: string;
  pickup_point: {
    id?: string;
    name: string;
    distance_meters: number;
  };
  items: OrderItem[];
  created_at: string;
  updated_at: string;
  courier?: CourierInfo;
  estimated_completion_at?: string;
}

interface OrderTrackingViewProps {
  initialOrder: OrderData;
}

export function OrderTrackingView({ initialOrder }: OrderTrackingViewProps) {
  const router = useRouter();
  const [isRepeating, setIsRepeating] = useState(false);

  const liveState = useOrderStatus(
    initialOrder.id,
    initialOrder.status,
    initialOrder.courier,
    initialOrder.estimated_completion_at,
  );

  const handleRepeatOrder = async () => {
    setIsRepeating(true);
    try {
      // Добавляем товары заказа обратно в корзину
      for (const item of initialOrder.items) {
        await apiFetch("/cart/items", {
          method: "POST",
          body: { product_id: item.product_id, quantity: item.quantity },
        });
      }
      router.push("/cart");
    } catch {
      // Ignore repeat errors
    } finally {
      setIsRepeating(false);
    }
  };

  const currentStatus = liveState.status;
  const isDelivered = currentStatus === "delivered";
  const isDelayed = currentStatus === "delivery_delayed";

  return (
    <div className="container mx-auto max-w-3xl px-4 sm:px-6 py-8 space-y-8">
      <div className="flex items-center justify-between">
        <Link
          href="/orders"
          className="inline-flex items-center gap-2 text-sm text-zinc-500 hover:text-zinc-900 dark:hover:text-zinc-100 transition-colors"
        >
          <ArrowLeft className="h-4 w-4" />
          <span>Ко всем заказам</span>
        </Link>

        {isDelivered && (
          <Button
            variant="outline"
            size="sm"
            className="rounded-xl gap-2 text-xs"
            onClick={handleRepeatOrder}
            isLoading={isRepeating}
          >
            <RefreshCw className="h-3.5 w-3.5" />
            <span>Повторить заказ</span>
          </Button>
        )}
      </div>

      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <div className="flex items-center gap-2 mb-1">
            <Badge variant="dopamine">INV-01: 10.00 ₽</Badge>
          </div>
          <h1 className="text-3xl font-black text-zinc-900 dark:text-zinc-50 tracking-tight">
            Заказ #{initialOrder.id.slice(0, 8)}
          </h1>
          <p className="text-xs text-zinc-400 mt-1">
            Оформлен {new Date(initialOrder.created_at).toLocaleString("ru-RU")}
          </p>
        </div>

        <Badge
          variant={isDelivered ? "success" : "default"}
          className="text-sm px-4 py-1.5 self-start sm:self-auto"
        >
          Статус: {currentStatus}
        </Badge>
      </div>

      {/* Celebration Confetti when Delivered */}
      <CelebrationConfetti show={isDelivered} orderNumber={initialOrder.id} />

      {/* Delayed Banner */}
      {isDelayed && <DeliveryDelayedBanner reason={liveState.reason} />}

      {/* Real-time Tracking & Progress */}
      <DeliveryProgress
        status={currentStatus}
        courier={liveState.courier}
        estimatedCompletionAt={liveState.estimatedCompletionAt}
        isLive={liveState.isLive}
      />

      {/* Pickup Point Details */}
      <Card className="border-zinc-200/80 dark:border-zinc-800">
        <CardContent className="p-6 flex items-start gap-4">
          <div className="flex h-12 w-12 items-center justify-center rounded-2xl bg-rose-50 text-rose-500 dark:bg-zinc-800 flex-shrink-0">
            <MapPin className="h-6 w-6" />
          </div>
          <div className="space-y-1">
            <h4 className="font-bold text-base text-zinc-900 dark:text-zinc-100">
              {initialOrder.pickup_point.name}
            </h4>
            <p className="text-xs text-zinc-500">
              Синтетический пункт выдачи заказов. Дистанция ~{Math.round(initialOrder.pickup_point.distance_meters)}м от вас.
            </p>
          </div>
        </CardContent>
      </Card>

      {/* Order Items */}
      <Card className="border-zinc-200/80 dark:border-zinc-800">
        <CardHeader>
          <CardTitle className="text-base">Состав заказа</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {initialOrder.items.map((item) => (
            <div
              key={item.product_id}
              className="flex items-center justify-between py-2 border-b border-zinc-100 dark:border-zinc-800 last:border-none"
            >
              <div className="space-y-0.5">
                <h5 className="font-bold text-sm text-zinc-900 dark:text-zinc-100">
                  {item.name}
                </h5>
                <span className="text-xs text-zinc-400">
                  {item.quantity} шт. × {formatPrice(item.price_rub)}
                </span>
              </div>
              <span className="font-semibold text-sm text-zinc-700 dark:text-zinc-300">
                {formatPrice(item.subtotal_rub)}
              </span>
            </div>
          ))}

          <div className="pt-4 border-t border-zinc-200 dark:border-zinc-800 flex justify-between items-center text-lg font-black text-zinc-900 dark:text-zinc-100">
            <span>Итого оплачено</span>
            <span className="text-rose-500">{formatPrice(initialOrder.total_amount_rub)}</span>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
