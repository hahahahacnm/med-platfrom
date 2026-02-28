import React, { useState, useEffect, useMemo } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import {
    Star,
    MessageSquare,
    AlertTriangle,
    ChevronDown,
    ChevronUp,
    Tag,
    Share2,
    Trash2,
    Flame
} from 'lucide-react';
import { cn } from '../lib/utils';
import api from '../lib/api';
import { QuestionItem } from './QuestionItem';
import { QuestionComments } from './QuestionComments';

interface QuestionCardProps {
    question: any;
    serialNumber: number;
    onAnswerResult?: (payload: { id: number, isCorrect: boolean | null }) => void;
    onRemoveMistake?: () => void;
    showRemoveBtn?: boolean;
}

export function QuestionCard({
    question,
    serialNumber,
    onAnswerResult,
    onRemoveMistake,
    showRemoveBtn
}: QuestionCardProps) {
    const [isFavorited, setIsFavorited] = useState(false);
    const [favLoading, setFavLoading] = useState(false);
    const [showNotes, setShowNotes] = useState(false);

    useEffect(() => {
        setIsFavorited(!!(question?.is_favorite || question?.IsFavorite));
    }, [question]);

    const toggleFavorite = async (e: React.MouseEvent) => {
        e.stopPropagation();
        if (favLoading) return;
        setFavLoading(true);
        try {
            const res: any = await api.post(`/favorites/${question.id}`);
            if (res.data) {
                setIsFavorited(res.data.is_favorite);
            }
        } catch (e) {
            console.error(e);
        } finally {
            setFavLoading(false);
        }
    };

    const hasChild = useMemo(() => question.children?.length > 0, [question.children]);

    const opts = useMemo(() => {
        if (!hasChild || !question.options) return null;
        return typeof question.options === 'string' ? JSON.parse(question.options) : question.options;
    }, [hasChild, question.options]);

    const parsedB1Opts = useMemo(() => {
        if (!question.type?.includes('B1') || !opts) return [];
        return Object.keys(opts).sort().map(k => ({
            key: k,
            value: opts[k]
        }));
    }, [question.type, opts]);

    const parsedStem = useMemo(() => {
        let txt = question.stem ? question.stem.replace(/【(共用主干|共用题干|案例描述)】/g, '').trim() : '';
        return txt;
    }, [question.stem]);

    return (
        <div className="bg-card rounded-3xl border border-border/50 shadow-sm overflow-hidden mb-6 hover:shadow-md transition-shadow">
            {/* Mistake Stats Banner */}
            {question.wrong_count > 0 && (
                <div className={cn(
                    "mx-8 mt-8 px-6 py-4 rounded-3xl flex items-center justify-between border-2 border-dashed animate-in fade-in slide-in-from-top-4 duration-500",
                    question.wrong_count >= 3 ? "bg-rose-500/10 border-rose-500/30 text-rose-600" : "bg-amber-500/10 border-amber-500/30 text-amber-600"
                )}>
                    <div className="flex items-center gap-4">
                        <div className={cn(
                            "w-10 h-10 rounded-2xl flex items-center justify-center shadow-lg shadow-current/10",
                            question.wrong_count >= 3 ? "bg-rose-500 text-white" : "bg-amber-500 text-white"
                        )}>
                            <Flame size={20} className="fill-current" />
                        </div>
                        <div className="flex flex-col">
                            <span className="text-[10px] font-black uppercase tracking-[0.2em] leading-none mb-1 opacity-70">深度回顾</span>
                            <span className="text-sm font-bold">该题您已累计答错 <span className="text-lg font-black font-mono mx-1">{question.wrong_count}</span> 次</span>
                        </div>
                    </div>
                    <div className="hidden sm:flex flex-col items-end">
                        <span className="text-[10px] font-black uppercase tracking-widest opacity-60">建议强度</span>
                        <div className="flex items-center gap-1">
                            {Array(Math.min(question.wrong_count, 5)).fill(0).map((_, i) => (
                                <div key={i} className={cn("w-1 h-3 rounded-full", question.wrong_count >= 3 ? "bg-rose-500" : "bg-amber-500")} />
                            ))}
                        </div>
                    </div>
                </div>
            )}

            <div className="p-8">
                {hasChild ? (
                    <div className="space-y-8">
                        <div className="flex items-center gap-3 mb-6">
                            <span className="px-3 py-1 bg-indigo-500/10 text-indigo-600 text-xs font-black rounded-full uppercase tracking-widest">
                                {question.type || '组合题'}
                            </span>
                            <span className="text-sm font-bold text-muted-foreground">(共 {question.children.length} 小题)</span>
                        </div>

                        {question.stem && !question.type?.includes('B1') && (
                            <div className="p-6 bg-muted/30 rounded-2xl border border-border/50 relative overflow-hidden group/stem">
                                <div className="absolute top-0 left-0 w-1.5 h-full bg-indigo-500/50" />
                                <div className="text-xs font-black text-indigo-500/60 uppercase tracking-widest mb-3 flex items-center gap-2">
                                    <Tag className="w-3 h-3" /> 主题干描述
                                </div>
                                <div className="text-lg font-bold text-foreground leading-relaxed" dangerouslySetInnerHTML={{ __html: parsedStem }} />
                            </div>
                        )}

                        {question.type?.includes('B1') && parsedB1Opts.length > 0 && (
                            <div className="p-6 bg-emerald-500/5 rounded-2xl border border-emerald-500/10 relative overflow-hidden">
                                <div className="absolute top-0 left-0 w-1.5 h-full bg-emerald-500/50" />
                                <div className="text-xs font-black text-emerald-600/60 uppercase tracking-widest mb-4 flex items-center gap-2">
                                    <Tag className="w-3 h-3" /> 共用备选答案
                                </div>
                                <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                                    {parsedB1Opts.map(opt => (
                                        <div key={opt.key} className="flex gap-3 text-sm font-medium text-foreground p-3 rounded-xl border border-border bg-background shadow-sm hover:border-emerald-500/50 transition-colors">
                                            <div className="w-6 h-6 rounded-full bg-emerald-500 text-white flex items-center justify-center font-bold text-xs shrink-0">{opt.key}</div>
                                            <div dangerouslySetInnerHTML={{ __html: opt.value }} />
                                        </div>
                                    ))}
                                </div>
                            </div>
                        )}

                        <div className="space-y-12">
                            {question.children.map((child: any, i: number) => (
                                <div key={child.id} className="relative">
                                    <div className="flex items-center gap-3 mb-4">
                                        <div className="px-4 py-1.5 bg-primary text-white text-[11px] font-black rounded-full shadow-lg shadow-primary/20">
                                            小题 {i + 1}
                                        </div>
                                    </div>
                                    <QuestionItem
                                        question={child}
                                        sharedOptions={opts}
                                        index={child.displayIndex}
                                        isChild
                                        onAnswerResult={onAnswerResult}
                                    />
                                    {i < question.children.length - 1 && (
                                        <div className="h-[1px] w-full bg-gradient-to-r from-transparent via-border to-transparent mt-12" />
                                    )}
                                </div>
                            ))}
                        </div>
                    </div>
                ) : (
                    <QuestionItem
                        question={question}
                        index={serialNumber}
                        showTypeTag
                        onAnswerResult={onAnswerResult}
                    />
                )}
            </div>

            {/* Card Footer Actions */}
            <div className="px-8 py-5 bg-muted/20 border-t border-border/50 flex flex-wrap items-center justify-between gap-4">
                <div className="flex items-center gap-2 sm:gap-4">
                    <button
                        onClick={toggleFavorite}
                        className={cn(
                            "flex items-center gap-2 px-4 py-2 rounded-2xl text-sm font-bold transition-all border shadow-sm",
                            isFavorited
                                ? "bg-amber-500 text-white border-amber-600 ring-2 ring-amber-500/20"
                                : "bg-card text-muted-foreground border-border hover:bg-muted/50"
                        )}
                    >
                        <Star className={cn("w-4 h-4", isFavorited && "fill-current")} />
                        {isFavorited ? "已经收藏" : "加入收藏"}
                    </button>

                    <button
                        onClick={() => setShowNotes(!showNotes)}
                        className={cn(
                            "flex items-center gap-2 px-4 py-2 rounded-2xl text-sm font-bold transition-all border shadow-sm bg-card",
                            showNotes ? "text-primary border-primary ring-2 ring-primary/20" : "text-muted-foreground border-border hover:bg-muted/50"
                        )}
                    >
                        <MessageSquare className="w-4 h-4" />
                        讨论区 <span className="bg-muted px-1.5 py-0.5 rounded-lg text-[10px] font-black ml-1 uppercase">{question.note_count || 0}</span>
                    </button>
                </div>

                <div className="flex items-center gap-2">
                    <button className="p-2.5 text-muted-foreground hover:bg-accent rounded-xl border border-transparent hover:border-border transition-all hover:text-foreground shadow-sm">
                        <AlertTriangle className="w-4 h-4" />
                    </button>
                    <button className="p-2.5 text-muted-foreground hover:bg-accent rounded-xl border border-transparent hover:border-border transition-all hover:text-foreground shadow-sm">
                        <Share2 className="w-4 h-4" />
                    </button>
                    {showRemoveBtn && (
                        <button
                            onClick={onRemoveMistake}
                            className="flex items-center gap-2 px-4 py-2 bg-rose-500/10 text-rose-600 rounded-2xl text-sm font-bold hover:bg-rose-500 hover:text-white transition-all border border-rose-500/20"
                        >
                            <Trash2 className="w-4 h-4" /> 移出错题本
                        </button>
                    )}
                </div>
            </div>

            {/* Discussion / Notes Section */}
            <AnimatePresence>
                {showNotes && (
                    <motion.div
                        initial={{ height: 0, opacity: 0 }}
                        animate={{ height: "auto", opacity: 1 }}
                        exit={{ height: 0, opacity: 0 }}
                        className="bg-muted/10 border-t border-border/50 overflow-hidden"
                    >
                        <div className="h-[600px]">
                            <QuestionComments
                                questionId={question.id}
                                onClose={() => setShowNotes(false)}
                            />
                        </div>
                    </motion.div>
                )}
            </AnimatePresence>
        </div>
    );
};
