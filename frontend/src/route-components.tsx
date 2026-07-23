import { Outlet } from '@tanstack/react-router'

export function RootLayout() {
  return <Outlet />
}

export function EmptyRoute() {
  return null
}
