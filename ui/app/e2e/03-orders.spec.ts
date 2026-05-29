import { test, expect } from '@playwright/test'
import { COPY, expectTableOrEmpty } from './helpers'

test.describe('Orders', () => {
  test('renders title + three tabs', async ({ page }) => {
    await page.goto('/orders')
    await expect(page.getByRole('heading', { name: COPY.ru['Orders'] })).toBeVisible()
    await expect(page.getByRole('tab', { name: COPY.ru['Active'] })).toBeVisible()
    await expect(page.getByRole('tab', { name: COPY.ru['Pending'] })).toBeVisible()
    await expect(page.getByRole('tab', { name: COPY.ru['Positions'] })).toBeVisible()
  })

  test('switches between tabs and each renders table-or-empty', async ({ page }) => {
    await page.goto('/orders')
    for (const tabName of [COPY.ru['Active'], COPY.ru['Pending'], COPY.ru['Positions']]) {
      await page.getByRole('tab', { name: tabName }).click()
      await page.waitForTimeout(800)
      await expectTableOrEmpty(page, new RegExp(`${COPY.ru['No orders']}|No orders`, 'i'))
    }
  })
})
