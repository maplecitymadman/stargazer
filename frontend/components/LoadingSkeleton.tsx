'use client';

/**
 * LoadingSkeleton component provides placeholder UI that mimics content structure
 * Better UX than spinners - users see the layout immediately
 */

interface SkeletonProps {
  variant?: 'text' | 'card' | 'table' | 'list' | 'metric';
  lines?: number;
  className?: string;
}

/**
 * Text skeleton - for paragraphs or text content
 */
export function TextSkeleton({ lines = 3, className = '' }: { lines?: number; className?: string }) {
  return (
    <div className={`space-y-2 ${className}`}>
      {Array.from({ length: lines }).map((_, i) => (
        <div
          key={i}
          className="h-4 bg-[#1a1a24] rounded animate-pulse"
          style={{ width: i === lines - 1 ? '75%' : '100%' }}
        />
      ))}
    </div>
  );
}

/**
 * Card skeleton - for card components
 */
export function CardSkeleton({ className = '' }: { className?: string }) {
  return (
    <div className={`card rounded-lg p-5 ${className}`}>
      <div className="h-6 bg-[#1a1a24] rounded w-1/3 mb-4 animate-pulse" />
      <div className="space-y-3">
        <div className="h-4 bg-[#1a1a24] rounded animate-pulse" />
        <div className="h-4 bg-[#1a1a24] rounded w-5/6 animate-pulse" />
        <div className="h-4 bg-[#1a1a24] rounded w-4/6 animate-pulse" />
      </div>
    </div>
  );
}

/**
 * Table skeleton - for table/list views
 */
export function TableSkeleton({ rows = 5, cols = 4, className = '' }: { rows?: number; cols?: number; className?: string }) {
  return (
    <div className={`space-y-2 ${className}`}>
      {/* Header */}
      <div className="grid gap-3" style={{ gridTemplateColumns: `repeat(${cols}, 1fr)` }}>
        {Array.from({ length: cols }).map((_, i) => (
          <div key={i} className="h-4 bg-[#1a1a24] rounded animate-pulse" />
        ))}
      </div>
      {/* Rows */}
      {Array.from({ length: rows }).map((_, rowIdx) => (
        <div key={rowIdx} className="grid gap-3" style={{ gridTemplateColumns: `repeat(${cols}, 1fr)` }}>
          {Array.from({ length: cols }).map((_, colIdx) => (
            <div
              key={colIdx}
              className="h-10 bg-[#1a1a24] rounded animate-pulse"
              style={{ animationDelay: `${rowIdx * 50}ms` }}
            />
          ))}
        </div>
      ))}
    </div>
  );
}

/**
 * List skeleton - for list items
 */
export function ListSkeleton({ items = 5, className = '' }: { items?: number; className?: string }) {
  return (
    <div className={`space-y-3 ${className}`}>
      {Array.from({ length: items }).map((_, i) => (
        <div key={i} className="flex items-center gap-3">
          <div className="h-10 w-10 bg-[#1a1a24] rounded animate-pulse" />
          <div className="flex-1 space-y-2">
            <div className="h-4 bg-[#1a1a24] rounded w-3/4 animate-pulse" />
            <div className="h-3 bg-[#1a1a24] rounded w-1/2 animate-pulse" />
          </div>
        </div>
      ))}
    </div>
  );
}

/**
 * Metric skeleton - for dashboard metrics
 */
export function MetricSkeleton({ className = '' }: { className?: string }) {
  return (
    <div className={`card rounded-lg p-5 ${className}`}>
      <div className="h-3 bg-[#1a1a24] rounded w-1/4 mb-3 animate-pulse" />
      <div className="h-8 bg-[#1a1a24] rounded w-1/3 mb-2 animate-pulse" />
      <div className="h-3 bg-[#1a1a24] rounded w-1/2 animate-pulse" />
      <div className="mt-4 h-1 bg-[#1a1a24] rounded-full animate-pulse" />
    </div>
  );
}

/**
 * Main LoadingSkeleton component - choose variant
 */
export default function LoadingSkeleton({ variant = 'text', lines = 3, className = '' }: SkeletonProps) {
  switch (variant) {
    case 'text':
      return <TextSkeleton lines={lines} className={className} />;
    case 'card':
      return <CardSkeleton className={className} />;
    case 'table':
      return <TableSkeleton className={className} />;
    case 'list':
      return <ListSkeleton className={className} />;
    case 'metric':
      return <MetricSkeleton className={className} />;
    default:
      return <TextSkeleton lines={lines} className={className} />;
  }
}
