import js from '@eslint/js'
import pluginVue from 'eslint-plugin-vue'
import tseslint from 'typescript-eslint'
import vueParser from 'vue-eslint-parser'
import globals from 'globals'

export default [
  {
    name: 'app/files-to-lint',
    files: ['**/*.{ts,mts,tsx,vue}'],
  },
  {
    name: 'app/files-to-ignore',
    ignores: ['**/dist/**', '**/node_modules/**'],
  },
  js.configs.recommended,
  ...tseslint.configs.recommended,
  ...pluginVue.configs['flat/essential'],
  {
    files: ['**/*.vue'],
    languageOptions: {
      parser: vueParser,
      parserOptions: {
        parser: tseslint.parser,
        sourceType: 'module',
      },
      globals: {
        ...globals.browser,
      },
    },
  },
  {
    files: ['**/*.ts'],
    languageOptions: {
      globals: {
        ...globals.browser,
      },
    },
  },
  {
    rules: {
      'no-console': 'off',
      'no-debugger': 'warn',
      '@typescript-eslint/no-unused-vars': ['warn', { argsIgnorePattern: '^_' }],
    },
  },
]
