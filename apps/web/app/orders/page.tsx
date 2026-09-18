import Link from "next/link";
import { Package, ArrowRight, Clock, MapPin } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
import { apiFetch } from "@/lib/api/client";
import { formatPrice } from "@/lib/utils";

export const dynamic = "force-dynamic";

interface OrderItem {
  product_id: string;
  name: string;
  category_name?: string;
  price_rub: string;
  quantity: number;
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
}

export default async function OrdersPage() {
  let orders: Order[] = [];
  try {
    const res = await apiFetch<{ orders: Order[] }>("/orders", { cache: "no-store" });
    orders = res.orders || [];
  } catch (err) {
    console.error("Error fetching orders:", err);
  }

  const getStatusBadge = (status: string) => {
    switch (status) {
      case "paid":
      case "assembling":
      case "courier_assigned":
      case "in_transit":
        return <Badge variant="default">{status}</Badge>;
      case "delivered":
        return <Badge variant="success">Доставлен</Badge>;
      case "payment_failed":
      case "cancelled":
        return <Badge variant="destructive">{status}</Badge>;
      default:
        return <Badge variant="secondary">{status}</Badge>;
    }
  };

  return (
    <div className="container mx-auto max-w-4xl px-4 sm:px-6 py-8 space-y-8">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-black text-zinc-900 dark:text-zinc-50 tracking-tight">
          Мои заказы ({orders.length})
        </h1>
        <Link href="/catalog" className="text-sm font-semibold text-rose-500 hover:text-rose-600">
          В каталог
        </Link>
      </div>

      {orders.length === 0 ? (
        <div className="flex min-h-[40vh] flex-col items-center justify-center text-center p-8 border border-dashed border-zinc-300 dark:border-zinc-800 rounded-3xl space-y-4">
          <div className="flex h-16 w-16 items-center justify-center rounded-2xl bg-zinc-100 text-zinc-400 dark:bg-zinc-800">
            <Package className="h-8 w-8" />
          </div>
          <h3 className="text-lg font-bold text-zinc-900 dark:text-zinc-100">
            У вас пока нет заказов
          </h3>
          <p className="text-sm text-zinc-500 max-w-sm">
            Оформите свой первый синтетический заказ всего за 10 рублей.
          </p>
          <Link href="/catalog">
            <Badge variant="default" className="px-4 py-2 cursor-pointer text-sm">
              Перейти к покупкам
            </Badge>
          </Link>
        </div>
      ) : (
        <div className="space-y-4">
          {orders.map((order) => (
            <Link key={order.id} href={`/orders/${order.id}`} className="block group">
              <Card className="border-zinc-200/80 dark:border-zinc-800 hover:border-rose-200 dark:hover:border-zinc-700 hover:shadow-lg transition-all">
                <CardContent className="p-6 space-y-4">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-3">
                      <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-rose-50 text-rose-500 dark:bg-zinc-800">
                        <Package className="h-5 w-5" />
                      </div>
                      <div>
                        <h4 className="font-bold text-sm text-zinc-900 dark:text-zinc-100">
                          Заказ #{order.id.slice(0, 8)}
                        </h4>
                        <span className="text-xs text-zinc-400 flex items-center gap-1 mt-0.5">
                          <Clock className="h-3 w-3" />
                          {new Date(order.created_at).toLocaleString("ru-RU")}
                        </span>
                      </div>
                    </div>

                    {getStatusBadge(order.status)}
                  </div>

                  <div className="flex items-center justify-between pt-2 border-t border-zinc-100 dark:border-zinc-800 text-sm">
                    <div className="flex items-center gap-2 text-zinc-500 text-xs">
                      <MapPin className="h-3.5 w-3.5 text-rose-500" />
                      <span>{order.pickup_point?.name}</span>
                    </div>

                    <div className="flex items-center gap-3">
                      <span className="font-black text-base text-zinc-900 dark:text-zinc-100">
                        {formatPrice(order.total_amount_rub)}
                      </span>
                      <ArrowRight className="h-4 w-4 text-zinc-400 group-hover:text-rose-500 group-hover:translate-x-0.5 transition-all" />
                    </div>
                  </div>
                </CardContent>
              </Card>
            </Link>
          ))}
        </div>
      )}
    </div>
  );
}
