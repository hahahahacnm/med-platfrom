import React, { useState, useEffect } from 'react';
import { motion } from 'framer-motion';
import { Search, Filter, MoreVertical, FileWarning, Star, BookMarked, BrainCircuit, Play } from 'lucide-react';
import { cn } from '../lib/utils';
import api from '../lib/api';
import { QuestionSession } from './QuestionSession';

export function ListItemsContent({ title, type }: { title: string, type: 'mistakes' | 'favorites' | 'notes' }) {
    const [loading, setLoading] = useState(true);
    const [items, setItems] = useState<any[]>([]);
    const [activeSession, setActiveSession] = useState(false);

    useEffect(() => {
        const fetchItems = async () => {
            setLoading(true);
            try {
                // Determine endpoint based on type
                let endpoint = '/mistakes/skeleton';
                if (type === 'favorites') endpoint = '/favorites';
                if (type === 'notes') endpoint = '/notes';

                const res: any = await api.get(endpoint);
                if (res.data) {
                    setItems(res.data);
                }
            } catch (err) {
                console.error(`Failed to fetch ${type}`, err);
            } finally {
                setLoading(false);
            }
        };
        fetchItems();
    }, [type]);

    const container = { hidden: { opacity: 0 }, show: { opacity: 1, transition: { staggerChildren: 0.1 } } };
    const itemAnim = { hidden: { opacity: 0, y: 15 }, show: { opacity: 1, y: 0, transition: { ease: "easeOut" } } };

    const getIcon = () => {
        switch (type) {
            case 'mistakes': return <FileWarning className="text-rose-500 w-5 h-5" />;
            case 'favorites': return <Star className="text-amber-500 w-5 h-5" />;
            case 'notes': return <BookMarked className="text-blue-500 w-5 h-5" />;
            default: return <BrainCircuit className="text-primary w-5 h-5" />;
        }
    };

    const getBg = () => {
        switch (type) {
            case 'mistakes': return "bg-rose-500/10";
            case 'favorites': return "bg-amber-500/10";
            case 'notes': return "bg-blue-500/10";
            default: return "bg-primary/10";
        }
    };

    return (
        <div className="p-6 lg:p-10 max-w-[1400px] mx-auto space-y-8">
            {activeSession ? (
                <QuestionSession
                    endpoint={type === 'mistakes' ? '/mistakes/skeleton' : (type === 'favorites' ? '/favorites/skeleton' : '/notes/skeleton')}
                    onExit={() => setActiveSession(false)}
                />
            ) : (
                <>
                    <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-6">
                        <div>
                            <h1 className="text-3xl font-bold tracking-tight text-foreground flex items-center gap-3">
                                <div className={cn("p-2 rounded-xl", getBg())}>{getIcon()}</div>
                                {title}
                            </h1>
                            <p className="text-muted-foreground mt-2 text-sm italic">
                                您共有 {items.length} 条{title}记录，保持复习是成功的关键。
                            </p>
                        </div>

                        <div className="flex gap-3 w-full md:w-auto">
                            <button
                                onClick={() => setActiveSession(true)}
                                className="flex-1 md:flex-none px-6 py-2.5 bg-primary text-primary-foreground font-bold rounded-xl shadow-lg shadow-primary/20 hover:scale-105 active:scale-95 transition-all flex items-center justify-center gap-2"
                            >
                                <Play className="w-4 h-4 fill-current" />
                                考前突击
                            </button>
                            <div className="relative hidden sm:block">
                                <Search className="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
                                <input placeholder="搜索记录..." className="w-full pl-9 pr-4 py-2.5 bg-card border border-border/50 rounded-xl text-sm focus:ring-2 focus:ring-primary/20 outline-none transition-all shadow-sm" />
                            </div>
                        </div>
                    </div>

                    {loading ? (
                        <div className="space-y-4">
                            {[1, 2, 3].map(i => <div key={i} className="h-24 bg-muted animate-pulse rounded-2xl" />)}
                        </div>
                    ) : items.length === 0 ? (
                        <div className="flex flex-col items-center justify-center py-20 text-center opacity-40">
                            <div className={cn("w-20 h-20 rounded-full flex items-center justify-center mb-6", getBg())}>
                                {getIcon()}
                            </div>
                            <p className="text-lg font-bold italic">暂时没有任何记录</p>
                        </div>
                    ) : (
                        <motion.div variants={container as any} initial="hidden" animate="show" className="space-y-4">
                            {items.map((item, i) => (
                                <motion.div
                                    key={item.id || i}
                                    variants={itemAnim as any}
                                    className="group flex flex-col sm:flex-row gap-4 sm:items-center bg-card p-5 rounded-3xl border border-border hover:border-primary/40 shadow-sm hover:shadow-xl transition-all cursor-pointer"
                                >
                                    <div className={cn("p-3 rounded-2xl flex-shrink-0 self-start sm:self-center transition-colors group-hover:bg-primary group-hover:text-white", getBg())}>
                                        {getIcon()}
                                    </div>
                                    <div className="flex-1 min-w-0">
                                        <h3 className="font-bold text-foreground text-sm sm:text-base mb-1 truncate group-hover:text-primary transition-colors">
                                            {(item.stem || item.content || '未命名条目').replace(/!\[.*?\]\(.*?\)/g, '[图片]')}
                                        </h3>
                                        <div className="flex items-center gap-3">
                                            <span className="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">{item.type || '未分类'}</span>
                                            {item.source && (
                                                <span className="text-[10px] font-black uppercase tracking-widest text-primary/80 bg-primary/5 px-1.5 py-0.5 rounded leading-none">
                                                    {item.source}
                                                </span>
                                            )}
                                            {item.wrong_count && (
                                                <span className="text-[10px] font-black uppercase text-rose-500 bg-rose-500/5 px-1.5 py-0.5 rounded leading-none flex items-center gap-1">
                                                    已被难住 {item.wrong_count} 次
                                                </span>
                                            )}
                                        </div>
                                    </div>
                                    <div className="self-end sm:self-center opacity-0 group-hover:opacity-100 transition-opacity">
                                        <div className="flex gap-2">
                                            <button className="p-2 text-muted-foreground hover:bg-muted rounded-xl transition-colors">
                                                <MoreVertical className="w-5 h-5" />
                                            </button>
                                        </div>
                                    </div>
                                </motion.div>
                            ))}
                        </motion.div>
                    )}
                </>
            )}
        </div>
    );
}
