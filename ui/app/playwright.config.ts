import { defineConfig, devices } from '@playwright/test'

// E2E tests run against a live stack — `make stack-up` must be green.
// Default target is the ui-app host port (Phase J port-remap = 13000).
// Override with E2E_BASE_URL=... for a different target.
const baseURL = process.env.E2E_BASE_URL ?? 'http://127.0.0.1:13000'

export default defineConfig({
  testDir: './e2e',
  fullyParallel: false,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: 1,
  reporter: process.env.CI ? 'github' : 'list',
  timeout: 30_000,
  expect: { timeout: 8_000 },
  use: {
    baseURL,
    headless: true,
    viewport: { width: 1440, height: 900 },
    ignoreHTTPSErrors: true,
    screenshot: 'only-on-failure',
    trace: 'retain-on-failure',
    // The Vuetify templates stamp `data-test` on every actionable
    // node; getByTestId looks for `data-testid` by default. Re-point
    // it once here so the entire suite reads `data-test` and the
    // templates stay clean.
    testIdAttribute: 'data-test',
    // Locale lock: Playwright's headless Chromium defaults to en-US,
    // which makes @nuxtjs/i18n's detectBrowserLanguage redirect every
    // landing-page hit to /en/. The tests deliberately exercise the
    // ru-default + /en/ prefix paths separately, so we pin the browser
    // preference to ru-RU here and let the per-test fixtures override
    // when they want to assert another locale's behaviour.
    locale: 'ru-RU',
    extraHTTPHeaders: {
      'Accept-Language': 'ru-RU,ru;q=0.9',
    },
  },
  projects: [
    { name: 'chromium', use: devices['Desktop Chrome'] },
  ],
})
