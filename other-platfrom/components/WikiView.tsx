
import React, { useMemo, useState, useEffect } from 'react';
import { motion } from 'framer-motion';
import { ArrowLeft, Book, Clock, Calendar, User, AlertCircle, ChevronRight, Hash, Star, CheckCircle2, Lock } from 'lucide-react';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { WikiCategory, Article } from '../types';
import { AccessGuard } from './AccessGuard';
import { useAppContext } from '../context';
import { api } from '../services/api';

const WikiView: React.FC = () => {
  const { wikiState, setWikiState, toggleBookmark, isBookmarked, hasAccess } = useAppContext();
  const { category: selectedCategory, article: selectedArticle } = wikiState;
  const [categories, setCategories] = useState<WikiCategory[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    api.wiki.getCategories().then(data => {
      setCategories(data);
      setLoading(false);
    }).catch(err => {
      console.error(err);
      setLoading(false);
    });
  }, []);

  // Sort categories: Owned items first
  const sortedCategories = useMemo(() => {
    return [...categories].sort((a, b) => {
      const aOwned = hasAccess(`wiki_${a.id}`) ? 1 : 0;
      const bOwned = hasAccess(`wiki_${b.id}`) ? 1 : 0;
      return bOwned - aOwned;
    });
  }, [categories, hasAccess]);

  // Level 1: Categories
  if (!selectedCategory) {
    return (
      <div className="animate-in fade-in slide-in-from-bottom-4 duration-500">
        <h2 className="text-2xl font-bold text-slate-900 mb-6">医学知识库</h2>
        {loading ? (
          <div className="text-center py-20 bg-white rounded-3xl border border-dashed border-slate-200 text-slate-400">
            <Book size={48} className="mx-auto mb-4 opacity-20 animate-pulse" />
            <p>正在加载知识库...</p>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            {sortedCategories.map(cat => {
              const isOwned = hasAccess(`wiki_${cat.id}`);
              return (
                <motion.div
                  key={cat.id}
                  whileHover={{ scale: 1.02 }}
                  onClick={() => setWikiState({ category: cat, article: null })}
                  className={`bg-white p-8 rounded-3xl border shadow-sm cursor-pointer transition-all relative overflow-hidden group ${isOwned
                    ? `border-emerald-200 ring-1 ring-emerald-50`
                    : 'border-slate-100 hover:shadow-xl'
                    }`}
                >
                  {isOwned ? (
                    <div className="absolute top-6 right-6 z-20 bg-emerald-100 text-emerald-600 px-2 py-1 rounded-full text-xs font-bold flex items-center gap-1">
                      <CheckCircle2 size={12} /> 已订阅
                    </div>
                  ) : (
                    <div className="absolute top-6 right-6 z-20 text-slate-200 group-hover:text-slate-300 transition-colors">
                      <Lock size={20} />
                    </div>
                  )}

                  <div className={`absolute top-0 right-0 w-32 h-32 bg-${cat.color}-50 rounded-bl-full -mr-8 -mt-8 transition-transform group-hover:scale-110`} />
                  <div className="relative z-10">
                    <span className={`text-${cat.color}-500 font-bold tracking-wider text-xs uppercase mb-2 block`}>KNOWLEDGE BASE</span>
                    <h3 className="text-2xl font-bold text-slate-900 mb-3">{cat.title}</h3>
                    <p className="text-slate-500 mb-6 max-w-sm">{cat.description}</p>
                    <div className="flex items-center gap-2 text-sm font-medium text-slate-400">
                      <Book size={16} /> {cat.articles.length} 篇文章
                    </div>
                  </div>
                </motion.div>
              );
            })}
          </div>
        )}
      </div>
    );
  }

  // Level 2: Article List (Protected)
  if (!selectedArticle) {
    return (
      <div className="animate-in fade-in slide-in-from-right-4 duration-300">
        <button
          onClick={() => setWikiState({ category: null, article: null })}
          className="flex items-center text-slate-500 hover:text-slate-900 mb-6 transition-colors"
        >
          <ArrowLeft size={20} className="mr-1" /> 返回分类
        </button>

        <div className="flex items-center gap-4 mb-8">
          <div className={`w-12 h-12 rounded-xl bg-${selectedCategory.color}-100 text-${selectedCategory.color}-600 flex items-center justify-center`}>
            <Book size={24} />
          </div>
          <div>
            <h2 className="text-3xl font-bold text-slate-900">{selectedCategory.title}</h2>
            <p className="text-slate-500">{selectedCategory.description}</p>
          </div>
        </div>

        {/* Access Guard now checks for 'wiki_' + id */}
        <AccessGuard accessId={`wiki_${selectedCategory.id}`} title={`查看${selectedCategory.title}文章`}>
          <div className="space-y-4">
            {selectedCategory.articles.length === 0 ? (
              <div className="text-center py-20 bg-white rounded-2xl border border-dashed border-slate-200 text-slate-400">
                <Book size={48} className="mx-auto mb-4 opacity-20" />
                <p>该分类下暂无文章</p>
              </div>
            ) : (
              selectedCategory.articles.map(article => (
                <motion.div
                  key={article.id}
                  initial={{ opacity: 0, y: 10 }}
                  animate={{ opacity: 1, y: 0 }}
                  onClick={() => setWikiState({ category: selectedCategory, article: article })}
                  className="bg-white p-6 rounded-2xl border border-slate-100 shadow-sm hover:shadow-md cursor-pointer transition-all flex justify-between items-center group"
                >
                  <div className="flex-1 pr-4">
                    <h3 className="text-lg font-bold text-slate-800 group-hover:text-blue-600 transition-colors mb-2">{article.title}</h3>
                    <p className="text-slate-500 text-sm mb-3 line-clamp-1">{article.excerpt}</p>
                    <div className="flex items-center gap-4 text-xs text-slate-400">
                      <span className="flex items-center gap-1"><User size={12} /> {article.author}</span>
                      <span className="flex items-center gap-1"><Clock size={12} /> {article.readTime}</span>
                      <div className="flex gap-1">
                        {article.tags.slice(0, 2).map(tag => (
                          <span key={tag} className="bg-slate-50 px-2 py-0.5 rounded text-slate-500">{tag}</span>
                        ))}
                      </div>
                    </div>
                  </div>
                  <div className={`w-10 h-10 rounded-full bg-slate-50 flex items-center justify-center text-slate-300 group-hover:bg-${selectedCategory.color}-50 group-hover:text-${selectedCategory.color}-500 transition-colors`}>
                    <ChevronRight size={20} />
                  </div>
                </motion.div>
              ))
            )}
          </div>
        </AccessGuard>
      </div>
    );
  }

  // Level 3: Article Content (Protected by Parent Guard implicitly)
  const bookmarkId = `wiki-${selectedArticle.id}`;
  const bookmarked = isBookmarked(bookmarkId);

  return (
    <div className="max-w-4xl mx-auto pb-20 animate-in fade-in zoom-in-95 duration-300">
      <button
        onClick={() => setWikiState({ category: selectedCategory, article: null })}
        className="flex items-center text-slate-500 hover:text-slate-900 mb-6 transition-colors"
      >
        <ArrowLeft size={20} className="mr-1" /> 返回列表
      </button>

      <article className="bg-white rounded-3xl shadow-lg border border-slate-100 overflow-hidden relative">
        {/* Article Header */}
        <div className={`bg-gradient-to-r from-${selectedCategory.color}-50 to-white p-8 md:p-12 border-b border-slate-100 relative`}>
          {/* Bookmark Button */}
          <button
            onClick={() => toggleBookmark({
              id: bookmarkId,
              type: 'article',
              title: selectedArticle.title,
              path: selectedCategory.title
            })}
            className={`absolute top-8 right-8 p-2 rounded-full transition-colors z-20 ${bookmarked ? 'text-amber-400 bg-white/80' : 'text-slate-300 hover:bg-white/50'}`}
          >
            <Star size={28} className={bookmarked ? "fill-amber-400" : ""} />
          </button>

          <div className="flex flex-wrap gap-2 mb-6">
            {selectedArticle.tags.map(t => (
              <span key={t} className={`px-3 py-1 bg-white/80 backdrop-blur rounded-full text-xs font-bold text-${selectedCategory.color}-700 border border-${selectedCategory.color}-100 uppercase tracking-wide shadow-sm`}>
                {t}
              </span>
            ))}
          </div>
          <h1 className="text-3xl md:text-5xl font-bold text-slate-900 mb-8 leading-tight">{selectedArticle.title}</h1>
          <div className="flex flex-wrap items-center gap-6 text-sm text-slate-500 font-medium">
            <span className="flex items-center gap-2 bg-white px-3 py-1.5 rounded-lg border border-slate-100 shadow-sm"><User size={16} className="text-blue-500" /> {selectedArticle.author}</span>
            <span className="flex items-center gap-2 bg-white px-3 py-1.5 rounded-lg border border-slate-100 shadow-sm"><Calendar size={16} className="text-orange-500" /> {selectedArticle.date}</span>
            <span className="flex items-center gap-2 bg-white px-3 py-1.5 rounded-lg border border-slate-100 shadow-sm"><Clock size={16} className="text-emerald-500" /> {selectedArticle.readTime} 阅读</span>
          </div>
        </div>

        {/* Article Content with Custom Markdown Components */}
        <div className="p-8 md:p-12 text-slate-800 leading-relaxed font-sans">
          <ReactMarkdown
            remarkPlugins={[remarkGfm]}
            components={{
              h1: ({ node, ...props }) => <h1 className="text-3xl font-bold mb-6 mt-2 text-slate-900" {...props} />,
              h2: ({ node, ...props }) => (
                <h2 className="text-2xl font-bold mt-10 mb-6 text-slate-800 flex items-center gap-3 pb-2 border-b border-slate-100" {...props}>
                  <span className={`w-1.5 h-6 bg-${selectedCategory.color}-500 rounded-full inline-block`}></span>
                  {props.children}
                </h2>
              ),
              h3: ({ node, ...props }) => <h3 className="text-xl font-bold mt-8 mb-4 text-slate-800 flex items-center gap-2" {...props}><Hash size={16} className="text-slate-300" />{props.children}</h3>,
              p: ({ node, ...props }) => <p className="mb-6 text-slate-600 leading-7 text-lg" {...props} />,
              ul: ({ node, ...props }) => <ul className="mb-6 space-y-2 text-slate-600" {...props} />,
              ol: ({ node, ...props }) => <ol className="list-decimal list-inside mb-6 space-y-2 text-slate-600 font-medium" {...props} />,
              li: ({ node, ...props }) => (
                <li className="flex gap-3 items-start ml-2" {...props}>
                  <span className={`mt-2 w-1.5 h-1.5 rounded-full bg-${selectedCategory.color}-400 shrink-0`}></span>
                  <span>{props.children}</span>
                </li>
              ),
              // Custom Blockquote for Medical Warnings/Notes
              blockquote: ({ node, children, ...props }) => (
                <div className="flex gap-4 my-8 bg-blue-50 border-l-4 border-blue-500 p-6 rounded-r-xl text-blue-900 shadow-sm">
                  <AlertCircle className="shrink-0 mt-0.5 text-blue-600" size={24} />
                  <div className="italic font-medium">{children}</div>
                </div>
              ),
              // Custom Table for Drug Dosage / Classifications
              table: ({ node, ...props }) => (
                <div className="my-8 overflow-hidden rounded-xl border border-slate-200 shadow-sm">
                  <div className="overflow-x-auto">
                    <table className="min-w-full divide-y divide-slate-200" {...props} />
                  </div>
                </div>
              ),
              thead: ({ node, ...props }) => <thead className="bg-slate-50" {...props} />,
              th: ({ node, ...props }) => <th className="px-6 py-4 text-left text-xs font-bold text-slate-500 uppercase tracking-wider" {...props} />,
              tbody: ({ node, ...props }) => <tbody className="bg-white divide-y divide-slate-100" {...props} />,
              tr: ({ node, ...props }) => <tr className="hover:bg-slate-50 transition-colors group" {...props} />,
              td: ({ node, ...props }) => <td className="px-6 py-4 whitespace-nowrap text-sm text-slate-600 group-hover:text-slate-900" {...props} />,
              a: ({ node, ...props }) => <a className={`text-${selectedCategory.color}-600 hover:underline font-medium cursor-pointer`} {...props} />,
              strong: ({ node, ...props }) => <strong className="font-bold text-slate-900 bg-slate-100 px-1 rounded" {...props} />,
              code: ({ node, ...props }) => <code className="bg-slate-100 px-1.5 py-0.5 rounded text-sm font-mono text-pink-600 border border-slate-200" {...props} />,
              hr: ({ node, ...props }) => <hr className="my-8 border-slate-100" {...props} />,
              img: ({ node, ...props }) => <img className="rounded-xl shadow-md my-6 w-full object-cover max-h-[400px]" {...props} />
            }}
          >
            {selectedArticle.content}
          </ReactMarkdown>
        </div>
      </article>
    </div>
  );
};

export default WikiView;
