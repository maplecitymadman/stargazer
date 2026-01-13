'use client';

import { useEffect } from 'react';
import { Icon } from '@/components/SpaceshipIcons';

export default function Error({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  useEffect(() => {
    // Log error to console in development
    if (process.env.NODE_ENV === 'development') {
      console.error('Application error:', error);
    }
  }, [error]);

  return (
    <div className="min-h-screen flex items-center justify-center bg-[#0a0a0f] p-4">
      <div className="card rounded-lg p-8 text-center max-w-md w-full">
        <Icon name="critical" className="text-red-400 text-4xl mb-4 mx-auto" />
        <h2 className="text-xl font-bold text-[#e4e4e7] mb-2">Something went wrong</h2>
        <p className="text-[#71717a] mb-4">
          {error.message || 'An unexpected error occurred'}
        </p>
        <button
          onClick={reset}
          className="px-4 py-2 bg-[#3b82f6] hover:bg-[#2563eb] text-white rounded-lg text-sm transition-all"
        >
          Try again
        </button>
      </div>
    </div>
  );
}
