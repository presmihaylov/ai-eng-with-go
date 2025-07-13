import React, { useState, useRef, useEffect } from 'react';
import { 
  ChatBubbleLeftRightIcon, 
  PaperAirplaneIcon, 
  TrashIcon,
  WrenchIcon,
  ChartBarIcon,
  UserIcon,
  CpuChipIcon,
  BookOpenIcon,
  AcademicCapIcon,
  LightBulbIcon,
  ClipboardDocumentListIcon
} from '@heroicons/react/24/outline';
import { agentApi } from '../api/flashcardsApi';
import toast from 'react-hot-toast';
import ReactMarkdown from 'react-markdown';

// CSS for hiding scrollbars
const scrollbarHideStyles = `
  .scrollbar-hide {
    -ms-overflow-style: none;
    scrollbar-width: none;
  }
  .scrollbar-hide::-webkit-scrollbar {
    display: none;
  }
`;

// Inject styles into head
if (typeof document !== 'undefined') {
  const styleSheet = document.createElement('style');
  styleSheet.type = 'text/css';
  styleSheet.innerText = scrollbarHideStyles;
  document.head.appendChild(styleSheet);
}

const AgentChat = () => {
  const [messages, setMessages] = useState([]);
  const [currentMessage, setCurrentMessage] = useState('');
  const [loading, setLoading] = useState(false);
  const [loadingMessage, setLoadingMessage] = useState('');
  
  // Ref for auto-scrolling
  const messagesEndRef = useRef(null);
  const textareaRef = useRef(null);

  // Loading messages for variety
  const loadingMessages = {
    default: [
      'Thinking through your question...',
      'Preparing response...',
      'Analyzing your request...',
      'Working on your study session...',
      'Processing your question...',
      'Getting ready to help...'
    ],
    withTools: [
      'Analyzing notes and concepts...',
      'Reviewing your study materials...',
      'Examining note content...',
      'Processing study topics...',
      'Organizing information...',
      'Checking knowledge areas...'
    ]
  };

  const getRandomLoadingMessage = (hasToolCalls) => {
    const messageArray = hasToolCalls ? loadingMessages.withTools : loadingMessages.default;
    return messageArray[Math.floor(Math.random() * messageArray.length)];
  };

  // Auto-scroll to bottom when messages update
  useEffect(() => {
    if (messagesEndRef.current) {
      messagesEndRef.current.scrollIntoView({ behavior: 'smooth' });
    }
  }, [messages]);

  const continueConversation = async (currentMessages) => {
    try {
      const response = await agentApi.chat(currentMessages);
      const newMessages = response.data.messages;
      
      setMessages(newMessages);
      
      // Simple logic: only continue if the last message is a tool result
      // This means we just got tool results and need the assistant's final response
      const lastMessage = newMessages[newMessages.length - 1];
      const shouldContinue = lastMessage?.role === 'tool';
      
      if (shouldContinue) {
        // Update loading message for tool processing
        setLoadingMessage(getRandomLoadingMessage(true));
        // Small delay to show the tool results to the user
        setTimeout(() => {
          continueConversation(newMessages);
        }, 800);
      } else {
        setLoading(false);
      }
      
    } catch (error) {
      console.error('Error in conversation:', error);
      toast.error('Failed to continue conversation. Please try again.');
      setLoading(false);
    }
  };

  const sendMessage = async () => {
    if (!currentMessage.trim() || loading) return;

    const userMessage = {
      role: 'user',
      content: currentMessage.trim()
    };

    // Add user message to conversation
    const updatedMessages = [...messages, userMessage];
    setMessages(updatedMessages);
    setCurrentMessage('');
    setLoading(true);
    setLoadingMessage(getRandomLoadingMessage(false));

    try {
      await continueConversation(updatedMessages);
    } catch (error) {
      // Remove the user message since the request failed
      setMessages(messages);
      setLoading(false);
    }
  };

  const clearChat = () => {
    setMessages([]);
    setCurrentMessage('');
  };

  const handleKeyDown = (e) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      sendMessage();
    }
  };

  const truncateText = (text, maxLength = 100) => {
    if (text.length <= maxLength) return text;
    return text.substring(0, maxLength) + '...';
  };

  const renderToolCall = (toolCall, index) => {
    const formatArgs = (args) => {
      const argString = JSON.stringify(args, null, 0);
      return truncateText(argString, 40);
    };

    return (
      <div key={index} className="bg-indigo-50 border border-indigo-200 rounded-md p-2 mb-1 last:mb-0">
        <div className="flex items-center gap-1.5">
          <WrenchIcon className="h-3 w-3 text-indigo-600" />
          <span className="text-xs font-medium text-indigo-800">{toolCall.name}</span>
          <span className="text-xs text-indigo-600">•</span>
          <span className="text-xs text-indigo-700">{formatArgs(toolCall.arguments)}</span>
        </div>
      </div>
    );
  };

  const renderToolResults = (toolResults, toolCalls) => {
    return toolResults.map((result, idx) => {
      // Find corresponding tool call to get the tool name
      const toolCall = toolCalls?.find(tc => tc.id === result.tool_call_id);
      const toolName = toolCall?.name || 'Unknown Tool';
      
      return (
        <div key={idx} className="bg-emerald-50 border border-emerald-200 rounded-md p-2 mb-1 last:mb-0">
          <div className="flex items-center gap-1.5">
            <ChartBarIcon className="h-3 w-3 text-emerald-600" />
            <span className="text-xs font-medium text-emerald-800">{toolName}</span>
            <span className="text-xs text-emerald-600">→</span>
            <span className="text-xs text-emerald-700">{truncateText(result.content, 80)}</span>
          </div>
        </div>
      );
    });
  };

  const renderMessage = (message, index) => {
    const isUser = message.role === 'user';
    const isAssistant = message.role === 'assistant';
    const isTool = message.role === 'tool';

    // Skip standalone tool messages - they'll be rendered inline with assistant messages
    if (isTool) {
      return null;
    }

    // Find all corresponding tool results for assistant messages
    const toolResults = [];
    if (isAssistant && message.tool_calls?.length > 0) {
      // Look for all consecutive tool messages after this assistant message
      let nextIndex = index + 1;
      while (nextIndex < messages.length && messages[nextIndex].role === 'tool') {
        toolResults.push(...messages[nextIndex].tool_results);
        nextIndex++;
      }
    }

    return (
      <div key={index} className={`flex items-start gap-3 mb-6 ${isUser ? 'flex-row-reverse' : ''}`}>
        <div className={`flex-shrink-0 w-10 h-10 rounded-full flex items-center justify-center ${
          isUser 
            ? 'bg-blue-600 text-white' 
            : 'bg-gradient-to-br from-indigo-600 to-purple-600 text-white'
        }`}>
          {isUser ? (
            <UserIcon className="h-5 w-5" />
          ) : (
            <AcademicCapIcon className="h-5 w-5" />
          )}
        </div>
        
        <div className={`flex-1 ${isUser ? 'flex flex-col items-end' : ''}`}>
          <div className={`text-sm text-gray-500 mb-2 ${isUser ? 'text-right' : ''}`}>
            {isUser ? 'You' : 'System Design Tutor'}
          </div>
          
          {message.content && (
            <div className={`inline-block max-w-3xl ${
              isUser 
                ? 'bg-blue-600 text-white rounded-2xl rounded-tr-md px-4 py-2 text-left'
                : 'bg-gray-100 text-gray-900 rounded-2xl rounded-tl-md px-4 py-2'
            }`}>
              <div className={`prose prose-sm max-w-none ${
                isUser ? 'prose-invert' : ''
              }`} style={isUser ? { whiteSpace: 'pre-wrap' } : {}}>
                <ReactMarkdown>{message.content}</ReactMarkdown>
              </div>
            </div>
          )}

          {/* Render tool calls and results below the message */}
          {(message.tool_calls?.length > 0 || toolResults.length > 0) && (
            <div className="mt-2 ml-8 space-y-1">
              {/* Tool Calls */}
              {message.tool_calls?.map((toolCall, idx) => renderToolCall(toolCall, idx))}
              
              {/* Tool Results */}
              {toolResults.length > 0 && renderToolResults(toolResults, message.tool_calls)}
            </div>
          )}
        </div>
      </div>
    );
  };

  return (
    <div className="max-w-4xl mx-auto px-6">
      {/* Header */}
      <div className="bg-white rounded-lg shadow-sm border border-gray-200 mb-6">
        <div className="px-6 py-4 border-b border-gray-200 bg-gradient-to-r from-indigo-50 to-purple-50">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="p-3 bg-gradient-to-br from-indigo-600 to-purple-600 rounded-xl shadow-lg">
                <AcademicCapIcon className="h-7 w-7 text-white" />
              </div>
              <div>
                <h2 className="text-2xl font-bold bg-gradient-to-r from-indigo-600 to-purple-600 bg-clip-text text-transparent">
                  System Design Study Assistant
                </h2>
                <p className="text-sm text-gray-600 font-medium">
                  Your personal tutor for mastering system design concepts
                </p>
              </div>
            </div>
            
            {messages.length > 0 && (
              <button
                onClick={clearChat}
                className="flex items-center gap-2 px-3 py-2 text-red-600 hover:text-red-700 hover:bg-red-50 rounded-lg transition-colors"
              >
                <TrashIcon className="h-4 w-4" />
                Clear Chat
              </button>
            )}
          </div>
        </div>

        {/* Chat Messages */}
        <div className="p-6">
          {messages.length === 0 ? (
            <div className="text-center py-12">
              <div className="mb-6">
                <div className="inline-flex items-center justify-center w-20 h-20 bg-gradient-to-br from-indigo-100 to-purple-100 rounded-full mb-4">
                  <AcademicCapIcon className="h-10 w-10 text-indigo-600" />
                </div>
                <h3 className="text-xl font-semibold text-gray-900 mb-2">Ready to Study System Design?</h3>
                <p className="text-gray-600 mb-8 max-w-lg mx-auto">
                  I'm here to help you learn and master system design concepts. Ask me anything about your study materials, request explanations, or test your knowledge.
                </p>
              </div>
            </div>
          ) : (
            <div className="space-y-4">
              {messages.map((message, index) => renderMessage(message, index))}
              
              {loading && (
                <div className="flex items-start gap-3 mb-6">
                  <div className="flex-shrink-0 w-10 h-10 bg-gradient-to-br from-indigo-600 to-purple-600 text-white rounded-full flex items-center justify-center">
                    <AcademicCapIcon className="h-5 w-5" />
                  </div>
                  <div className="flex-1">
                    <div className="text-sm text-gray-500 mb-2">System Design Tutor</div>
                    <div className="bg-gray-100 rounded-2xl rounded-tl-md px-4 py-2">
                      <div className="flex items-center gap-2 text-gray-600">
                        <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-indigo-600"></div>
                        <span>{loadingMessage}</span>
                      </div>
                    </div>
                  </div>
                </div>
              )}
              
              <div ref={messagesEndRef} />
            </div>
          )}
        </div>

        {/* Message Input */}
        <div className="px-6 py-4 border-t border-gray-200">
          <div className="flex gap-3 items-end">
            <textarea
              value={currentMessage}
              onChange={(e) => setCurrentMessage(e.target.value)}
              onKeyDown={handleKeyDown}
              placeholder="Type your message... (Press Enter to send, Shift+Enter for new line)"
              className="flex-1 resize-none border border-gray-300 rounded-lg px-3 py-2 focus:ring-2 focus:ring-blue-500 focus:border-transparent min-h-[40px] max-h-32 overflow-y-auto scrollbar-hide"
              rows="1"
              style={{
                height: 'auto',
                minHeight: '40px',
                maxHeight: '128px',
                scrollbarWidth: 'none',
                msOverflowStyle: 'none'
              }}
              onInput={(e) => {
                e.target.style.height = 'auto';
                e.target.style.height = Math.min(e.target.scrollHeight, 128) + 'px';
              }}
              ref={(textarea) => {
                if (textarea) {
                  textarea.style.height = 'auto';
                  textarea.style.height = Math.min(textarea.scrollHeight, 128) + 'px';
                }
              }}
              disabled={loading}
            />
            <button
              onClick={sendMessage}
              disabled={!currentMessage.trim() || loading}
              className="flex-shrink-0 bg-gradient-to-r from-indigo-600 to-purple-600 text-white px-3 py-2 rounded-lg hover:from-indigo-700 hover:to-purple-700 disabled:opacity-50 disabled:cursor-not-allowed transition-all duration-200 flex items-center gap-1 shadow-lg h-10"
            >
              <PaperAirplaneIcon className="h-4 w-4" />
              <span className="text-sm">Send</span>
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

export default AgentChat;