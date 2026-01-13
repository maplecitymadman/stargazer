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

      return (
        <div className="card rounded-lg p-8 text-center">
          <Icon name="critical" className="text-red-400 text-4xl mb-4" />
          <h2 className="text-xl font-bold text-[#e4e4e7] mb-2">Something went wrong</h2>
          <p className="text-[#71717a] mb-4">
            {this.state.error && this.state.error.message ? this.state.error.message : 'An unexpected error occurred'}
          </p>
          <button
            onClick={() => this.setState({ hasError: false, error: null })}
            className="px-4 py-2 bg-[#3b82f6] hover:bg-[#2563eb] text-white rounded-lg text-sm transition-all"
          >
            Try again
          </button>
        </div>
      );
    }

    return this.props.children;
  }
}

export { ErrorBoundary };
