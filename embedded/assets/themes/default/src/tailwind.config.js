/** @type {import('tailwindcss').Config} */
module.exports = {
  daisyui: {
    themes: [
      {
        letemcook: {
          "base-content": "#493b4b",
          "secondary-content": "#493b4b",
          "primary": "#493b4b",
          "secondary": "#fff",
          "neutral": "#bda59e",
          "accent": "#d2dddb",
          "base-100": "#f1c484",
          "info": "#67e8f9",
          "success": "#a3e635",
          "warning": "#facc15",
          "error": "#e11d48",
          "--color-primary": "#493b4b",
          "--color-content-secondary": "#f1f1f1",
          "--color-inputs-secondary": "#fff",
          "--color-ck-border": "#bdbdbd",
          "--color-monitor-border": "#bdbdbd",
          "--color-navbar-bg": "#f3f4f6",
          "--color-accent": "#e3b87c",
          "--color-accent-focus": "#d6af76",
          "--outline-accent": "#e3b87c",
          "--overlay-light-alpha": "rgba(255, 255, 255, 0.2)",
          "--overlay-light-stronger-alpha": "rgba(255, 255, 255, 0.25)",
          "--shadow-accent": "#bc6d00",
          "--background-progress-alpha": "rgba(255, 209, 118, 0.32)",
          "--effect-overlay1-alpha": "rgba(255, 202, 0, 0.2)",
          "--effect-overlay2-alpha": "rgba(255, 214, 255, 0)",
          "--shadow-effect-alpha": "rgba(0, 0, 0, 0.08)",
        },
      },
    ],
  },
  content: [
    '../../../../../views/**/*.templ',
  ],
  theme: {
    extend: {
      screens: {
        'navbreak': '1100px',
      },
      fontFamily: {
        sans: ['"Rye"', 'sans-serif'],
      },
    },
    backgroundImage: {
      'bg-lemc': "url('/themes/default/public/imgs/bg.png')",
      'bg-lemc-navbar': "url('/themes/default/public/imgs/bg-nav.svg')",
      'bg-lemc-logo-up': "url('/themes/default/public/imgs/logo-up.png')",
      'bg-lemc-logo-down': "url('/themes/default/public/imgs/logo-down.png')",
      'bg-lemc-logo-normal': "url('/themes/default/public/imgs/logo-normal.png')"
    },
  },
  plugins: [
    require('daisyui'),
  ],
}

