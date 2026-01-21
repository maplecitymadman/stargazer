'use client';

// Spaceship-style icon system
export const Icons = {
  // Status indicators
  healthy: '◉',
  degraded: '◈',
  critical: '▴',
  warning: '▾',
  info: '▸',

  // Actions
  refresh: '⟐',
  scan: '◉',
  execute: '▸',
  loading: '⟐',

  // Systems
  pods: '▸',
  deployments: '◈',
  events: '⟐',
  terminal: '▸',
  output: '▸',
  history: '▸',
  time: '▸',
  network: '◈',
  connections: '⟐',
  check: '◉',
  close: '✕',

  // Navigation
  all: '◈',

  // Branding
  logo: '◈',
  stargazer: '◈',
};

interface IconProps {
  name: keyof typeof Icons;
  className?: string;
  size?: 'sm' | 'md' | 'lg';
}

export function Icon({ name, className = '', size = 'md' }: IconProps) {
  const sizeClasses = {
    sm: 'text-sm',
    md: 'text-base',
    lg: 'text-lg',
  };

  return (
    <span className={`inline-block ${sizeClasses[size]} ${className}`}>
      {Icons[name]}
    </span>
  );
}
