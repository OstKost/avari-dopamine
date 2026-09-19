"use client";

import { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { ArrowLeft, MapPin, RefreshCw, Sparkles, Tag, Package } from "lucide-react";
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

  // Calculate items raw subtotal
  const rawSubtotal = initialOrder.items.reduce((sum, item) => {
    const itemSub = parseFloat(item.subtotal_rub) || (parseFloat(item.price_rub) * item.quantity);
    return sum + (isNaN(itemSub) ? 0 : itemSub);
  }, 0);

  const discountAmount = Math.max(0, rawSubtotal - 10.00);

  return (
    <div className="container mx-auto max-w-3xl px-4 sm:px-6 py-8 space-y-8">
      <div className="flex items-center justify-between">
        <Link
          href="/orders"
          className="inline-flex items-center gap-2 text-sm text-[#9FB3C4] hover:text-[#F4F1E8] transition-colors"
        >
          <ArrowLeft className="h-4 w-4" />
          <span>Ко всем заказам</span>
        </Link>

        {isDelivered && (
          <Button
            variant="outline"
            size="sm"
            className="rounded-xl gap-2 text-xs border-[#1E3A50] text-[#F4F1E8] hover:bg-[#1E3A50]"
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
          <div className="flex items-center gap-2 mb-1.5">
            <span className="px-2.5 py-0.5 rounded-full text-xs font-bold bg-amber-400/15 border border-amber-400/30 text-amber-400">
              Оплачено: 10.00 ₽
            </span>
            <span className="px-2.5 py-0.5 rounded-full text-xs font-bold bg-teal-500/15 border border-teal-500/30 text-teal-300 flex items-center gap-1">
              <Sparkles className="h-3 w-3 text-teal-400" />
              +15 XP
            </span>
          </div>
          <h1 className="text-3xl font-black text-[#F4F1E8] tracking-tight">
            Заказ #{initialOrder.id.slice(0, 8)}
          </h1>
          <p className="text-xs text-[#5E7488] mt-1">
            Оформлен {new Date(initialOrder.created_at).toLocaleString("ru-RU")}
          </p>
        </div>

        <Badge
          variant={isDelivered ? "success" : "default"}
          className="text-sm px-4 py-1.5 self-start sm:self-auto uppercase tracking-wide"
        >
          {isDelivered ? "Доставлен в ПВЗ" : `Статус: ${currentStatus}`}
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
      <Card className="border-[#1E3A50] bg-[#0B1622] shadow-lg">
        <CardContent className="p-6 flex items-start gap-4">
          <div className="flex h-12 w-12 items-center justify-center rounded-2xl bg-teal-500/15 border border-teal-500/30 text-teal-400 flex-shrink-0">
            <MapPin className="h-6 w-6" />
          </div>
          <div className="space-y-1">
            <h4 className="font-bold text-base text-[#F4F1E8]">
              {initialOrder.pickup_point.name}
            </h4>
            <p className="text-xs text-[#9FB3C4]">
              Пункт выдачи заказов. Дистанция ~{Math.round(initialOrder.pickup_point.distance_meters)}м от вас.
            </p>
          </div>
        </CardContent>
      </Card>

      {/* Order Items & Breakdown */}
      <Card className="border-[#1E3A50] bg-[#0B1622] shadow-lg">
        <CardHeader className="pb-3 border-b border-[#1E3A50]">
          <CardTitle className="text-base text-[#F4F1E8] flex items-center gap-2">
            <Package className="h-4 w-4 text-amber-400" />
            <span>Состав заказа</span>
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4 pt-4">
          {initialOrder.items.map((item) => (
            <div
              key={item.product_id}
              className="flex items-center justify-between py-2 border-b border-[#152B3D] last:border-none"
            >
              <div className="space-y-0.5">
                <h5 className="font-bold text-sm text-[#F4F1E8]">
                  {item.name}
                </h5>
                <span className="text-xs text-[#5E7488]">
                  {item.quantity} шт. × {formatPrice(item.price_rub)}
                </span>
              </div>
              <span className="font-semibold text-sm text-[#9FB3C4]">
                {formatPrice(item.subtotal_rub || (parseFloat(item.price_rub) * item.quantity).toFixed(2))}
              </span>
            </div>
          ))}

          {/* Pricing & Promo Code breakdown */}
          <div className="pt-4 border-t border-[#1E3A50] space-y-2 text-sm">
            <div className="flex justify-between text-[#9FB3C4]">
              <span>Сумма товаров</span>
              <span>{formatPrice(rawSubtotal.toFixed(2))}</span>
            </div>

            <div className="flex justify-between text-amber-400 font-semibold items-center">
              <span className="flex items-center gap-1.5">
                <Tag className="h-3.5 w-3.5" />
                <span>Скидка по промокоду</span>
              </span>
              <span>-{formatPrice(discountAmount.toFixed(2))}</span>
            </div>

            <div className="pt-3 border-t border-[#1E3A50] flex justify-between items-center text-lg font-black text-[#F4F1E8]">
              <span>Итого оплачено</span>
              <span className="text-2xl font-black text-amber-400">
                {formatPrice(initialOrder.total_amount_rub || "10.00")}
              </span>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
