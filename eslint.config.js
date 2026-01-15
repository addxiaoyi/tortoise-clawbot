import eslint from "@eslint/js";
import tseslint from "typescript-eslint";
import { fileURLToPath } from "node:url";
import path from "node:path";

const rootDir = path.dirname(fileURLToPath(import.meta.url));

// Lint 范围：src 下全部 .ts、scripts 下全部 .ts、vitest.config.ts；.test.ts 单独放宽规则。
const allTs = [
  "src/**/*.ts",
  "scripts/**/*.ts",
  "extensions/tohelp-openclaw/**/*.ts",
  "vitest.config.ts",
];

const pragmaticTypescriptRules = {
  "preserve-caught-error": "off",
  "@typescript-eslint/no-explicit-any": "off",
  "@typescript-eslint/ban-ts-comment": "off",
  "@typescript-eslint/no-unused-vars": [
    "error",
    {
      argsIgnorePattern: "^_",
      varsIgnorePattern: "^_",
      caughtErrors: "none",
    },
  ],
  "@typescript-eslint/no-require-imports": "off",
  "no-empty": "off",
};

export default tseslint.config(
  eslint.configs.recommended,
  ...tseslint.configs.recommended,
  {
    ignores: [
      "**/node_modules/**",
      "openclaw-main/**",
      "dist/**",
      "coverage/**",
      "cloud/**",
      "templates/**",
      "tortoise/**",
    ],
  },
  {
    files: allTs,
    languageOptions: {
      parserOptions: {
        projectService: true,
        tsconfigRootDir: rootDir,
      },
    },
    rules: pragmaticTypescriptRules,
  },
  {
    files: ["**/*.test.ts"],
    rules: {
      "@typescript-eslint/no-unused-vars": "off",
    },
  },
);
