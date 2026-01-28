
import React, { useState, useMemo, useEffect } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { ArrowLeft, CheckCircle2, XCircle, AlertCircle, ChevronRight, BrainCircuit, Layers, FileText, Settings2, Grid, RotateCcw, PlayCircle, Zap, GraduationCap, FileCheck, Star, Lock, MessageSquare, Send } from 'lucide-react';
import { useAppContext } from '../context';
import { api } from '../services/api';
import { Subject, Question, Chapter, QuizMode, Comment } from '../types';
import { AccessGuard } from './AccessGuard';

// Helper to parse raw text into structured chapters (Restored for local usage/MistakesView)
export const parseRawDataToChapters = (rawContent?: string): Chapter[] => {
  if (!rawContent) return [];

  const chapters: Chapter[] = [];
  const lines = rawContent.split('\n');
  let currentChapter: Chapter = { id: 'default', title: '默认章节', questions: [] };
  let currentQuestion: Partial<Question> | null = null;

  const saveQuestion = () => {
    if (currentQuestion && currentQuestion.text) {
      currentChapter.questions.push(currentQuestion as Question);
      currentQuestion = null;
    }
  };

  const saveChapter = () => {
    if (currentChapter.questions.length > 0) {
      if (currentChapter.id === 'default') currentChapter.id = `chap_${chapters.length + 1}`;
      chapters.push({ ...currentChapter, questions: [...currentChapter.questions] });
    }
    currentChapter = { id: `chap_${chapters.length + 2}`, title: '默认章节', questions: [] };
  };

  lines.forEach((line) => {
    const trimmed = line.trim();
    if (!trimmed) return;

    if (trimmed.startsWith('###')) {
      saveQuestion();
      saveChapter();
      currentChapter.title = trimmed.replace(/^###\s*/, '').trim();
      currentChapter.id = currentChapter.title;
    } else if (/^\d+\./.test(trimmed)) {
      saveQuestion();
      const dotIndex = trimmed.indexOf('.');
      const qId = parseInt(trimmed.substring(0, dotIndex), 10);
      const qText = trimmed.substring(dotIndex + 1).trim();
      currentQuestion = {
        id: isNaN(qId) ? 0 : qId,
        text: qText,
        options: [],
        correctAnswers: [],
        explanation: ''
      };
    } else if (/^[A-Z]\./.test(trimmed)) {
      if (currentQuestion) {
        const dotIndex = trimmed.indexOf('.');
        const optId = trimmed.substring(0, dotIndex).trim();
        const optText = trimmed.substring(dotIndex + 1).trim();
        currentQuestion.options = currentQuestion.options || [];
        currentQuestion.options.push({ id: optId, text: optText });
      }
    } else if (trimmed.startsWith('答案：') || trimmed.startsWith('答案:')) {
      if (currentQuestion) {
        const ans = trimmed.replace(/^答案[:：]/, '').trim();
        currentQuestion.correctAnswers = ans.split('').map(x => x.trim());
      }
    } else if (trimmed.startsWith('解析：') || trimmed.startsWith('解析:')) {
      if (currentQuestion) {
        currentQuestion.explanation = trimmed.replace(/^解析[:：]/, '').trim();
      }
    }
  });

  saveQuestion();
  saveChapter();

  return chapters;
};


// --- Main Quiz View ---
const QuizView: React.FC = () => {
  const { hasAccess } = useAppContext();
  const [subjects, setSubjects] = useState<Subject[]>([]);
  const [selectedSubject, setSelectedSubject] = useState<Subject | null>(null);
  const [selectedChapter, setSelectedChapter] = useState<Chapter | null>(null);
  const [selectedMode, setSelectedMode] = useState<QuizMode | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const loadSubjects = async () => {
      try {
        const data = await api.quiz.getSubjects();
        setSubjects(data);
      } catch (err) {
        console.error("Failed to load subjects", err);
      } finally {
        setLoading(false);
      }
    };
    loadSubjects();
  }, []);

  // Sort subjects: Owned items first
  const sortedSubjects = useMemo(() => {
    return [...subjects].sort((a, b) => {
      const aOwned = hasAccess(`quiz_${a.id}`) ? 1 : 0;
      const bOwned = hasAccess(`quiz_${b.id}`) ? 1 : 0;
      return bOwned - aOwned;
    });
  }, [hasAccess, subjects]);

  // --- Step 1: Subject Selection ---
  if (!selectedSubject) {
    return (
      <div>
        <h2 className="text-2xl font-bold text-slate-900 mb-6">选择题库科目</h2>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {sortedSubjects.map(sub => {
            const isOwned = hasAccess(`quiz_${sub.id}`);
            return (
              <motion.div
                key={sub.id}
                whileHover={{ y: -5 }}
                onClick={() => setSelectedSubject(sub)}
                className={`bg-white p-6 rounded-2xl border transition-all cursor-pointer group relative overflow-hidden ${isOwned
                  ? 'border-emerald-200 shadow-md ring-1 ring-emerald-50'
                  : `border-slate-100 shadow-sm hover:shadow-lg hover:border-${sub.color}-200`
                  }`}
              >
                {isOwned ? (
                  <div className="absolute top-0 right-0 bg-emerald-500 text-white text-[10px] font-bold px-2 py-1 rounded-bl-lg shadow-sm z-10 flex items-center gap-1">
                    <CheckCircle2 size={10} /> 已解锁
                  </div>
                ) : (
                  <div className="absolute top-4 right-4 text-slate-200 group-hover:text-slate-300 transition-colors">
                    <Lock size={16} />
                  </div>
                )}

                <div className={`w-12 h-12 rounded-xl bg-${sub.color}-50 text-${sub.color}-600 flex items-center justify-center text-2xl mb-4 group-hover:scale-110 transition-transform`}>
                  {sub.icon}
                </div>
                <h3 className="text-lg font-bold text-slate-900">{sub.title}</h3>
                <p className="text-sm text-slate-500 mt-2">{sub.description}</p>
                <div className="mt-4 flex items-center text-xs font-semibold text-slate-400">
                  <BrainCircuit size={14} className="mr-1" /> 点击进入章节
                </div>
              </motion.div>
            );
          })}
        </div>
      </div>
    );
  }

  // --- Wrapper with Access Guard ---
  return (
    <div>
      <AccessGuard
        accessId={`quiz_${selectedSubject.id}`}
        title={`解锁${selectedSubject.title}题库`}
        onBack={() => setSelectedSubject(null)}
        backLabel="返回科目列表"
      >
        {!selectedChapter ? (
          <ChapterSelection
            subject={selectedSubject}
            onSelect={(chapter, mode) => {
              setSelectedChapter(chapter);
              setSelectedMode(mode);
            }}
            onBack={() => setSelectedSubject(null)}
          />
        ) : selectedMode ? (
          <ActiveQuiz
            subject={selectedSubject}
            chapter={selectedChapter}
            mode={selectedMode}
            onFinish={() => {
              setSelectedChapter(null);
              setSelectedMode(null);
            }}
          />
        ) : null}
      </AccessGuard>
    </div>
  );
};

