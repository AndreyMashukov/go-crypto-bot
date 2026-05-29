// Deep config flow — exercises every form control on /config:
//   - frame interval dropdown
//   - history interval dropdown
//   - profit-option array editor (add / unit dropdown / trigger switch / remove)
//   - extra-charge array editor (add / remove)
//   - is-enabled switch
//   - bot settings tab: hold score input + trade-stack-sorting dropdown + save
//
// All API calls go through page.route fulfill so the test runs identically
// whether the bot is up, restarting, or returning 502 — what matters is the
// shape of the payload the UI sends.

import { test, expect, type Route, type Request } from '@playwright/test'

type Recorded = { method: string; url: string; body: unknown }

function record(route: Route, log: Recorded[]): void {
  const req = route.request()
  const body = req.postData()
  log.push({
    method: req.method(),
    url: req.url(),
    body: body ? JSON.parse(body) as unknown : null,
  })
  route.fulfill({ status: 200, contentType: 'application/json', body: '{}' })
}

const SEED_LIMIT = {
  symbol: 'BTCUSDT',
  USDTLimit: 100,
  minPrice: 0.0001,
  minQuantity: 0.0001,
  minNotional: 10,
  isEnabled: true,
  minPriceMinutesPeriod: 200,
  frameInterval: '2h',
  framePeriod: 20,
  buyPriceHistoryCheckInterval: '1d',
  buyPriceHistoryCheckPeriod: 14,
  profitOptions: [],
  extraChargeOptions: [],
}

const HEALTH = {
  bot: {
    id: 1,
    botUuid: '00000000-0000-0000-0000-000000000001',
    exchange: 'binance',
    isMasterBot: false,
    tradeStackSorting: 'percent',
    holdScore: 75,
  },
  binanceStatus: 'ok', dbStatus: 'ok', redisStatus: 'ok', mlStatus: 'ready',
  updates: {}, orderBook: {}, dateTimeNow: '',
}

async function wireMocks(page: Parameters<Parameters<typeof test>[1]>[0]['page']) {
  const recorded: Recorded[] = []

  // Catch-all first; more specific routes registered after it take
  // priority because Playwright tries handlers in reverse registration
  // order. The catch-all only fires for paths we didn't explicitly
  // mock so a flaky live backend can't fail the test.
  await page.route('**/api/bot/**', (route: Route, request: Request) => {
    const u = new URL(request.url())
    if (request.method() === 'GET' && (u.pathname.endsWith('/list') || u.pathname.endsWith('/check'))) {
      route.fulfill({ status: 200, contentType: 'application/json', body: '[]' })
      return
    }
    route.fulfill({ status: 200, contentType: 'application/json', body: '{}' })
  })

  await page.route('**/api/bot/trade/limit/list*', (r) => {
    r.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify([SEED_LIMIT]) })
  })
  await page.route('**/api/bot/trade/limit/create*', (r) => record(r, recorded))
  await page.route('**/api/bot/trade/limit/update*', (r) => record(r, recorded))
  await page.route('**/api/bot/trade/limit/switch/**', (r) => record(r, recorded))
  await page.route('**/api/bot/health/check*', (r) => {
    r.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify(HEALTH) })
  })
  await page.route('**/api/bot/bot/update*', (r) => record(r, recorded))

  return recorded
}

