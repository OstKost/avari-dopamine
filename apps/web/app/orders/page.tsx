import Link from "next/link";
import { Package, ArrowRight, Clock, MapPin, Sparkles } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
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
        return (
          <span className="px-3 py-1 rounded-full text-xs font-bold bg-teal-500/15 border border-teal-500/30 text-teal-300">
            {status === "in_transit" ? "В пути" : status === "assembling" ? "Собирается" : status}
          </span>
        );
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
        <div className="space-y-1">
          <div className="flex items-center gap-2">
            <span className="px-2.5 py-0.5 rounded-full text-xs font-bold bg-amber-400/15 border border-amber-400/30 text-amber-400">
              Фикс-цена: 10 ₽
            </span>
          </div>
          <h1 className="text-3xl font-black text-[#F4F1E8] tracking-tight">
            Мои заказы ({orders.length})
          </h1>
        </div>
        <Link href="/catalog">
          <Button variant="outline" size="sm" className="rounded-xl text-xs border-[#1E3A50] text-[#F4F1E8] hover:bg-[#1E3A50]">
            В каталог
          </Button>
        </Link>
      </div>

      {orders.length === 0 ? (
        <div className="flex min-h-[45vh] flex-col items-center justify-center text-center p-8 border border-dashed border-[#1E3A50] bg-[#0B1622]/50 rounded-3xl space-y-4">
          <div className="flex h-16 w-16 items-center justify-center rounded-2xl bg-[#0E1B29] border border-[#1E3A50] text-amber-400">
            <Package className="h-8 w-8" />
          </div>
          <div className="space-y-1">
            <h3 className="text-lg font-bold text-[#F4F1E8]">
              У вас пока нет заказов
            </h3>
            <p className="text-sm text-[#9FB3C4] max-w-sm">
              Оформите свой первый заказ в каталоге всего за 10 рублей с применением промокода.
            </p>
          </div>
          <Link href="/catalog">
            <Button variant="reward" className="px-6 py-2.5 rounded-2xl text-sm font-bold shadow-lg shadow-amber-500/20">
              <Sparkles className="h-4 w-4 mr-2" />
              <span>Перейти к покупкам</span>
            </Button>
          </Link>
        </div>
      ) : (
        <div className="space-y-4">
          {orders.map((order) => (
            <Link key={order.id} href={`/orders/${order.id}`} className="block group">
              <Card className="border-[#1E3A50] bg-[#0B1622] hover:border-amber-400/50 hover:shadow-xl hover:shadow-amber-500/5 transition-all">
                <CardContent className="p-6 space-y-4">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-3">
                      <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-amber-500/10 border border-amber-500/20 text-amber-400">
                        <Package className="h-5 w-5" />
                      </div>
                      <div>
                        <h4 className="font-bold text-sm text-[#F4F1E8] group-hover:text-amber-400 transition-colors">
                          Заказ #{order.id.slice(0, 8)}
                        </h4>
                        <span className="text-xs text-[#5E7488] flex items-center gap-1 mt-0.5">
                          <Clock className="h-3 w-3" />
                          {new Date(order.created_at).toLocaleString("ru-RU")}
                        </span>
                      </div>
                    </div>

                    {getStatusBadge(order.status)}
                  </div>

                  <div className="flex items-center justify-between pt-3 border-t border-[#152B3D] text-sm">
                    <div className="flex items-center gap-2 text-[#9FB3C4] text-xs">
                      <MapPin className="h-3.5 w-3.5 text-teal-400" />
                      <span>{order.pickup_point?.name}</span>
                    </div>

                    <div className="flex items-center gap-3">
                      <span className="font-black text-base text-amber-400">
                        {formatPrice(order.total_amount_rub || "10.00")}
                      </span>
                      <ArrowRight className="h-4 w-4 text-[#5E7488] group-hover:text-amber-400 group-hover:translate-x-1 transition-all" />
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
