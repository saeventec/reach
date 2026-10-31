import tailwindcss from '@tailwindcss/vite'
import react from '@vitejs/plugin-react'
import { defineConfig } from 'vite'

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    react(),
    // Creates the CSS for the Tailwind classes used in our code.
    // Learn more: docs/learn/tailwind.md
    tailwindcss(),
  ],
  server: {
    // Listen on all network interfaces, not only "localhost". Inside a
    // container, "localhost" is the container itself, so without this
    // Docker could not forward your browser's requests to Vite.
    host: true,
  },
})
