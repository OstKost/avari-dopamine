import { notFound } from "next/navigation";
import { apiFetch } from "@/lib/api/client";
import { OrderTrackingView } from "@/components/features/order/OrderTrackingView";

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

  return <OrderTrackingView initialOrder={order} />;
}
