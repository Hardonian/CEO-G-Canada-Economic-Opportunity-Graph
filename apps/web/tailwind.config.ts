import type { Config } from "tailwindcss";

const config: Config = {
  content: [
    "./app/**/*.{js,ts,jsx,tsx,mdx}",
    "./components/**/*.{js,ts,jsx,tsx,mdx}",
  ],
  darkMode: "class",
  theme: {
    extend: {
      colors: {
        background: "#050B08",     // Deep Boreal Obsidian (Not generic dark blue)
        surface: "#0B1511",        // Rich Forest Surface
        card: "#101F19",           // Boreal Obsidian Card
        cardHover: "#162C23",      // Interactive Hover Card
        border: "#1D382D",         // Emerald-tinted Slate Border
        borderSubtle: "#12251D",   // Subdued Border
        primary: {
          DEFAULT: "#10B981",      // Radiant Aurora Emerald
          hover: "#059669",
          light: "#34D399",
        },
        aurora: {
          DEFAULT: "#00F5A0",      // Electric Aurora Green
          glow: "#10B981",
          emerald: "#059669",
          teal: "#14B8A6",
          mint: "#6EE7B7",
        },
        accent: {
          cyan: "#00F5A0",         // Upgraded to Electric Aurora
          gold: "#F59E0B",         // Canadian Mineral Gold
          green: "#10B981",        // Pure Aurora Green
          purple: "#A855F7",       // Northern Light Violet
          copper: "#D97706",       // Canadian Copper
          crimson: "#E11D48",      // Sovereign Maple Crimson
        },
        gold: {
          DEFAULT: "#F59E0B",      // Canadian Mineral Gold
          light: "#FBBF24",
          dark: "#D97706",
          copper: "#B45309",
        },
        crimson: {
          DEFAULT: "#E11D48",      // Canadian Maple Sovereign Crimson
          light: "#FB7185",
          dark: "#BE123C",
        },
        text: {
          main: "#F2F6F4",         // Quartz White
          muted: "#9CB3A8",        // Boreal Muted
          subtle: "#5F786C",       // Subdued Sage
        }
      },
      fontFamily: {
        sans: ["Inter", "-apple-system", "BlinkMacSystemFont", "Segoe UI", "sans-serif"],
        mono: ["JetBrains Mono", "SFMono-Regular", "Menlo", "monospace"],
      },
      backgroundImage: {
        "aurora-mesh": "radial-gradient(ellipse 80% 50% at 50% -20%, rgba(0, 245, 160, 0.15), rgba(5, 11, 8, 0))",
        "gold-glow": "radial-gradient(ellipse 60% 40% at 50% 120%, rgba(245, 158, 11, 0.12), rgba(5, 11, 8, 0))",
        "radial-highlight": "radial-gradient(circle at 50% 0%, rgba(16, 185, 129, 0.2) 0%, transparent 70%)",
      },
      animation: {
        "pulse-slow": "pulse 4s cubic-bezier(0.4, 0, 0.6, 1) infinite",
        "float": "float 6s ease-in-out infinite",
        "shimmer": "shimmer 2.5s linear infinite",
      },
      keyframes: {
        float: {
          "0%, 100%": { transform: "translateY(0)" },
          "50%": { transform: "translateY(-6px)" },
        },
        shimmer: {
          "0%": { backgroundPosition: "-200% 0" },
          "100%": { backgroundPosition: "200% 0" },
        }
      }
    },
  },
  plugins: [],
};

export default config;
