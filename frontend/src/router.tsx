import {
  createRootRoute,
  createRoute,
  createRouter,
  lazyRouteComponent,
} from '@tanstack/react-router'
import { LandingPage } from './pages/landing-page'
import { RootLayout } from './route-components'

const rootRoute = createRootRoute({
  component: RootLayout,
})

const landingRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/',
  component: LandingPage,
})

const cutterRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: 'cutter',
  component: lazyRouteComponent(() => import('./features/cutter/pages/CutterPage'), 'CutterPage'),
})

const routeTree = rootRoute.addChildren([landingRoute, cutterRoute])

export const router = createRouter({ routeTree })

declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router
  }
}
