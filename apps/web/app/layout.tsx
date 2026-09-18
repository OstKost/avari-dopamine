import type { Metadata } from "next";
import { Header } from "@/components/layout/header";
import { Footer } from "@/components/layout/footer";
import { BottomNav } from "@/components/layout/bottom-nav";
import "./globals.css";

export const metadata: Metadata = {
  title: "Avari Dopamine — Маркетплейс мгновенного дофамина",
  description: "Любой заказ всего за 10₽. Мгновенная синтетическая доставка, геймификация, стрики и яркие эмоции.",
  icons: {
    icon: "/favicon.png",
    apple: "/logo-minimal-star.png",
  },
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="ru" className="dark">
      <body className="flex min-h-screen flex-col bg-[#050B14] text-[#F4F1E8] antialiased font-sans">
        <Header />
        <main className="flex-1 pb-16 md:pb-0">{children}</main>
        <Footer />
        <BottomNav />
      </body>
    </html>
  );
}
