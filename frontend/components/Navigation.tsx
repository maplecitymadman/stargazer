'use client';

import { useState } from 'react';
import { Icon } from './SpaceshipIcons';
import StargazerLogo from './StargazerLogo';
import ContextSelector from './ContextSelector';
import NamespaceSelector from './NamespaceSelector';

interface NavigationProps {
  currentSection: string;
  onSectionChange: (section: string) => void;
  namespace: string;
  onNamespaceChange: (namespace: string) => void;
  onContextChange?: () => void;
}

export default function Navigation({ currentSection, onSectionChange, namespace, onNamespaceChange, onContextChange }: NavigationProps) {
  const [isOpen, setIsOpen] = useState(true);

  const sections = [
    { id: 'dashboard', label: 'Dashboard', icon: 'info' as const },
    { id: 'topology', label: 'Service Topology', icon: 'network' as const },
    { id: 'events', label: 'Events', icon: 'events' as const },
    { id: 'settings', label: 'Settings', icon: 'info' as const },
  ];

  return (
    <>
      {/* Mobile toggle button */}
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="fixed top-4 left-4 z-50 lg:hidden px-3 py-2 glass border border-[rgba(255,255,255,0.08)] rounded-md text-[#e4e4e7] hover:border-[rgba(255,255,255,0.12)] transition-all"
      >
        <Icon name="info" />
      </button>

      {/* Sidebar Navigation */}
      <nav
        className={`fixed left-0 top-0 h-full w-64 glass border-r border-[rgba(255,255,255,0.08)] z-40 transition-transform duration-300 ease-out ${
          isOpen ? 'translate-x-0' : '-translate-x-full'
        } lg:translate-x-0`}
      >
        <div className="flex flex-col h-full">
          {/* Logo/Header */}
          <div className="p-5 border-b border-[rgba(255,255,255,0.08)]">
            <div className="mb-3">
              <StargazerLogo size="sm" />
            </div>
            <div className="text-xs text-[#71717a] mb-3 tracking-wide">
              K8s Cluster Observatory
            </div>
            <div className="text-xs text-[#3b82f6]/60 mb-3 font-medium">
            </div>
            <div className="space-y-2">
              <ContextSelector onContextChange={onContextChange} />
              <NamespaceSelector 
                currentNamespace={namespace} 
                onNamespaceChange={onNamespaceChange}
              />
            </div>
          </div>

          {/* Navigation Items */}
          <div className="flex-1 overflow-y-auto p-3 space-y-1">
            {sections.map((section) => (
              <button
                key={section.id}
                onClick={() => {
                  onSectionChange(section.id);
                  setIsOpen(false);
                }}
                className={`w-full flex items-center gap-3 px-4 py-2.5 rounded-md text-sm transition-all ${
                  currentSection === section.id
                    ? 'bg-[#3b82f6] text-white'
                    : 'text-[#a1a1aa] hover:text-[#e4e4e7] hover:bg-[#111118]'
                }`}
              >
                <Icon
                  name={section.icon}
                  className={currentSection === section.id ? 'text-white' : 'text-[#71717a]'}
                  size="sm"
                />
                <span className="font-medium">{section.label}</span>
              </button>
            ))}
          </div>

          {/* Footer */}
          <div className="p-4 border-t border-[rgba(255,255,255,0.08)]">
            <div className="text-xs text-[#71717a] text-center">
              <div className="mb-1">Stargazer v0.1.0</div>
            </div>
          </div>
        </div>
      </nav>

      {/* Overlay for mobile */}
      {isOpen && (
        <div
          className="fixed inset-0 bg-black/60 backdrop-blur-sm z-30 lg:hidden transition-opacity"
          onClick={() => setIsOpen(false)}
        />
      )}
    </>
  );
}
