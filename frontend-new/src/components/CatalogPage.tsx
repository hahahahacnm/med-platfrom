import React, { useState, useEffect } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import {
    Layers, Search, Home, ChevronRight, Loader2,
    BookOpen, Target, Hash, Info, Settings2,
    Database, Flame, Heart, BookMarked, Menu
} from 'lucide-react';
import { cn } from '../lib/utils';
import api from '../lib/api';
import { TreeNavigation } from './TreeNavigation';
import { QuestionSession } from './QuestionSession';

interface CatalogPageProps {
    type: 'quiz' | 'mistakes' | 'favorites' | 'notes';
    title: string;
    description: string;
    icon: React.ElementType;
    treeEndpoint: string;
    skeletonEndpoint: string;
    colorClass: string;
}

export function CatalogPage({
    type, title, description, icon: Icon,
    treeEndpoint, skeletonEndpoint, colorClass
}: CatalogPageProps) {
    const [loading, setLoading] = useState(true);
    const [banks, setBanks] = useState<string[]>([]);
    const [currentBank, setCurrentBank] = useState('');
    const [selectedCategory, setSelectedCategory] = useState<{ id: number, name: string, full: string } | null>(null);
    const [activeSession, setActiveSession] = useState<{ cat: string, source: string } | null>(null);
    const [mobileTreeOpen, setMobileTreeOpen] = useState(false);
    const [isSidebarPinned, setIsSidebarPinned] = useState(true);

    // Initial load for banks
    useEffect(() => {
        const init = async () => {
            try {
                const res: any = await api.get('/banks');
                if (res.data && res.data.length > 0) {
                    setBanks(res.data);
                    const storageKey = `last_${type}_bank`;
                    const savedBank = typeof window !== 'undefined' ? localStorage.getItem(storageKey) : null;
                    const initialBank = savedBank && res.data.includes(savedBank) ? savedBank : res.data[0];
                    setCurrentBank(initialBank);
                }
            } catch (err) {
                console.error('Failed to fetch banks', err);
            } finally {
                setLoading(false);
            }
        };
        init();
    }, [type]);

    useEffect(() => {
        if (currentBank && typeof window !== 'undefined') {
            localStorage.setItem(`last_${type}_bank`, currentBank);
        }
    }, [currentBank, type]);

    const handleNodeSelect = (node: any) => {
        setSelectedCategory({ id: node.id, name: node.name, full: node.full });
        setActiveSession({ cat: node.full, source: currentBank });
        setMobileTreeOpen(false); // Close tree on select for mobile
        setIsSidebarPinned(false); // Auto-collapse the tree strictly for cleaner view
    };

    const bankSelectorNode = (
        <div className="flex bg-muted/30 p-1 mt-2 mb-2 rounded-xl border border-border/50 shadow-inner overflow-x-auto no-scrollbar scroll-smooth w-full">
            {banks.map(bank => (
                <button
                    key={bank}
                    onClick={() => { setCurrentBank(bank); setSelectedCategory(null); }}
                    className={cn(
                        "px-3 py-1.5 flex-1 rounded-lg text-[10px] font-black transition-all uppercase tracking-widest whitespace-nowrap text-center",
                        currentBank === bank
                            ? "bg-card text-primary shadow-sm ring-1 ring-border/10"
                            : "text-muted-foreground hover:bg-card/50"
                    )}
                >
                    {bank}
                </button>
            ))}
        </div>
    );

    return (
        <div className="flex h-screen md:h-[calc(100vh-64px)] overflow-hidden bg-background">
            {/* Desktop Sidebar - Tree */}
            <div className="hidden md:block h-full">
                {!loading && (
                    <TreeNavigation
                        type={type}
                        source={currentBank}
                        endpoint={treeEndpoint}
                        onSelect={handleNodeSelect}
                        activeCategoryId={selectedCategory?.id}
                        isPinned={isSidebarPinned}
                        onPinChange={setIsSidebarPinned}
                        topWidget={banks.length > 0 ? bankSelectorNode : null}
                    />
                )}
            </div>

            {/* Mobile Drawer - Tree */}
            <AnimatePresence>
                {mobileTreeOpen && (
                    <div className="md:hidden fixed inset-0 z-[100] flex">
                        <motion.div
                            initial={{ opacity: 0 }}
                            animate={{ opacity: 1 }}
                            exit={{ opacity: 0 }}
                            onClick={() => setMobileTreeOpen(false)}
                            className="absolute inset-0 bg-black/40 backdrop-blur-sm"
                        />
                        <motion.div
                            initial={{ x: "-100%" }}
                            animate={{ x: 0 }}
                            exit={{ x: "-100%" }}
                            transition={{ type: "spring", damping: 25, stiffness: 200 }}
                            className="relative w-[80%] max-w-xs h-full bg-card"
                        >
                            <TreeNavigation
                                type={type}
                                source={currentBank}
                                endpoint={treeEndpoint}
                                onSelect={handleNodeSelect}
                                activeCategoryId={selectedCategory?.id}
                                topWidget={banks.length > 0 ? bankSelectorNode : null}
                            />
                        </motion.div>
                    </div>
                )}
            </AnimatePresence>

            {/* Main Content Area */}
            <div className="flex-1 flex flex-col overflow-hidden relative">
                {/* Mobile Top Bar */}
                {!activeSession && (
                    <div className="md:hidden shrink-0 flex items-center px-4 py-3 bg-card/80 backdrop-blur-md border-b border-border/50 sticky top-0 z-10 transition-all">
                        <button
                            onClick={() => setMobileTreeOpen(true)}
                            className="p-2 -ml-2 rounded-xl text-muted-foreground hover:bg-muted transition-colors mr-3 active:scale-95"
                        >
                            <Menu size={20} />
                        </button>
                        <div className="flex items-center gap-2 min-w-0">
                            <Icon className={cn("w-4 h-4 shrink-0", colorClass)} />
                            <div className="text-sm font-black tracking-tight uppercase truncate">{title}</div>
                        </div>
                    </div>
                )}

                {/* Content */}
                <div className="flex-1 overflow-hidden relative">
                    {!selectedCategory ? (
                        <div className="h-full overflow-y-auto p-6 md:p-10 no-scrollbar flex flex-col items-center justify-center space-y-8 md:space-y-10 group">
                            <div className="relative">
                                <motion.div
                                    className={cn("w-24 h-24 md:w-32 md:h-32 rounded-[2rem] md:rounded-[2.5rem] bg-opacity-5 flex items-center justify-center shadow-2xl", colorClass.replace('text-', 'bg-'))}
                                    animate={{ rotate: [0, 5, -5, 0] }}
                                    transition={{ duration: 4, repeat: Infinity, ease: "easeInOut" }}
                                >
                                    <Icon className={cn("w-10 h-10 md:w-14 md:h-14", colorClass)} />
                                </motion.div>
                                <div className="absolute -bottom-1 md:-bottom-2 -right-1 md:-right-2 w-8 h-8 md:w-10 md:h-10 bg-card rounded-xl md:rounded-2xl shadow-lg border border-border/50 flex items-center justify-center text-primary group-hover:scale-110 transition-transform">
                                    <Info size={16} />
                                </div>
                            </div>

                            <div className="text-center space-y-4 max-w-md px-4">
                                <h3 className="font-black text-xl md:text-2xl text-foreground uppercase tracking-widest">
                                    <span className="md:hidden">点击左侧按钮选择章节</span>
                                    <span className="hidden md:inline">请从左侧选择章节</span>
                                </h3>
                                <p className="text-xs md:text-sm font-medium text-muted-foreground leading-relaxed">
                                    我们已按层级为您整理好了所有内容。展开左侧目录树，点击具体的知识点或章节即可立即进入学习模式。
                                </p>
                            </div>

                            <div className="grid grid-cols-2 gap-3 md:gap-4 w-full max-w-sm px-4">
                                <div className="p-3 md:p-4 bg-card border border-border/50 rounded-2xl flex flex-col items-center text-center space-y-1 md:space-y-2">
                                    <Database className="w-4 h-4 md:w-5 md:h-5 text-blue-500" />
                                    <span className="text-[9px] md:text-[10px] font-black uppercase tracking-widest">海量资源</span>
                                </div>
                                <div className="p-3 md:p-4 bg-card border border-border/50 rounded-2xl flex flex-col items-center text-center space-y-1 md:space-y-2">
                                    <Hash className="w-4 h-4 md:w-5 md:h-5 text-emerald-500" />
                                    <span className="text-[9px] md:text-[10px] font-black uppercase tracking-widest">精准定位</span>
                                </div>
                            </div>
                        </div>
                    ) : !activeSession ? (
                        <div className="h-full flex flex-col items-center justify-center space-y-4">
                            <Loader2 className="w-8 h-8 animate-spin text-primary opacity-20" />
                            <span className="text-xs font-black uppercase tracking-widest text-muted-foreground animate-pulse">
                                正在初始化学习会话...
                            </span>
                        </div>
                    ) : (
                        <div className="absolute inset-0">
                            <QuestionSession
                                category={activeSession.cat}
                                source={activeSession.source}
                                endpoint={skeletonEndpoint}
                                onExit={() => {
                                    setActiveSession(null);
                                    setIsSidebarPinned(true);
                                }}
                            />
                        </div>
                    )}
                </div>
            </div>
        </div>
    );
}
