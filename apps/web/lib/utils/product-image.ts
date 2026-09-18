/**
 * Deterministic product image generator (ADR-012).
 * Generates beautiful, responsive, offline-ready SVG gradients and product silhouettes
 * based on product seed and category, eliminating external third-party CDN 403/downtime issues.
 */

const PALETTES = [
  { bg1: "#8B5CF6", bg2: "#EC4899", accent: "#F472B6", icon: "✨" }, // Purple - Pink
  { bg1: "#3B82F6", bg2: "#06B6D4", accent: "#67E8F9", icon: "⚡" }, // Blue - Cyan
  { bg1: "#10B981", bg2: "#059669", accent: "#6EE7B7", icon: "🌱" }, // Emerald
  { bg1: "#F59E0B", bg2: "#EF4444", accent: "#FCA5A5", icon: "🔥" }, // Amber - Red
  { bg1: "#6366F1", bg2: "#8B5CF6", accent: "#C4B5FD", icon: "💎" }, // Indigo - Violet
  { bg1: "#EC4899", bg2: "#F43F5E", accent: "#FDA4AF", icon: "❤️" }, // Pink - Rose
  { bg1: "#14B8A6", bg2: "#3B82F6", accent: "#93C5FD", icon: "🌊" }, // Teal - Blue
  { bg1: "#84CC16", bg2: "#10B981", accent: "#A7F3D0", icon: "🍀" }, // Lime - Green
  { bg1: "#F97316", bg2: "#FB923C", accent: "#FED7AA", icon: "🎯" }, // Orange
  { bg1: "#6D28D9", bg2: "#4F46E5", accent: "#A5B4FC", icon: "🔮" }, // Deep Purple
];

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
  const title = name ? (name.length > 28 ? name.slice(0, 26) + "…" : name) : "Dopamine Item";
  const cat = category || "10.00 ₽";

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
    <linearGradient id="grad-${hash}" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" stop-color="${palette.bg1}" />
      <stop offset="100%" stop-color="${palette.bg2}" />
    </linearGradient>
    <filter id="glow-${hash}" x="-20%" y="-20%" width="140%" height="140%">
      <feGaussianBlur stdDeviation="15" result="blur" />
      <feComposite in="SourceGraphic" in2="blur" operator="over" />
    </filter>
  </defs>

  <!-- Background -->
  <rect width="400" height="300" fill="url(#grad-${hash})" />

  <!-- Decorative geometric background patterns -->
  <circle cx="340" cy="60" r="100" fill="white" fill-opacity="0.08" />
  <circle cx="60" cy="240" r="80" fill="white" fill-opacity="0.06" />
  <path d="M 0 200 Q 150 120 400 240 L 400 300 L 0 300 Z" fill="white" fill-opacity="0.05" />

  <!-- Center Card / Icon Platter -->
  <g transform="translate(200, 120)">
    <circle cx="0" cy="0" r="54" fill="white" fill-opacity="0.18" />
    <circle cx="0" cy="0" r="44" fill="white" fill-opacity="0.25" filter="url(#glow-${hash})" />
    <text x="0" y="16" font-size="42" text-anchor="middle" font-family="-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif">${palette.icon}</text>
  </g>

  <!-- Product Badge Pill -->
  <g transform="translate(20, 24)">
    <rect width="90" height="24" rx="12" fill="white" fill-opacity="0.22" />
    <text x="45" y="16" fill="white" font-size="11" font-weight="700" text-anchor="middle" font-family="-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif">10.00 ₽</text>
  </g>

  <!-- Product Category Label -->
  <text x="200" y="222" fill="white" fill-opacity="0.85" font-size="12" font-weight="500" text-anchor="middle" font-family="-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif">${safeCategory}</text>

  <!-- Product Name Label -->
  <text x="200" y="248" fill="white" font-size="16" font-weight="700" text-anchor="middle" font-family="-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif">${safeTitle}</text>
</svg>`;

  return `data:image/svg+xml;charset=utf-8,${encodeURIComponent(svg)}`;
}