// --- Sub-component: Chapter List ---
const ChapterSelection: React.FC<{
  subject: Subject;
  onSelect: (c: Chapter, m: QuizMode) => void;
  onBack: () => void;
}> = ({ subject, onSelect, onBack }) => {
  const { user, resetChapterProgress } = useAppContext();
  const chapters = useMemo(() => subject.chapters || [], [subject]);
  const [modalChapter, setModalChapter] = useState<Chapter | null>(null);

  // Helper to get progress
  const getProgress = (title: string) => {
    return user.chapterProgress[`${subject.id}_${title}`];
  };

  return (
    <div className="max-w-4xl mx-auto">
      <button onClick={onBack} className="flex items-center text-slate-500 hover:text-slate-900 mb-6 transition-colors">
        <ArrowLeft size={20} className="mr-1" /> 返回科目列表
      </button>

      <div className="mb-8">
        <h2 className="text-2xl font-bold text-slate-900 flex items-center gap-2">
          <span className={`text-${subject.color}-600`}>{subject.icon}</span>
          {subject.title} - 章节练习
        </h2>
        <p className="text-slate-500 mt-1">请选择章节并设定答题模式</p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {chapters.map((chapter, index) => {
          const progress = getProgress(chapter.title);
          const doneCount = progress ? Object.keys(progress.history).length : 0;

          return (
            <motion.div
              key={index}
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: index * 0.05 }}
              onClick={() => setModalChapter(chapter)}
              className="bg-white p-5 rounded-xl border border-slate-100 shadow-sm hover:shadow-md hover:border-blue-200 cursor-pointer flex items-center justify-between group transition-all relative overflow-hidden"
            >
              <div className="flex items-center gap-4 relative z-10">
                <div className={`w-10 h-10 rounded-full flex items-center justify-center font-bold text-sm transition-colors ${progress ? 'bg-blue-100 text-blue-600' : 'bg-slate-50 text-slate-400'}`}>
                  {index + 1}
                </div>
                <div>
                  <h3 className="font-bold text-slate-800 group-hover:text-blue-600 transition-colors">{chapter.title}</h3>
                  <div className="flex items-center gap-3 mt-1">
                    <span className="text-xs text-slate-400 flex items-center gap-1">
                      <FileText size={12} /> {chapter.questions.length} 题
                    </span>
                    {progress && (
                      <span className="text-xs font-bold text-blue-600 flex items-center gap-1 bg-blue-50 px-1.5 py-0.5 rounded">
                        <RotateCcw size={10} /> 继续做 ({doneCount}/{chapter.questions.length})
                      </span>
                    )}
                  </div>
                </div>
              </div>
              <ChevronRight size={20} className="text-slate-300 group-hover:translate-x-1 transition-transform relative z-10" />
            </motion.div>
          );
        })}
      </div>

      {/* Mode Selection Modal */}
      <AnimatePresence>
        {modalChapter && (
          <div className="fixed inset-0 z-50 flex items-center justify-center px-4">
            <motion.div
              initial={{ opacity: 0 }} animate={{ opacity: 1 }} exit={{ opacity: 0 }}
              className="absolute inset-0 bg-slate-900/40 backdrop-blur-sm"
              onClick={() => setModalChapter(null)}
            />
            <motion.div
              initial={{ scale: 0.9, opacity: 0 }} animate={{ scale: 1, opacity: 1 }} exit={{ scale: 0.9, opacity: 0 }}
              className="bg-white rounded-3xl p-6 md:p-8 w-full max-w-lg relative z-10 shadow-2xl"
            >
              <h3 className="text-xl font-bold text-slate-900 mb-2">选择答题模式</h3>
              <p className="text-slate-500 mb-6 text-sm">章节：{modalChapter.title}</p>

              <div className="grid grid-cols-2 gap-4 mb-6">
                <ModeCard
                  title="练习模式"
                  desc="做完显示答案，需手动下一题"
                  icon={<PlayCircle size={24} />}
                  color="blue"
                  onClick={() => onSelect(modalChapter, 'practice')}
                />
                <ModeCard
                  title="快刷模式"
                  desc="答对自动下一题，答错停留"
                  icon={<Zap size={24} />}
                  color="amber"
                  onClick={() => onSelect(modalChapter, 'fast')}
                />
                <ModeCard
                  title="测试模式"
                  desc="全真模拟，交卷后显示结果"
                  icon={<FileCheck size={24} />}
                  color="rose"
                  onClick={() => onSelect(modalChapter, 'test')}
                />
                <ModeCard
                  title="背题模式"
                  desc="直接显示答案与解析"
                  icon={<GraduationCap size={24} />}
                  color="emerald"
                  onClick={() => onSelect(modalChapter, 'study')}
                />
              </div>

              {getProgress(modalChapter.title) && (
                <div className="border-t border-slate-100 pt-6">
                  <h4 className="text-xs font-bold text-slate-400 uppercase mb-3">记录管理</h4>
                  <div className="flex gap-4">
                    <div
                      onClick={() => onSelect(modalChapter, 'practice')}
                      className="flex-1 bg-blue-50 border border-blue-100 p-3 rounded-xl cursor-pointer hover:bg-blue-100 transition-colors flex items-center gap-3"
                    >
                      <div className="bg-blue-500 text-white p-1.5 rounded-lg"><RotateCcw size={16} /></div>
                      <div>
                        <div className="text-xs text-blue-800 font-bold">继续刷题</div>
                        <div className="text-xs text-blue-600">从上次进度开始</div>
                      </div>
                    </div>

                    <div
                      onClick={() => {
                        if (confirm('确定要清除本章节的做题记录吗？')) {
                          resetChapterProgress(subject.id, modalChapter.title);
                          setModalChapter(null);
                        }
                      }}
                      className="flex-1 bg-red-50 border border-red-100 p-3 rounded-xl cursor-pointer hover:bg-red-100 transition-colors flex items-center gap-3"
                    >
                      <div className="bg-red-500 text-white p-1.5 rounded-lg"><XCircle size={16} /></div>
                      <div>
                        <div className="text-xs text-red-800 font-bold">清除记录</div>
                        <div className="text-xs text-red-600">重置所有进度</div>
                      </div>
                    </div>
                  </div>
                </div>
              )}

              <button
                onClick={() => setModalChapter(null)}
                className="absolute top-4 right-4 text-slate-400 hover:text-slate-600 p-2"
              >
                <XCircle size={24} />
              </button>
            </motion.div>
          </div>
        )}
      </AnimatePresence>
    </div>
  );
};

