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
        <div className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full bg-amber-400/15 border border-amber-400/30 text-amber-400 text-xs font-bold">
          <Sparkles className="h-4 w-4" />
          <span>Синтетические ПВЗ (INV-03)</span>
        </div>
        <h1 className="text-3xl sm:text-4xl font-black text-[#F4F1E8] tracking-tight">
          Выберите удобный пункт выдачи
        </h1>
        <p className="text-sm sm:text-base text-[#9FB3C4]">
          Мы автоматически рассчитаем точки в радиусе 100–500 метров от вашей позиции.
        </p>
      </div>

      {error && (
        <div className="rounded-2xl bg-rose-950/40 border border-rose-500/30 p-4 text-sm text-rose-300">
          {error}
        </div>
      )}

      {/* Geolocation & City Form */}
      <Card className="shadow-xl border-[#1E3A50] bg-[#0B1622]">
        <CardHeader>
          <CardTitle className="text-lg text-[#F4F1E8]">Определение местоположения</CardTitle>
          <CardDescription className="text-[#9FB3C4]">
            Разрешите браузеру доступ к геолокации или введите город вручную
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <Button
            variant="default"
            size="lg"
            className="w-full gap-2 text-base rounded-xl bg-teal-500 hover:bg-teal-400 text-[#050B14] font-bold"
            isLoading={isLoadingGeo || isGenerating}
            onClick={handleUseGeolocation}
          >
            <Navigation className="h-5 w-5" />
            <span>Определить по геолокации</span>
          </Button>

          <div className="relative flex items-center justify-center">
            <div className="border-t border-[#1E3A50] w-full" />
            <span className="bg-[#0B1622] px-3 text-xs text-[#5E7488] uppercase font-bold">
              или
            </span>
          </div>

          <div className="flex gap-2">
            <Input
              placeholder="Город (например, Ростов-на-Дону)"
              value={city}
              onChange={(e) => setCity(e.target.value)}
              className="rounded-xl border-[#1E3A50] bg-[#0E1B29] text-[#F4F1E8]"
            />
            <Button
              variant="secondary"
              isLoading={isGenerating}
              onClick={() => handleGenerate(undefined, undefined, city)}
              className="rounded-xl px-6 border-[#1E3A50] bg-[#122234] text-[#F4F1E8] hover:bg-[#1E3A50]"
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
            <h2 className="text-xl font-bold text-[#F4F1E8]">
              Ближайшие пункты выдачи ({points.length})
            </h2>
            <span className="text-xs text-teal-400 font-medium">Радиус 100-500м</span>
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
                      ? "border-amber-400 bg-[#122234] shadow-lg shadow-amber-500/10"
                      : "border-[#1E3A50] bg-[#0B1622] hover:border-[#5E7488]"
                  }`}
                >
                  <div className="flex items-start gap-3">
                    <div
                      className={`flex h-10 w-10 items-center justify-center rounded-xl transition-colors ${
                        isSelected
                          ? "bg-amber-400 text-[#050B14] shadow-sm font-bold"
                          : "bg-[#0E1B29] text-teal-400 border border-[#1E3A50]"
                      }`}
                    >
                      <MapPin className="h-5 w-5" />
                    </div>
                    <div>
                      <h4 className="font-bold text-sm text-[#F4F1E8]">
                        {pt.name}
                      </h4>
                      <p className="text-xs text-[#9FB3C4]">
                        Синтетический ПВЗ Avari Dopamine
                      </p>
                    </div>
                  </div>

                  <div className="flex items-center gap-3">
                    <Badge variant={isSelected ? "default" : "secondary"} className={isSelected ? "bg-amber-400 text-[#050B14]" : ""}>
                      {Math.round(pt.distance_meters)} м
                    </Badge>

                    <div
                      className={`flex h-6 w-6 items-center justify-center rounded-full border ${
                        isSelected
                          ? "border-amber-400 bg-amber-400 text-[#050B14]"
                          : "border-[#1E3A50]"
                      }`}
                    >
                      {isSelected && <Check className="h-3.5 w-3.5 stroke-[3]" />}
                    </div>
                  </div>
                </div>
              );
            })}
          </div>

          <Button
            size="lg"
            variant="reward"
            className="w-full gap-2 text-base rounded-2xl mt-4 shadow-lg shadow-amber-500/20"
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
