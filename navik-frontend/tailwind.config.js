/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./App.{js,jsx,ts,tsx}",
    "./app/**/*.{js,jsx,ts,tsx}",
    "./components/**/*.{js,jsx,ts,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        "uber-black": "#000000",
        "uber-gray": "#f8f8f8",
        "uber-blue": "#3b82f6",
      },
    },
  },
};
