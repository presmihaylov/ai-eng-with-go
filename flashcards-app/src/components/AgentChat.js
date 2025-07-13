import React, { useState, useRef, useEffect } from 'react';
import { 
  ChatBubbleLeftRightIcon, 
  PaperAirplaneIcon, 
  TrashIcon,
  WrenchIcon,
  ChartBarIcon,
  UserIcon,
  CpuChipIcon
} from '@heroicons/react/24/outline';
import { agentApi } from '../api/flashcardsApi';
import toast from 'react-hot-toast';
import ReactMarkdown from 'react-markdown';

const AgentChat = () => {
  const [messages, setMessages] = useState([]);
  const [currentMessage, setCurrentMessage] = useState('');
  const [loading, setLoading] = useState(false);
  
  // Ref for auto-scrolling
  const messagesEndRef = useRef(null);

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

  const handleKeyPress = (e) => {
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
      <div key={index} className="bg-orange-50 border border-orange-200 rounded-md p-2 mb-1 last:mb-0">
        <div className="flex items-center gap-1.5">
          <WrenchIcon className="h-3 w-3 text-orange-600" />
          <span className="text-xs font-medium text-orange-800">{toolCall.name}</span>
          <span className="text-xs text-orange-600">•</span>
          <span className="text-xs text-orange-700">{formatArgs(toolCall.arguments)}</span>
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
        <div key={idx} className="bg-green-50 border border-green-200 rounded-md p-2 mb-1 last:mb-0">
          <div className="flex items-center gap-1.5">
            <ChartBarIcon className="h-3 w-3 text-green-600" />
            <span className="text-xs font-medium text-green-800">{toolName}</span>
            <span className="text-xs text-green-600">→</span>
            <span className="text-xs text-green-700">{truncateText(result.content, 80)}</span>
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
            : 'bg-gray-800 text-white'
        }`}>
          {isUser ? (
            <UserIcon className="h-5 w-5" />
          ) : (
            <CpuChipIcon className="h-5 w-5" />
          )}
        </div>
        
        <div className={`flex-1 ${isUser ? 'text-right' : ''}`}>
          <div className={`text-sm text-gray-500 mb-2 ${isUser ? 'text-right' : ''}`}>
            {isUser ? 'You' : 'AI Assistant'}
          </div>
          
          <div className={`inline-block max-w-3xl ${
            isUser 
              ? 'bg-blue-600 text-white rounded-2xl rounded-tr-md p-4'
              : 'bg-gray-100 text-gray-900 rounded-2xl rounded-tl-md p-4'
          }`}>
            {message.content && (
              <div className={`prose prose-sm max-w-none ${
                isUser ? 'prose-invert' : ''
              }`}>
                <ReactMarkdown>{message.content}</ReactMarkdown>
              </div>
            )}
          </div>

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
        <div className="px-6 py-4 border-b border-gray-200">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="p-2 bg-indigo-100 rounded-lg">
                <CpuChipIcon className="h-6 w-6 text-indigo-600" />
              </div>
              <div>
                <h2 className="text-xl font-bold text-gray-900">AI Assistant</h2>
                <p className="text-sm text-gray-500">
                  Intelligent assistant with tool capabilities • Weather information and more
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
              <CpuChipIcon className="h-12 w-12 text-gray-300 mx-auto mb-4" />
              <h3 className="text-lg font-medium text-gray-900 mb-2">Start a conversation</h3>
              <p className="text-gray-500 mb-6">
                Ask the AI assistant anything - it has access to tools and can help with various tasks.
              </p>
              <div className="space-y-2 text-sm text-gray-400">
                <p><strong>Try:</strong> "What is the weather in New York City?"</p>
                <p><strong>Try:</strong> "Check weather in NYC and London"</p>
                <p><strong>Try:</strong> "What is 15 * 7?"</p>
              </div>
            </div>
          ) : (
            <div className="space-y-4">
              {messages.map((message, index) => renderMessage(message, index))}
              
              {loading && (
                <div className="flex items-start gap-3 mb-6">
                  <div className="flex-shrink-0 w-10 h-10 bg-gray-800 text-white rounded-full flex items-center justify-center">
                    <CpuChipIcon className="h-5 w-5" />
                  </div>
                  <div className="flex-1">
                    <div className="text-sm text-gray-500 mb-2">AI Assistant</div>
                    <div className="bg-gray-100 rounded-2xl rounded-tl-md p-4">
                      <div className="flex items-center gap-2 text-gray-600">
                        <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-gray-600"></div>
                        <span>{messages.some(m => m.tool_calls?.length > 0) ? 'Processing tools...' : 'Thinking...'}</span>
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
          <div className="flex gap-3">
            <textarea
              value={currentMessage}
              onChange={(e) => setCurrentMessage(e.target.value)}
              onKeyPress={handleKeyPress}
              placeholder="Type your message... (Press Enter to send, Shift+Enter for new line)"
              className="flex-1 resize-none border border-gray-300 rounded-lg px-3 py-2 focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              rows="2"
              disabled={loading}
            />
            <button
              onClick={sendMessage}
              disabled={!currentMessage.trim() || loading}
              className="flex-shrink-0 bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors flex items-center gap-2"
            >
              <PaperAirplaneIcon className="h-4 w-4" />
              Send
            </button>
          </div>
        </div>
      </div>

      {/* Features Info */}
      <div className="bg-indigo-50 border border-indigo-200 rounded-lg p-4">
        <h3 className="font-medium text-indigo-900 mb-2">✨ Assistant Capabilities</h3>
        <div className="text-sm text-indigo-800 space-y-1">
          <p>🌤️ <strong>Weather Information:</strong> Get weather for any city worldwide</p>
          <p>💬 <strong>General Conversation:</strong> Questions, math, explanations, and more</p>
          <p>🔄 <strong>Multi-task Requests:</strong> Handle multiple operations in one message</p>
          <p>🧠 <strong>Context Awareness:</strong> Remembers conversation history</p>
        </div>
      </div>
    </div>
  );
};

export default AgentChat;