/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    './pages/**/*.{js,ts,jsx,tsx,mdx}',
    './components/**/*.{js,ts,jsx,tsx,mdx}',
    './app/**/*.{js,ts,jsx,tsx,mdx}',
  ],
  theme: {
    extend: {
      colors: {
        space: {
          dark: '#0a0e27',
          darker: '#050714',
          blue: '#1e3a8a',
          purple: '#6b21a8',
          cyan: '#06b6d4',
          star: '#fbbf24',
          nebula: '#7c3aed',
          text: '#e0e7ff',
          'text-dim': '#a5b4fc',
        },
      },
      animation: {
        'pulse-slow': 'pulse 3s cubic-bezier(0.4, 0, 0.6, 1) infinite',
      },
    },
  },
  plugins: [],
}
