'use client';

import { useState } from 'react';
import { Icon } from './SpaceshipIcons';
import StargazerLogo from './StargazerLogo';
import ContextSelector from './ContextSelector';
import NamespaceSelector from './NamespaceSelector';
import GlobalResourceSearch from './GlobalResourceSearch';

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
  const [showSearch, setShowSearch] = useState(false);

  const sections: NavSection[] = [
    {
      id: 'dashboard',
      label: 'Dashboard',
      icon: 'info'
    },
    {
      id: 'traffic-analysis',
      label: 'Traffic Analysis',
      icon: 'network',
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
      icon: 'scan',
      subsections: [
        { id: 'view', label: 'View Policies' },
        { id: 'build', label: 'Build Policy' },
        { id: 'test', label: 'Test Policy' },
      ]
    },
    {
      id: 'compliance',
      label: 'Compliance',
      icon: 'scan',
      subsections: [
        { id: 'score', label: 'Compliance Score' },
        { id: 'recommendations', label: 'Recommendations' },
        { id: 'cost', label: 'Cost Optimization' },
      ]
    },
    {
      id: 'troubleshooting',
      label: 'Troubleshooting',
      icon: 'critical',
      subsections: [
        { id: 'blocked', label: 'Blocked Connections' },
        { id: 'services', label: 'Service Health & Drift' },
      ]
    },
    {
      id: 'events',
      label: 'Events',
      icon: 'events'
    },
    {
      id: 'settings',
      label: 'Settings',
      icon: 'info'
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
        className={`fixed left-0 top-0 h-full w-64 bg-[#0a0a0f]/90 backdrop-blur-xl border-r border-[rgba(255,255,255,0.06)] z-40 transition-transform duration-300 ease-out ${
          isOpen ? 'translate-x-0' : '-translate-x-full'
        } lg:translate-x-0 flex flex-col`}
      >
        <div className="flex flex-col h-full">
          {/* Logo/Header */}
          <div className="p-6 border-b border-[rgba(255,255,255,0.06)] relative overflow-hidden">
            <div className="absolute top-0 left-0 w-full h-[1px] bg-gradient-to-r from-transparent via-[rgba(59,130,246,0.5)] to-transparent opacity-50"></div>
            <div className="mb-4">
              <StargazerLogo size="sm" />
            </div>
            <div className="text-[10px] text-[#a1a1aa] uppercase tracking-[0.2em] font-medium mb-4 pl-1 opacity-70">
              K8S CLUSTER OBSERVATORY
            </div>
            <div className="space-y-3">
              <ContextSelector onContextChange={onContextChange} />
              <NamespaceSelector
                currentNamespace={namespace}
                onNamespaceChange={onNamespaceChange}
              />
              <button
                onClick={() => setShowSearch(true)}
                className="w-full flex items-center gap-2 px-3 py-2 bg-[rgba(255,255,255,0.03)] hover:bg-[rgba(255,255,255,0.06)] border border-[rgba(255,255,255,0.06)] hover:border-[rgba(255,255,255,0.1)] rounded-md text-[#a1a1aa] hover:text-[#e4e4e7] transition-all text-sm group shadow-sm hover:shadow-md hover:shadow-blue-900/10"
              >
                <Icon name="scan" size="sm" className="opacity-70 group-hover:opacity-100 transition-opacity" />
                <span className="flex-1 text-left text-xs font-medium tracking-wide">SEARCH</span>
                <span className="text-[10px] bg-[rgba(255,255,255,0.05)] px-1.5 py-0.5 rounded border border-[rgba(255,255,255,0.05)] text-[#71717a] group-hover:text-[#a1a1aa] group-hover:border-[rgba(255,255,255,0.1)] transition-colors">⌘K</span>
              </button>
            </div>
          </div>

          {/* Navigation Items */}
          <div className="flex-1 overflow-y-auto p-3 space-y-1 scrollbar-hide">
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
                        onSectionChange(section.id);
                      } else {
                        onSectionChange(section.id);
                        setIsOpen(false);
                      }
                    }}
                    className={`w-full flex items-center justify-between px-3 py-2.5 rounded-lg text-sm transition-all cursor-pointer group ${
                      isActive && !currentSubsection
                        ? 'bg-gradient-to-r from-blue-600/20 to-blue-600/5 border border-blue-500/20 text-blue-100 shadow-[0_0_15px_-3px_rgba(59,130,246,0.3)]'
                        : 'text-[#a1a1aa] hover:text-[#e4e4e7] hover:bg-[rgba(255,255,255,0.03)] border border-transparent'
                    }`}
                    aria-label={`Navigate to ${section.label}`}
                  >
                    <div className="flex items-center gap-3">
                      <div className={`p-1 rounded ${isActive && !currentSubsection ? 'bg-blue-500/20 text-blue-400' : 'bg-transparent text-[#71717a] group-hover:text-[#a1a1aa]'}`}>
                        <Icon
                          name={section.icon as any}
                          size="sm"
                        />
                      </div>
                      <span className={`font-medium tracking-wide ${isActive && !currentSubsection ? 'text-shadow-sm' : ''}`}>{section.label}</span>
                    </div>
                    {hasSubsections && (
                      <span className={`text-[#52525b] transition-transform duration-200 ${isExpanded ? 'rotate-90 text-[#a1a1aa]' : ''}`}>
                         <Icon name="info" size="sm" className="w-3 h-3" />
                      </span>
                    )}
                  </button>

                  {/* Sub-sections */}
                  {hasSubsections && isExpanded && (
                    <div className="ml-4 mt-1 mb-2 space-y-0.5 border-l border-[rgba(255,255,255,0.06)] pl-3 relative">
                      {section.subsections!.map((subsection) => {
                        const isSubActive = isActive && currentSubsection === subsection.id;
                        return (
                          <button
                            key={subsection.id}
                            onClick={() => {
                              onSectionChange(section.id, subsection.id);
                              setIsOpen(false);
                            }}
                            className={`w-full flex items-center gap-3 px-3 py-2 rounded-md text-xs transition-all cursor-pointer relative overflow-hidden ${
                              isSubActive
                                ? 'text-blue-300 bg-blue-500/10 font-medium'
                                : 'text-[#71717a] hover:text-[#e4e4e7] hover:bg-[rgba(255,255,255,0.02)]'
                            }`}
                            aria-label={`Navigate to ${subsection.label}`}
                          >
                            {isSubActive && (
                                <div className="absolute left-0 top-1/2 -translate-y-1/2 w-0.5 h-4 bg-blue-500 rounded-r shadow-[0_0_8px_rgba(59,130,246,0.8)]"></div>
                            )}
                            <span className={`w-1 h-1 rounded-full transition-all ${isSubActive ? 'bg-blue-400 scale-125 shadow-[0_0_5px_rgba(59,130,246,0.8)]' : 'bg-[#52525b]'}`}></span>
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
          <div className="p-4 border-t border-[rgba(255,255,255,0.06)] bg-[rgba(0,0,0,0.2)]">
            <div className="flex items-center justify-between text-[10px] text-[#52525b] mb-1">
               <span>STARGAZER SYSTEM</span>
               <span className="w-1.5 h-1.5 rounded-full bg-green-500/50 animate-pulse"></span>
            </div>
            <div className="text-[10px] text-[#404045] font-mono">
              v0.1.0 • CONNECTED
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
      {/* Search Modal */}
      {showSearch && (
        <GlobalResourceSearch
          onClose={() => setShowSearch(false)}
          onNavigate={(type, name, namespace) => {
            // Basic navigation logic
            if (type === 'service') {
              onSectionChange('traffic-analysis', 'topology');
            } else if (type === 'pod') {
              onSectionChange('troubleshooting', 'services');
            } else {
              onSectionChange('dashboard');
            }
          }}
        />
      )}
    </>
  );
}
