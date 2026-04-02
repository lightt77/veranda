import React from 'react';

interface SidebarProps {
  activeTab: string;
  onTabChange: (tab: string) => void;
}

const tabs = [
  { id: 'timers', label: 'Timers', icon: '⏱️' },
  { id: 'stopwatches', label: 'Stopwatches', icon: '⏲️' },
  { id: 'todos', label: 'Todos', icon: '✓' },
];

export const Sidebar: React.FC<SidebarProps> = ({ activeTab, onTabChange }) => {
  return (
    <nav className="w-64 bg-gray-900 border-r border-gray-800 flex flex-col">
      <div className="p-6">
        <h1 className="text-2xl font-bold text-white">Veranda</h1>
        <p className="text-gray-400 text-sm mt-1">Stay productive</p>
      </div>

      <div className="flex-1 px-4 py-2">
        {tabs.map((tab) => (
          <button
            key={tab.id}
            onClick={() => onTabChange(tab.id)}
            className={`w-full flex items-center gap-3 px-4 py-3 rounded-lg mb-2 transition-colors text-left ${
              activeTab === tab.id
                ? 'bg-blue-600 text-white'
                : 'text-gray-300 hover:bg-gray-800'
            }`}
          >
            <span className="text-xl">{tab.icon}</span>
            <span className="font-medium">{tab.label}</span>
          </button>
        ))}
      </div>

      <div className="p-4 border-t border-gray-800">
        <button className="w-full flex items-center gap-3 px-4 py-2 text-gray-400 hover:text-white transition-colors">
          <span>⚙️</span>
          <span>Settings</span>
        </button>
      </div>
    </nav>
  );
};
