import React, { useState, useEffect, useCallback } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import {
    ChevronRight, ChevronDown, Folder, FileText,
    Search, Database, Loader2, Pin, PinOff
} from 'lucide-react';
import { cn } from '../lib/utils';
import api from '../lib/api';

interface CategoryNode {
    id: number;
    name: string;
    full: string;
    count?: number;
    total_count?: number;
    done_count?: number;
    isLeaf: boolean;
    level: number;
    children?: CategoryNode[];
    expanded?: boolean;
    loading?: boolean;
}

interface TreeNavigationProps {
    type: 'quiz' | 'mistakes' | 'favorites' | 'notes';
    source: string;
    onSelect: (cat: CategoryNode) => void;
    activeCategoryId?: number;
    endpoint: string;
    isPinned?: boolean;
    onPinChange?: (pinned: boolean) => void;
    topWidget?: React.ReactNode;
}

export function TreeNavigation({ type, source, onSelect, activeCategoryId, endpoint, isPinned: externalIsPinned, onPinChange, topWidget }: TreeNavigationProps) {
    const [treeData, setTreeData] = useState<CategoryNode[]>([]);
    const [loading, setLoading] = useState(false);
    const [searchQuery, setSearchQuery] = useState('');
    const [internalIsPinned, setInternalIsPinned] = useState(true);

    const isPinned = externalIsPinned !== undefined ? externalIsPinned : internalIsPinned;
    const setIsPinned = (val: boolean) => {
        if (onPinChange) onPinChange(val);
        setInternalIsPinned(val);
    };

    const fetchNodes = useCallback(async (parentId: number = 0) => {
        try {
            const res: any = await api.get(endpoint, {
                params: { source, parent_id: parentId || undefined }
            });
            if (res.data) {
                return res.data.map((item: any) => ({
                    ...item,
                    isLeaf: item.isLeaf || item.is_leaf,
                    count: item.count,
                    total_count: item.total_count,
                    done_count: item.done_count,
                    expanded: false,
                    loading: false,
                    children: []
                }));
            }
            return [];
        } catch (err) {
            console.error('Failed to fetch tree nodes', err);
            return [];
        }
    }, [source, endpoint]);

    useEffect(() => {
        const loadRoot = async () => {
            if (!source) return;
            setLoading(true);
            const nodes = await fetchNodes(0);
            setTreeData(nodes);
            setLoading(false);
        };
        loadRoot();
    }, [source, fetchNodes]);

    const toggleNode = async (nodeId: number) => {
        const updateTree = async (nodes: CategoryNode[]): Promise<CategoryNode[]> => {
            return Promise.all(nodes.map(async (n) => {
                if (n.id === nodeId) {
                    if (n.isLeaf) return n;
                    const newExpanded = !n.expanded;
                    if (newExpanded && n.children?.length === 0) {
                        const children = await fetchNodes(n.id);
                        return { ...n, expanded: true, children };
                    }
                    return { ...n, expanded: newExpanded };
                }
                if (n.children && n.children.length > 0) {
                    return { ...n, children: await updateTree(n.children) };
                }
                return n;
            }));
        };
        const newData = await updateTree(treeData);
        setTreeData(newData);
    };

    const renderNode = (node: CategoryNode) => {
        const isActive = activeCategoryId === node.id;
        const hasChildren = !node.isLeaf;

        return (
            <div key={node.id} className="select-none">
                <div
                    className={cn(
                        "flex items-center gap-2 px-3 py-2 rounded-xl transition-all duration-200 cursor-pointer group",
                        isActive ? "bg-primary text-white shadow-md shadow-primary/20" : "hover:bg-muted/50 text-muted-foreground hover:text-foreground"
                    )}
                    onClick={() => {
                        if (hasChildren && !node.expanded) toggleNode(node.id);
                        onSelect(node);
                    }}
                    style={{ paddingLeft: `${node.level * 16 + 12}px` }}
                >
                    <div
                        className="w-5 h-5 flex items-center justify-center shrink-0 hover:bg-background/20 rounded-md cursor-pointer transition-colors"
                        onClick={(e) => {
                            if (hasChildren) {
                                e.stopPropagation();
                                toggleNode(node.id);
                            }
                        }}
                    >
                        {node.loading ? (
                            <Loader2 className="w-3 h-3 animate-spin" />
                        ) : hasChildren ? (
                            node.expanded ? <ChevronDown size={14} /> : <ChevronRight size={14} />
                        ) : (
                            <FileText size={14} className={cn(isActive ? "text-white" : "text-primary/60")} />
                        )}
                    </div>

                    <span className={cn(
                        "flex-1 text-sm font-medium truncate",
                        isActive ? "text-white" : "text-foreground/80 group-hover:text-foreground"
                    )}>
                        {node.name}
                    </span>

                    {node.total_count !== undefined ? (
                        node.total_count > 0 && (
                            <span className={cn(
                                "text-[10px] font-black px-1.5 py-0.5 rounded-md whitespace-nowrap shadow-sm transition-colors font-mono",
                                isActive ? "bg-white/20 text-white border border-white/20" :
                                    (node.done_count && node.total_count && node.done_count >= node.total_count) ? "bg-emerald-500/10 text-emerald-600 border border-emerald-500/20" :
                                        "bg-muted text-muted-foreground border border-border/50"
                            )}>
                                {node.done_count || 0}<span className={cn("opacity-50 font-normal px-0.5", isActive ? "text-white" : "text-muted-foreground")}>/</span>{node.total_count}
                            </span>
                        )
                    ) : node.count !== undefined && node.count > 0 ? (
                        <span className={cn(
                            "text-[10px] font-black px-1.5 py-0.5 rounded-md whitespace-nowrap shadow-sm border font-mono",
                            isActive ? "bg-white/20 text-white border-white/20" : "bg-card text-foreground/70 border-border/50"
                        )}>
                            {node.count}
                        </span>
                    ) : null}
                </div>

                <AnimatePresence>
                    {node.expanded && node.children && node.children.length > 0 && (
                        <motion.div
                            initial={{ height: 0, opacity: 0 }}
                            animate={{ height: "auto", opacity: 1 }}
                            exit={{ height: 0, opacity: 0 }}
                            transition={{ duration: 0.2 }}
                            className="overflow-hidden"
                        >
                            {node.children.map(renderNode)}
                        </motion.div>
                    )}
                </AnimatePresence>
            </div>
        );
    };

    return (
        <div className={cn(
            "flex flex-col bg-card border-r border-border/50 h-full transition-all duration-300",
            isPinned ? "w-[300px]" : "w-12 overflow-hidden"
        )}>
            {isPinned ? (
                <>
                    <div className="p-4 border-b border-border/50 space-y-4">
                        <div className="flex items-center justify-between">
                            <span className="text-sm font-bold uppercase tracking-widest text-primary">目录导航</span>
                            <button onClick={() => setIsPinned(false)} className="hidden md:block text-muted-foreground hover:text-primary transition-colors">
                                <Pin size={16} />
                            </button>
                        </div>
                        {topWidget}
                        <div className="relative group">
                            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-muted-foreground group-focus-within:text-primary transition-colors" />
                            <input
                                placeholder="快速定位章节..."
                                value={searchQuery}
                                onChange={(e) => setSearchQuery(e.target.value)}
                                className="w-full pl-9 pr-3 py-2 bg-muted/50 border border-border/50 rounded-xl text-xs font-medium focus:ring-2 focus:ring-primary/20 outline-none transition-all"
                            />
                        </div>
                    </div>
                    <div className="flex-1 overflow-y-auto p-2 no-scrollbar">
                        {loading && treeData.length === 0 ? (
                            <div className="flex flex-col items-center justify-center py-20 space-y-3 opacity-20">
                                <Loader2 className="w-8 h-8 animate-spin" />
                                <span className="text-xs font-bold uppercase tracking-widest">正在载入目录树...</span>
                            </div>
                        ) : treeData.length > 0 ? (
                            <div className="space-y-0.5">
                                {treeData.map(renderNode)}
                            </div>
                        ) : (
                            <div className="flex flex-col items-center justify-center py-20 space-y-3 opacity-20">
                                <Database className="w-8 h-8" />
                                <span className="text-xs font-bold uppercase tracking-widest">暂无目录数据</span>
                            </div>
                        )}
                    </div>
                </>
            ) : (
                <div className="flex flex-col items-center py-4 space-y-4">
                    <button onClick={() => setIsPinned(true)} className="text-muted-foreground hover:text-primary transition-colors">
                        <PinOff size={16} />
                    </button>
                    <div className="w-[1px] h-full bg-border/50" />
                </div>
            )}
        </div>
    );
}
