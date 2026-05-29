import { test, expect } from '@playwright/test'
import { COPY } from './helpers'

// Charts page renders a fixed set of PromQL panels; we only assert
// each panel is in a recognisable state (chart canvas OR loading/state
// overlay) without making any claim about the live Prometheus series.
const PANEL_TITLES_RU = [
  'Сбои публикации тика',
  'Сброшенные тики',
  'Переподключения pub/sub',
  'Задержка обработки тика, сек',
  'Время обработки, сек',
  'Время размещения ордера, сек',
  'Реализованный P&L',
  'Открытие позиции, сек',
]

test.describe('Charts (Prometheus)', () => {
  test('renders 8 panels with locale-correct titles', async ({ page }) => {
    await page.goto('/charts')
    await expect(page.getByRole('heading', { name: COPY.ru['Prometheus metrics'] })).toBeVisible()
    for (const title of PANEL_TITLES_RU) {
      await expect(page.getByText(title)).toBeVisible()
    }
  })

  test('each panel resolves to chart OR state overlay', async ({ page }) => {
    await page.goto('/charts')
    await page.waitForTimeout(3000)

    const states = await page.locator('.metric-chart').evaluateAll((nodes) => {
      return nodes.map(n => {
        const stateEl = n.querySelector('.metric-chart__state')
        if (stateEl) return 'state'
        const canvas = n.querySelector('canvas')
        if (canvas) return 'chart'
        return 'unknown'
      })
    })

    expect(states.length).toBeGreaterThanOrEqual(8)
    for (const s of states) {
      expect(['state', 'chart']).toContain(s)
    }
  })
})
