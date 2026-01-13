'use client';

export default function StargazerLogo({ size = 'lg' }: { size?: 'sm' | 'md' | 'lg' }) {
  const sizeClasses = {
    sm: 'text-lg',
    md: 'text-2xl',
    lg: 'text-4xl',
  };

  return (
    <div className={`${sizeClasses[size]} font-semibold text-[#3b82f6] flex items-center gap-2 tracking-tight`}>
      <span className="text-[#60a5fa]">◈</span>
      <span>STARGAZER</span>
    </div>
  );
}
