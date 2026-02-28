import React, { useState, useEffect, useRef } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import {
    ChevronLeft,
    ChevronRight,
    LayoutGrid,
    ArrowLeft,
    X,
    Target,
    Trophy,
    History,
    AlertCircle,
    Flame,
    RefreshCw,
    CheckCircle
} from 'lucide-react';
import { cn } from '../lib/utils';
import api from '../lib/api';
import { QuestionCard } from './QuestionCard';

interface QuestionSessionProps {
    category?: string;
    source?: string;
    endpoint?: string;
    onExit: () => void;
}

export function QuestionSession({ category, source, endpoint = '/questions/skeleton', onExit }: QuestionSessionProps) {
    const [loading, setLoading] = useState(true);
    const [skeletonList, setSkeletonList] = useState<any[]>([]);
    const [currentIndex, setCurrentIndex] = useState(0);
    const [currentDetail, setCurrentDetail] = useState<any>(null);
    const [loadingDetail, setLoadingDetail] = useState(false);
    const [summary, setSummary] = useState<any>(null);
    const [showSheetModal, setShowSheetModal] = useState(false);
    const [isRightSheetOpen, setIsRightSheetOpen] = useState(false);
    const [autoRemove, setAutoRemove] = useState(false);
    const scrollRef = useRef<HTMLDivElement>(null);
    const isMistakeMode = endpoint.includes('mistakes');

    useEffect(() => {
        if (typeof window !== 'undefined') {
            setAutoRemove(localStorage.getItem('mistake_auto_remove') === 'true');
        }

        const fetchSkeleton = async () => {
            setLoading(true);
            try {
                const res: any = await api.get(endpoint, { params: { category, source } });
                if (res.data) {
                    setSkeletonList(res.data.map((q: any, idx: number) => ({
                        ...q,
                        displayIndex: idx + 1,
                        status: q.status || 'unfilled'
                    })));
                    setSummary(res.summary);
                }
            } catch (err) {
                console.error('Failed to fetch skeleton', err);
            } finally {
                setLoading(false);
            }
        };

        if (category && source) {
            fetchSkeleton();
        }
    }, [category, source, endpoint]);

    useEffect(() => {
        const fetchDetail = async () => {
            if (skeletonList.length === 0 || currentIndex < 0 || currentIndex >= skeletonList.length) return;
            const target = skeletonList[currentIndex];
            setLoadingDetail(true);
            try {
                const res: any = await api.get(`/questions/${target.id}`);
                if (res.data) {
                    setCurrentDetail({
                        ...res.data,
                        displayIndex: target.displayIndex,
                        _wrongCount: target.wrong_count || 1
                    });
                }
                scrollRef.current?.scrollTo({ top: 0, behavior: 'smooth' });
            } catch (err) {
                console.error('Failed to load question detail', err);
            } finally {
                setLoadingDetail(false);
            }
        };
        fetchDetail();
    }, [currentIndex, skeletonList]);

    const onAnswerResult = async (payload: { id: number, isCorrect: boolean | null }) => {
        setSkeletonList(prev => prev.map(q => {
            if (q.id === payload.id) {
                return { ...q, status: payload.isCorrect === true ? 'correct' : (payload.isCorrect === false ? 'wrong' : 'unfilled') };
            }
            return q;
        }));

        if (payload.isCorrect === true && autoRemove && isMistakeMode) {
            setTimeout(() => {
                handleRemoveMistake(payload.id, true);
            }, 800);
        }

        if (!payload.isCorrect && isMistakeMode) {
            try {
                const res: any = await api.get('/mistakes/skeleton', { params: { category, source } });
                const latest = res.data?.find((q: any) => q.id === payload.id);
                if (latest) {
                    setSkeletonList(prev => prev.map(q => q.id === payload.id ? { ...q, wrong_count: latest.wrong_count } : q));
                    if (currentDetail?.id === payload.id) {
                        setCurrentDetail((prev: any) => ({ ...prev, _wrongCount: latest.wrong_count }));
                    }
                }
            } catch (err) { }
        }
    };

    const handleRemoveMistake = async (id: number, silent = false) => {
        try {
            await api.delete(`/mistakes/${id}`);
            setSkeletonList(prev => {
                const newList = prev.filter(q => q.id !== id);
                if (newList.length > 0) {
                    if (currentIndex >= newList.length) setCurrentIndex(newList.length - 1);
                } else {
                    setCurrentDetail(null);
                }
                return newList;
            });
            if (!silent) alert('🎉 已移出错题本');
        } catch (err) {
            if (!silent) alert('移除失败');
        }
    };

    const nextQuestion = () => {
        if (currentIndex < skeletonList.length - 1) {
            setCurrentIndex(currentIndex + 1);
        }
    };

    const prevQuestion = () => {
        if (currentIndex > 0) {
            setCurrentIndex(currentIndex - 1);
        }
    };

    const jumpTo = (index: number) => {
        setCurrentIndex(index);
        setShowSheetModal(false);
    };

    if (loading) {
        return (
            <div className="flex flex-col items-center justify-center h-full w-full gap-6 animate-pulse">
                <div className="w-16 h-16 border-4 border-primary/20 border-t-primary rounded-full animate-spin" />
                <p className="text-sm font-black text-muted-foreground uppercase tracking-widest italic">正在同步题库资源...</p>
            </div>
        );
    }

    if (skeletonList.length === 0) {
        return (
            <div className="flex flex-col items-center justify-center h-full w-full p-10 bg-card">
                <div className="w-20 h-20 bg-muted/50 rounded-full flex items-center justify-center mb-6">
                    <AlertCircle className="w-10 h-10 text-muted-foreground/30" />
                </div>
                <h3 className="text-xl font-bold mb-2">章节内容为空</h3>
                <p className="text-muted-foreground mb-8 text-sm">该章节暂时还没有上传任何题目，请尝试其他章节。</p>
                <button onClick={onExit} className="px-6 py-2.5 bg-primary text-primary-foreground font-bold rounded-xl shadow-md flex items-center gap-2 transition-all hover:scale-105 active:scale-95">
                    <ArrowLeft className="w-4 h-4" /> 返回题库列表
                </button>
            </div>
        );
    }

    const SheetContent = () => (
        <div className="flex flex-col h-full bg-card">
            <div className="p-4 md:p-6 border-b border-border/50 flex items-center justify-between shrink-0 bg-muted/10">
                <div className="flex items-center gap-3">
                    <div className="w-10 h-10 bg-primary/10 rounded-xl flex items-center justify-center text-primary shadow-sm">
                        <LayoutGrid className="w-5 h-5" />
                    </div>
                    <div>
                        <h2 className="text-base font-black text-foreground uppercase tracking-tight">章节答题卡</h2>
                        <p className="text-[10px] font-bold text-muted-foreground mt-0.5 tracking-widest">点击序号或滑动跳题</p>
                    </div>
                </div>
            </div>

            <div className="flex-1 overflow-y-auto p-4 md:p-6 custom-scrollbar">
                <div className="flex flex-wrap gap-2 md:gap-3">
                    {skeletonList.map((q, idx) => (
                        <button
                            key={q.id}
                            onClick={() => jumpTo(idx)}
                            className={cn(
                                "relative w-10 h-10 md:w-11 md:h-11 rounded-xl text-sm font-black font-mono transition-all flex items-center justify-center border-2 border-transparent shadow-sm hover:scale-105 active:scale-95",
                                currentIndex === idx && "border-primary text-primary scale-110 shadow-lg shadow-primary/20 bg-primary/5 ring-2 ring-primary/20 z-10",
                                q.status === 'correct' && currentIndex !== idx && "bg-emerald-500/10 text-emerald-600 border-emerald-500/20",
                                q.status === 'wrong' && currentIndex !== idx && "bg-rose-500/10 text-rose-600 border-rose-500/20",
                                q.status === 'unfilled' && currentIndex !== idx && "bg-muted/50 text-muted-foreground hover:bg-muted"
                            )}
                        >
                            {idx + 1}
                            {isMistakeMode && q.wrong_count > 1 && (
                                <span className="absolute -top-1 -right-1 w-4 h-4 bg-rose-500 text-white text-[8px] leading-none rounded-full flex items-center justify-center border-2 border-card shadow-sm z-20 font-sans">
                                    {q.wrong_count}
                                </span>
                            )}
                        </button>
                    ))}
                </div>
            </div>

            <div className="p-4 md:p-6 bg-muted/20 border-t border-border/50 shrink-0">
                <div className="flex justify-between items-center mb-2">
                    <div className="flex items-center gap-1.5">
                        <div className="w-2.5 h-2.5 rounded-full bg-emerald-500 shadow-sm" />
                        <span className="text-[10px] font-bold text-muted-foreground">正确</span>
                    </div>
                    <div className="flex items-center gap-1.5">
                        <div className="w-2.5 h-2.5 rounded-full bg-rose-500 shadow-sm" />
                        <span className="text-[10px] font-bold text-muted-foreground">错误</span>
                    </div>
                    <div className="flex items-center gap-1.5">
                        <div className="w-2.5 h-2.5 rounded-full bg-muted border border-border/50 shadow-sm" />
                        <span className="text-[10px] font-bold text-muted-foreground">未答</span>
                    </div>
                </div>
                <div className="flex items-center justify-between">
                    <span className="text-[10px] font-black uppercase text-primary tracking-widest flex items-center gap-1 px-2 py-1 bg-primary/10 rounded-lg">
                        <Target className="w-3 h-3" /> 进度 {Math.round((summary?.attempted_num || 0) / skeletonList.length * 100)}%
                    </span>
                    <span className="text-[10px] font-black uppercase text-emerald-500 tracking-widest flex items-center gap-1 px-2 py-1 bg-emerald-500/10 rounded-lg">
                        <Trophy className="w-3 h-3" /> 正确率 {summary?.accuracy_rate || 0}%
                    </span>
                </div>
            </div>
        </div>
    );

    return (
        <div className="flex flex-col xl:flex-row h-full w-full bg-background overflow-hidden relative">

            {/* Main Content Pane */}
            <div className="flex-1 flex flex-col h-full overflow-hidden relative">

                {/* Embedded Header */}
                <div className="shrink-0 sticky top-0 z-30 bg-card/80 backdrop-blur-xl border-b border-border/50 px-4 md:px-8 py-3 flex items-center justify-between shadow-sm">
                    <div className="flex items-center gap-2 md:gap-4 overflow-hidden">
                        <button onClick={onExit} className="p-2 bg-muted/50 hover:bg-muted rounded-xl transition-colors text-muted-foreground hover:text-foreground shrink-0 shadow-sm">
                            <ArrowLeft className="w-4 h-4 md:w-5 md:h-5" />
                        </button>
                        <div className="flex flex-col min-w-0">
                            <div className="text-[9px] md:text-[10px] font-black text-primary uppercase tracking-widest leading-none mb-1 truncate">
                                {source} · {isMistakeMode ? '错题专练' : '章节练习'}
                            </div>
                            <h1 className="text-sm md:text-base font-bold text-foreground tracking-tight truncate">{category}</h1>
                        </div>
                    </div>

                    <div className="flex items-center gap-3 shrink-0">
                        <button
                            onClick={async () => {
                                if (confirm('确定要清空本章的答题记录重新开始吗？')) {
                                    try {
                                        await api.post('/questions/reset-chapter', { category, source });
                                        window.location.reload();
                                    } catch (e) {
                                        console.error(e);
                                    }
                                }
                            }}
                            className="hidden md:flex items-center gap-1.5 px-3 py-1.5 bg-rose-500/10 hover:bg-rose-500 text-rose-600 hover:text-white border border-rose-500/20 rounded-xl text-[10px] font-black transition-all"
                        >
                            <History className="w-3.5 h-3.5" /> 清空记录
                        </button>

                        {/* Screen-dependent trigger for Sheet */}
                        <button
                            onClick={() => {
                                if (window.innerWidth >= 1280) {
                                    setIsRightSheetOpen(!isRightSheetOpen);
                                } else {
                                    setShowSheetModal(true);
                                }
                            }}
                            className={cn(
                                "flex items-center gap-2 px-3 py-1.5 rounded-xl text-[10px] font-black transition-all border",
                                isRightSheetOpen
                                    ? "bg-primary text-white border-primary shadow-md"
                                    : "bg-primary/10 hover:bg-primary text-primary hover:text-white border-primary/20"
                            )}
                        >
                            <LayoutGrid className="w-3.5 h-3.5" /> 答题卡
                        </button>
                    </div>
                </div>

                {/* Auto Remove Toolbar (Mistake Mode) */}
                {isMistakeMode && (
                    <div className="shrink-0 bg-rose-500/5 border-b border-rose-500/10 px-4 md:px-8 py-2 flex justify-between items-center text-[10px] font-black uppercase tracking-widest">
                        <div className="flex items-center gap-1.5 text-rose-600/70">
                            <Flame className="w-3.5 h-3.5 animate-pulse" />
                            重拳出击，消灭错题
                        </div>
                        <button
                            onClick={() => {
                                const val = !autoRemove;
                                setAutoRemove(val);
                                if (typeof window !== 'undefined') localStorage.setItem('mistake_auto_remove', String(val));
                            }}
                            className={cn(
                                "flex items-center gap-1.5 py-1 px-3 rounded-lg border transition-all shadow-sm",
                                autoRemove ? "bg-rose-500 text-white border-rose-600" : "bg-card border-border text-muted-foreground hover:bg-muted"
                            )}
                        >
                            <RefreshCw className={cn("w-3 h-3", autoRemove && "animate-spin-slow")} />
                            答对自动移出
                        </button>
                    </div>
                )}

                {/* Question Area */}
                <div ref={scrollRef} className="flex-1 overflow-y-auto px-4 md:px-8 py-6 custom-scrollbar bg-card/30">
                    <div className="max-w-4xl mx-auto h-full">
                        <AnimatePresence mode="wait">
                            {currentDetail ? (
                                <motion.div
                                    key={currentDetail.id}
                                    initial={{ opacity: 0, scale: 0.98, y: 10 }}
                                    animate={{ opacity: 1, scale: 1, y: 0 }}
                                    exit={{ opacity: 0, scale: 0.98, y: -10 }}
                                    transition={{ duration: 0.2, ease: "easeOut" }}
                                    className={cn("transition-opacity duration-300", loadingDetail && "opacity-50 pointer-events-none")}
                                >
                                    <QuestionCard
                                        question={currentDetail}
                                        serialNumber={currentDetail.displayIndex}
                                        onAnswerResult={onAnswerResult}
                                        showRemoveBtn={isMistakeMode}
                                        onRemoveMistake={() => handleRemoveMistake(currentDetail.id)}
                                    />
                                </motion.div>
                            ) : (
                                <div className="py-20 flex justify-center">
                                    <div className="w-8 h-8 border-3 border-primary/20 border-t-primary rounded-full animate-spin" />
                                </div>
                            )}
                        </AnimatePresence>
                    </div>
                </div>

                {/* Action Bar (Replaces floating bar) */}
                <div className="shrink-0 sticky bottom-0 z-20 bg-card border-t border-border/50 px-4 md:px-8 py-3 flex items-center justify-between shadow-[0_-10px_40px_-10px_rgba(0,0,0,0.05)]">
                    <button
                        disabled={currentIndex === 0}
                        onClick={prevQuestion}
                        className="flex items-center gap-2 px-4 py-2.5 bg-muted/50 hover:bg-muted text-foreground font-bold rounded-xl transition-all disabled:opacity-30 disabled:scale-100 hover:scale-105 active:scale-95 text-xs tracking-widest"
                    >
                        <ChevronLeft className="w-4 h-4" /> <span className="hidden sm:inline">上一题</span>
                    </button>

                    <div className="flex flex-col items-center">
                        <div className="flex items-baseline gap-1">
                            <span className="text-xl font-black text-primary font-mono leading-none tracking-tighter">{currentIndex + 1}</span>
                            <span className="text-sm text-muted-foreground font-bold tracking-tighter">/ {skeletonList.length}</span>
                        </div>
                    </div>

                    <button
                        disabled={currentIndex === skeletonList.length - 1}
                        onClick={nextQuestion}
                        className="flex items-center gap-2 px-4 py-2.5 bg-primary hover:bg-primary/90 text-primary-foreground font-bold rounded-xl shadow-lg shadow-primary/20 transition-all disabled:opacity-30 disabled:scale-100 disabled:shadow-none hover:scale-105 active:scale-95 text-xs tracking-widest"
                    >
                        <span className="hidden sm:inline">下一题</span> <ChevronRight className="w-4 h-4" />
                    </button>
                </div>

            </div>

            {/* Right Pane (Desktop Answer Sheet) */}
            <div className={cn(
                "hidden xl:flex shrink-0 transition-all duration-300 ease-in-out z-20 relative border-l border-border/50",
                isRightSheetOpen ? "w-[320px] 2xl:w-[360px] opacity-100 shadow-[-10px_0_40px_-20px_rgba(0,0,0,0.05)]" : "w-0 opacity-0 overflow-hidden border-transparent"
            )}>
                <div className="w-[320px] 2xl:w-[360px] h-full flex flex-col shrink-0">
                    <SheetContent />
                </div>
            </div>

            {/* Mobile/Tablet Modal Answer Sheet */}
            <AnimatePresence>
                {showSheetModal && (
                    <div className="xl:hidden fixed inset-0 z-[100] flex justify-end">
                        <motion.div
                            initial={{ opacity: 0 }}
                            animate={{ opacity: 1 }}
                            exit={{ opacity: 0 }}
                            onClick={() => setShowSheetModal(false)}
                            className="absolute inset-0 bg-black/40 backdrop-blur-sm"
                        />
                        <motion.div
                            initial={{ x: "100%" }}
                            animate={{ x: 0 }}
                            exit={{ x: "100%" }}
                            transition={{ type: "spring", damping: 25, stiffness: 200 }}
                            className="relative w-[320px] max-w-[85vw] h-full shadow-2xl z-10"
                        >
                            <SheetContent />
                            <button
                                onClick={() => setShowSheetModal(false)}
                                className="absolute top-4 right-4 w-8 h-8 flex items-center justify-center hover:bg-border/50 rounded-full transition-all text-muted-foreground"
                            >
                                <X className="w-5 h-5" />
                            </button>
                        </motion.div>
                    </div>
                )}
            </AnimatePresence>

        </div>
    );
}
