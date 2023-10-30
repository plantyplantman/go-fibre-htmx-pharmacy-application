/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./public/**/*.html","./public/**/*.tpl"],
  theme: {
    extend: {},
  },
  plugins: [require("@tailwindcss/typography"), require("daisyui")],
}

