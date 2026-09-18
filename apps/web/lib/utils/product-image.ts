/**
 * Deterministic product image generator (ADR-012).
 * Generates beautiful, responsive, offline-ready SVG gradients and product silhouettes
 * based on product seed and category, eliminating external third-party CDN 403/downtime issues.
 */

const PALETTES = [
  { bg1: "#071828", bg2: "#0A283B", glow: "#54ACBF", accent: "#FFD37A" }, // Deep Midnight Teal
  { bg1: "#101B2E", bg2: "#192B45", glow: "#3B82F6", accent: "#93C5FD" }, // Deep Navy
  { bg1: "#1A1728", bg2: "#2A2042", glow: "#8B5CF6", accent: "#F472B6" }, // Deep Violet
  { bg1: "#1C141E", bg2: "#2F1926", glow: "#EC4899", accent: "#FDA4AF" }, // Deep Crimson
  { bg1: "#0C1F1D", bg2: "#133835", glow: "#10B981", accent: "#6EE7B7" }, // Deep Emerald
  { bg1: "#1F1A0E", bg2: "#352B14", glow: "#F2B84B", accent: "#FFD37A" }, // Deep Amber Gold
];

function detectIcon(name?: string, category?: string): string {
  const combined = `${name || ""} ${category || ""}`.toLowerCase();
  if (combined.includes("пицц")) return "🍕";
  if (combined.includes("донат") || combined.includes("пончик")) return "🍩";
  if (combined.includes("кола") || combined.includes("газиров")) return "🥤";
  if (combined.includes("бургер") || combined.includes("чизбургер")) return "🍔";
  if (combined.includes("чипс") || combined.includes("снек") || combined.includes("сухарик")) return "🍟";
  if (combined.includes("сок") || combined.includes("лимонад") || combined.includes("смузи")) return "🧃";
  if (combined.includes("кофе") || combined.includes("капучино") || combined.includes("латте")) return "☕";
  if (combined.includes("шоколад") || combined.includes("конфет")) return "🍫";
  if (combined.includes("суши") || combined.includes("ролл")) return "🍣";
  if (combined.includes("морожен") || combined.includes("пломбир")) return "🍦";
  if (combined.includes("наушник") || combined.includes("аудио")) return "🎧";
  if (combined.includes("свеч") || combined.includes("лампа")) return "🕯️";
  if (combined.includes("игр") || combined.includes("кубик") || combined.includes("антистресс")) return "🎮";
  if (combined.includes("книг") || combined.includes("блокнот")) return "📖";
  if (combined.includes("чай") || combined.includes("травы")) return "🍵";
  if (combined.includes("ягод") || combined.includes("фрукт") || combined.includes("витамин")) return "🍓";
  if (combined.includes("выпечк") || combined.includes("круассан")) return "🥐";
  return "✨";
}

function hashString(str: string): number {
  let hash = 0;
  for (let i = 0; i < str.length; i++) {
    hash = (hash << 5) - hash + str.charCodeAt(i);
    hash |= 0;
  }
  return Math.abs(hash);
}

export function getProductImageUrl(seed: string, name?: string, category?: string): string {
  const cleanSeed = seed || "dopamine-item";
  const hash = hashString(cleanSeed);
  const palette = PALETTES[hash % PALETTES.length];
  const icon = detectIcon(name, category);
  const title = name ? (name.length > 26 ? name.slice(0, 24) + "…" : name) : "Dopamine Item";
  const cat = category || "Синтетический маркет";

  // Escape special XML characters
  const escapeXml = (unsafe: string) =>
    unsafe
      .replace(/&/g, "&amp;")
      .replace(/</g, "&lt;")
      .replace(/>/g, "&gt;")
      .replace(/"/g, "&quot;")
      .replace(/'/g, "&apos;");

  const safeTitle = escapeXml(title);
  const safeCategory = escapeXml(cat);

  const svg = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 400 300" width="400" height="300">
  <defs>
    <linearGradient id="bgGrad-${hash}" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" stop-color="${palette.bg1}" />
      <stop offset="100%" stop-color="${palette.bg2}" />
    </linearGradient>
    <radialGradient id="ambientGlow-${hash}" cx="50%" cy="45%" r="60%">
      <stop offset="0%" stop-color="${palette.glow}" stop-opacity="0.35" />
      <stop offset="70%" stop-color="${palette.glow}" stop-opacity="0.05" />
      <stop offset="100%" stop-color="${palette.glow}" stop-opacity="0" />
    </radialGradient>
    <radialGradient id="starGlow" cx="50%" cy="50%" r="50%">
      <stop offset="0%" stop-color="#FFD37A" stop-opacity="0.8" />
      <stop offset="100%" stop-color="#F2B84B" stop-opacity="0" />
    </radialGradient>
  </defs>

  <!-- Background Base -->
  <rect width="400" height="300" fill="url(#bgGrad-${hash})" />
  <rect width="400" height="300" fill="url(#ambientGlow-${hash})" />

  <!-- Ambient star particles -->
  <circle cx="65" cy="45" r="1.5" fill="#FFD37A" opacity="0.7" />
  <circle cx="330" cy="70" r="2" fill="#54ACBF" opacity="0.8" />
  <circle cx="345" cy="220" r="1.5" fill="#FFD37A" opacity="0.6" />
  <circle cx="45" cy="235" r="2" fill="#A7EBF2" opacity="0.7" />
  <polygon points="320,40 323,45 328,45 324,48 326,53 320,50 314,53 316,48 312,45 317,45" fill="#FFD37A" opacity="0.4" />

  <!-- Main Center Icon Platter with layered golden halo -->
  <g transform="translate(200, 115)">
    <circle cx="0" cy="0" r="62" fill="#050B14" fill-opacity="0.6" />
    <circle cx="0" cy="0" r="54" fill="#0B1622" stroke="${palette.glow}" stroke-width="1.5" stroke-opacity="0.4" />
    <circle cx="0" cy="0" r="46" fill="${palette.glow}" fill-opacity="0.12" />
    
    <!-- Central Icon -->
    <text x="0" y="19" font-size="46" text-anchor="middle" font-family="-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif">${icon}</text>
    
    <!-- Star sparkle on top right of platter -->
    <g transform="translate(38, -32) scale(0.7)">
      <polygon points="0,-12 3,-3 12,0 3,3 0,12 -3,3 -12,0 -3,-3" fill="#FFD37A" />
    </g>
  </g>

  <!-- Card Border outline -->
  <rect x="1" y="1" width="398" height="298" rx="16" fill="none" stroke="#1E3A50" stroke-width="1.5" stroke-opacity="0.6" />

  <!-- Category Badge Pill -->
  <g transform="translate(200, 218)">
    <text x="0" y="0" fill="#9FB3C4" font-size="12" font-weight="600" text-anchor="middle" font-family="-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif" letter-spacing="0.5">${safeCategory}</text>
  </g>

  <!-- Product Name Title -->
  <g transform="translate(200, 246)">
    <text x="0" y="0" fill="#F4F1E8" font-size="16" font-weight="800" text-anchor="middle" font-family="-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif">${safeTitle}</text>
  </g>
</svg>`;

  return `data:image/svg+xml;charset=utf-8,${encodeURIComponent(svg)}`;
}
