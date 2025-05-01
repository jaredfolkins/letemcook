/** @type {import('tailwindcss').Config} */
module.exports = {
  daisyui: {
    themes: [
      {
        banilla: {
          "base-content": "#000000",
          "secondary-content": "#171717",
          "primary": "#c5c5c5",
          "secondary": "#a78bfa",
          "neutral": "#f3f4f6",
          "accent": "#f3f4f6",
          "base-100": "#ffffff",
          "info": "#e1e1e1",
          "success": "#22c55e",
          "warning": "#f59e0b",
          "error": "#ef4444",
          "--color-primary": "#e1e1e1",
          "--color-content-secondary": "#000",
          "--color-inputs-secondary": "#fff",
          "--color-ck-border": "#bdbdbd",
          "--color-monitor-border": "#bdbdbd",
          "--color-navbar-bg": "#f3f4f6",
          "--color-accent": "#fff",
          "--color-accent-focus": "#ececec",
          "--outline-accent": "#f3f4f6",
          "--overlay-light-alpha": "rgba(255, 255, 255, 0.2)",
          "--overlay-light-stronger-alpha": "rgba(255, 255, 255, 0.25)",
          "--shadow-accent": "#f3f4f6",
          "--background-progress-alpha": "rgba(11,11,11,0.32)",
          "--effect-overlay1-alpha": "rgba(20, 184, 166, 0.2)",
          "--effect-overlay2-alpha": "rgba(167, 139, 250, 0)",
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
        sans: ["Inter", "sans-serif"],
      },
    },
    backgroundImage: {
      'bg-lemc': "url('/themes/banilla/public/imgs/bg.png')",
      'bg-lemc-navbar': "url('/themes/banilla/public/imgs/bg-nav.png')",
      'bg-lemc-logo-up': "url('/themes/banilla/public/imgs/logo-up.png')",
      'bg-lemc-logo-down': "url('/themes/banilla/public/imgs/logo-down.png')",
      'bg-lemc-logo-normal': "url('/themes/banilla/public/imgs/logo-normal.png')"
    },
  },
  plugins: [
    require('daisyui'),
  ],
}


