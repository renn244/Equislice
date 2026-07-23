import { BaseStyles, ThemeProvider } from '@primer/react'
import { RouterProvider } from '@tanstack/react-router'
import { router } from './router'

function App() {
  return (
    <ThemeProvider colorMode="day">
      <BaseStyles className="app-root">
        <RouterProvider router={router} />
      </BaseStyles>
    </ThemeProvider>
  )
}

export default App
