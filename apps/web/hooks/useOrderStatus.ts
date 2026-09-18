"use client";

import { useEffect, useState, useRef, useCallback } from "react";
import { apiFetch } from "@/lib/api/client";

export interface CourierInfo {
  name: string;
  rating: number;
}

export interface OrderLiveState {
  orderId: string;
  status: string;
  courier?: CourierInfo;
  estimatedCompletionAt?: string;
  reason?: string;
  isLive: boolean;
  error?: string;
}

export function useOrderStatus(
  initialOrderId: string,
  initialStatus: string,
  initialCourier?: CourierInfo,
  initialEstimatedCompletionAt?: string,
) {
  const [state, setState] = useState<OrderLiveState>({
    orderId: initialOrderId,
    status: initialStatus,
    courier: initialCourier,
    estimatedCompletionAt: initialEstimatedCompletionAt,
    isLive: false,
  });

  const eventSourceRef = useRef<EventSource | null>(null);
  const pollingIntervalRef = useRef<NodeJS.Timeout | null>(null);

  const fallbackPoll = useCallback(async () => {
    try {
      const order = await apiFetch<{
        id: string;
        status: string;
        created_at: string;
      }>(`/orders/${initialOrderId}`);

      setState((prev) => ({
        ...prev,
        status: order.status,
        isLive: false,
      }));
    } catch {
      // Ignore polling errors
    }
  }, [initialOrderId]);

  useEffect(() => {
    // If order is terminal (delivered or cancelled), no need to stream
    if (state.status === "delivered" || state.status === "cancelled" || state.status === "payment_failed") {
      return;
    }

    const apiUrl = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
    const sseUrl = `${apiUrl}/orders/${initialOrderId}/events`;

    try {
      const es = new EventSource(sseUrl, { withCredentials: true });
      eventSourceRef.current = es;

      es.onopen = () => {
        setState((prev) => ({ ...prev, isLive: true, error: undefined }));
        if (pollingIntervalRef.current) {
          clearInterval(pollingIntervalRef.current);
          pollingIntervalRef.current = null;
        }
      };

      es.addEventListener("snapshot", (e: MessageEvent) => {
        try {
          const data = JSON.parse(e.data);
          setState((prev) => ({
            ...prev,
            status: data.status || prev.status,
            courier: data.courier || prev.courier,
            estimatedCompletionAt: data.estimated_completion_at || prev.estimatedCompletionAt,
            isLive: true,
          }));
        } catch {
          // ignore parse error
        }
      });

      es.addEventListener("status_changed", (e: MessageEvent) => {
        try {
          const data = JSON.parse(e.data);
          setState((prev) => ({
            ...prev,
            status: data.status || prev.status,
            courier: data.courier || prev.courier,
            reason: data.reason || prev.reason,
            estimatedCompletionAt: data.estimated_completion_at || prev.estimatedCompletionAt,
            isLive: true,
          }));
        } catch {
          // ignore parse error
        }
      });

      es.onerror = () => {
        es.close();
        setState((prev) => ({ ...prev, isLive: false }));

        // Graceful fallback to short-polling every 5 seconds (EPIC-11)
        if (!pollingIntervalRef.current) {
          pollingIntervalRef.current = setInterval(fallbackPoll, 5000);
        }
      };
    } catch {
      // If EventSource is not supported or fails
      if (!pollingIntervalRef.current) {
        pollingIntervalRef.current = setInterval(fallbackPoll, 5000);
      }
    }

    return () => {
      if (eventSourceRef.current) {
        eventSourceRef.current.close();
        eventSourceRef.current = null;
      }
      if (pollingIntervalRef.current) {
        clearInterval(pollingIntervalRef.current);
        pollingIntervalRef.current = null;
      }
    };
  }, [initialOrderId, state.status, fallbackPoll]);

  return state;
}
