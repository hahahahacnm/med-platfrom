
import React, { useState } from 'react';
import { useAppContext } from '../context';
import { Star, BookOpen, BrainCircuit, Trash2, ChevronRight, AlertCircle, ArrowRight } from 'lucide-react';
import { WIKI_CATEGORIES } from '../data';
import { WikiCategory, Article } from '../types';

const FavoritesView: React.FC = () => {
  const { user, toggleBookmark, setView, setWikiState } = useAppContext();
  const [filter, setFilter] = useState<'all' | 'question' | 'article'>('all');

  const filteredBookmarks = user.bookmarks.filter(b => 
    filter === 'all' ? true : b.type === filter
  );

  const handleWikiJump = (articleId: string) => {
    // Find category and article from static data
    let foundCat: WikiCategory | undefined;
    let foundArt: Article | undefined;

    for (const cat of WIKI_CATEGORIES) {
      const art = cat.articles.find(a => a.id === articleId);
      if (art) {
        foundCat = cat;
        foundArt = art;
        break;
      }
    }

    if (foundCat && foundArt) {
      setWikiState({ category: foundCat, article: foundArt });
      setView('wiki');
    } else {
      alert("文章数据未找到");
    }
  };

  return (
    <div className="max-w-4xl mx-auto space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-slate-900 flex items-center gap-2">
            <Star className="text-amber-400 fill-amber-400" size={28} /> 我的收藏
          </h2>
          <p className="text-slate-500 mt-1">
            共 {user.bookmarks.length} 条收藏内容
          </p>
        </div>
        
        <div className="bg-white p-1 rounded-lg border border-slate-200 flex text-sm font-medium">
          <button 
            onClick={() => setFilter('all')}
            className={`px-4 py-1.5 rounded-md transition-colors ${filter === 'all' ? 'bg-slate-100 text-slate-900' : 'text-slate-500 hover:text-slate-700'}`}
          >
            全部
          </button>
          <button 
            onClick={() => setFilter('question')}
            className={`px-4 py-1.5 rounded-md transition-colors ${filter === 'question' ? 'bg-indigo-50 text-indigo-700' : 'text-slate-500 hover:text-slate-700'}`}
          >
            题目
          </button>
          <button 
            onClick={() => setFilter('article')}
            className={`px-4 py-1.5 rounded-md transition-colors ${filter === 'article' ? 'bg-rose-50 text-rose-700' : 'text-slate-500 hover:text-slate-700'}`}
          >
            文章
          </button>
        </div>
      </div>

      {filteredBookmarks.length === 0 ? (
        <div className="text-center py-20 bg-white rounded-3xl border border-dashed border-slate-200">
           <Star size={48} className="mx-auto mb-4 text-slate-200" />
           <p className="text-slate-400">暂无收藏内容</p>
           <p className="text-sm text-slate-300 mt-1">在答题或阅读时点击星号收藏</p>
        </div>
      ) : (
        <div className="grid gap-4">
          {filteredBookmarks.map(bookmark => {
            if (bookmark.type === 'article') {
              return (
                <div key={bookmark.id} className="bg-white p-6 rounded-2xl border border-slate-100 shadow-sm hover:shadow-md transition-all group">
                   <div className="flex justify-between items-start mb-2">
                      <div className="flex items-center gap-2 text-xs font-bold text-rose-500 bg-rose-50 px-2 py-1 rounded mb-2 w-fit">
                        <BookOpen size={12} /> 知识库文章
                      </div>
                      <button 
                        onClick={(e) => { e.stopPropagation(); toggleBookmark(bookmark); }}
                        className="text-amber-400 hover:text-slate-300 transition-colors"
                      >
                        <Star size={20} className="fill-amber-400" />
                      </button>
                   </div>
                   
                   <h3 className="text-lg font-bold text-slate-900 mb-1">{bookmark.title}</h3>
                   <div className="text-xs text-slate-400 mb-4">{bookmark.path}</div>
                   
                   <button 
                     onClick={() => handleWikiJump(bookmark.id.replace('wiki-', ''))}
                     className="text-sm font-bold text-rose-600 flex items-center gap-1 hover:gap-2 transition-all"
                   >
                     查看全文 <ArrowRight size={16} />
                   </button>
                </div>
              );
            }

            // Question Item
            const qData = bookmark.data;
            return (
               <div key={bookmark.id} className="bg-white p-6 rounded-2xl border border-slate-100 shadow-sm hover:shadow-md transition-all">
                  <div className="flex justify-between items-start mb-3">
                      <div className="flex items-center gap-2 text-xs font-bold text-indigo-500 bg-indigo-50 px-2 py-1 rounded w-fit">
                        <BrainCircuit size={12} /> 题目收藏
                      </div>
                      <button 
                        onClick={() => toggleBookmark(bookmark)}
                        className="text-amber-400 hover:text-slate-300 transition-colors"
                      >
                        <Star size={20} className="fill-amber-400" />
                      </button>
                   </div>

                   <h3 className="font-bold text-slate-800 mb-4 text-lg">{bookmark.title}</h3>
                   
                   {qData && (
                     <div className="space-y-2 mb-4">
                        {qData.options.map((opt: any) => (
                          <div key={opt.id} className={`p-3 rounded-lg text-sm border flex items-center gap-3 ${
                             qData.correctAnswers.includes(opt.id) 
                               ? 'bg-emerald-50 border-emerald-200 text-emerald-800 font-medium'
                               : 'bg-slate-50 border-slate-100 text-slate-500'
                          }`}>
                             <span className={`w-6 h-6 rounded-full flex items-center justify-center text-xs border ${
                                qData.correctAnswers.includes(opt.id) ? 'border-emerald-500 bg-emerald-500 text-white' : 'border-slate-300 bg-white'
                             }`}>
                               {opt.id}
                             </span>
                             {opt.text}
                          </div>
                        ))}
                     </div>
                   )}

                   {qData && (
                     <div className="bg-slate-50 p-4 rounded-xl text-sm border border-slate-100">
                        <div className="font-bold text-slate-700 mb-1 flex items-center gap-1">
                           <AlertCircle size={14} /> 解析
                        </div>
                        <p className="text-slate-600">{qData.explanation}</p>
                     </div>
                   )}
                   
                   <div className="mt-4 pt-4 border-t border-slate-50 text-xs text-slate-400 flex justify-between">
                      <span>来源: {bookmark.path}</span>
                      <span>收藏于 {new Date(bookmark.timestamp).toLocaleDateString()}</span>
                   </div>
               </div>
            );
          })}
        </div>
      )}
    </div>
  );
};

export default FavoritesView;
