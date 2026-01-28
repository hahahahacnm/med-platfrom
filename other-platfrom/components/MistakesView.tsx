import React, { useMemo, useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { Eraser, AlertCircle, ChevronDown, ChevronUp, CheckCircle, XCircle, Trash2 } from 'lucide-react';
import { useAppContext } from '../context';
import { SUBJECTS } from '../data';
import { parseRawDataToChapters } from './QuizView';
import { Question } from '../types';

interface MistakeItem {
  subjectId: string;
  subjectTitle: string;
  chapterTitle: string;
  question: Question;
  originalIndex: number;
}

interface MistakeGroup {
  subjectTitle: string;
  items: MistakeItem[];
}

const MistakesView: React.FC = () => {
  const { user, updateChapterProgress } = useAppContext();
  const [expandedItems, setExpandedItems] = useState<Record<string, boolean>>({});

  // Dynamically calculate mistakes from progress history
  const groupedMistakes = useMemo<Record<string, MistakeGroup>>(() => {
    const mistakes: Record<string, MistakeGroup> = {};

    SUBJECTS.forEach(subject => {
      const chapters = parseRawDataToChapters(subject.rawContent);

      chapters.forEach(chapter => {
        const progressKey = `${subject.id}_${chapter.title}`;
        const history = user.chapterProgress[progressKey]?.history;

        if (history) {
          chapter.questions.forEach((q, index) => {
            if (history[index] === false) {
              // Found a mistake
              if (!mistakes[subject.id]) {
                mistakes[subject.id] = { subjectTitle: subject.title, items: [] };
              }
              mistakes[subject.id].items.push({
                subjectId: subject.id,
                subjectTitle: subject.title,
                chapterTitle: chapter.title,
                question: q,
                originalIndex: index
              });
            }
          });
        }
      });
    });

    return mistakes;
  }, [user.chapterProgress]);

  const totalMistakes = Object.values(groupedMistakes).reduce((acc: number, curr: MistakeGroup) => acc + curr.items.length, 0);

  const toggleExpand = (id: string) => {
    setExpandedItems(prev => ({ ...prev, [id]: !prev[id] }));
  };

  const removeMistake = (subjectId: string, chapterTitle: string, index: number) => {
    // To "remove" a mistake, we conceptually remove the 'false' record. 
    // We can set it to 'undefined' or delete the key, but updateChapterProgress usually expects setting a value.
    // However, our updateChapterProgress logic in App.tsx doesn't support 'delete'.
    // A workaround is to treat it as 'cleared' (maybe strictly we shouldn't change history unless re-answered correct).
    // For this feature, let's assume we can remove the entry from the history object in App.tsx context logic,
    // but since we only exposed `updateChapterProgress` which sets values, we might just set it to `true` (Correct) 
    // to "fix" the mistake, OR we just alert user they need to retake the quiz.
    // But users expect "Clear" button. Let's assume we simply hide it by setting it to undefined in a real app.
    // Given current constraints, we'll mark it as "Correct" (Green) to clear it from the "Mistakes" list (which filters === false).
    if (confirm("确定要移除这道错题吗？(这将标记该题为已掌握)")) {
      updateChapterProgress(subjectId, chapterTitle, index, true);
    }
  };

  return (
    <div className="max-w-4xl mx-auto space-y-8 animate-in fade-in duration-500">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-slate-900 flex items-center gap-2">
            <Eraser className="text-rose-500" size={28} /> 错题集
          </h2>
          <p className="text-slate-500 mt-1">
            自动收录练习中的错题，共 {totalMistakes} 道待复习
          </p>
        </div>
      </div>

      {totalMistakes === 0 ? (
        <div className="text-center py-20 bg-white rounded-3xl border border-dashed border-slate-200">
          <CheckCircle size={48} className="mx-auto mb-4 text-emerald-200" />
          <p className="text-slate-400">太棒了！暂无错题记录。</p>
          <p className="text-sm text-slate-300 mt-1">继续保持，多做练习巩固知识点。</p>
        </div>
      ) : (
        <div className="space-y-8">
          {Object.entries(groupedMistakes).map(([subjectId, group]: [string, MistakeGroup]) => (
            <div key={subjectId} className="space-y-4">
              <h3 className="font-bold text-lg text-slate-800 flex items-center gap-2 border-l-4 border-slate-300 pl-3">
                {group.subjectTitle} <span className="text-xs bg-slate-100 text-slate-500 px-2 py-0.5 rounded-full">{group.items.length}题</span>
              </h3>

              <div className="grid gap-4">
                {group.items.map((item) => {
                  const uniqueId = `${item.subjectId}-${item.chapterTitle}-${item.question.id}`;
                  const isExpanded = expandedItems[uniqueId];

                  return (
                    <div key={uniqueId} className="bg-white rounded-2xl border border-slate-100 shadow-sm overflow-hidden transition-all hover:shadow-md">
                      <div
                        className="p-5 cursor-pointer flex justify-between items-start gap-4"
                        onClick={() => toggleExpand(uniqueId)}
                      >
                        <div className="flex-1">
                          <div className="flex items-center gap-2 mb-2">
                            <span className="text-xs font-bold text-slate-400 bg-slate-50 px-2 py-1 rounded border border-slate-100">
                              {item.chapterTitle}
                            </span>
                            <span className="text-xs font-bold text-rose-500 bg-rose-50 px-2 py-1 rounded">
                              错题
                            </span>
                          </div>
                          <h4 className="font-bold text-slate-800 text-lg leading-relaxed">{item.question.text}</h4>
                        </div>
                        <div className="text-slate-300">
                          {isExpanded ? <ChevronUp /> : <ChevronDown />}
                        </div>
                      </div>

                      <AnimatePresence>
                        {isExpanded && (
                          <motion.div
                            initial={{ height: 0, opacity: 0 }}
                            animate={{ height: 'auto', opacity: 1 }}
                            exit={{ height: 0, opacity: 0 }}
                            className="bg-slate-50 border-t border-slate-100 overflow-hidden"
                          >
                            <div className="p-5 space-y-4">
                              {/* Options */}
                              <div className="space-y-2">
                                {item.question.options.map(opt => (
                                  <div key={opt.id} className={`p-3 rounded-lg border flex items-center gap-3 text-sm ${item.question.correctAnswers.includes(opt.id)
                                      ? 'bg-emerald-50 border-emerald-200 text-emerald-800 font-bold'
                                      : 'bg-white border-slate-200 text-slate-500'
                                    }`}>
                                    <span className={`w-6 h-6 rounded-full flex items-center justify-center text-xs border ${item.question.correctAnswers.includes(opt.id)
                                        ? 'border-emerald-500 bg-emerald-500 text-white'
                                        : 'border-slate-300 bg-slate-50'
                                      }`}>
                                      {opt.id}
                                    </span>
                                    {opt.text}
                                  </div>
                                ))}
                              </div>

                              {/* Explanation */}
                              <div className="bg-white p-4 rounded-xl border border-slate-200">
                                <div className="flex items-center gap-2 font-bold text-slate-800 mb-2">
                                  <AlertCircle size={16} className="text-blue-500" /> 解析
                                </div>
                                <p className="text-slate-600 text-sm leading-relaxed">
                                  {item.question.explanation}
                                </p>
                              </div>

                              {/* Actions */}
                              <div className="flex justify-end pt-2">
                                <button
                                  onClick={(e) => { e.stopPropagation(); removeMistake(item.subjectId, item.chapterTitle, item.originalIndex); }}
                                  className="flex items-center gap-1 text-sm text-slate-400 hover:text-rose-500 transition-colors px-3 py-1.5 rounded-lg hover:bg-rose-50"
                                >
                                  <Trash2 size={16} /> 移除此题
                                </button>
                              </div>
                            </div>
                          </motion.div>
                        )}
                      </AnimatePresence>
                    </div>
                  );
                })}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

export default MistakesView;