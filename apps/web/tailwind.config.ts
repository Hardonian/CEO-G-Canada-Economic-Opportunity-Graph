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
        background: "#080D1A",
        surface: "#0F172A",
        card: "#131E36",
        cardHover: "#182746",
        border: "#202E4C",
        borderSubtle: "#16223B",
        primary: {
          DEFAULT: "#D32F2F", // Canadian Flag Red Accent
          hover: "#B71C1C",
          light: "#FFEBEE",
        },
        accent: {
          cyan: "#00F2FE",
          gold: "#F59E0B",
          green: "#10B981",
          purple: "#8B5CF6",
        },
        text: {
          main: "#F8FAFC",
          muted: "#94A3B8",
          subtle: "#64748B",
        }
      },
      fontFamily: {
        sans: ["Inter", "-apple-system", "BlinkMacSystemFont", "Segoe UI", "sans-serif"],
        mono: ["JetBrains Mono", "SFMono-Regular", "Menlo", "monospace"],
      }
    },
  },
  plugins: [],
};

export default config;
