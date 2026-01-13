'use client';

interface Breadcrumb {
  label: string;
  onClick?: () => void;
}

interface BreadcrumbsProps {
  items: Breadcrumb[];
}

export default function Breadcrumbs({ items }: BreadcrumbsProps) {
  return (
    <nav className="flex items-center gap-2 text-sm mb-2">
      {items.map((item, index) => (
        <div key={index} className="flex items-center gap-2">
          {index > 0 && <span className="text-[#71717a]">/</span>}
          {item.onClick ? (
            <button
              onClick={item.onClick}
              className="text-[#71717a] hover:text-[#3b82f6] transition-colors"
            >
              {item.label}
            </button>
          ) : (
            <span className={index === items.length - 1 ? 'text-[#e4e4e7] font-medium' : 'text-[#71717a]'}>
              {item.label}
            </span>
          )}
        </div>
      ))}
    </nav>
  );
}
