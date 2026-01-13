'use client';

import { useState, useEffect } from 'react';
import apiClient from '@/lib/api';
import { Icon } from './SpaceshipIcons';

interface Event {
  name: string;
  namespace: string;
  type: string;
  reason: string;
  message: string;
  object_kind: string;
  object_name: string;
  timestamp: string;
  age: string;
}

interface EventsDetailProps {
  namespace?: string;
  onClose?: () => void;
}

export default function EventsDetail({ namespace, onClose }: EventsDetailProps) {
  const [events, setEvents] = useState<Event[]>([]);
  const [loading, setLoading] = useState(true);
  const [filter, setFilter] = useState<'all' | 'warning' | 'error'>('all');

  useEffect(() => {
    loadEvents();
  }, [namespace]);

  const loadEvents = async () => {
    try {
      setLoading(true);
      const data = await apiClient.getEvents(namespace);
      setEvents(data.events);
    } catch (error) {
      // Error loading events - will show empty state
    } finally {
      setLoading(false);
    }
  };

  const filteredEvents = events.filter(event => {
    if (filter === 'all') return true;
    if (filter === 'warning') return event.type === 'Warning';
    if (filter === 'error') return event.reason.includes('Error') || event.reason.includes('Fail');
    return true;
  });

  const getEventColor = (type: string, reason: string) => {
    if (type === 'Warning' || reason.includes('Error') || reason.includes('Fail')) {
      return 'text-red-400';
    }
    return 'text-yellow-400';
  };

  const mainContent = (
    <div className={`card rounded-lg p-6 w-full ${onClose ? 'max-w-6xl max-h-[90vh]' : ''} overflow-hidden flex flex-col`}>
      <div className="flex items-center justify-between mb-5 pb-4 border-b border-[rgba(255,255,255,0.08)]">
        <div className="flex items-center gap-3">
          <Icon name="events" className="text-[#71717a]" size="sm" />
          <h2 className="text-2xl font-bold text-[#e4e4e7]">Events</h2>
          <span className="px-2.5 py-1 rounded-full bg-[#1a1a24] border border-[rgba(255,255,255,0.08)] text-[#71717a] text-xs">
            {events.length} TOTAL
          </span>
        </div>
        {onClose && (
          <button
            onClick={onClose}
            className="px-3 py-1.5 bg-red-600/90 hover:bg-red-700 text-white rounded-lg text-sm transition-all"
          >
            CLOSE
          </button>
        )}
      </div>

      <div className="flex gap-2 mb-4">
        {[
          { id: 'all', label: 'All' },
          { id: 'warning', label: 'Warnings' },
          { id: 'error', label: 'Errors' },
        ].map(f => (
          <button
            key={f.id}
            onClick={() => setFilter(f.id as any)}
            className={`px-4 py-2 rounded-lg text-sm transition-all ${
              filter === f.id
                ? 'bg-[#3b82f6] text-white'
                : 'bg-[#1a1a24] text-[#71717a] hover:bg-[#27272a]'
            }`}
          >
            {f.label}
          </button>
        ))}
      </div>

      <div className="flex-1 overflow-auto space-y-2.5">
        {loading ? (
          <div className="text-center py-10">
            <Icon name="loading" className="text-blue-400/80 animate-pulse text-3xl" />
            <p className="text-[#71717a] mt-2 text-sm">Loading events...</p>
          </div>
        ) : filteredEvents.length === 0 ? (
          <div className="text-center py-10">
            <Icon name="healthy" className="text-emerald-400/90 text-3xl" />
            <p className="text-[#71717a] mt-2 text-sm">No events found</p>
          </div>
        ) : (
          filteredEvents.map(event => (
            <div
              key={event.name}
              className="card card-hover rounded-lg p-3.5 border border-purple-500/15"
            >
              <div className="flex items-start justify-between">
                <div className="flex-1">
                  <div className="flex items-center gap-2 mb-2">
                    <Icon name="events" className={getEventColor(event.type, event.reason)} size="sm" />
                    <h3 className="font-bold text-[#e4e4e7] text-base">{event.reason}</h3>
                    <span className={`px-2 py-0.5 rounded text-xs font-semibold ${getEventColor(event.type, event.reason)}`}>
                      {event.type.toUpperCase()}
                    </span>
                  </div>
                  <div className="grid grid-cols-2 md:grid-cols-4 gap-2.5 text-xs mb-2">
                    <div>
                      <span className="text-[#71717a]">Object:</span>
                      <div className="text-[#3b82f6]">{event.object_kind}/{event.object_name}</div>
                    </div>
                    <div>
                      <span className="text-[#71717a]">Namespace:</span>
                      <div className="text-[#e4e4e7]">{event.namespace}</div>
                    </div>
                    <div>
                      <span className="text-[#71717a]">Time:</span>
                      <div className="text-[#e4e4e7]">{new Date(event.timestamp).toLocaleString()}</div>
                    </div>
                    <div>
                      <span className="text-[#71717a]">Age:</span>
                      <div className="text-[#e4e4e7]">{event.age}</div>
                    </div>
                  </div>
                  <div className="mt-2 p-2.5 bg-[#1a1a24] rounded border border-[rgba(255,255,255,0.08)]">
                    <p className="text-[#71717a] text-xs leading-relaxed">{event.message}</p>
                  </div>
                </div>
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  );

  if (onClose) {
    return (
      <div className="fixed inset-0 bg-black/80 backdrop-blur-sm z-50 flex items-center justify-center p-4">
        {mainContent}
      </div>
    );
  }

  return mainContent;
}
