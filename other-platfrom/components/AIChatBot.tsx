
import React, { useState, useRef, useEffect } from 'react';
import { Send, Bot, User, Sparkles } from 'lucide-react';
import { generateMedicalResponse } from '../services/geminiService';

interface ChatMessage {
  id: string;
  role: 'user' | 'model';
  text: string;
  isError?: boolean;
}

export const AIChatBot: React.FC = () => {
  const [input, setInput] = useState('');
  const [messages, setMessages] = useState<ChatMessage[]>([
    {
      id: 'welcome',
      role: 'model',
      text: '你好！我是您的 题酷 智能助教。有关医学知识、考试重点或平台订阅的问题都可以问我。'
    }
  ]);
  const [isLoading, setIsLoading] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  const handleSend = async () => {
    if (!input.trim() || isLoading) return;

    const userMsg: ChatMessage = {
      id: Date.now().toString(),
      role: 'user',
      text: input.trim()
    };

    setMessages(prev => [...prev, userMsg]);
    setInput('');
    setIsLoading(true);

    try {
      const responseText = await generateMedicalResponse(userMsg.text);
      const botMsg: ChatMessage = {
        id: (Date.now() + 1).toString(),
        role: 'model',
        text: responseText
      };
      setMessages(prev => [...prev, botMsg]);
    } catch (error) {
      const errorMsg: ChatMessage = {
        id: (Date.now() + 1).toString(),
        role: 'model',
        text: '网络连接异常，请稍后再试。',
        isError: true
      };
      setMessages(prev => [...prev, errorMsg]);
    } finally {
      setIsLoading(false);
    }
  };

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  return (
    <div className="flex flex-col h-full w-full bg-slate-50 relative">
      {/* Header */}
      <div className="bg-white px-6 py-4 border-b border-slate-200 flex items-center justify-between shrink-0 sticky top-0 z-10">
        <div className="flex items-center gap-4">
          <div className="bg-blue-600 p-2.5 rounded-xl text-white shadow-lg shadow-blue-500/20">
            <Bot size={24} />
          </div>
          <div>
            <h3 className="font-bold text-lg text-slate-900 flex items-center gap-2">
              智能助教
              <span className="bg-blue-50 text-blue-600 text-[10px] px-2 py-0.5 rounded-full border border-blue-100 flex items-center gap-1 font-bold tracking-wide">
                <Sparkles size={10} /> AI Powered
              </span>
            </h3>
            <p className="text-xs text-slate-500">24/7 在线为您解答医学疑难</p>
          </div>
        </div>
      </div>

      {/* Messages */}
      <div className="flex-1 overflow-y-auto p-4 md:p-8 space-y-6 scroll-smooth">
        {messages.map((msg) => (
          <div key={msg.id} className={`flex gap-4 max-w-4xl mx-auto ${msg.role === 'user' ? 'flex-row-reverse' : ''}`}>
            <div className={`w-10 h-10 rounded-full flex items-center justify-center shrink-0 border-2 ${msg.role === 'user' ? 'bg-blue-100 text-blue-600 border-white shadow-sm' : 'bg-teal-100 text-teal-600 border-white shadow-sm'
              }`}>
              {msg.role === 'user' ? <User size={18} /> : <Bot size={18} />}
            </div>
            <div className={`max-w-[85%] md:max-w-[75%] p-4 md:p-5 rounded-3xl text-sm leading-relaxed shadow-sm ${msg.role === 'user'
              ? 'bg-blue-600 text-white rounded-tr-none shadow-blue-500/20'
              : 'bg-white text-slate-700 border border-slate-200 rounded-tl-none'
              } ${msg.isError ? 'bg-red-50 text-red-500 border-red-100' : ''}`}>
              {msg.text}
            </div>
          </div>
        ))}
        {isLoading && (
          <div className="flex gap-4 max-w-4xl mx-auto">
            <div className="w-10 h-10 rounded-full bg-teal-100 text-teal-600 flex items-center justify-center shrink-0 border-2 border-white shadow-sm">
              <Bot size={18} />
            </div>
            <div className="bg-white px-6 py-4 rounded-3xl rounded-tl-none border border-slate-200 shadow-sm flex items-center gap-1.5">
              <span className="w-2 h-2 bg-slate-400 rounded-full animate-bounce" style={{ animationDelay: '0ms' }}></span>
              <span className="w-2 h-2 bg-slate-400 rounded-full animate-bounce" style={{ animationDelay: '150ms' }}></span>
              <span className="w-2 h-2 bg-slate-400 rounded-full animate-bounce" style={{ animationDelay: '300ms' }}></span>
            </div>
          </div>
        )}
        <div ref={messagesEndRef} />
      </div>

      {/* Input */}
      <div className="p-4 pb-6 md:p-6 bg-white border-t border-slate-200 shrink-0">
        <div className="flex gap-3 max-w-4xl mx-auto">
          <div className="flex-1 bg-slate-50 rounded-2xl p-1 focus-within:ring-2 focus-within:ring-blue-500/20 focus-within:bg-white transition-all border border-slate-200">
            <input
              type="text"
              value={input}
              onChange={(e) => setInput(e.target.value)}
              onKeyDown={handleKeyPress}
              placeholder="输入医学问题或询问平台功能..."
              className="w-full bg-transparent border-none outline-none px-4 py-3 text-slate-700 placeholder:text-slate-400"
            />
          </div>
          <button
            onClick={handleSend}
            disabled={!input.trim() || isLoading}
            className={`p-4 rounded-2xl transition-all duration-300 flex items-center justify-center ${input.trim() && !isLoading
              ? 'bg-blue-600 text-white shadow-lg hover:bg-blue-700 active:scale-95'
              : 'bg-slate-100 text-slate-400 cursor-not-allowed'
              }`}
          >
            <Send size={20} />
          </button>
        </div>
      </div>
    </div>
  );
};
