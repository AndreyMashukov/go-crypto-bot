import { test, expect } from '@playwright/test'
import { COPY, expectTableOrEmpty } from './helpers'

test.describe('Dashboard', () => {
  test('renders three cards regardless of backend data state', async ({ page }) => {
    await page.goto('/dashboard')
    await expect(page.getByRole('heading', { name: COPY.ru['Bot state'] })).toBeVisible()
    await expect(page.getByText(COPY.ru['USDT balance'])).toBeVisible()
    await expect(page.getByText(COPY.ru['Open positions'])).toBeVisible()
    await expect(page.getByText(COPY.ru['Recent trades'])).toBeVisible()
  })

  test('balance card holds either em-dash or a USDT figure', async ({ page }) => {
    await page.goto('/dashboard')
    const balanceBlock = page.getByText(COPY.ru['USDT balance']).locator('xpath=ancestor::div[contains(@class,"v-card")]')
    await expect(balanceBlock).toBeVisible()
    const text = (await balanceBlock.innerText()).trim()
    expect(text.includes('—') || /\d+\.\d{2}\s*USDT/.test(text)).toBeTruthy()
  })

  test('positions + recent trades render as table-or-empty', async ({ page }) => {
    await page.goto('/dashboard')
    await page.getByText(COPY.ru['Open positions']).waitFor()
    await expectTableOrEmpty(page, new RegExp(`${COPY.ru['No data']}|No data`, 'i'))
  })
})
