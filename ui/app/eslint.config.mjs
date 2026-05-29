// @ts-check
import withNuxt from './.nuxt/eslint.config.mjs'
import messDetector from '@amashukov/eslint-plugin-mess-detector'

export default withNuxt(
  {
    files: ['app/**/*.{vue,ts,js,mjs}', 'tests/**/*.ts'],
    plugins: { 'mess-detector': messDetector },
    rules: messDetector.configs.recommended.rules,
  },
  {
    rules: {
      '@typescript-eslint/no-explicit-any': 'error',
      '@typescript-eslint/no-unused-vars': ['error', { argsIgnorePattern: '^_' }],
      'vue/no-unused-vars': 'error',
      'vue/no-mutating-props': 'error',
      'vue/multi-word-component-names': 'off',
      'no-console': ['error', { allow: ['warn', 'error'] }],
      'no-debugger': 'error',
      'prefer-const': 'error',
      'no-var': 'error',
      'eqeqeq': ['error', 'always'],
    },
  },
)
