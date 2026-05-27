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
)
