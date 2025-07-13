import React, { useState } from 'react';
import { BookOpenIcon, AcademicCapIcon, HeartIcon, ChevronDownIcon, ChatBubbleLeftRightIcon } from '@heroicons/react/24/outline';

const Navigation = ({ activeTab, setActiveTab }) => {
  const [showOldVersions, setShowOldVersions] = useState(false);

  const tabs = [
    { id: 'notes', name: 'Notes', icon: BookOpenIcon },
    { id: 'agent', name: 'Agent', icon: ChatBubbleLeftRightIcon },
  ];

  const oldVersions = [
    { id: 'interactive', name: 'Interactive Quiz (v2)', icon: AcademicCapIcon },
    { id: 'interactive-old', name: 'Interactive Quiz (v1)', icon: AcademicCapIcon },
  ];

  return (
    <nav className="bg-white shadow-sm border-b border-gray-200">
      <div className="max-w-4xl mx-auto px-6">
        <div className="flex justify-between items-center py-4">
          <div className="flex items-center gap-2">
            <BookOpenIcon className="h-8 w-8 text-blue-600" />
            <h1 className="text-xl font-bold text-gray-800">Flashcards App</h1>
          </div>
          
          <div className="flex space-x-1">
            {tabs.map((tab) => {
              const Icon = tab.icon;
              return (
                <button
                  key={tab.id}
                  onClick={() => setActiveTab(tab.id)}
                  className={`flex items-center gap-2 px-4 py-2 rounded-lg font-medium transition-colors ${
                    activeTab === tab.id
                      ? 'bg-blue-100 text-blue-700'
                      : 'text-gray-600 hover:text-gray-800 hover:bg-gray-100'
                  }`}
                >
                  <Icon className="h-5 w-5" />
                  {tab.name}
                </button>
              );
            })}
            
            {/* Old Versions Dropdown */}
            <div 
              className="relative"
              onMouseEnter={() => setShowOldVersions(true)}
              onMouseLeave={() => setShowOldVersions(false)}
            >
              <button
                className={`flex items-center gap-2 px-4 py-2 rounded-lg font-medium transition-colors ${
                  oldVersions.some(version => version.id === activeTab)
                    ? 'bg-blue-100 text-blue-700'
                    : 'text-gray-600 hover:text-gray-800 hover:bg-gray-100'
                }`}
              >
                <ChevronDownIcon className="h-5 w-5" />
                Old Versions
              </button>
              
              {showOldVersions && (
                <div className="absolute top-full left-0 -mt-1 pt-1 bg-white border border-gray-200 rounded-lg shadow-lg z-50 min-w-48">
                  {oldVersions.map((version) => {
                    const Icon = version.icon;
                    return (
                      <button
                        key={version.id}
                        onClick={() => {
                          setActiveTab(version.id);
                          setShowOldVersions(false);
                        }}
                        className={`flex items-center gap-2 w-full px-4 py-2 text-left font-medium transition-colors first:rounded-t-lg last:rounded-b-lg ${
                          activeTab === version.id
                            ? 'bg-blue-100 text-blue-700'
                            : 'text-gray-600 hover:text-gray-800 hover:bg-gray-100'
                        }`}
                      >
                        <Icon className="h-5 w-5" />
                        {version.name}
                      </button>
                    );
                  })}
                </div>
              )}
            </div>
          </div>
          
          <div className="flex items-center gap-2 text-sm text-gray-500">
            <span>Made with</span>
            <HeartIcon className="h-4 w-4 text-red-500" />
            <span>& AI</span>
          </div>
        </div>
      </div>
    </nav>
  );
};

export default Navigation;