test.describe('Deep trade-pair configuration', () => {
  test('create a new pair: fills every dropdown + array editor + saves', async ({ page }) => {
    const recorded = await wireMocks(page)
    await page.goto('/config')

    await page.getByTestId('add-pair-btn').click()
    await expect(page.getByTestId('pair-dialog')).toBeVisible()

    // scalar inputs
    await page.getByTestId('form-symbol').locator('input').fill('DOTUSDT')
    await page.getByTestId('form-usdt-limit').locator('input').fill('250')
    await page.getByTestId('form-frame-period').locator('input').fill('30')
    await page.getByTestId('form-history-period').locator('input').fill('21')

    // frame interval dropdown — pick "4h"
    await page.getByTestId('form-frame-interval').click()
    await page.getByRole('option', { name: '4h', exact: true }).click()

    // history interval dropdown — pick "1w"
    await page.getByTestId('form-history-interval').click()
    await page.getByRole('option', { name: '1w', exact: true }).click()

    // add a profit option, fill values, flip trigger
    await page.getByTestId('add-profit-option').click()
    await page.getByTestId('profit-value-0').locator('input').fill('2')
    await page.getByTestId('profit-unit-0').click()
    await page.getByRole('option', { name: 'h', exact: true }).click()
    await page.getByTestId('profit-percent-0').locator('input').fill('1.5')
    await page.getByTestId('profit-trigger-0').locator('input').check()

    // second profit option
    await page.getByTestId('add-profit-option').click()
    await page.getByTestId('profit-percent-1').locator('input').fill('3')

    // add extra-charge option
    await page.getByTestId('add-extra-charge').click()
    await page.getByTestId('extra-percent-0').locator('input').fill('-2.5')
    await page.getByTestId('extra-amount-0').locator('input').fill('25')

    await page.getByTestId('form-save').click()
    await expect(page.getByTestId('pair-dialog')).not.toBeVisible()

    const create = recorded.find(r => r.url.includes('/trade/limit/create'))
    expect(create, 'create request not recorded').toBeTruthy()
    expect(create!.method).toBe('POST')
    const body = create!.body as Record<string, unknown>
    expect(body.symbol).toBe('DOTUSDT')
    expect(body.USDTLimit).toBe(250)
    expect(body.frameInterval).toBe('4h')
    expect(body.framePeriod).toBe(30)
    expect(body.buyPriceHistoryCheckInterval).toBe('1w')
    expect(body.buyPriceHistoryCheckPeriod).toBe(21)
    expect(body.isEnabled).toBe(true)

    const profitOptions = body.profitOptions as Array<Record<string, unknown>>
    expect(profitOptions).toHaveLength(2)
    expect(profitOptions[0]).toMatchObject({
      optionValue: 2, optionUnit: 'h', optionPercent: 1.5, isTriggerOption: true,
    })
    expect(profitOptions[1]).toMatchObject({ optionPercent: 3 })

    const extra = body.extraChargeOptions as Array<Record<string, unknown>>
    expect(extra).toHaveLength(1)
    expect(extra[0]).toMatchObject({ percent: -2.5, amountUsdt: 25 })
  })

  test('edit existing pair preserves symbol (disabled), accepts new values', async ({ page }) => {
    const recorded = await wireMocks(page)
    await page.goto('/config')

    await page.getByTestId('edit-BTCUSDT').click()
    await expect(page.getByTestId('pair-dialog')).toBeVisible()
    await expect(page.getByTestId('form-symbol').locator('input')).toBeDisabled()

    await page.getByTestId('form-usdt-limit').locator('input').fill('500')
    await page.getByTestId('form-frame-interval').click()
    await page.getByRole('option', { name: '6h', exact: true }).click()
    await page.getByTestId('form-save').click()

    const upd = recorded.find(r => r.url.includes('/trade/limit/update'))
    expect(upd).toBeTruthy()
    const body = upd!.body as Record<string, unknown>
    expect(body.symbol).toBe('BTCUSDT')
    expect(body.USDTLimit).toBe(500)
    expect(body.frameInterval).toBe('6h')
  })

  test('toggle is-enabled switch in the table calls /switch/{symbol}', async ({ page }) => {
    const recorded = await wireMocks(page)
    await page.goto('/config')

    await page.getByTestId('toggle-BTCUSDT').locator('input').click()
    await page.waitForTimeout(500)

    const sw = recorded.find(r => r.url.includes('/trade/limit/switch/BTCUSDT'))
    expect(sw).toBeTruthy()
    expect(sw!.method).toBe('PUT')
    expect((sw!.body as { isEnabled: boolean }).isEnabled).toBe(false)
  })

  test('remove a profit option drops it from the payload', async ({ page }) => {
    const recorded = await wireMocks(page)
    await page.goto('/config')

    await page.getByTestId('add-pair-btn').click()
    await page.getByTestId('form-symbol').locator('input').fill('XRPUSDT')
    await page.getByTestId('add-profit-option').click()
    await page.getByTestId('add-profit-option').click()
    await page.getByTestId('remove-profit-0').click()
    await page.getByTestId('form-save').click()

    const create = recorded.find(r => r.url.includes('/trade/limit/create'))
    const body = create!.body as Record<string, unknown>
    expect((body.profitOptions as unknown[])).toHaveLength(1)
  })
})

test.describe('Deep bot-settings flow', () => {
  test('change hold score + trade-stack-sorting dropdown, save → PUT /bot/update', async ({ page }) => {
    const recorded = await wireMocks(page)
    await page.goto('/config')

    await page.getByTestId('tab-bot').click()
    await expect(page.getByTestId('hold-score-input')).toBeVisible()

    const holdInput = page.getByTestId('hold-score-input').locator('input')
    await expect(holdInput).toHaveValue('75')

    await holdInput.fill('60')

    await page.getByTestId('stack-sorting-select').click()
    await page.getByRole('option', { name: 'diff', exact: true }).click()

    await page.getByTestId('bot-save-btn').click()

    const upd = recorded.find(r => r.url.includes('/bot/bot/update'))
    expect(upd, 'bot/update not recorded').toBeTruthy()
    const body = upd!.body as Record<string, unknown>
    expect(body.holdScore).toBe(60)
    expect(body.tradeStackSorting).toBe('diff')
  })

  test('dropdown options match the model.TradeStackSorting enum', async ({ page }) => {
    await wireMocks(page)
    await page.goto('/config')
    await page.getByTestId('tab-bot').click()

    await page.getByTestId('stack-sorting-select').click()
    const items = page.getByRole('option')
    await expect(items).toHaveCount(2)
    await expect(items.nth(0)).toHaveText(/percent/)
    await expect(items.nth(1)).toHaveText(/diff/)
  })
})
