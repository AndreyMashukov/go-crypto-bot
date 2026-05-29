import { test, expect } from '@playwright/test'
import { COPY } from './helpers'

test.describe('Config — surface', () => {
  test('renders title + add-pair button + 2 tabs', async ({ page }) => {
    await page.goto('/config')
    await expect(page.getByRole('heading', { name: COPY.ru['Configuration'] })).toBeVisible()
    await expect(page.getByRole('tab', { name: COPY.ru['Trade pairs'] })).toBeVisible()
    await expect(page.getByRole('tab', { name: COPY.ru['Bot settings'] })).toBeVisible()
    await expect(page.getByRole('button', { name: COPY.ru['Add pair'] })).toBeVisible()
  })

  test('add-pair dialog opens with empty form fields and closes on cancel', async ({ page }) => {
    await page.goto('/config')
    await page.getByTestId('add-pair-btn').click()
    await expect(page.getByTestId('pair-dialog')).toBeVisible()
    await expect(page.getByTestId('form-symbol').locator('input')).toHaveValue('')
    await page.getByTestId('form-cancel').click()
    await expect(page.getByTestId('pair-dialog')).not.toBeVisible()
  })

  test('table headers render in Russian regardless of row count', async ({ page }) => {
    await page.goto('/config')
    for (const en of ['Symbol', 'USDT limit', 'Frame interval', 'Profit options', 'Enabled'] as const) {
      await expect(page.getByRole('columnheader', { name: COPY.ru[en] })).toBeVisible()
    }
  })
})
