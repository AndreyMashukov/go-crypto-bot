import { test, expect } from '@playwright/test'
import { COPY } from './helpers'

test.describe('Landing & navigation', () => {
  test('root redirects to dashboard in default locale (ru)', async ({ page }) => {
    await page.goto('/', { waitUntil: 'networkidle' })
    await expect(page).toHaveURL(/\/dashboard$/)
    await expect(page.getByRole('heading', { name: COPY.ru['Bot state'] })).toBeVisible()
  })

  test('sidebar exposes 4 navigation items', async ({ page }) => {
    await page.goto('/dashboard')
    for (const en of ['Dashboard', 'Orders', 'Configuration', 'Charts'] as const) {
      await expect(page.getByRole('link', { name: COPY.ru[en] })).toBeVisible()
    }
  })

  test('clicking each nav item lands on the matching page', async ({ page }) => {
    await page.goto('/dashboard')
    await page.getByRole('link', { name: COPY.ru['Orders'] }).click()
    await expect(page).toHaveURL(/\/orders$/)
    await page.getByRole('link', { name: COPY.ru['Configuration'] }).click()
    await expect(page).toHaveURL(/\/config$/)
    await page.getByRole('link', { name: COPY.ru['Charts'] }).click()
    await expect(page).toHaveURL(/\/charts$/)
    await page.getByRole('link', { name: COPY.ru['Dashboard'] }).click()
    await expect(page).toHaveURL(/\/dashboard$/)
  })
})
