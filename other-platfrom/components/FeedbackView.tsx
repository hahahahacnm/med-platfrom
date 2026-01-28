
import React, { useState } from 'react';
import { Send, MessageSquare, AlertTriangle, Lightbulb, Bug, FileQuestion, CheckCircle2 } from 'lucide-react';
import { motion } from 'framer-motion';
import { api } from '../services/api';

const FeedbackView: React.FC = () => {
  const [type, setType] = useState('suggestion');
  const [content, setContent] = useState('');
  const [contact, setContact] = useState('');
  const [submitted, setSubmitted] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!content.trim()) return;

    try {
      await api.feedback.submit({ type, content, contact });
      setSubmitted(true);
      setContent('');
      setContact('');
    } catch (err) {
      alert('提交失败，请重试');
    }
  };

  if (submitted) {
    return (
      <motion.div
        initial={{ opacity: 0, scale: 0.95 }}
        animate={{ opacity: 1, scale: 1 }}
        className="max-w-xl mx-auto mt-12 md:mt-20 text-center"
      >
        <div className="w-24 h-24 bg-emerald-100 text-emerald-600 rounded-full flex items-center justify-center mx-auto mb-8 shadow-lg shadow-emerald-200">
          <CheckCircle2 size={48} />
        </div>
        <h2 className="text-3xl font-bold text-slate-900 mb-3">反馈提交成功</h2>
        <p className="text-slate-500 mb-10 text-lg">感谢您的宝贵意见，我们会尽快分析并优化。</p>
        <button
          onClick={() => setSubmitted(false)}
          className="px-10 py-4 bg-slate-900 text-white rounded-2xl font-bold hover:bg-blue-600 transition-all shadow-xl shadow-slate-200 hover:shadow-blue-500/20 active:scale-95"
        >
          再次反馈
        </button>
      </motion.div>
    );
  }

  // Helper for option styling to ensure Tailwind JIT picks up classes safely
  const getOptionClasses = (id: string, activeColorClass: string, activeBorderClass: string, activeBgClass: string) => {
    const isSelected = type === id;
    return `p-5 rounded-2xl border-2 flex flex-col items-center justify-center gap-4 transition-all cursor-pointer h-40 group relative overflow-hidden ${isSelected
      ? `${activeBgClass} ${activeBorderClass} ${activeColorClass} shadow-md scale-[1.02]`
      : 'bg-white border-slate-200 text-slate-500 hover:border-slate-300 hover:bg-slate-50'
      }`;
  };

  return (
    <motion.div
      initial={{ opacity: 0, y: 10 }}
      animate={{ opacity: 1, y: 0 }}
      className="max-w-5xl mx-auto space-y-10 pb-10"
    >
      {/* Header Section */}
      <div className="text-center md:text-left">
        <h2 className="text-3xl md:text-4xl font-bold text-slate-900 flex items-center justify-center md:justify-start gap-4 mb-3">
          <div className="p-3 bg-blue-600 rounded-2xl text-white shadow-lg shadow-blue-500/30">
            <MessageSquare size={32} />
          </div>
          问题与反馈
        </h2>
        <p className="text-slate-500 text-lg pl-1 md:pl-20">
          您的建议是我们前进的动力。无论是 Bug 汇报还是功能吐槽，我们都洗耳恭听。
        </p>
      </div>

      <form onSubmit={handleSubmit} className="space-y-10">

        {/* Section 1: Type Selection Grid */}
        <section>
          <div className="flex items-center gap-3 mb-6">
            <span className="flex items-center justify-center w-8 h-8 rounded-full bg-slate-200 text-slate-600 font-bold text-sm">1</span>
            <h3 className="text-xl font-bold text-slate-800">您想反馈什么类型的问题？</h3>
          </div>

          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 md:gap-6">
            <button
              type="button"
              onClick={() => setType('suggestion')}
              className={getOptionClasses('suggestion', 'text-amber-700', 'border-amber-400', 'bg-amber-50')}
            >
              <div className={`p-4 rounded-full transition-colors ${type === 'suggestion' ? 'bg-amber-200 text-amber-700' : 'bg-slate-100 group-hover:bg-slate-200'}`}>
                <Lightbulb size={28} />
              </div>
              <span className="font-bold text-lg">功能建议</span>
              {type === 'suggestion' && <motion.div layoutId="check" className="absolute top-3 right-3 text-amber-600"><CheckCircle2 size={20} /></motion.div>}
            </button>

            <button
              type="button"
              onClick={() => setType('bug')}
              className={getOptionClasses('bug', 'text-rose-700', 'border-rose-400', 'bg-rose-50')}
            >
              <div className={`p-4 rounded-full transition-colors ${type === 'bug' ? 'bg-rose-200 text-rose-700' : 'bg-slate-100 group-hover:bg-slate-200'}`}>
                <Bug size={28} />
              </div>
              <span className="font-bold text-lg">系统 Bug</span>
              {type === 'bug' && <motion.div layoutId="check" className="absolute top-3 right-3 text-rose-600"><CheckCircle2 size={20} /></motion.div>}
            </button>

            <button
              type="button"
              onClick={() => setType('content')}
              className={getOptionClasses('content', 'text-indigo-700', 'border-indigo-400', 'bg-indigo-50')}
            >
              <div className={`p-4 rounded-full transition-colors ${type === 'content' ? 'bg-indigo-200 text-indigo-700' : 'bg-slate-100 group-hover:bg-slate-200'}`}>
                <FileQuestion size={28} />
              </div>
              <span className="font-bold text-lg">内容错误</span>
              {type === 'content' && <motion.div layoutId="check" className="absolute top-3 right-3 text-indigo-600"><CheckCircle2 size={20} /></motion.div>}
            </button>

            <button
              type="button"
              onClick={() => setType('other')}
              className={getOptionClasses('other', 'text-slate-700', 'border-slate-400', 'bg-slate-200')}
            >
              <div className={`p-4 rounded-full transition-colors ${type === 'other' ? 'bg-slate-300 text-slate-700' : 'bg-slate-100 group-hover:bg-slate-200'}`}>
                <AlertTriangle size={28} />
              </div>
              <span className="font-bold text-lg">其他问题</span>
              {type === 'other' && <motion.div layoutId="check" className="absolute top-3 right-3 text-slate-600"><CheckCircle2 size={20} /></motion.div>}
            </button>
          </div>
        </section>

        {/* Section 2: Details Input */}
        <section>
          <div className="flex items-center gap-3 mb-6">
            <span className="flex items-center justify-center w-8 h-8 rounded-full bg-slate-200 text-slate-600 font-bold text-sm">2</span>
            <h3 className="text-xl font-bold text-slate-800">请详细描述您遇到的情况</h3>
          </div>

          <div className="relative group">
            <textarea
              required
              value={content}
              onChange={(e) => setContent(e.target.value)}
              className="w-full h-56 p-6 bg-white border-2 border-slate-200 rounded-3xl focus:outline-none focus:border-blue-500 focus:ring-4 focus:ring-blue-500/10 transition-all resize-none text-slate-700 placeholder:text-slate-400 text-lg shadow-sm group-hover:border-slate-300"
              placeholder="例如：在做《内科学》第三章第5题时，发现解析图片无法加载..."
            />
            <div className="absolute bottom-6 right-6 text-slate-400 text-sm font-medium bg-slate-50 px-3 py-1 rounded-full border border-slate-100">
              {content.length} 字
            </div>
          </div>
        </section>

        {/* Section 3: Contact */}
        <section>
          <div className="flex items-center gap-3 mb-6">
            <span className="flex items-center justify-center w-8 h-8 rounded-full bg-slate-200 text-slate-600 font-bold text-sm">3</span>
            <h3 className="text-xl font-bold text-slate-800">联系方式 <span className="text-sm text-slate-400 font-normal ml-2">(选填，方便我们给您回信)</span></h3>
          </div>
          <input
            type="text"
            value={contact}
            onChange={(e) => setContact(e.target.value)}
            className="w-full p-6 bg-white border-2 border-slate-200 rounded-2xl focus:outline-none focus:border-blue-500 focus:ring-4 focus:ring-blue-500/10 transition-all text-slate-700 text-lg shadow-sm hover:border-slate-300"
            placeholder="请输入邮箱或手机号码"
          />
        </section>

        {/* Submit Button */}
        <div className="pt-6">
          <button
            type="submit"
            disabled={!content.trim()}
            className="w-full bg-slate-900 text-white py-6 rounded-3xl font-bold text-xl hover:bg-blue-600 transition-all shadow-xl shadow-slate-200 hover:shadow-blue-500/30 disabled:opacity-50 disabled:shadow-none flex items-center justify-center gap-3 transform active:scale-[0.99] group"
          >
            <Send size={24} className="group-hover:translate-x-1 group-hover:-translate-y-1 transition-transform" />
            提交反馈
          </button>
        </div>

      </form>
    </motion.div>
  );
};

export default FeedbackView;
