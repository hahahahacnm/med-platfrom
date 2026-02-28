import React, { useState, useEffect, useRef } from 'react';
import {
    MessageSquare,
    Send,
    Image as ImageIcon,
    Smile,
    Globe,
    Lock,
    ThumbsUp,
    Star,
    MoreVertical,
    Trash2,
    AlertTriangle,
    Edit2,
    X,
    ChevronDown,
    Loader2,
    CornerDownRight,
    Flame,
    Clock
} from 'lucide-react';
import { motion, AnimatePresence } from 'framer-motion';
import api from '../lib/api';
import { formatDistanceToNow } from 'date-fns';
import { zhCN } from 'date-fns/locale';
import { cn } from '../lib/utils';

interface User {
    id: number;
    username: string;
    nickname: string;
    avatar: string;
    role: string;
}

interface Note {
    id: number;
    user_id: number;
    question_id: number;
    content: string;
    is_public: boolean;
    images: string[] | null;
    like_count: number;
    report_count: number;
    created_at: string;
    user: User;
    parent?: Note;
    is_liked?: boolean;
    is_collected?: boolean;
}

interface QuestionCommentsProps {
    questionId: number;
    onClose?: () => void;
}

const EMOJI_LIST = ['😀', '😃', '😄', '😁', '😆', '😅', '😂', '🤣', '😊', '😇', '🙂', '🙃', '😉', '😌', '😍', '🥰', '😘', '😗', '😙', '😚', '😋', '😛', '😝', '😜', '🤪', '🤨', '🧐', '🤓', '😎', '🤩', '🥳', '😏', '😒', '😞', '😔', '😟', '😕', '🙁', '☹️', '😣', '😖', '😫', '😩', '🥺', '😢', '😭', '😤', '😠', '😡', '🤬', '🤯', '😳', '🥵', '🥶', '😱', '😨', '😰', '😥', '😓', '🤗', '🤔', '🤭', '🤫', '🤥', '😶', '😐', '😑', '😬', '🙄', '😯', '😦', '😧', '😮', '😲', '😴', '🤤', '😪', '😵', '🤐', '🥴', '🤢', '🤮', '🤧', '😷', '🤒', '🤕', '🤑', '🤠', '😈', '👿', '👹', '👺', '🤡', '💩', '👻', '💀', '☠️', '👽', '👾', '🤖', '🎃', '😺', '😸', '😹', '😻', '😼', '😽', '🙀', '😿', '😾'];

