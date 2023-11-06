/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./public/**/*.html","./public/**/*.tpl", "./api/**/*.go"],
  theme: {
    extend: {},
  },
  plugins: [require("@tailwindcss/typography"), require("daisyui")],
}

