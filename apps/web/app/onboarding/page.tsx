"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { MapPin, Navigation, Sparkles, Check, ArrowRight } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { apiFetch } from "@/lib/api/client";

interface PickupPoint {
  id: string;
  name: string;
  latitude: number;
  longitude: number;
  distance_meters: number;
}

export default function OnboardingPage() {
  const router = useRouter();
  const [city, setCity] = useState("Ростов-на-Дону");
  const [isLoadingGeo, setIsLoadingGeo] = useState(false);
  const [isGenerating, setIsGenerating] = useState(false);
  const [points, setPoints] = useState<PickupPoint[]>([]);
  const [selectedPointId, setSelectedPointId] = useState<string | null>(null);
  const [isSaving, setIsSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleGenerate = async (latitude?: number, longitude?: number, cityName?: string) => {
    setError(null);
    setIsGenerating(true);

    try {
      const res = await apiFetch<{ points: PickupPoint[] }>("/pickup/generate", {
        method: "POST",
        body: {
          latitude,
          longitude,
          city: cityName || city,
        },
      });

      setPoints(res.points || []);
      if (res.points && res.points.length > 0) {
        setSelectedPointId(res.points[0].id);
      }
    } catch (err: unknown) {
      if (err instanceof Error) {
        if (err.message.includes("401")) {
          router.push("/login?next=/onboarding");
          return;
        }
        setError(err.message);
      } else {
        setError("Не удалось сгенерировать пункты выдачи.");
      }
    } finally {
      setIsGenerating(false);
    }
  };

  const handleUseGeolocation = () => {
    setIsLoadingGeo(true);
    setError(null);

    if (!navigator.geolocation) {
      setError("Геолокация не поддерживается вашим браузером. Используйте выбор по городу.");
      setIsLoadingGeo(false);
      return;
    }

    navigator.geolocation.getCurrentPosition(
      (pos) => {
        setIsLoadingGeo(false);
        handleGenerate(pos.coords.latitude, pos.coords.longitude);
      },
      () => {
        setIsLoadingGeo(false);
        // Fallback on city
        handleGenerate(undefined, undefined, city);
      },
      { timeout: 5000 }
    );
  };

  const handleSavePickupPoint = async () => {
    if (!selectedPointId) return;

    setIsSaving(true);
    try {
      await apiFetch("/cart/pickup-point", {
        method: "PUT",
        body: {
          pickup_point_id: selectedPointId,
        },
      });

      router.push("/catalog");
    } catch (err: unknown) {
      if (err instanceof Error) {
        setError(err.message);
      }
    } finally {
      setIsSaving(false);
    }
  };

  return (
    <div className="container mx-auto max-w-2xl px-4 sm:px-6 py-12 space-y-8">
      <div className="text-center space-y-3">
        <Badge variant="dopamine" className="gap-1.5 px-3 py-1">
          <Sparkles className="h-4 w-4 text-rose-500" />
          Синтетические ПВЗ (INV-03)
        </Badge>
        <h1 className="text-3xl sm:text-4xl font-black text-zinc-900 dark:text-zinc-50 tracking-tight">
          Выберите удобный пункт выдачи
        </h1>
        <p className="text-sm sm:text-base text-zinc-600 dark:text-zinc-400">
          Мы автоматически рассчитаем точки в радиусе 100–500 метров от вашей позиции.
        </p>
      </div>

      {error && (
        <div className="rounded-2xl bg-red-50 p-4 text-sm text-red-600 dark:bg-red-950/40 dark:text-red-300">
          {error}
        </div>
      )}

      {/* Geolocation & City Form */}
      <Card className="shadow-lg border-zinc-200/80 dark:border-zinc-800">
        <CardHeader>
          <CardTitle className="text-lg">Определение местоположения</CardTitle>
          <CardDescription>
            Разрешите браузеру доступ к геолокации или введите город вручную
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <Button
            variant="default"
            size="lg"
            className="w-full gap-2 text-base rounded-xl"
            isLoading={isLoadingGeo || isGenerating}
            onClick={handleUseGeolocation}
          >
            <Navigation className="h-5 w-5" />
            <span>Определить по геолокации</span>
          </Button>

          <div className="relative flex items-center justify-center">
            <div className="border-t border-zinc-200 dark:border-zinc-800 w-full" />
            <span className="bg-white dark:bg-zinc-900 px-3 text-xs text-zinc-400 uppercase font-bold">
              или
            </span>
          </div>

          <div className="flex gap-2">
            <Input
              placeholder="Город (например, Ростов-на-Дону)"
              value={city}
              onChange={(e) => setCity(e.target.value)}
              className="rounded-xl"
            />
            <Button
              variant="secondary"
              isLoading={isGenerating}
              onClick={() => handleGenerate(undefined, undefined, city)}
              className="rounded-xl px-6"
            >
              Найти
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* Generated Pickup Points List */}
      {points.length > 0 && (
        <div className="space-y-4 pt-4">
          <div className="flex items-center justify-between">
            <h2 className="text-xl font-bold text-zinc-900 dark:text-zinc-100">
              Ближайшие пункты выдачи ({points.length})
            </h2>
            <span className="text-xs text-zinc-500 font-medium">Радиус 100-500м</span>
          </div>

          <div className="grid grid-cols-1 gap-3">
            {points.map((pt) => {
              const isSelected = selectedPointId === pt.id;
              return (
                <div
                  key={pt.id}
                  onClick={() => setSelectedPointId(pt.id)}
                  className={`flex items-center justify-between p-4 rounded-2xl border cursor-pointer transition-all ${
                    isSelected
                      ? "border-rose-500 bg-rose-50/50 dark:bg-rose-950/20 shadow-md"
                      : "border-zinc-200/80 dark:border-zinc-800 bg-white dark:bg-zinc-900 hover:border-zinc-300"
                  }`}
                >
                  <div className="flex items-start gap-3">
                    <div
                      className={`flex h-10 w-10 items-center justify-center rounded-xl transition-colors ${
                        isSelected
                          ? "bg-rose-500 text-white shadow-sm"
                          : "bg-zinc-100 text-zinc-500 dark:bg-zinc-800"
                      }`}
                    >
                      <MapPin className="h-5 w-5" />
                    </div>
                    <div>
                      <h4 className="font-bold text-sm text-zinc-900 dark:text-zinc-100">
                        {pt.name}
                      </h4>
                      <p className="text-xs text-zinc-500">
                        Синтетический ПВЗ Dopamine Market
                      </p>
                    </div>
                  </div>

                  <div className="flex items-center gap-3">
                    <Badge variant={isSelected ? "dopamine" : "secondary"}>
                      {Math.round(pt.distance_meters)} м
                    </Badge>

                    <div
                      className={`flex h-6 w-6 items-center justify-center rounded-full border ${
                        isSelected
                          ? "border-rose-500 bg-rose-500 text-white"
                          : "border-zinc-300 dark:border-zinc-700"
                      }`}
                    >
                      {isSelected && <Check className="h-3.5 w-3.5" />}
                    </div>
                  </div>
                </div>
              );
            })}
          </div>

          <Button
            size="lg"
            variant="glow"
            className="w-full gap-2 text-base rounded-2xl mt-4"
            disabled={!selectedPointId}
            isLoading={isSaving}
            onClick={handleSavePickupPoint}
          >
            <span>Подтвердить выбор и перейти в каталог</span>
            <ArrowRight className="h-5 w-5" />
          </Button>
        </div>
      )}
    </div>
  );
}
