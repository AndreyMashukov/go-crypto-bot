import { test, expect } from '@playwright/test'
import { COPY } from './helpers'

test.describe('i18n locale switching', () => {
  test('default locale is Russian, /dashboard shows Russian title', async ({ page }) => {
    await page.goto('/dashboard')
    await expect(page.getByRole('heading', { name: COPY.ru['Bot state'] })).toBeVisible()
  })

  test('/en/dashboard renders English copy', async ({ page }) => {
    await page.goto('/en/dashboard')
    await expect(page.getByRole('heading', { name: COPY.en['Bot state'] })).toBeVisible()
    await expect(page.getByText(COPY.en['USDT balance'])).toBeVisible()
    await expect(page.getByText(COPY.en['Open positions'])).toBeVisible()
    await expect(page.getByText(COPY.en['Recent trades'])).toBeVisible()
  })

  test('/en sidebar shows English nav labels', async ({ page }) => {
    await page.goto('/en/dashboard')
    for (const en of ['Dashboard', 'Orders', 'Configuration', 'Charts'] as const) {
      await expect(page.getByRole('link', { name: COPY.en[en] })).toBeVisible()
    }
  })

  test('every advertised locale serves its own dashboard URL with HTTP 200', async ({ page }) => {
    const locales = ['en', 'de', 'fr', 'es', 'ja', 'zh-hans', 'ko']
    for (const code of locales) {
      const response = await page.goto(`/${code}/dashboard`)
      expect(response?.status(), `/${code}/dashboard`).toBeLessThan(400)
    }
  })
})
