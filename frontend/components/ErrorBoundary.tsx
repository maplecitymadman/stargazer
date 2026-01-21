'use client';

import React, { Component, ErrorInfo, ReactNode } from 'react';
import { Icon } from './SpaceshipIcons';

interface Props {
  children: ReactNode;
  fallback?: ReactNode;
}

interface State {
  hasError: boolean;
  error: Error | null;
}

class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    // Error logged silently - user sees friendly error message
  }

  render() {
    if (this.state.hasError) {
      if (this.props.fallback) {
        return this.props.fallback;
      }

      const errorMessage = this.state.error?.message || 'An unexpected error occurred';
      const isNetworkError = errorMessage.includes('fetch') || errorMessage.includes('network') || errorMessage.includes('Failed to fetch');
      const isTimeoutError = errorMessage.includes('timeout') || errorMessage.includes('timed out');

      return (
        <div className="card rounded-lg p-8 text-center max-w-md mx-auto">
          <Icon name="critical" className="text-[#ef4444] text-4xl mb-4 mx-auto" />
          <h2 className="text-xl font-bold text-[#e4e4e7] mb-2">Something went wrong</h2>
          <p className="text-[#71717a] mb-2 text-sm">
            {errorMessage}
          </p>

          {/* Specific error guidance */}
          {isNetworkError && (
            <div className="mb-4 p-3 bg-[#f59e0b]/10 border border-[#f59e0b]/20 rounded text-left text-xs text-[#f59e0b]">
              <p className="font-semibold mb-1">Network Connection Issue</p>
              <p>• Check your internet connection</p>
              <p>• Verify the API server is running</p>
              <p>• Try refreshing the page</p>
            </div>
          )}

          {isTimeoutError && (
            <div className="mb-4 p-3 bg-[#f59e0b]/10 border border-[#f59e0b]/20 rounded text-left text-xs text-[#f59e0b]">
              <p className="font-semibold mb-1">Request Timed Out</p>
              <p>• The server may be under heavy load</p>
              <p>• Try again in a few moments</p>
              <p>• Check cluster connectivity</p>
            </div>
          )}

          <div className="flex gap-3 justify-center mt-6">
            <button
              onClick={() => {
                this.setState({ hasError: false, error: null });
                window.location.reload();
              }}
              className="px-4 py-2 bg-[#3b82f6] hover:bg-[#2563eb] text-white rounded-lg text-sm transition-all font-medium"
            >
              Reload Page
            </button>
            <button
              onClick={() => this.setState({ hasError: false, error: null })}
              className="px-4 py-2 border border-[rgba(255,255,255,0.08)] hover:bg-[#111118] text-[#a1a1aa] rounded-lg text-sm transition-all"
            >
              Try Again
            </button>
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}

export { ErrorBoundary };