const ModeCard = ({ title, desc, icon, color, onClick }: any) => (
  <button
    onClick={onClick}
    className={`flex flex-col items-start p-4 rounded-xl border border-slate-100 bg-slate-50 hover:bg-${color}-50 hover:border-${color}-200 transition-all group text-left h-full`}
  >
    <div className={`text-${color}-500 mb-2 group-hover:scale-110 transition-transform`}>{icon}</div>
    <div className="font-bold text-slate-800 text-sm mb-1">{title}</div>
    <div className="text-xs text-slate-500 leading-tight">{desc}</div>
  </button>
);

// --- Sub-component: Active Quiz Interface ---
const ActiveQuiz: React.FC<{
  subject: Subject;
  chapter: Chapter;
  mode: QuizMode;
  onFinish: () => void
}> = ({ subject, chapter, mode, onFinish }) => {
  const { user, addQuizResult, updateChapterProgress, toggleBookmark, isBookmarked } = useAppContext();

  // Progress Key
  const progressKey = `${subject.id}_${chapter.title}`;
  const savedProgress = user.chapterProgress[progressKey];

  const [questions] = useState(chapter.questions);
  const [currentIndex, setCurrentIndex] = useState(savedProgress?.lastIndex || 0);
  const [selected, setSelected] = useState<string[]>([]);
  const [isAnswered, setIsAnswered] = useState(false);
  const [showGrid, setShowGrid] = useState(false);

  // Use useMemo to ensure history object reference is stable
  const history = useMemo(() => savedProgress?.history || {}, [savedProgress]);

  // Ref to track history locally for immediate access
  const historyRef = React.useRef(history);
  // Sync ref with history from context
  useEffect(() => {
    historyRef.current = { ...historyRef.current, ...history };
  }, [history]);

  const currentQ = questions[currentIndex];
  // Determine if question is multiple choice
  const isMulti = useMemo(() => currentQ.correctAnswers.length > 1, [currentQ]);

  // Construct Unique Bookmark ID
  const bookmarkId = `quiz-${subject.id}-${chapter.title}-${currentQ.id}`;
  const bookmarked = isBookmarked(bookmarkId);

  // Mode-specific initial state
  useEffect(() => {
    // If Study mode, we are always "Answered" conceptually to show analysis
    if (mode === 'study') {
      setIsAnswered(true);
    } else if (history[currentIndex] !== undefined && mode !== 'test') {
      // If we visited this question before and it's recorded (and not test mode where history is hidden)
      setIsAnswered(true);
    } else {
      setIsAnswered(false);
      setSelected([]);
    }
  }, [currentIndex, mode, history]);

  const handleSelect = (id: string) => {
    if (isAnswered && mode !== 'test') return;

    if (mode === 'study') return;

    if (isMulti) {
      // Multiple choice toggle
      if (selected.includes(id)) setSelected(prev => prev.filter(i => i !== id));
      else setSelected(prev => [...prev, id]);
    } else {
      // Single choice switch: always replace selection
      setSelected([id]);
    }
  };

  const processResult = () => {
    const isCorrect = selected.sort().join('') === currentQ.correctAnswers.sort().join('');
    // Update local Ref immediately
    historyRef.current = { ...historyRef.current, [currentIndex]: isCorrect };
    // Update Global Progress
    updateChapterProgress(subject.id, chapter.title, currentIndex, isCorrect);
    return isCorrect;
  };

  const handleNext = () => {
    if (currentIndex < questions.length - 1) {
      const nextIdx = currentIndex + 1;
      setCurrentIndex(nextIdx);
      // Update progress index only
      updateChapterProgress(subject.id, chapter.title, nextIdx);
    } else {
      finishQuiz();
    }
  };

  const handleSubmit = () => {
    if (selected.length === 0) return;

    const isCorrect = processResult();

    if (mode === 'test') {
      handleNext();
    } else if (mode === 'fast') {
      setIsAnswered(true);
      if (isCorrect) {
        // Delay slightly then auto next
        setTimeout(() => {
          handleNext();
        }, 500);
      }
      // If wrong, stay on page to read analysis
    } else {
      // Practice
      setIsAnswered(true);
    }
  };

  const handlePrev = () => {
    if (currentIndex > 0) {
      const prevIdx = currentIndex - 1;
      setCurrentIndex(prevIdx);
      updateChapterProgress(subject.id, chapter.title, prevIdx);
    }
  };

  const jumpToQuestion = (index: number) => {
    setCurrentIndex(index);
    updateChapterProgress(subject.id, chapter.title, index);
    setShowGrid(false);
  };

  const finishQuiz = () => {
    // Calculate Score based on History Ref (latest data)
    const correctCount = Object.values(historyRef.current).filter(v => v === true).length;
    addQuizResult({
      subjectId: `${subject.title} - ${chapter.title} [${mode}]`,
      total: questions.length,
      correct: correctCount,
      date: new Date().toLocaleDateString()
    });
    alert(`练习结束！\n正确率: ${correctCount}/${questions.length}`);
    onFinish();
  };



  // Check if it is an Essay/Simple Answer question
  const isEssay = useMemo(() => (!currentQ.options || currentQ.options.length === 0), [currentQ]);

  // Handle Essay Self-Check
  const handleSelfCheck = (isCorrect: boolean) => {
    historyRef.current = { ...historyRef.current, [currentIndex]: isCorrect };
    updateChapterProgress(subject.id, chapter.title, currentIndex, isCorrect);
    // Don't auto-next
  };

  // Determine styles based on state
  const getOptionStyle = (optId: string) => {
    if (mode === 'test') {
      return selected.includes(optId) ? "border-blue-500 bg-blue-50 text-blue-700 shadow-md" : "border-slate-200 hover:bg-slate-50";
    }
    if (mode === 'study') {
      const isCorrect = currentQ.correctAnswers.includes(optId);
      return isCorrect ? "border-emerald-500 bg-emerald-50 text-emerald-700 font-bold" : "border-slate-100 opacity-60";
    }

    if (isAnswered) {
      const isCorrect = currentQ.correctAnswers.includes(optId);
      const isSelected = selected.includes(optId);
      if (isCorrect) return "border-emerald-500 bg-emerald-50 text-emerald-700";
      if (isSelected && !isCorrect) return "border-red-500 bg-red-50 text-red-700";
      return "border-slate-100 opacity-50";
    } else {
      return selected.includes(optId) ? "border-blue-500 bg-blue-50 text-blue-700 shadow-md" : "border-slate-200 hover:bg-slate-50";
    }
  };

  return (
    <div className="max-w-3xl mx-auto relative min-h-[80vh] flex flex-col pt-0">
      {/* Header */}
      <div className="flex items-center justify-between mb-2">
        <button onClick={onFinish} className="p-2 -ml-2 text-slate-400 hover:text-slate-600">
          <ArrowLeft size={24} />
        </button>

        <div
          onClick={() => setShowGrid(true)}
          className="flex flex-col items-center cursor-pointer hover:bg-slate-100 px-4 py-1 rounded-xl transition-colors"
        >
          <span className="text-xs text-slate-400 font-bold tracking-wider uppercase mb-0.5">{mode === 'study' ? '背题模式' : mode === 'test' ? '测试模式' : mode === 'fast' ? '快刷模式' : '练习模式'}</span>
          <div className="flex items-center gap-1 font-bold text-slate-800">
            <span className="text-xl">{currentIndex + 1}</span>
            <span className="text-slate-300">/</span>
            <span className="text-sm">{questions.length}</span>
            <ChevronRight size={14} className="ml-1 text-slate-400 rotate-90" />
          </div>
        </div>

        <button onClick={() => setShowGrid(true)} className="p-2 -mr-2 text-slate-400 hover:text-blue-600">
          <Grid size={24} />
        </button>
      </div>

      {/* Progress Bar */}
      <div className="h-1 w-full bg-slate-100 rounded-full mb-4 overflow-hidden">
        <motion.div
          className={`h-full ${mode === 'test' ? 'bg-rose-500' : 'bg-blue-500'}`}
          initial={{ width: 0 }}
          animate={{ width: `${((currentIndex + 1) / questions.length) * 100}%` }}
        />
      </div>

      {/* Question Card */}
      <div className="bg-white rounded-3xl shadow-sm border border-slate-200 p-5 md:p-8 flex-1 relative">
        {/* Bookmark Button */}
        <button
          onClick={() => toggleBookmark({
            id: bookmarkId,
            type: 'question',
            title: currentQ.text,
            path: `${subject.title} > ${chapter.title}`,
            data: currentQ
          })}
          className={`absolute top-6 right-6 p-2 rounded-full transition-colors ${bookmarked ? 'text-amber-400 bg-amber-50' : 'text-slate-300 hover:bg-slate-50'}`}
        >
          <Star size={24} className={bookmarked ? "fill-amber-400" : ""} />
        </button>

        <div className="mb-3 pr-12">
          <span className={`inline-block px-2 py-0.5 rounded text-xs font-bold mb-3 ${isEssay ? 'bg-amber-100 text-amber-700' : isMulti ? 'bg-purple-100 text-purple-700' : 'bg-blue-100 text-blue-700'}`}>
            {isEssay ? '问答题' : isMulti ? '多选题' : '单选题'}
          </span>
          <h3 className="text-base font-bold text-slate-800 leading-relaxed whitespace-pre-wrap">{currentQ.text}</h3>
        </div>

        {isEssay ? (
          <div className="mb-8">
            <textarea
              className="w-full h-40 p-4 border border-slate-200 rounded-xl focus:outline-none focus:border-blue-500 bg-slate-50 text-slate-700 resize-none"
              placeholder="在此输入您的思考或答案..."
              readOnly={mode === 'study' || (isAnswered && mode !== 'test')}
            ></textarea>
          </div>
        ) : (
          <div className="space-y-3 mb-6">
            {currentQ.options.map(opt => (
              <button
                key={opt.id}
                onClick={() => handleSelect(opt.id)}
                disabled={mode === 'study' || (isAnswered && mode !== 'test')}
                className={`w-full text-left p-4 rounded-xl border-2 transition-all flex items-center justify-between group ${getOptionStyle(opt.id)}`}
              >
                <div className="flex items-center gap-4">
                  <span className={`w-8 h-8 rounded-full flex items-center justify-center font-bold text-sm border-2 ${getOptionStyle(opt.id).includes('emerald') || getOptionStyle(opt.id).includes('blue') ? 'border-transparent bg-white/20' : 'border-slate-200 bg-white'}`}>
                    {opt.id}
                  </span>
                  <span className="text-base font-medium">{opt.text}</span>
                </div>
                {(mode === 'study' || (isAnswered && mode !== 'test')) && currentQ.correctAnswers.includes(opt.id) && <CheckCircle2 className="text-emerald-600" size={20} />}
                {(isAnswered && mode !== 'test') && selected.includes(opt.id) && !currentQ.correctAnswers.includes(opt.id) && <XCircle className="text-red-500" size={20} />}
              </button>
            ))}
          </div>
        )}

        {/* Navigation Controls (Moved above Analysis) */}
        <div className="flex justify-between items-center mb-6">
          <button
            onClick={handlePrev}
            disabled={currentIndex === 0}
            className="px-6 py-2.5 rounded-xl font-bold text-slate-500 hover:bg-slate-100 disabled:opacity-30 transition-colors border border-slate-100"
          >
            上一题
          </button>

          {mode !== 'study' && !isAnswered && mode !== 'test' ? (
            <button
              onClick={() => {
                if (isEssay) {
                  setIsAnswered(true);
                } else {
                  handleSubmit();
                }
              }}
              disabled={!isEssay && selected.length === 0}
              className="bg-slate-900 text-white px-10 py-2.5 rounded-xl font-bold hover:bg-blue-600 transition-colors shadow-lg shadow-slate-200 disabled:opacity-50 disabled:shadow-none"
            >
              {isEssay ? '查看答案' : '提交'}
            </button>
          ) : (
            <button
              onClick={handleNext}
              className="bg-blue-600 text-white px-10 py-2.5 rounded-xl font-bold hover:bg-blue-700 transition-colors shadow-lg shadow-blue-200 flex items-center gap-2"
            >
              {currentIndex === questions.length - 1 ? '完成' : '下一题'} <ChevronRight size={18} />
            </button>
          )}
        </div>

        {/* Analysis Area */}
        {(mode === 'study' || (isAnswered && mode !== 'test')) && (
          <motion.div
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            className="bg-slate-50 p-6 rounded-2xl border border-slate-100"
          >
            <div className="flex items-center gap-2 font-bold text-slate-800 mb-3">
              <AlertCircle size={18} className="text-blue-500" />
              解析
            </div>
            {isEssay ? (
              <div className="mb-4">
                <div className="text-sm font-bold text-slate-900 mb-1">参考答案：</div>
                <div className="text-emerald-700 bg-emerald-50 p-3 rounded-lg text-sm leading-relaxed whitespace-pre-wrap">
                  {currentQ.correctAnswers.join('\n')}
                </div>
              </div>
            ) : (
              <div className="text-sm font-bold text-slate-900 mb-2">
                正确答案：<span className="text-emerald-600">{currentQ.correctAnswers.join('')}</span>
              </div>
            )}
            <p className="text-slate-600 text-sm leading-relaxed">{currentQ.explanation || "暂无详细解析"}</p>

            {isEssay && mode !== 'study' && (
              <div className="mt-6 pt-4 border-t border-slate-200">
                <p className="text-sm text-slate-500 mb-3 text-center">请根据参考答案自评：</p>
                <div className="flex gap-4 max-w-xs mx-auto">
                  <button
                    onClick={() => handleSelfCheck(true)}
                    className={`flex-1 py-3 rounded-lg font-bold border-2 transition-all flex items-center justify-center gap-2 ${history[currentIndex] === true
                      ? 'border-emerald-500 bg-emerald-50 text-emerald-700 ring-2 ring-emerald-200 ring-offset-2'
                      : 'border-slate-200 text-slate-500 hover:border-emerald-300 hover:text-emerald-600'
                      }`}
                  >
                    <CheckCircle2 size={18} /> 我答对了
                  </button>
                  <button
                    onClick={() => handleSelfCheck(false)}
                    className={`flex-1 py-3 rounded-lg font-bold border-2 transition-all flex items-center justify-center gap-2 ${history[currentIndex] === false
                      ? 'border-red-500 bg-red-50 text-red-700 ring-2 ring-red-200 ring-offset-2'
                      : 'border-slate-200 text-slate-500 hover:border-red-300 hover:text-red-600'
                      }`}
                  >
                    <XCircle size={18} /> 我答错了
                  </button>
                </div>
              </div>
            )}
          </motion.div>
        )}

        {/* Comment Section */}
        {((mode === 'study' || (isAnswered && mode !== 'test'))) && (
          <CommentSection questionId={currentQ.id} />
        )}
      </div>



      {/* Grid Drawer */}
      <AnimatePresence>
        {showGrid && (
          <>
            <motion.div
              initial={{ opacity: 0 }} animate={{ opacity: 1 }} exit={{ opacity: 0 }}
              className="fixed inset-0 bg-slate-900/20 backdrop-blur-sm z-40"
              onClick={() => setShowGrid(false)}
            />
            <motion.div
              initial={{ y: '100%' }} animate={{ y: 0 }} exit={{ y: '100%' }}
              transition={{ type: "spring", damping: 25, stiffness: 200 }}
              className="fixed bottom-0 left-0 right-0 bg-white rounded-t-3xl z-50 p-8 max-h-[70vh] overflow-y-auto shadow-2xl"
            >
              <div className="flex justify-between items-center mb-6">
                <h3 className="text-lg font-bold text-slate-900">题目列表</h3>
                <button onClick={() => setShowGrid(false)}><XCircle className="text-slate-300 hover:text-slate-500" /></button>
              </div>

              <div className="flex gap-4 mb-6 text-xs text-slate-500 font-medium">
                <div className="flex items-center gap-1"><span className="w-3 h-3 rounded-full bg-emerald-500"></span> 正确</div>
                <div className="flex items-center gap-1"><span className="w-3 h-3 rounded-full bg-red-500"></span> 错误</div>
                <div className="flex items-center gap-1"><span className="w-3 h-3 rounded-full bg-slate-100 border border-slate-300"></span> 未做</div>
                <div className="flex items-center gap-1"><span className="w-3 h-3 rounded-full bg-blue-600"></span> 当前</div>
              </div>

              <div className="grid grid-cols-5 md:grid-cols-8 gap-3">
                {questions.map((_, idx) => {
                  const status = history[idx]; // true, false, or undefined
                  let bg = "bg-slate-50 text-slate-500 border border-slate-200";
                  if (idx === currentIndex) bg = "bg-blue-600 text-white border-blue-600 ring-2 ring-offset-2 ring-blue-200";
                  else if (status === true) bg = "bg-emerald-100 text-emerald-700 border-emerald-200";
                  else if (status === false) bg = "bg-red-100 text-red-700 border-red-200";

                  return (
                    <button
                      key={idx}
                      onClick={() => jumpToQuestion(idx)}
                      className={`h-12 rounded-xl font-bold text-sm flex items-center justify-center transition-all ${bg}`}
                    >
                      {idx + 1}
                    </button>
                  )
                })}
              </div>
            </motion.div>
          </>
        )}
      </AnimatePresence>

    </div>
  );
};

