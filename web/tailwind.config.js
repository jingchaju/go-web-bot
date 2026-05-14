/** @type {import('tailwindcss').Config} */
export default {
  content: [
    './index.html',
    './src/**/*.{vue,js,ts,jsx,tsx}',
  ],
  theme: {
    extend: {
      colors: {
        brand: '#2563eb',
        ink: '#0f172a',
      },
    },
  },
  plugins: [],
}
