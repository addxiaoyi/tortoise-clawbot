/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        tortoise: {
          50: '#f0fdf4',
          100: '#dcfce7',
          200: '#bbf7d0',
          300: '#86efac',
          400: '#4ade80',
          500: '#22c55e',
          600: '#16a34a',
          700: '#15803d',
          800: '#166534',
          900: '#14532d',
        },
        dark: {
          100: '#1e293b',
          200: '#1a1f2e',
          300: '#151922',
          400: '#111318',
          500: '#0d0f14',
        }
      },
      fontFamily: {
        sans: ['Inter', 'system-ui', 'sans-serif'],
      },
      animation: {
        'pulse-slow': 'pulse 3s cubic-bezier(0.4, 0, 0.6, 1) infinite',
        'typing': 'typing 1.4s ease-in-out infinite',
      },
      keyframes: {
        typing: {
          '0%, 100%': { opacity: '0' },
          '50%': { opacity: '1' },
        }
      }
    },
  },
  plugins: [],
}
