const js = require("@eslint/js");
const tsEslint = require("@typescript-eslint/eslint-plugin");
const tsParser = require("@typescript-eslint/parser");

module.exports = [
  js.configs.recommended,
  {
    files: ["**/*.ts", "**/*.tsx"],
    languageOptions: {
      parser: tsParser,
      parserOptions: {
        ecmaVersion: 2020,
        sourceType: "module",
      },
      globals: {
        console: "readonly",
        process: "readonly",
        Buffer: "readonly",
        __dirname: "readonly",
        __filename: "readonly",
        global: "readonly",
        module: "readonly",
        require: "readonly",
        exports: "readonly",
        setTimeout: "readonly",
        clearTimeout: "readonly",
        setInterval: "readonly",
        clearInterval: "readonly",
        AbortController: "readonly"
      }
    },
    plugins: {
      "@typescript-eslint": tsEslint
    },
    rules: {
      "semi": "warn",
      "curly": "off", 
      "eqeqeq": "warn",
      "no-throw-literal": "warn",
      "no-unused-vars": "off",
      "no-useless-catch": "off",
      "no-useless-escape": "off",
      "@typescript-eslint/no-unused-vars": ["warn", { "argsIgnorePattern": "^_", "varsIgnorePattern": "^_" }]
    }
  },
  {
    ignores: ["out/**", "dist/**", "**/*.d.ts"]
  }
];