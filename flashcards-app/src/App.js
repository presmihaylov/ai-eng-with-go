import React, { useState } from 'react';
import { Toaster } from 'react-hot-toast';
import Navigation from './components/Navigation';
import NotesManager from './components/NotesManager';
import InteractiveQuiz from './components/InteractiveQuiz';
import QuizV2 from './components/QuizV2';
import AgentChat from './components/AgentChat';

function App() {
  const [activeTab, setActiveTab] = useState('notes');

  const renderActiveComponent = () => {
    switch (activeTab) {
      case 'notes':
        return <NotesManager />;
      case 'agent':
        return <AgentChat />;
      case 'interactive':
        return <QuizV2 />;
      case 'interactive-old':
        return <InteractiveQuiz />;
      default:
        return <NotesManager />;
    }
  };

  return (
    <div className="min-h-screen bg-gray-50">
      <Toaster 
        position="top-right"
        toastOptions={{
          duration: 3000,
          style: {
            background: '#363636',
            color: '#fff',
          },
          success: {
            duration: 3000,
            iconTheme: {
              primary: '#10B981',
              secondary: '#fff',
            },
          },
          error: {
            duration: 4000,
            iconTheme: {
              primary: '#EF4444',
              secondary: '#fff',
            },
          },
        }}
      />
      
      <Navigation activeTab={activeTab} setActiveTab={setActiveTab} />
      
      <main className="py-8">
        {renderActiveComponent()}
      </main>
      
      <footer className="bg-white border-t border-gray-200 py-6 mt-16">
        <div className="max-w-4xl mx-auto px-6 text-center text-gray-500 text-sm">
          <p>Flashcards App - Powered by Go backend & React frontend</p>
        </div>
      </footer>
    </div>
  );
}

export default App;
