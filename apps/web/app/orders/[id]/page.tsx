import Link from "next/link";
import { notFound } from "next/navigation";
import { ArrowLeft, MapPin, CheckCircle2, Clock, Truck, Home } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { ProgressBar } from "@/components/ui/progress-bar";
import { apiFetch } from "@/lib/api/client";
import { formatPrice } from "@/lib/utils";

export const dynamic = "force-dynamic";

interface OrderItem {
  product_id: string;
  name: string;
  category_name?: string;
  price_rub: string;
  quantity: number;
  subtotal_rub: string;
}

interface Order {
  id: string;
  user_id: string;
  status: string;
  total_amount_rub: string;
  pickup_point: {
    name: string;
    distance_meters: number;
  };
  items: OrderItem[];
  created_at: string;
  updated_at: string;
}

export default async function OrderDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;

  let order: Order | null = null;
  try {
    order = await apiFetch<Order>(`/orders/${id}`, { cache: "no-store" });
  } catch {
    // handled below
  }

  if (!order) {
    notFound();
  }

  const getProgress = (status: string) => {
    switch (status) {
      case "created":
      case "payment_pending":
        return 15;
      case "paid":
        return 30;
      case "assembling":
        return 50;
      case "courier_assigned":
        return 70;
      case "in_transit":
        return 85;
      case "delivered":
        return 100;
      default:
        return 0;
    }
  };

  const progress = getProgress(order.status);

  return (
    <div className="container mx-auto max-w-3xl px-4 sm:px-6 py-8 space-y-8">
      <Link href="/orders" className="inline-flex items-center gap-2 text-sm text-zinc-500 hover:text-zinc-900 dark:hover:text-zinc-100 transition-colors">
        <ArrowLeft className="h-4 w-4" />
        <span>Ко всем заказам</span>
      </Link>

      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <div className="flex items-center gap-2 mb-1">
            <Badge variant="dopamine">INV-01: 10.00 ₽</Badge>
          </div>
          <h1 className="text-3xl font-black text-zinc-900 dark:text-zinc-50 tracking-tight">
            Заказ #{order.id.slice(0, 8)}
          </h1>
          <p className="text-xs text-zinc-400 mt-1">
            Оформлен {new Date(order.created_at).toLocaleString("ru-RU")}
          </p>
        </div>

        <Badge variant={order.status === "delivered" ? "success" : "default"} className="text-sm px-4 py-1.5 self-start sm:self-auto">
          Статус: {order.status}
        </Badge>
      </div>

      {/* Progress & Tracking */}
      <Card className="border-zinc-200/80 dark:border-zinc-800 shadow-md">
        <CardHeader className="pb-3">
          <CardTitle className="text-base flex items-center justify-between">
            <span>Прогресс доставки</span>
            <span className="text-xs text-rose-500 font-bold">{progress}%</span>
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-6">
          <ProgressBar progress={progress} />

          <div className="grid grid-cols-4 gap-2 text-center text-xs">
            <div className={`space-y-1 ${progress >= 15 ? "text-rose-500 font-bold" : "text-zinc-400"}`}>
              <Clock className="h-4 w-4 mx-auto" />
              <span>Создан</span>
            </div>
            <div className={`space-y-1 ${progress >= 30 ? "text-rose-500 font-bold" : "text-zinc-400"}`}>
              <CheckCircle2 className="h-4 w-4 mx-auto" />
              <span>Оплачен</span>
            </div>
            <div className={`space-y-1 ${progress >= 85 ? "text-rose-500 font-bold" : "text-zinc-400"}`}>
              <Truck className="h-4 w-4 mx-auto" />
              <span>В пути</span>
            </div>
            <div className={`space-y-1 ${progress >= 100 ? "text-emerald-500 font-bold" : "text-zinc-400"}`}>
              <Home className="h-4 w-4 mx-auto" />
              <span>Доставлен</span>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Pickup Point Details */}
      <Card className="border-zinc-200/80 dark:border-zinc-800">
        <CardContent className="p-6 flex items-start gap-4">
          <div className="flex h-12 w-12 items-center justify-center rounded-2xl bg-rose-50 text-rose-500 dark:bg-zinc-800 flex-shrink-0">
            <MapPin className="h-6 w-6" />
          </div>
          <div className="space-y-1">
            <h4 className="font-bold text-base text-zinc-900 dark:text-zinc-100">
              {order.pickup_point.name}
            </h4>
            <p className="text-xs text-zinc-500">
              Синтетический пункт выдачи заказов. Дистанция ~{Math.round(order.pickup_point.distance_meters)}м от вас.
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
          {order.items.map((item) => (
            <div key={item.product_id} className="flex items-center justify-between py-2 border-b border-zinc-100 dark:border-zinc-800 last:border-none">
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
            <span className="text-rose-500">{formatPrice(order.total_amount_rub)}</span>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
