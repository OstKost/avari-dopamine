"use client";

import { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { apiFetch } from "@/lib/api/client";

const registerSchema = z
  .object({
    email: z.string().email("Введите корректный email адрес"),
    password: z.string().min(8, "Пароль должен содержать минимум 8 символов"),
    confirmPassword: z.string().min(8, "Повторите пароль"),
  })
  .refine((data) => data.password === data.confirmPassword, {
    message: "Пароли не совпадают",
    path: ["confirmPassword"],
  });

type RegisterFormData = z.infer<typeof registerSchema>;

export default function RegisterPage() {
  const router = useRouter();
  const [serverError, setServerError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<RegisterFormData>({
    resolver: zodResolver(registerSchema),
  });

  const onSubmit = async (data: RegisterFormData) => {
    setServerError(null);
    setIsLoading(true);

    try {
      await apiFetch("/auth/register", {
        method: "POST",
        body: {
          email: data.email,
          password: data.password,
        },
      });

      router.push("/");
      router.refresh();
    } catch (err: unknown) {
      if (err instanceof Error) {
        setServerError(err.message);
      } else {
        setServerError("Не удалось зарегистрироваться. Попробуйте другой email.");
      }
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <Card className="shadow-2xl border-[#1E3A50] bg-[#0B1622]">
      <CardHeader className="space-y-1 text-center">
        <CardTitle className="text-2xl text-[#F4F1E8]">Создать аккаунт</CardTitle>
        <CardDescription className="text-[#9FB3C4]">
          Зарегистрируйтесь, чтобы оформлять заказы по 10 ₽
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

          <Input
            label="Подтверждение пароля"
            type="password"
            placeholder="••••••••"
            error={errors.confirmPassword?.message}
            {...register("confirmPassword")}
          />
        </CardContent>

        <CardFooter className="flex flex-col gap-4">
          <Button type="submit" variant="reward" className="w-full text-base font-bold py-3 rounded-xl shadow-lg shadow-amber-500/20" isLoading={isLoading}>
            Зарегистрироваться
          </Button>

          <p className="text-center text-xs text-[#9FB3C4]">
            Уже есть аккаунт?{" "}
            <Link href="/login" className="font-semibold text-amber-400 hover:text-amber-300 underline underline-offset-2">
              Войти
            </Link>
          </p>
        </CardFooter>
      </form>
    </Card>
  );
}
