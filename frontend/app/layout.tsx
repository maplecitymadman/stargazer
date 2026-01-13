import type { Metadata } from 'next'
import './globals.css'

export const metadata: Metadata = {
  title: 'Stargazer - Kubernetes Troubleshooting',
  description: 'Modern Kubernetes troubleshooting dashboard',
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en">
      <body>
        {children}
      </body>
    </html>
  )
}