const CommentSection: React.FC<{ questionId: number }> = ({ questionId }) => {
  const { user } = useAppContext();
  const [comments, setComments] = useState<Comment[]>([]);
  const [newComment, setNewComment] = useState('');
  const [loading, setLoading] = useState(false);
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    loadComments();
  }, [questionId]);

  const loadComments = async () => {
    setLoading(true);
    try {
      const data = await api.quiz.getComments(questionId);
      setComments(data);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const handleSubmit = async () => {
    if (!newComment.trim()) return;
    setSubmitting(true);
    try {
      await api.quiz.addComment(questionId, newComment);
      setNewComment('');
      loadComments();
    } catch (err) {
      alert('评论发表失败');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="mt-8 border-t border-slate-100 pt-8">
      <h3 className="font-bold text-slate-900 mb-6 flex items-center gap-2">
        <MessageSquare size={20} className="text-blue-500" />
        讨论区 ({comments.length})
      </h3>

      {/* Input */}
      <div className="flex gap-4 mb-8">
        <div className="w-10 h-10 rounded-full bg-slate-100 flex items-center justify-center text-slate-400 font-bold shrink-0 overflow-hidden">
          {user.avatar ? <img src={user.avatar} className="w-full h-full object-cover" /> : user.name[0]}
        </div>
        <div className="flex-1">
          <textarea
            value={newComment}
            onChange={e => setNewComment(e.target.value)}
            placeholder="写下你的想法..."
            className="w-full p-4 rounded-xl border border-slate-200 focus:outline-none focus:border-blue-500 bg-slate-50 min-h-[100px] resize-none text-sm"
          />
          <div className="flex justify-end mt-2">
            <button
              onClick={handleSubmit}
              disabled={submitting || !newComment.trim()}
              className="bg-blue-600 text-white px-6 py-2 rounded-lg text-sm font-bold hover:bg-blue-700 transition-colors disabled:opacity-50 flex items-center gap-2"
            >
              <Send size={16} /> 发表评论
            </button>
          </div>
        </div>
      </div>

      {/* List */}
      <div className="space-y-6">
        {loading ? (
          <div className="text-center py-8 text-slate-400">加载中...</div>
        ) : comments.length === 0 ? (
          <div className="text-center py-8 bg-slate-50 rounded-xl text-slate-400 text-sm">
            暂无评论，快来抢沙发吧~
          </div>
        ) : (
          comments.map((comment) => (
            <div key={comment.id} className="flex gap-4">
              <div className="w-10 h-10 rounded-full bg-slate-100 flex items-center justify-center text-slate-500 font-bold shrink-0 overflow-hidden border border-slate-100">
                {comment.user.avatar ? <img src={comment.user.avatar} className="w-full h-full object-cover" /> : comment.user.name[0]}
              </div>
              <div className="flex-1">
                <div className="flex items-center gap-2 mb-1">
                  <span className="font-bold text-slate-800 text-sm">{comment.user.name}</span>
                  <span className="text-xs text-slate-400">{new Date(comment.createdAt).toLocaleString()}</span>
                </div>
                <p className="text-slate-600 text-sm leading-relaxed bg-slate-50 p-3 rounded-lg rounded-tl-none inline-block">
                  {comment.content}
                </p>
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  );
};

export default QuizView;
