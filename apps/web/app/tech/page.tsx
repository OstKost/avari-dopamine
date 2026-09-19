import Link from "next/link";
import {
  Cpu,
  Database,
  Layers,
  Shield,
  Activity,
  FileCheck2,
  Terminal,
  GitBranch,
  CheckCircle2,
  ExternalLink,
  Lock,
  Compass
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";

export const metadata = {
  title: "Архитектура и стек | Avari Dopamine (For Recruiters & Tech Leads)",
  description: "Технический разбор архитектуры Dopamine Market: Modular Monolith, Clean Architecture, Transactional Outbox, SSE, Go + Next.js 15.",
};

export default function TechPage() {
  return (
    <div className="container mx-auto max-w-6xl px-4 sm:px-6 py-10 sm:py-14 space-y-12">
      {/* Hero Header */}
      <div className="space-y-4">
        <div className="flex flex-wrap items-center gap-2">
          <Badge variant="gold" className="gap-1.5 px-3 py-1 text-xs shadow-glow-amber">
            <Cpu className="h-3.5 w-3.5 text-amber-300" />
            <span>Architecture & Engineering Deep Dive</span>
          </Badge>
          <span className="text-xs text-[#5E7488]">· Для техлидов, архитекторов и рекрутеров</span>
        </div>

        <h1 className="text-3xl sm:text-5xl font-black text-[#F4F1E8] tracking-tight max-w-4xl leading-tight">
          Архитектура и стек{" "}
          <span className="bg-gradient-to-r from-[#F2B84B] via-[#FFD37A] to-[#F2B84B] bg-clip-text text-transparent">
            Dopamine Market
          </span>
        </h1>

        <p className="text-base sm:text-lg text-[#9FB3C4] max-w-3xl leading-relaxed">
          Высоконагруженный модульный монолит (Modular Monolith) на Go с чистой гексагональной архитектурой,
          атомарным transactional outbox, отказоустойчивой стейт-машиной доставки и Next.js 15 frontend.
        </p>
      </div>

      {/* Tech Stack Matrix */}
      <div className="space-y-4">
        <h2 className="text-xl sm:text-2xl font-black text-[#F4F1E8] flex items-center gap-2">
          <Layers className="h-5 w-5 text-amber-400" />
          <span>Технологический стек</span>
        </h2>

        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
          <Card className="border-[#1E3A50] bg-[#0B1622]/90">
            <CardHeader className="p-4 pb-2">
              <span className="text-xs font-bold text-amber-400 uppercase tracking-wider">Backend Core</span>
              <CardTitle className="text-base font-bold text-[#F4F1E8]">Go 1.23+</CardTitle>
            </CardHeader>
            <CardContent className="p-4 pt-1 text-xs text-[#9FB3C4] space-y-1">
              <p>• Chi HTTP router + middleware chain</p>
              <p>• pgx/v5 (connection pool + typed SQL)</p>
              <p>• sqlc для типобезопасного SQL</p>
              <p>• Goose миграции в embed.FS</p>
            </CardContent>
          </Card>

          <Card className="border-[#1E3A50] bg-[#0B1622]/90">
            <CardHeader className="p-4 pb-2">
              <span className="text-xs font-bold text-teal-400 uppercase tracking-wider">Data & Messaging</span>
              <CardTitle className="text-base font-bold text-[#F4F1E8]">PostgreSQL & Kafka</CardTitle>
            </CardHeader>
            <CardContent className="p-4 pt-1 text-xs text-[#9FB3C4] space-y-1">
              <p>• PostgreSQL (изолированные схемы на модуль)</p>
              <p>• Apache Kafka (асинхронные события)</p>
              <p>• Redis (сессии, корзина, rate limiting)</p>
              <p>• Transactional Outbox relay worker</p>
            </CardContent>
          </Card>

          <Card className="border-[#1E3A50] bg-[#0B1622]/90">
            <CardHeader className="p-4 pb-2">
              <span className="text-xs font-bold text-amber-400 uppercase tracking-wider">Frontend Engine</span>
              <CardTitle className="text-base font-bold text-[#F4F1E8]">Next.js 15 & React 19</CardTitle>
            </CardHeader>
            <CardContent className="p-4 pt-1 text-xs text-[#9FB3C4] space-y-1">
              <p>• App Router + Server Components (RSC)</p>
              <p>• TypeScript strict mode (0 any)</p>
              <p>• Tailwind CSS 3.4 + Canvas Confetti</p>
              <p>• OpenAPI $\rightarrow$ typed schema generation</p>
            </CardContent>
          </Card>

          <Card className="border-[#1E3A50] bg-[#0B1622]/90">
            <CardHeader className="p-4 pb-2">
              <span className="text-xs font-bold text-teal-400 uppercase tracking-wider">Observability</span>
              <CardTitle className="text-base font-bold text-[#F4F1E8]">Metrics & Tracing</CardTitle>
            </CardHeader>
            <CardContent className="p-4 pt-1 text-xs text-[#9FB3C4] space-y-1">
              <p>• Prometheus `/metrics` эндпоинт</p>
              <p>• OpenTelemetry W3C traceparent context</p>
              <p>• Structured slog JSON + GELF UDP</p>
              <p>• Архитектурный линтинг (depguard)</p>
            </CardContent>
          </Card>
        </div>
      </div>

      {/* Architectural Pillars */}
      <div className="space-y-6">
        <h2 className="text-xl sm:text-2xl font-black text-[#F4F1E8] flex items-center gap-2">
          <Shield className="h-5 w-5 text-teal-400" />
          <span>Ключевые инженерные решения и паттерны</span>
        </h2>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          {/* Pillar 1: Modular Monolith */}
          <div className="p-6 rounded-3xl bg-[#0B1622] border border-[#1E3A50] space-y-3">
            <div className="flex items-center gap-3">
              <div className="h-10 w-10 rounded-xl bg-amber-400/15 border border-amber-400/30 flex items-center justify-center text-amber-300">
                <Lock className="h-5 w-5" />
              </div>
              <div>
                <h3 className="text-base font-bold text-[#F4F1E8]">Modular Monolith с чистыми границами</h3>
                <span className="text-xs text-[#5E7488]">ADR-001 & ADR-004 · CI depguard rules</span>
              </div>
            </div>
            <p className="text-xs sm:text-sm text-[#9FB3C4] leading-relaxed">
              Запрещены прямые импорты между модулями (<code className="text-amber-300">internal/modules/X</code> $\to$ <code className="text-amber-300">internal/modules/Y</code>).
              Любое синхронное взаимодействие строго через <code className="text-teal-300">internal/contracts/*</code>, 
              а асинхронное — через события Kafka.
            </p>
          </div>

          {/* Pillar 2: Transactional Outbox */}
          <div className="p-6 rounded-3xl bg-[#0B1622] border border-[#1E3A50] space-y-3">
            <div className="flex items-center gap-3">
              <div className="h-10 w-10 rounded-xl bg-teal-400/15 border border-teal-400/30 flex items-center justify-center text-teal-300">
                <Database className="h-5 w-5" />
              </div>
              <div>
                <h3 className="text-base font-bold text-[#F4F1E8]">Transactional Outbox (Exactly-Once)</h3>
                <span className="text-xs text-[#5E7488]">ADR-003 · Dual-write prevention</span>
              </div>
            </div>
            <p className="text-xs sm:text-sm text-[#9FB3C4] leading-relaxed">
              Любая запись бизнес-агрегата и публикация доменного события выполняются в единой ACID-транзакции в таблицу <code className="text-teal-300">outbox_messages</code>.
              Фоновый воркер осуществляет relay в Kafka с идемпотентным дедуплицированием на стороне получателей.
            </p>
          </div>

          {/* Pillar 3: Resilient Delivery State Machine */}
          <div className="p-6 rounded-3xl bg-[#0B1622] border border-[#1E3A50] space-y-3">
            <div className="flex items-center gap-3">
              <div className="h-10 w-10 rounded-xl bg-teal-400/15 border border-teal-400/30 flex items-center justify-center text-teal-300">
                <Activity className="h-5 w-5" />
              </div>
              <div>
                <h3 className="text-base font-bold text-[#F4F1E8]">Отказоустойчивая стейт-машина доставки</h3>
                <span className="text-xs text-[#5E7488]">ADR-005 · No in-memory sleep goroutines</span>
              </div>
            </div>
            <p className="text-xs sm:text-sm text-[#9FB3C4] leading-relaxed">
              Никаких in-memory <code className="text-amber-300">time.Sleep</code> таймеров. Запланированные переходы хранятся в БД (<code className="text-teal-300">delivery.scheduled_transitions</code>) 
              и исполняются планировщиком через <code className="text-teal-300">SELECT ... FOR UPDATE SKIP LOCKED</code>, гарантируя выживание при рестартах серверов.
            </p>
          </div>

          {/* Pillar 4: Geodesic & SSE */}
          <div className="p-6 rounded-3xl bg-[#0B1622] border border-[#1E3A50] space-y-3">
            <div className="flex items-center gap-3">
              <div className="h-10 w-10 rounded-xl bg-amber-400/15 border border-amber-400/30 flex items-center justify-center text-amber-300">
                <Compass className="h-5 w-5" />
              </div>
              <div>
                <h3 className="text-base font-bold text-[#F4F1E8]">Геодезическая модель и Realtime SSE</h3>
                <span className="text-xs text-[#5E7488]">ADR-006 & ADR-008 · In-Process Pub/Sub</span>
              </div>
            </div>
            <p className="text-xs sm:text-sm text-[#9FB3C4] leading-relaxed">
              Генерация ПВЗ использует формулу гаверсинусов (Haversine) и прямую геодезическую задачу на сфере. 
              Статус заказа транслируется клиенту в реальном времени через Server-Sent Events (SSE) с keep-alive и автореконнектом.
            </p>
          </div>
        </div>
      </div>

      {/* Architectural Invariants Table */}
      <div className="space-y-4">
        <h2 className="text-xl sm:text-2xl font-black text-[#F4F1E8] flex items-center gap-2">
          <FileCheck2 className="h-5 w-5 text-amber-400" />
          <span>Системные инварианты (Invariants Matrix)</span>
        </h2>

        <div className="overflow-x-auto rounded-2xl border border-[#1E3A50] bg-[#0B1622]">
          <table className="w-full text-left text-xs sm:text-sm text-[#9FB3C4]">
            <thead className="bg-[#0E1B29] text-[#F4F1E8] border-b border-[#1E3A50] font-bold">
              <tr>
                <th className="p-3.5 sm:p-4">ID</th>
                <th className="p-3.5 sm:p-4">Правило инварианта</th>
                <th className="p-3.5 sm:p-4">Механизм обеспечения</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-[#1E3A50]/60 font-medium">
              <tr>
                <td className="p-3.5 sm:p-4 font-mono font-bold text-amber-400">INV-01</td>
                <td className="p-3.5 sm:p-4 text-[#F4F1E8]">Сумма к оплате за заказ ВСЕГДА 10.00 RUB</td>
                <td className="p-3.5 sm:p-4">Доменный конструктор заказа фиксирует сумму 10.00 RUB независимо от корзины</td>
              </tr>
              <tr>
                <td className="p-3.5 sm:p-4 font-mono font-bold text-amber-400">INV-02</td>
                <td className="p-3.5 sm:p-4 text-[#F4F1E8]">Один активный payment_pending на пользователя</td>
                <td className="p-3.5 sm:p-4">Conditional unique constraint / DB level atomic check перед созданием заказа</td>
              </tr>
              <tr>
                <td className="p-3.5 sm:p-4 font-mono font-bold text-amber-400">INV-03</td>
                <td className="p-3.5 sm:p-4 text-[#F4F1E8]">Пункт выдачи строго в радиусе [100м, 500м]</td>
                <td className="p-3.5 sm:p-4">Геодезический алгоритм вычисления координат точки назначения (DestinationPoint)</td>
              </tr>
              <tr>
                <td className="p-3.5 sm:p-4 font-mono font-bold text-amber-400">INV-04</td>
                <td className="p-3.5 sm:p-4 text-[#F4F1E8]">Заказ неизменяем после перехода в paid</td>
                <td className="p-3.5 sm:p-4">Иммутабельный доменный агрегат и валидация допустимых переходов стейт-машины</td>
              </tr>
              <tr>
                <td className="p-3.5 sm:p-4 font-mono font-bold text-amber-400">INV-05</td>
                <td className="p-3.5 sm:p-4 text-[#F4F1E8]">Exactly-once доставка событий потребителям</td>
                <td className="p-3.5 sm:p-4">Transactional Outbox + таблица processed_events для дедупликации</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      {/* Directory Structure & Quality Gates */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="p-6 rounded-3xl bg-[#0B1622] border border-[#1E3A50] space-y-4">
          <div className="flex items-center gap-2 text-amber-400">
            <Terminal className="h-5 w-5" />
            <h3 className="font-bold text-base text-[#F4F1E8]">Структура слоев модуля (Hexagonal)</h3>
          </div>
          <pre className="p-4 rounded-xl bg-[#050B14] border border-[#1E3A50] text-[11px] font-mono text-[#9FB3C4] overflow-x-auto leading-relaxed">
{`internal/modules/{name}/
├── domain/       # Чистые сущности, VO, правила (0 внешних deps)
├── port/         # Интерфейсы репозиториев, провайдеров
├── usecase/      # Бизнес-сценарии с инъекцией портов
└── adapter/
    ├── postgres/ # sqlc репозитории, миграции
    ├── httpapi/  # Chi хендлеры, DTO, парсинг
    └── kafka/    # Продюсеры событий, консьюмеры`}
          </pre>
        </div>

        <div className="p-6 rounded-3xl bg-[#0B1622] border border-[#1E3A50] space-y-4">
          <div className="flex items-center gap-2 text-teal-400">
            <GitBranch className="h-5 w-5" />
            <h3 className="font-bold text-base text-[#F4F1E8]">Quality Gates & CI/CD</h3>
          </div>
          <div className="space-y-2 text-xs sm:text-sm text-[#9FB3C4]">
            <p className="flex items-center gap-2">
              <CheckCircle2 className="h-4 w-4 text-teal-400 flex-shrink-0" />
              <span><strong>make lint-arch:</strong> проверка границ модулей линтером depguard</span>
            </p>
            <p className="flex items-center gap-2">
              <CheckCircle2 className="h-4 w-4 text-teal-400 flex-shrink-0" />
              <span><strong>Unit & Integration Tests:</strong> 100% покрытие доменного слоя и usecase mocks</span>
            </p>
            <p className="flex items-center gap-2">
              <CheckCircle2 className="h-4 w-4 text-teal-400 flex-shrink-0" />
              <span><strong>TypeScript Strict:</strong> полная типобезопасность фронтенда без any</span>
            </p>
            <p className="flex items-center gap-2">
              <CheckCircle2 className="h-4 w-4 text-teal-400 flex-shrink-0" />
              <span><strong>Single VPS Compose:</strong> воспроизводимый деплой всей инфраструктуры</span>
            </p>
          </div>
        </div>
      </div>

      {/* Navigation Links */}
      <div className="flex flex-wrap items-center justify-between gap-4 pt-4 border-t border-[#1E3A50]">
        <Link href="/about">
          <Button variant="outline" className="rounded-xl border-[#1E3A50] text-[#F4F1E8] hover:bg-[#1E3A50]">
            ← О концепции сервиса
          </Button>
        </Link>
        <Link href="/catalog">
          <Button variant="gold" className="rounded-xl gap-2 font-bold shadow-glow-amber">
            <span>Протестировать каталог</span>
            <ExternalLink className="h-4 w-4" />
          </Button>
        </Link>
      </div>
    </div>
  );
}
