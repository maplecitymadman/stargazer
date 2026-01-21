import type { Metadata } from 'next'
import './globals.css'
import Toaster from '@/components/Toaster'

export const metadata: Metadata = {
  title: 'Stargazer - K8S CLUSTER OBSERVATORY',
  description: 'Premium Kubernetes Cluster Observatory',
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
        <Toaster />
      </body>
    </html>
  )
}
