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
        background: "#050B08",     // Deep Boreal Obsidian
        surface: "#0C1812",        // High-contrast forest surface
        card: "#12231B",           // Institutional Boreal panel
        cardHover: "#193126",      // Clear interactive hover state
        border: "#5C7D6D",         // 3:1+ control boundary on supported panels
        borderSubtle: "#314B40",   // Decorative divider, not a sole status cue
        primary: {
          DEFAULT: "#22D39A",      // Radiant Aurora Emerald
          hover: "#35E8AE",
          light: "#6EF1C5",
        },
        aurora: {
          DEFAULT: "#00F5A0",      // Electric Aurora Green
          glow: "#22D39A",
          emerald: "#18BF8A",
          teal: "#35D8C1",
          mint: "#6EF1C5",
        },
        accent: {
          cyan: "#00F5A0",         // Upgraded to Electric Aurora
          gold: "#F6C453",         // Canadian Mineral Gold
          green: "#22D39A",        // Pure Aurora Green
          purple: "#C5A3FF",       // Northern Light Violet
          copper: "#E8A24B",       // Canadian Copper
          crimson: "#FF5D73",      // Sovereign Maple Crimson
        },
        gold: {
          DEFAULT: "#F6C453",      // Canadian Mineral Gold
          light: "#FFE08A",
          dark: "#E9A43B",
          copper: "#E8A24B",
        },
        crimson: {
          DEFAULT: "#FF5D73",      // Canadian Maple accent; use sparingly
          light: "#FF91A0",
          dark: "#D9435A",
        },
        text: {
          main: "#F7FAF8",         // Quartz White
          muted: "#BBCBC3",        // 8:1+ across supported panels
          subtle: "#91AA9E",       // 5.5:1+ across supported panels
        }
      },
      fontFamily: {
        sans: ["Inter", "-apple-system", "BlinkMacSystemFont", "Segoe UI", "sans-serif"],
        mono: ["JetBrains Mono", "SFMono-Regular", "Menlo", "monospace"],
      },
      backgroundImage: {
        "aurora-mesh": "radial-gradient(ellipse 80% 50% at 50% -20%, rgba(0, 245, 160, 0.14), rgba(5, 11, 8, 0))",
        "gold-glow": "radial-gradient(ellipse 60% 40% at 50% 120%, rgba(246, 196, 83, 0.1), rgba(5, 11, 8, 0))",
        "radial-highlight": "radial-gradient(circle at 50% 0%, rgba(34, 211, 154, 0.18) 0%, transparent 70%)",
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
