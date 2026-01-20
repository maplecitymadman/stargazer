'use client';

import { useState } from 'react';
import { Icon } from './SpaceshipIcons';
import StargazerLogo from './StargazerLogo';
import ContextSelector from './ContextSelector';
import NamespaceSelector from './NamespaceSelector';

interface NavigationProps {
  currentSection: string;
  currentSubsection?: string;
  onSectionChange: (section: string, subsection?: string) => void;
  namespace: string;
  onNamespaceChange: (namespace: string) => void;
  onContextChange?: () => void;
}

interface NavSection {
  id: string;
  label: string;
  icon: string;
  subsections?: { id: string; label: string }[];
}

export default function Navigation({ 
  currentSection, 
  currentSubsection,
  onSectionChange, 
  namespace, 
  onNamespaceChange, 
  onContextChange 
}: NavigationProps) {
  const [isOpen, setIsOpen] = useState(true);
  const [expandedSections, setExpandedSections] = useState<Set<string>>(new Set([currentSection]));

  const sections: NavSection[] = [
    { 
      id: 'dashboard', 
      label: 'Dashboard', 
      icon: 'info' as const 
    },
    { 
      id: 'traffic-analysis', 
      label: 'Traffic Analysis', 
      icon: 'network' as const,
      subsections: [
        { id: 'topology', label: 'Service Topology' },
        { id: 'path-trace', label: 'Path Tracer' },
        { id: 'ingress', label: 'Ingress Traffic' },
        { id: 'egress', label: 'Egress Traffic' },
      ]
    },
    { 
      id: 'network-policies', 
      label: 'Network Policies', 
      icon: 'scan' as const,
      subsections: [
        { id: 'view', label: 'View Policies' },
        { id: 'build', label: 'Build Policy' },
        { id: 'test', label: 'Test Policy' },
      ]
    },
    { 
      id: 'compliance', 
      label: 'Compliance', 
      icon: 'scan' as const,
      subsections: [
        { id: 'score', label: 'Compliance Score' },
        { id: 'recommendations', label: 'Recommendations' },
      ]
    },
    { 
      id: 'troubleshooting', 
      label: 'Troubleshooting', 
      icon: 'critical' as const,
      subsections: [
        { id: 'blocked', label: 'Blocked Connections' },
        { id: 'services', label: 'Services with Issues' },
      ]
    },
    { 
      id: 'events', 
      label: 'Events', 
      icon: 'events' as const 
    },
    { 
      id: 'settings', 
      label: 'Settings', 
      icon: 'info' as const 
    },
  ];

  const toggleSection = (sectionId: string) => {
    const newExpanded = new Set(expandedSections);
    if (newExpanded.has(sectionId)) {
      newExpanded.delete(sectionId);
    } else {
      newExpanded.add(sectionId);
    }
    setExpandedSections(newExpanded);
  };

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
              Network Troubleshooting
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
            {sections.map((section) => {
              const isActive = currentSection === section.id;
              const isExpanded = expandedSections.has(section.id);
              const hasSubsections = section.subsections && section.subsections.length > 0;

              return (
                <div key={section.id}>
                  <button
                    onClick={() => {
                      if (hasSubsections) {
                        toggleSection(section.id);
                        // Navigate to section overview (no subsection)
                        onSectionChange(section.id);
                      } else {
                        onSectionChange(section.id);
                        setIsOpen(false);
                      }
                    }}
                    className={`w-full flex items-center justify-between px-4 py-2.5 rounded-md text-sm transition-all cursor-pointer active:scale-[0.98] ${
                      isActive && !currentSubsection
                        ? 'bg-[#3b82f6] text-white'
                        : 'text-[#a1a1aa] hover:text-[#e4e4e7] hover:bg-[#111118]'
                    }`}
                    aria-label={`Navigate to ${section.label}`}
                  >
                    <div className="flex items-center gap-3">
                      <Icon
                        name={section.icon as any}
                        className={isActive && !currentSubsection ? 'text-white' : 'text-[#71717a]'}
                        size="sm"
                      />
                      <span className="font-medium">{section.label}</span>
                    </div>
                    {hasSubsections && (
                      <span className={`text-[#71717a] transition-transform inline-block ${isExpanded ? 'rotate-90' : ''}`}>
                        <Icon name="info" size="sm" />
                      </span>
                    )}
                  </button>

                  {/* Sub-sections */}
                  {hasSubsections && isExpanded && (
                    <div className="ml-4 mt-1 space-y-1 border-l border-[rgba(255,255,255,0.08)] pl-3">
                      {section.subsections!.map((subsection) => {
                        const isSubActive = isActive && currentSubsection === subsection.id;
                        return (
                          <button
                            key={subsection.id}
                            onClick={() => {
                              onSectionChange(section.id, subsection.id);
                              setIsOpen(false);
                            }}
                            className={`w-full flex items-center gap-2 px-3 py-2 rounded-md text-xs transition-all cursor-pointer active:scale-[0.98] ${
                              isSubActive
                                ? 'bg-[#3b82f6]/20 text-[#3b82f6] border border-[#3b82f6]/30'
                                : 'text-[#71717a] hover:text-[#e4e4e7] hover:bg-[#111118]'
                            }`}
                            aria-label={`Navigate to ${subsection.label}`}
                          >
                            <span className="w-1 h-1 rounded-full bg-current opacity-50"></span>
                            {subsection.label}
                          </button>
                        );
                      })}
                    </div>
                  )}
                </div>
              );
            })}
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
