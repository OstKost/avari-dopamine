"use client";

import { useState } from "react";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { apiFetch } from "@/lib/api/client";

const loginSchema = z.object({
  email: z.string().email("Введите корректный email адрес"),
  password: z.string().min(8, "Пароль должен содержать минимум 8 символов"),
});

type LoginFormData = z.infer<typeof loginSchema>;

export default function LoginPage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const next = searchParams.get("next") || "/";

  const [serverError, setServerError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<LoginFormData>({
    resolver: zodResolver(loginSchema),
  });

  const onSubmit = async (data: LoginFormData) => {
    setServerError(null);
    setIsLoading(true);

    try {
      await apiFetch("/auth/login", {
        method: "POST",
        body: data,
      });

      router.push(next);
      router.refresh();
    } catch (err: unknown) {
      if (err instanceof Error) {
        setServerError(err.message);
      } else {
        setServerError("Не удалось войти. Проверьте данные и попробуйте снова.");
      }
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <Card className="shadow-2xl border-[#1E3A50] bg-[#0B1622]">
      <CardHeader className="space-y-1 text-center">
        <CardTitle className="text-2xl text-[#F4F1E8]">Вход в аккаунт</CardTitle>
        <CardDescription className="text-[#9FB3C4]">
          Введите email и пароль для доступа к заказам
        </CardDescription>
      </CardHeader>
      <form onSubmit={handleSubmit(onSubmit)}>
        <CardContent className="space-y-4">
          {serverError && (
            <div className="rounded-xl bg-rose-950/40 border border-rose-500/30 p-3 text-sm text-rose-300">
              {serverError}
            </div>
          )}

          <Input
            label="Email"
            type="email"
            placeholder="user@example.com"
            error={errors.email?.message}
            {...register("email")}
          />

          <Input
            label="Пароль"
            type="password"
            placeholder="••••••••"
            error={errors.password?.message}
            {...register("password")}
          />
        </CardContent>

        <CardFooter className="flex flex-col gap-4">
          <Button type="submit" variant="reward" className="w-full text-base font-bold py-3 rounded-xl shadow-lg shadow-amber-500/20" isLoading={isLoading}>
            Войти
          </Button>

          <p className="text-center text-xs text-[#9FB3C4]">
            Нет аккаунта?{" "}
            <Link href="/register" className="font-semibold text-amber-400 hover:text-amber-300 underline underline-offset-2">
              Зарегистрироваться
            </Link>
          </p>
        </CardFooter>
      </form>
    </Card>
  );
}
