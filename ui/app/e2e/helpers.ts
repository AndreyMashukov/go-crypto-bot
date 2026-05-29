import { type Page, expect } from '@playwright/test'

// Localised copy. Keys mirror the hunico i18n convention — the i18n
// dictionary key IS the literal English text, and other locales just
// supply the translation. The COPY map below is therefore "look up the
// English phrase, get back the locale-specific rendering" — exactly
// what the test needs when asserting visible strings on the page.
const RU: Record<string, string> = {
  'Bot state':          'Состояние бота',
  'USDT balance':       'Баланс USDT',
  'Open positions':     'Открытые позиции',
  'Recent trades':      'Последние сделки',
  'No data':            'Нет данных',
  'Orders':             'Ордера',
  'Active':             'Активные',
  'Pending':            'Ожидающие',
  'Positions':          'Позиции',
  'No orders':          'Ордеров нет',
  'Trade pairs':        'Торговые пары',
  'Bot settings':       'Настройки бота',
  'Configuration':      'Конфигурация',
  'Add pair':           'Добавить пару',
  'Symbol':             'Символ',
  'USDT limit':         'Лимит USDT',
  'Frame interval':     'Интервал кадра',
  'Profit options':     'Опции прибыли',
  'Enabled':            'Активна',
  'Cancel':             'Отмена',
  'Save':               'Сохранить',
  'Prometheus metrics': 'Метрики Prometheus',
  'Dashboard':          'Дашборд',
  'Charts':             'Графики',
}

const EN: Record<string, string> = Object.fromEntries(
  Object.keys(RU).map(k => [k, k]),
)

export const COPY = { ru: RU, en: EN } as const

// Adaptive helper — table either has data rows or shows the empty
// message. Uses `.first()` to be tolerant of multi-card pages (the
// dashboard has two empty cards side by side).
export async function expectTableOrEmpty(page: Page, emptyText: string | RegExp) {
  const hasRows = await page.locator('.v-data-table__tr, tbody tr').count()
  if (hasRows > 0) return
  await expect(page.getByText(emptyText).first()).toBeVisible()
}
