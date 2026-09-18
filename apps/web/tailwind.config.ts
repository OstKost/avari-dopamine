import type { Config } from "tailwindcss";

const config: Config = {
  darkMode: ["class"],
  content: [
    "./app/**/*.{js,ts,jsx,tsx,mdx}",
    "./components/**/*.{js,ts,jsx,tsx,mdx}",
    "./lib/**/*.{js,ts,jsx,tsx,mdx}",
  ],
  theme: {
    extend: {
      colors: {
        background: "hsl(var(--background))",
        foreground: "hsl(var(--foreground))",
        // Avari Dopamine Design System Tokens
        "bg-base": "#050B14",
        "bg-surface": "#0B1622",
        "bg-elevated": "#122234",
        "bg-input": "#0E1B29",
        "bg-glass": "rgba(18, 34, 52, 0.65)",
        "border-default": "#1E3A50",
        "border-subtle": "#152B3D",
        "text-primary": "#F4F1E8",
        "text-secondary": "#9FB3C4",
        "text-muted": "#5E7488",
        // Amber Gold (Rewards, XP, Streaks, Primary CTA)
        amber: {
          DEFAULT: "#F2B84B",
          50: "#FFFBEB",
          100: "#FEF3C7",
          200: "#FDE68A",
          300: "#FCD34D",
          400: "#FFD37A",
          500: "#F2B84B",
          600: "#C6912F",
          700: "#B45309",
          800: "#92400E",
          900: "#78350F",
        },
        // Teal (Live tracking, movement, delivery)
        teal: {
          DEFAULT: "#54ACBF",
          50: "#F0FDFA",
          100: "#CCFBF1",
          200: "#99F6E4",
          300: "#5EEAD4",
          400: "#A7EBF2",
          500: "#54ACBF",
          600: "#0D9488",
          700: "#0F766E",
          800: "#115E59",
          900: "#003848",
        },
        card: {
          DEFAULT: "hsl(var(--card))",
          foreground: "hsl(var(--card-foreground))",
        },
        popover: {
          DEFAULT: "hsl(var(--popover))",
          foreground: "hsl(var(--popover-foreground))",
        },
        primary: {
          DEFAULT: "hsl(var(--primary))",
          foreground: "hsl(var(--primary-foreground))",
          hover: "hsl(var(--primary-hover))",
        },
        secondary: {
          DEFAULT: "hsl(var(--secondary))",
          foreground: "hsl(var(--secondary-foreground))",
        },
        muted: {
          DEFAULT: "hsl(var(--muted))",
          foreground: "hsl(var(--muted-foreground))",
        },
        accent: {
          DEFAULT: "hsl(var(--accent))",
          foreground: "hsl(var(--accent-foreground))",
        },
        destructive: {
          DEFAULT: "hsl(var(--destructive))",
          foreground: "hsl(var(--destructive-foreground))",
        },
        border: "hsl(var(--border))",
        input: "hsl(var(--input))",
        ring: "hsl(var(--ring))",
      },
      borderRadius: {
        lg: "var(--radius)",
        md: "calc(var(--radius) - 2px)",
        sm: "calc(var(--radius) - 4px)",
      },
      boxShadow: {
        "glow-amber": "0 0 24px rgba(242, 184, 75, 0.35)",
        "glow-amber-lg": "0 0 40px rgba(242, 184, 75, 0.5)",
        "glow-teal": "0 0 24px rgba(84, 172, 191, 0.35)",
        "glow-teal-lg": "0 0 40px rgba(84, 172, 191, 0.5)",
        "glass": "0 8px 32px 0 rgba(0, 0, 0, 0.37)",
      },
      backgroundImage: {
        "gradient-hero": "linear-gradient(135deg, #050B14 0%, #0B1622 40%, #122234 70%, #1B3A4D 100%)",
        "gradient-reward": "linear-gradient(135deg, #F2B84B 0%, #FFD37A 100%)",
        "gradient-live": "linear-gradient(135deg, #0E2A38 0%, #54ACBF 120%)",
        "gradient-gold-cta": "linear-gradient(180deg, #FFD37A 0%, #F2B84B 50%, #C6912F 100%)",
      },
    },
  },
  plugins: [],
};

export default config;
