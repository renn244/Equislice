import * as Sentry from '@sentry/react'
import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import App from './App.tsx'

const sentryDsn = import.meta.env.VITE_SENTRY_DSN
const apiOrigin = new URL(import.meta.env.VITE_API_URL ?? window.location.origin, window.location.origin).origin

if (import.meta.env.PROD && sentryDsn) {
  Sentry.init({
    dsn: sentryDsn,
    environment: import.meta.env.MODE,
    enableLogs: true,
    integrations: [
      Sentry.browserTracingIntegration(),
      Sentry.replayIntegration({
        blockAllMedia: true,
        maskAllInputs: true,
        maskAllText: true,
        networkCaptureBodies: false,
        networkDetailAllowUrls: [],
        networkDetailDenyUrls: [/.*/],
      }),
    ],
    replaysOnErrorSampleRate: 1,
    replaysSessionSampleRate: 0.05,
    sendDefaultPii: false,
    tracePropagationTargets: [window.location.origin, apiOrigin],
    tracesSampleRate: 0.1,
    beforeSend(event) {
      if (!event.request) return event

      delete event.request.cookies
      delete event.request.data

      if (event.request.headers) {
        delete event.request.headers.Authorization
        delete event.request.headers.authorization
      }

      if (event.request.url) event.request.url = event.request.url.split('?')[0]

      return event
    },
    beforeSendSpan(span) {
      for (const [key, value] of Object.entries(span.data)) {
        if (key.includes('url') && typeof value === 'string') span.data[key] = value.split('?')[0]
      }

      return span
    },
  })
}

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App />
  </StrictMode>,
)
