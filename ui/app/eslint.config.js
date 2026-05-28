import mess from '@amashukov/eslint-plugin-mess-detector'
import withNuxt from './.nuxt/eslint.config.mjs'

export default withNuxt(
  {
    ignores: [
      '.nuxt/**',
      '.output/**',
      'dist/**',
      'node_modules/**',
      'public/**',
    ],
  },
  {
    files: ['**/*.{js,mjs,cjs,ts,vue}'],
    plugins: { 'mess-detector': mess },
    rules: mess.configs.recommended.rules,
  },
  {
    files: ['**/*.{ts,vue}'],
    rules: {
      '@typescript-eslint/no-explicit-any': 'warn',
    },
  },
  {
    files: ['nuxt.config.ts'],
    rules: {
      'mess-detector/no-suppression-comments': 'off',
      'mess-detector/no-env-branch': 'off',
      'mess-detector/no-process-env-outside-config': 'off',
    },
  },
)
