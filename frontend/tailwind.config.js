/** @type {import('tailwindcss').Config} */
export default {
  darkMode: 'class',
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        mui: {
          blue: {
            50: 'var(--mui-blue-50)',
            100: 'var(--mui-blue-100)',
            200: 'var(--mui-blue-200)',
            300: 'var(--mui-blue-300)',
            400: 'var(--mui-blue-400)',
            500: 'var(--mui-blue-500)',
            600: 'var(--mui-blue-600)',
            700: 'var(--mui-blue-700)',
            800: 'var(--mui-blue-800)',
            900: 'var(--mui-blue-900)',
          },
          grey: {
            50: 'var(--mui-grey-50)',
            100: 'var(--mui-grey-100)',
            200: 'var(--mui-grey-200)',
            300: 'var(--mui-grey-300)',
            400: 'var(--mui-grey-400)',
            500: 'var(--mui-grey-500)',
            600: 'var(--mui-grey-600)',
            700: 'var(--mui-grey-700)',
            800: 'var(--mui-grey-800)',
            900: 'var(--mui-grey-900)',
          },
          'blue-grey': {
            50: 'var(--mui-blue-grey-50)',
            100: 'var(--mui-blue-grey-100)',
            200: 'var(--mui-blue-grey-200)',
            300: 'var(--mui-blue-grey-300)',
            400: 'var(--mui-blue-grey-400)',
            500: 'var(--mui-blue-grey-500)',
            600: 'var(--mui-blue-grey-600)',
            700: 'var(--mui-blue-grey-700)',
            800: 'var(--mui-blue-grey-800)',
            900: 'var(--mui-blue-grey-900)',
          },
          dark: {
            bg: 'var(--mui-dark-bg)',
            paper: 'var(--mui-dark-paper)',
          }
        }
      }
    },
  },
  plugins: [],
}
