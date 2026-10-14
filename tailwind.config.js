// CSS compilado en build (Dockerfile, etapa "css") en lugar del runtime de
// Tailwind Play en el navegador, que en el móvil recalculaba estilos en cada
// cambio del DOM. Mismo tema que el antiguo tailwind.config en línea de
// layout.html. Local: npx tailwindcss@3.4.1 -c tailwind.config.js -o static/app.css -m
/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    './templates/**/*.html',
    './static/app.js',
    './static/offline-storage.js',
    './static/ui-scale.js',
    './static/viewport.js',
  ],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        primary: '#f9a8d4',
        uncertain: '#fef3c7',
      },
    },
  },
}
