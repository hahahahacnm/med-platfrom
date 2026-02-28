import React, { useState, useEffect, useMemo } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { RefreshCw, CheckCircle2, XCircle, Info, ChevronDown, ChevronUp } from 'lucide-react';
import { cn } from '../lib/utils';
import api from '../lib/api';

interface QuestionItemProps {
    question: any;
    sharedOptions?: any;
    index?: number;
    isChild?: boolean;
    showSharedHeader?: boolean;
    showTypeTag?: boolean;
    onAnswerResult?: (payload: { id: number, isCorrect: boolean | null }) => void;
}

export function QuestionItem({
    question,
    sharedOptions,
    index,
    isChild,
    showTypeTag,
    onAnswerResult
}: QuestionItemProps) {
    const [selectedOption, setSelectedOption] = useState('');
    const [multiSelection, setMultiSelection] = useState<string[]>([]);
    const [result, setResult] = useState<any>(null);
    const [submitting, setSubmitting] = useState(false);
    const [showAnswer, setShowAnswer] = useState(false);

    const isMultiChoice = useMemo(() => {
        const t = (question.type || '').toUpperCase();
        return t.includes('X') || t.includes('多选');
    }, [question.type]);

    const isSubjective = useMemo(() => {
        const t = (question.type || '').trim();
        const explicitTypes = ['简答', '论述', '名词解释', '案例分析', '问答'];
        if (explicitTypes.some(type => t.includes(type))) return true;
        const hasSelfOpts = question.options && Object.keys(question.options).length > 0;
        const hasSharedOpts = sharedOptions && Object.keys(sharedOptions).length > 0;
        return !hasSelfOpts && !hasSharedOpts;
    }, [question, sharedOptions]);

    useEffect(() => {
        if (question.user_record) {
            const record = question.user_record;
            if (isMultiChoice && record.choice) {
                setMultiSelection(record.choice.split(''));
                setSelectedOption(record.choice);
            } else {
                setSelectedOption(record.choice || '');
            }

            if (typeof record.is_correct === 'boolean') {
                setResult({
                    is_correct: record.is_correct,
                    correct_answer: question.correct,
                    analysis: question.analysis
                });
            }
        } else {
            setSelectedOption('');
            setMultiSelection([]);
            setResult(null);
        }
    }, [question, isMultiChoice]);

    const formatText = (text: string) => {
        if (!text) return '';
        let res = text.replace(/!\[(.*?)\]\((.*?)\)/g, (_, __, url) => {
            const fullUrl = url.startsWith('http') ? url : `http://localhost:8080${url}`;
            return `<img src="${fullUrl}" class="max-w-[150px] max-h-[150px] rounded-lg border border-border my-2 cursor-zoom-in inline-block align-bottom" />`;
        });
        res = res.replace(/\[图片:(.*?)\]/g, (_, url) => {
            const fullUrl = url.startsWith('http') ? url : `http://localhost:8080${url}`;
            return `<img src="${fullUrl}" class="max-w-[150px] max-h-[150px] rounded-lg border border-border my-2 cursor-zoom-in inline-block align-bottom" />`;
        });
        return res;
    };

    const parsedStem = useMemo(() => formatText(question.stem), [question.stem]);
    const parsedAnalysis = useMemo(() => formatText(result?.analysis || question.analysis || '暂无解析'), [result, question.analysis]);
    const parsedCorrect = useMemo(() => formatText(question.correct || '略'), [question.correct]);

    const displayOptions = useMemo(() => {
        const opts = question.options || sharedOptions;
        if (!opts) return [];
        return Object.keys(opts).sort().map(key => ({
            key,
            value: opts[key],
            parsedValue: formatText(opts[key])
        }));
    }, [question.options, sharedOptions]);

    const handleRedo = async (e: React.MouseEvent) => {
        e.stopPropagation();
        try {
            await api.delete(`/questions/${question.id}/reset`);
            setSelectedOption('');
            setMultiSelection([]);
            setResult(null);
            setShowAnswer(false);
            onAnswerResult?.({ id: question.id, isCorrect: null });
        } catch (e) {
            console.error(e);
        }
    };

    const submitAnswer = async (answerStr: string) => {
        setSelectedOption(answerStr);
        setSubmitting(true);
        try {
            const res: any = await api.post(`/questions/${question.id}/submit`, { choice: answerStr });
            if (res.data) {
                setResult(res.data);
                onAnswerResult?.({ id: question.id, isCorrect: res.data.is_correct });
            }
        } catch (e) {
            console.error(e);
        } finally {
            setSubmitting(false);
        }
    };

    const handleOptionClick = (key: string) => {
        if (result) return;
        if (isMultiChoice) {
            setMultiSelection(prev => {
                const next = prev.includes(key) ? prev.filter(k => k !== key) : [...prev, key];
                return next;
            });
        } else {
            submitAnswer(key);
        }
    };

    const submitMultiChoice = (e: React.MouseEvent) => {
        e.stopPropagation();
        if (multiSelection.length === 0) return;
        submitAnswer([...multiSelection].sort().join(''));
    };

    const getOptionClass = (key: string) => {
        const isSelected = isMultiChoice ? multiSelection.includes(key) : selectedOption === key;
        if (!result) return isSelected ? 'bg-primary/10 border-primary/30 ring-1 ring-primary/20' : 'hover:bg-muted/50 border-transparent';

        const isKeyInCorrect = result.correct_answer.includes(key);
        if (isKeyInCorrect) return 'bg-emerald-500/10 border-emerald-500/30 text-emerald-700 font-bold';
        if (isSelected && !isKeyInCorrect) return 'bg-rose-500/10 border-rose-500/30 text-rose-700 font-bold';
        return 'opacity-60 border-transparent';
    };

    const getCircleClass = (key: string) => {
        const isSelected = isMultiChoice ? multiSelection.includes(key) : selectedOption === key;
        if (!result) return isSelected ? 'bg-primary text-white border-primary shadow-md scale-110' : 'bg-background text-muted-foreground border-border group-hover:border-primary/50 group-hover:text-primary';

        const isKeyInCorrect = result.correct_answer.includes(key);
        if (isKeyInCorrect) return 'bg-emerald-500 text-white border-emerald-500 shadow-md';
        if (isSelected && !isKeyInCorrect) return 'bg-rose-500 text-white border-rose-500 shadow-md';
        return 'bg-background text-muted-foreground border-border';
    };

    return (
        <div className="py-6 group/item relative">
            {/* Question Header */}
            <div className="flex justify-between items-center mb-6">
                <div className="flex items-center gap-3">
                    {showTypeTag && (
                        <span className="px-2.5 py-0.5 bg-primary/10 text-primary text-[11px] font-bold rounded-full uppercase tracking-wider">
                            {question.type || '题型'}
                        </span>
                    )}
                    {isMultiChoice && (
                        <span className="px-2.5 py-0.5 bg-amber-500/10 text-amber-600 text-[11px] font-bold rounded-full">
                            多项选择
                        </span>
                    )}
                </div>
                {result && (
                    <button
                        onClick={handleRedo}
                        className="flex items-center gap-1.5 px-3 py-1.5 bg-muted/50 hover:bg-amber-500 hover:text-white rounded-xl text-xs font-bold transition-all text-muted-foreground shadow-sm"
                    >
                        <RefreshCw className="w-3 h-3" /> 重新练习
                    </button>
                )}
            </div>

            {/* Stem */}
            <div className="flex gap-4 mb-8">
                <span className="text-xl font-black text-primary/40 leading-none shrink-0 font-mono">
                    {index ? (isChild ? `(${index})` : `${index}.`) : ''}
                </span>
                <div className="text-lg font-bold text-foreground leading-relaxed tracking-wide" dangerouslySetInnerHTML={{ __html: parsedStem }} />
            </div>

            {/* Options Container */}
            {!isSubjective && (
                <div className={cn(
                    "space-y-3",
                    sharedOptions ? "flex flex-wrap gap-4" : "flex flex-col"
                )}>
                    {displayOptions.map((opt) => (
                        <div
                            key={opt.key}
                            onClick={() => handleOptionClick(opt.key)}
                            className={cn(
                                "group transition-all duration-200 cursor-pointer border rounded-2xl flex items-center",
                                sharedOptions ? "p-2 w-[50px] h-[50px] justify-center" : "p-4 gap-4",
                                getOptionClass(opt.key)
                            )}
                        >
                            <div className={cn(
                                "flex-shrink-0 w-8 h-8 rounded-full border-2 flex items-center justify-center font-bold transition-all",
                                getCircleClass(opt.key)
                            )}>
                                {opt.key}
                            </div>
                            {!sharedOptions && (
                                <div className="text-sm font-medium leading-relaxed" dangerouslySetInnerHTML={{ __html: opt.parsedValue }} />
                            )}
                        </div>
                    ))}
                </div>
            )}

            {isMultiChoice && !result && (
                <button
                    onClick={submitMultiChoice}
                    disabled={multiSelection.length === 0}
                    className="mt-6 px-8 py-2.5 bg-primary text-primary-foreground font-black rounded-xl shadow-lg hover:scale-105 active:scale-95 disabled:opacity-50 disabled:scale-100 transition-all text-sm uppercase tracking-widest"
                >
                    确认结果
                </button>
            )}

            {isSubjective && (
                <button
                    onClick={() => setShowAnswer(!showAnswer)}
                    className="mt-4 flex items-center gap-2 px-4 py-2 bg-muted/50 rounded-xl text-sm font-bold text-muted-foreground hover:bg-muted"
                >
                    {showAnswer ? <ChevronUp className="w-4 h-4" /> : <ChevronDown className="w-4 h-4" />}
                    {showAnswer ? "收起答案" : "查看参考答案"}
                </button>
            )}

            {/* Analysis Panel */}
            <AnimatePresence>
                {(result || (isSubjective && showAnswer)) && (
                    <motion.div
                        initial={{ opacity: 0, y: 10 }}
                        animate={{ opacity: 1, y: 0 }}
                        exit={{ opacity: 0, y: -10 }}
                        className="mt-8 overflow-hidden rounded-3xl border border-border/50 bg-card shadow-lg shadow-indigo-500/5"
                    >
                        <div className={cn(
                            "px-6 py-4 flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-border/50",
                            result?.is_correct ? "bg-emerald-500/5" : "bg-rose-500/5"
                        )}>
                            <div className="flex items-center gap-6">
                                <div className="flex flex-col">
                                    <span className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider mb-1">参考答案</span>
                                    <div className="flex items-center gap-2">
                                        <span className="text-2xl font-black text-emerald-600 font-mono tracking-tighter">{result?.correct_answer || question.correct}</span>
                                        {sharedOptions && result?.correct_answer && (
                                            <span className="text-xs text-muted-foreground bg-muted px-2 py-0.5 rounded font-medium">
                                                {sharedOptions[result.correct_answer]}
                                            </span>
                                        )}
                                    </div>
                                </div>
                                {!isSubjective && (
                                    <div className="flex flex-col">
                                        <span className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider mb-1">我的作答</span>
                                        <span className={cn(
                                            "text-2xl font-black font-mono tracking-tighter",
                                            result?.is_correct ? "text-emerald-500" : "text-rose-500"
                                        )}>
                                            {selectedOption || '未答'}
                                        </span>
                                    </div>
                                )}
                            </div>
                            <div className="flex items-center gap-3">
                                {result?.is_correct ? (
                                    <div className="flex items-center gap-2 text-emerald-600 bg-emerald-100/50 px-4 py-2 rounded-2xl border border-emerald-200 shadow-sm animate-bounce-subtle">
                                        <CheckCircle2 className="w-5 h-5" />
                                        <span className="text-sm font-black">回答正确</span>
                                    </div>
                                ) : !isSubjective && result ? (
                                    <div className="flex items-center gap-2 text-rose-600 bg-rose-100/50 px-4 py-2 rounded-2xl border border-rose-200 shadow-sm">
                                        <XCircle className="w-5 h-5" />
                                        <span className="text-sm font-black">错误</span>
                                    </div>
                                ) : null}
                            </div>
                        </div>

                        <div className="p-8 space-y-6">
                            {isSubjective && (
                                <div className="space-y-3">
                                    <h4 className="flex items-center gap-2 font-black text-foreground text-sm uppercase tracking-widest">
                                        <div className="w-1.5 h-4 bg-primary rounded-full" />
                                        参考范文
                                    </h4>
                                    <div className="text-sm font-medium text-muted-foreground leading-loose" dangerouslySetInnerHTML={{ __html: parsedCorrect }} />
                                </div>
                            )}
                            <div className="space-y-3">
                                <h4 className="flex items-center gap-2 font-black text-foreground text-sm uppercase tracking-widest">
                                    <div className="w-1.5 h-4 bg-indigo-500 rounded-full" />
                                    名师解析
                                </h4>
                                <div className="text-sm font-medium text-muted-foreground leading-loose" dangerouslySetInnerHTML={{ __html: parsedAnalysis }} />
                            </div>

                            {/* Metadata */}
                            <div className="flex flex-wrap gap-4 pt-4 border-t border-border/50">
                                {question.difficulty && (
                                    <div className="flex items-center gap-1.5 px-3 py-1 bg-muted/30 rounded-full border border-border/50 text-[11px] font-bold text-muted-foreground">
                                        <Info className="w-3 h-3" /> 难度: {question.difficulty}
                                    </div>
                                )}
                                {question.syllabus && (
                                    <div className="flex items-center gap-1.5 px-3 py-1 bg-blue-500/5 rounded-full border border-blue-500/20 text-[11px] font-bold text-blue-600">
                                        考纲: {question.syllabus}
                                    </div>
                                )}
                            </div>
                        </div>
                    </motion.div>
                )}
            </AnimatePresence>
        </div>
    );
}