export const QuestionComments: React.FC<QuestionCommentsProps> = ({ questionId, onClose }) => {
    const [notes, setNotes] = useState<Note[]>([]);
    const [loading, setLoading] = useState(true);
    const [loadingMore, setLoadingMore] = useState(false);
    const [hasMore, setHasMore] = useState(false);
    const [page, setPage] = useState(1);
    const [sortMode, setSortMode] = useState<'hot' | 'time'>('hot');

    // Editor State
    const [content, setContent] = useState('');
    const [isPublic, setIsPublic] = useState(true);
    const [images, setImages] = useState<string[]>([]);
    const [parentId, setParentId] = useState<number | null>(null);
    const [replyUser, setReplyUser] = useState<string | null>(null);
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [showEmoji, setShowEmoji] = useState(false);
    const [editingNoteId, setEditingNoteId] = useState<number | null>(null);
    const [currentUserId, setCurrentUserId] = useState<number | null>(null);

    const scrollRef = useRef<HTMLDivElement>(null);
    const editorRef = useRef<HTMLTextAreaElement>(null);

    useEffect(() => {
        const uid = localStorage.getItem('id');
        if (uid) setCurrentUserId(parseInt(uid));
        fetchNotes(1, true);
    }, [questionId, sortMode]);

    const fetchNotes = async (targetPage: number, reset = false) => {
        if (reset) {
            setLoading(true);
            setPage(1);
        } else {
            setLoadingMore(true);
        }

        try {
            const res = await api.get('/notes', {
                params: {
                    question_id: questionId,
                    page: targetPage,
                    page_size: 10,
                    sort: sortMode
                }
            });

            if (reset) {
                setNotes(res.data.data || []);
            } else {
                setNotes(prev => [...prev, ...(res.data.data || [])]);
            }
            setHasMore(res.data.has_more);
            setPage(targetPage);
        } catch (err) {
            console.error('Failed to fetch notes', err);
        } finally {
            setLoading(false);
            setLoadingMore(false);
        }
    };

    const handleSortChange = (mode: 'hot' | 'time') => {
        if (mode === sortMode) return;
        setSortMode(mode);
    };

    const handleUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
        const file = e.target.files?.[0];
        if (!file) return;

        if (images.length >= 5) {
            alert('最多上传5张图片');
            return;
        }

        const formData = new FormData();
        formData.append('file', file);

        try {
            const res = await api.post('/notes/upload', formData, {
                headers: { 'Content-Type': 'multipart/form-data' }
            });
            setImages(prev => [...prev, res.data.url]);
        } catch (err) {
            console.error('Upload failed', err);
            alert('图片上传失败');
        }
    };

    const handleSubmit = async () => {
        if (!content.trim() && images.length === 0) return;
        if (content.length > 200) {
            alert('内容不能超过200字');
            return;
        }

        setIsSubmitting(true);
        try {
            const payload = {
                id: editingNoteId || 0,
                question_id: questionId,
                content,
                is_public: isPublic,
                parent_id: parentId,
                images
            };

            await api.post('/notes', payload);

            // Reset editor
            setContent('');
            setImages([]);
            setParentId(null);
            setReplyUser(null);
            setEditingNoteId(null);
            setShowEmoji(false);

            // Refresh
            fetchNotes(1, true);
        } catch (err: any) {
            alert(err.response?.data?.error || '发布失败');
        } finally {
            setIsSubmitting(false);
        }
    };

    const handleLike = async (note: Note) => {
        try {
            const res = await api.post(`/notes/${note.id}/like`);
            setNotes(prev => prev.map(n => {
                if (n.id === note.id) {
                    return { ...n, is_liked: res.data.is_liked, like_count: res.data.like_count };
                }
                return n;
            }));
        } catch (err) {
            console.error(err);
        }
    };

    const handleCollect = async (note: Note) => {
        try {
            const res = await api.post(`/notes/${note.id}/collect`);
            setNotes(prev => prev.map(n => {
                if (n.id === note.id) {
                    return { ...n, is_collected: res.data.is_collected };
                }
                return n;
            }));
        } catch (err) {
            console.error(err);
        }
    };

    const handleDelete = async (noteId: number) => {
        if (!confirm('确定要删除这条笔记吗？')) return;
        try {
            await api.delete(`/notes/${noteId}`);
            setNotes(prev => prev.filter(n => n.id !== noteId));
        } catch (err) {
            console.error(err);
        }
    };

    const handleReport = async (noteId: number) => {
        const reason = prompt('请输入举报原因:');
        if (!reason) return;
        try {
            await api.post(`/notes/${noteId}/report`, { reason });
            alert('举报已提交，我们会尽快处理');
        } catch (err) {
            console.error(err);
        }
    };

    const startReply = (note: Note) => {
        setParentId(note.id);
        setReplyUser(note.user.nickname || note.user.username);
        setEditingNoteId(null);
        editorRef.current?.focus();
    };

    const startEdit = (note: Note) => {
        setEditingNoteId(note.id);
        setContent(note.content);
        setImages(note.images || []);
        setIsPublic(note.is_public);
        setParentId(null);
        setReplyUser(null);
        editorRef.current?.focus();
    };

    return (
        <div className="flex flex-col h-full bg-background border-l border-border/50">
            {/* Header */}
            <div className="flex items-center justify-between px-6 py-4 border-b border-border/50">
                <div className="flex items-center gap-2">
                    <MessageSquare className="w-5 h-5 text-primary" />
                    <h2 className="text-lg font-bold tracking-tight">讨论区 & 笔记</h2>
                </div>
                <div className="flex items-center gap-4">
                    <div className="flex bg-muted p-1 rounded-lg">
                        <button
                            onClick={() => handleSortChange('hot')}
                            className={cn(
                                "flex items-center gap-1.5 px-3 py-1 text-xs font-bold rounded-md transition-all",
                                sortMode === 'hot' ? "bg-card text-primary shadow-sm" : "text-muted-foreground hover:text-foreground"
                            )}
                        >
                            <Flame className="w-3.5 h-3.5" /> 最热
                        </button>
                        <button
                            onClick={() => handleSortChange('time')}
                            className={cn(
                                "flex items-center gap-1.5 px-3 py-1 text-xs font-bold rounded-md transition-all",
                                sortMode === 'time' ? "bg-card text-primary shadow-sm" : "text-muted-foreground hover:text-foreground"
                            )}
                        >
                            <Clock className="w-3.5 h-3.5" /> 最新
                        </button>
                    </div>
                    {onClose && (
                        <button onClick={onClose} className="p-2 hover:bg-muted rounded-full transition-colors">
                            <X className="w-5 h-5 text-muted-foreground" />
                        </button>
                    )}
                </div>
            </div>

            {/* Content List */}
            <div ref={scrollRef} className="flex-1 overflow-y-auto p-4 custom-scrollbar">
                {loading ? (
                    <div className="flex flex-col items-center justify-center h-64 gap-3">
                        <Loader2 className="w-8 h-8 text-primary animate-spin" />
                        <p className="text-sm text-muted-foreground animate-pulse">正在加载讨论内容...</p>
                    </div>
                ) : notes.length === 0 ? (
                    <div className="flex flex-col items-center justify-center h-64 text-center">
                        <div className="w-16 h-16 bg-muted rounded-full flex items-center justify-center mb-4">
                            <MessageSquare className="w-8 h-8 text-muted-foreground/30" />
                        </div>
                        <h3 className="font-bold text-foreground mb-1">暂无讨论</h3>
                        <p className="text-sm text-muted-foreground">抢个沙发，分享你的解题心得吧！</p>
                    </div>
                ) : (
                    <div className="space-y-6">
                        {notes.map((note, idx) => (
                            <motion.div
                                key={note.id}
                                initial={{ opacity: 0, y: 20 }}
                                animate={{ opacity: 1, y: 0 }}
                                transition={{ delay: idx * 0.05 }}
                                className="group"
                            >
                                <div className="flex gap-3">
                                    <div className="flex-shrink-0">
                                        <div className="w-10 h-10 rounded-full overflow-hidden border border-border bg-muted">
                                            {note.user.avatar ? (
                                                <img src={note.user.avatar.startsWith('http') ? note.user.avatar : `${api.defaults.baseURL?.replace('/api/v1', '')}${note.user.avatar}`} alt={note.user.nickname} className="w-full h-full object-cover" />
                                            ) : (
                                                <div className="w-full h-full flex items-center justify-center text-sm font-bold bg-primary/10 text-primary">
                                                    {(note.user.nickname || note.user.username || '?')[0].toUpperCase()}
                                                </div>
                                            )}
                                        </div>
                                    </div>
                                    <div className="flex-1 min-w-0">
                                        <div className="flex items-center justify-between mb-1">
                                            <div className="flex items-center gap-2">
                                                <span className="font-bold text-sm text-foreground truncate">
                                                    {note.user.nickname || note.user.username}
                                                </span>
                                                {note.user.role === 'admin' && (
                                                    <span className="px-1.5 py-0.5 bg-rose-500/10 text-rose-500 text-[10px] font-black rounded uppercase">Staff</span>
                                                )}
                                                <span className="text-[10px] text-muted-foreground font-medium">
                                                    {formatDistanceToNow(new Date(note.created_at), { addSuffix: true, locale: zhCN })}
                                                </span>
                                                {!note.is_public && (
                                                    <Lock className="w-3 h-3 text-amber-500" />
                                                )}
                                            </div>
                                            <div className="opacity-0 group-hover:opacity-100 transition-opacity flex items-center gap-1">
                                                <button onClick={() => startReply(note)} className="p-1.5 hover:bg-muted rounded-md text-muted-foreground hover:text-primary transition-colors">
                                                    <MessageSquare className="w-4 h-4" />
                                                </button>
                                                <div className="relative dropdown-container">
                                                    <button className="p-1.5 hover:bg-muted rounded-md text-muted-foreground transition-colors peer">
                                                        <MoreVertical className="w-4 h-4" />
                                                    </button>
                                                    <div className="absolute right-0 top-full mt-1 w-32 bg-card border border-border rounded-xl shadow-xl py-1 z-10 hidden group-hover:block peer-focus:block hover:block">
                                                        <button onClick={() => handleReport(note.id)} className="w-full text-left px-3 py-1.5 text-xs font-medium text-muted-foreground hover:bg-muted hover:text-rose-500 flex items-center gap-2">
                                                            <AlertTriangle className="w-3.5 h-3.5" /> 举报
                                                        </button>
                                                        {currentUserId === note.user_id && (
                                                            <>
                                                                <button onClick={() => startEdit(note)} className="w-full text-left px-3 py-1.5 text-xs font-medium text-muted-foreground hover:bg-muted hover:text-primary flex items-center gap-2">
                                                                    <Edit2 className="w-3.5 h-3.5" /> 编辑
                                                                </button>
                                                                <button onClick={() => handleDelete(note.id)} className="w-full text-left px-3 py-1.5 text-xs font-medium text-muted-foreground hover:bg-muted hover:text-rose-500 flex items-center gap-2">
                                                                    <Trash2 className="w-3.5 h-3.5" /> 删除
                                                                </button>
                                                            </>
                                                        )}
                                                    </div>
                                                </div>
                                            </div>
                                        </div>

                                        {/* Reference Quote */}
                                        {note.parent && (
                                            <div className="mb-2 p-2.5 bg-muted/50 rounded-xl border-l-2 border-primary/30 flex items-start gap-2">
                                                <CornerDownRight className="w-3.5 h-3.5 text-primary/50 mt-0.5 flex-shrink-0" />
                                                <div className="min-w-0">
                                                    <span className="text-[11px] font-bold text-primary mr-1">@{note.parent.user?.nickname || '用户'}</span>
                                                    <p className="text-xs text-muted-foreground line-clamp-2">{note.parent.content}</p>
                                                </div>
                                            </div>
                                        )}

                                        <p className="text-sm text-foreground/90 leading-relaxed whitespace-pre-wrap break-words">
                                            {note.content}
                                        </p>

                                        {/* Images */}
                                        {note.images && note.images.length > 0 && (
                                            <div className="flex flex-wrap gap-2 mt-3">
                                                {note.images.map((img, i) => {
                                                    const fullUrl = img.startsWith('http') ? img : `${api.defaults.baseURL?.replace('/api/v1', '')}${img}`;
                                                    return (
                                                        <a key={i} href={fullUrl} target="_blank" rel="noopener noreferrer" className="relative w-20 h-20 rounded-lg overflow-hidden border border-border/50 bg-muted hover:ring-2 hover:ring-primary/50 transition-all">
                                                            <img src={fullUrl} alt="Note Attachment" className="w-full h-full object-cover" />
                                                        </a>
                                                    );
                                                })}
                                            </div>
                                        )}

                                        <div className="flex items-center gap-4 mt-3">
                                            <button
                                                onClick={() => handleLike(note)}
                                                className={cn(
                                                    "flex items-center gap-1.5 text-xs font-bold transition-colors",
                                                    note.is_liked ? "text-rose-500" : "text-muted-foreground hover:text-rose-500"
                                                )}
                                            >
                                                <ThumbsUp className={cn("w-3.5 h-3.5", note.is_liked && "fill-current")} />
                                                {note.like_count > 0 && note.like_count}
                                            </button>
                                            <button
                                                onClick={() => handleCollect(note)}
                                                className={cn(
                                                    "flex items-center gap-1.5 text-xs font-bold transition-colors",
                                                    note.is_collected ? "text-amber-500" : "text-muted-foreground hover:text-amber-500"
                                                )}
                                            >
                                                <Star className={cn("w-3.5 h-3.5", note.is_collected && "fill-current")} />
                                                {note.is_collected ? '已存' : '存笔记本'}
                                            </button>
                                        </div>
                                    </div>
                                </div>
                            </motion.div>
                        ))}

                        {hasMore && (
                            <div className="flex justify-center pt-4 pb-8">
                                <button
                                    onClick={() => fetchNotes(page + 1)}
                                    disabled={loadingMore}
                                    className="flex items-center gap-2 px-6 py-2 bg-card hover:bg-muted border border-border text-xs font-black text-muted-foreground hover:text-foreground rounded-full transition-all"
                                >
                                    {loadingMore ? <Loader2 className="w-3.5 h-3.5 animate-spin" /> : <ChevronDown className="w-3.5 h-3.5" />}
                                    查看更多评论
                                </button>
                            </div>
                        )}
                    </div>
                )}
            </div>

            {/* Input Area */}
            <div className="p-4 border-t border-border shadow-[0_-8px_32px_-12px_rgba(0,0,0,0.1)]">
                {parentId && (
                    <div className="mb-3 px-3 py-2 bg-primary/5 border border-primary/10 rounded-xl flex items-center justify-between animate-in slide-in-from-bottom-2">
                        <div className="flex items-center gap-2 text-xs">
                            <CornerDownRight className="w-3.5 h-3.5 text-primary" />
                            <span className="font-bold text-primary">回复 @{replyUser}</span>
                        </div>
                        <button onClick={() => { setParentId(null); setReplyUser(null); }} className="p-1 hover:bg-muted rounded-full">
                            <X className="w-3.5 h-3.5 text-muted-foreground" />
                        </button>
                    </div>
                )}

                {editingNoteId && (
                    <div className="mb-3 px-3 py-2 bg-amber-500/5 border border-amber-500/10 rounded-xl flex items-center justify-between animate-in slide-in-from-bottom-2">
                        <div className="flex items-center gap-2 text-xs">
                            <Edit2 className="w-3.5 h-3.5 text-amber-500" />
                            <span className="font-bold text-amber-500">正在修改笔记</span>
                        </div>
                        <button onClick={() => { setEditingNoteId(null); setContent(''); setImages([]); }} className="p-1 hover:bg-muted rounded-full">
                            <X className="w-3.5 h-3.5 text-muted-foreground" />
                        </button>
                    </div>
                )}

                <div className="relative bg-muted rounded-2xl border border-border/50 focus-within:border-primary/30 transition-all">
                    <textarea
                        ref={editorRef}
                        value={content}
                        onChange={(e) => setContent(e.target.value)}
                        placeholder={parentId ? "友善回复，传播知识..." : "我的解题思路、难点笔记..."}
                        className="w-full bg-transparent p-4 text-sm resize-none focus:outline-none min-h-[100px] max-h-[300px]"
                        maxLength={200}
                    />

                    {/* Image Preview */}
                    {images.length > 0 && (
                        <div className="flex gap-2 p-3 pt-0 overflow-x-auto">
                            {images.map((url, i) => (
                                <div key={i} className="relative w-16 h-16 rounded-lg overflow-hidden border border-border group/img">
                                    <img src={url.startsWith('http') ? url : `${api.defaults.baseURL?.replace('/api/v1', '')}${url}`} alt="Preview" className="w-full h-full object-cover" />
                                    <button
                                        onClick={() => setImages(prev => prev.filter((_, idx) => idx !== i))}
                                        className="absolute top-1 right-1 p-0.5 bg-black/60 text-white rounded-full opacity-0 group-hover/img:opacity-100 transition-opacity"
                                    >
                                        <X className="w-3 h-3" />
                                    </button>
                                </div>
                            ))}
                        </div>
                    )}

                    <div className="flex items-center justify-between px-4 py-2 bg-muted/50 rounded-b-2xl border-t border-border/10">
                        <div className="flex items-center gap-2">
                            <input type="file" id="comment-image" className="hidden" accept="image/*" onChange={handleUpload} />
                            <label htmlFor="comment-image" className="p-2 hover:bg-card rounded-lg text-muted-foreground hover:text-primary transition-all cursor-pointer">
                                <ImageIcon className="w-5 h-5" />
                            </label>

                            <div className="relative">
                                <button
                                    onClick={() => setShowEmoji(!showEmoji)}
                                    className={cn(
                                        "p-2 rounded-lg transition-all",
                                        showEmoji ? "bg-card text-primary" : "text-muted-foreground hover:text-primary hover:bg-card"
                                    )}
                                >
                                    <Smile className="w-5 h-5" />
                                </button>

                                <AnimatePresence>
                                    {showEmoji && (
                                        <motion.div
                                            initial={{ opacity: 0, y: -10, scale: 0.95 }}
                                            animate={{ opacity: 1, y: 0, scale: 1 }}
                                            exit={{ opacity: 0, y: -10, scale: 0.95 }}
                                            className="absolute bottom-full left-0 mb-2 w-72 bg-card border border-border rounded-2xl shadow-2xl p-3 z-50"
                                        >
                                            <div className="grid grid-cols-8 gap-1 h-48 overflow-y-auto custom-scrollbar">
                                                {EMOJI_LIST.map((e, idx) => (
                                                    <button
                                                        key={idx}
                                                        onClick={() => { setContent(prev => prev + e); setShowEmoji(false); }}
                                                        className="text-xl p-1 hover:bg-muted rounded-lg transition-colors"
                                                    >
                                                        {e}
                                                    </button>
                                                ))}
                                            </div>
                                        </motion.div>
                                    )}
                                </AnimatePresence>
                            </div>

                            <button
                                onClick={() => setIsPublic(!isPublic)}
                                className={cn(
                                    "flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-bold transition-all",
                                    isPublic ? "text-primary hover:bg-primary/5" : "text-amber-500 hover:bg-amber-500/5"
                                )}
                            >
                                {isPublic ? <Globe className="w-3.5 h-3.5" /> : <Lock className="w-3.5 h-3.5" />}
                                {isPublic ? '公开发布' : '私密笔记'}
                            </button>
                        </div>

                        <div className="flex items-center gap-3">
                            <span className={cn(
                                "text-[10px] font-mono font-bold",
                                content.length > 180 ? "text-rose-500" : "text-muted-foreground/50"
                            )}>
                                {content.length}/200
                            </span>
                            <button
                                onClick={handleSubmit}
                                disabled={isSubmitting || (!content.trim() && images.length === 0)}
                                className="flex items-center gap-2 px-5 py-2 bg-primary disabled:bg-primary/50 text-primary-foreground font-black rounded-xl shadow-lg shadow-primary/20 hover:scale-105 active:scale-95 transition-all"
                            >
                                {isSubmitting ? <Loader2 className="w-4 h-4 animate-spin" /> : <Send className="w-4 h-4" />}
                                {editingNoteId ? '更新' : '发布'}
                            </button>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
};
