
import React, { useState, useEffect, useRef } from 'react';
import { api } from '../services/api';
import {
    FileText, Folder, Plus, Edit3, Trash2, Save, Upload,
    CheckCircle2, XCircle, ChevronLeft, Eye, Layout
} from 'lucide-react';
import { WikiCategory } from '../types';

interface WikiManagerProps {
    onBack?: () => void; // Optional if needed
}

export const WikiManager: React.FC<WikiManagerProps> = () => {
    const [categories, setCategories] = useState<WikiCategory[]>([]);
    const [view, setView] = useState<'list' | 'editor'>('list');
    const [selectedCategory, setSelectedCategory] = useState<WikiCategory | null>(null);
    const [editingArticle, setEditingArticle] = useState<any>(null);
    const [loading, setLoading] = useState(false);

    // Editor State
    const [editorTitle, setEditorTitle] = useState('');
    const [editorContent, setEditorContent] = useState('');
    const [editorCategoryId, setEditorCategoryId] = useState('');
    const [editorStatus, setEditorStatus] = useState('published');

    // File Upload Ref
    const fileInputRef = useRef<HTMLInputElement>(null);

    useEffect(() => {
        loadCategories();
    }, []);

    const loadCategories = async () => {
        try {
            setLoading(true);
            const data = await api.wiki.getCategories();
            setCategories(data);
            if (data.length > 0 && !selectedCategory) {
                // Optional: Auto-select first category? No, let user choose.
            }
        } catch (err) {
            console.error('Failed to load categories', err);
        } finally {
            setLoading(false);
        }
    };

    const handleCreateArticle = () => {
        setEditingArticle(null);
        setEditorTitle('');
        setEditorContent('');
        setEditorStatus('published');
        setEditorCategoryId(selectedCategory?.id || (categories[0]?.id || ''));
        setView('editor');
    };

    const handleEditArticle = async (article: any) => {
        setLoading(true);
        try {
            // Fetch full content
            const fullArticle = await api.wiki.getArticle(article.id);
            setEditingArticle(fullArticle);
            setEditorTitle(fullArticle.title);
            setEditorContent(fullArticle.content);
            setEditorCategoryId(fullArticle.category?.id || selectedCategory?.id);
            setEditorStatus(fullArticle.status || 'published');
            setView('editor');
        } catch (err) {
            alert('Failed to load article details');
        } finally {
            setLoading(false);
        }
    };

    const handleSaveArticle = async () => {
        if (!editorTitle || !editorContent || !editorCategoryId) {
            alert('请填写标题、内容并选择分类');
            return;
        }

        try {
            const payload = {
                title: editorTitle,
                content: editorContent,
                categoryId: editorCategoryId,
                status: editorStatus,
                excerpt: editorContent.slice(0, 100) + '...', // Simple excerpt
                readTime: Math.ceil(editorContent.length / 500) + ' min',
            };

            if (editingArticle) {
                await api.wiki.updateArticle(editingArticle.id, payload);
            } else {
                await api.wiki.createArticle(payload);
            }

            // Reload and go back
            await loadCategories();
            setView('list');
        } catch (err) {
            alert('保存失败');
            console.error(err);
        }
    };

    const handleDeleteArticle = async (id: string, e: React.MouseEvent) => {
        e.stopPropagation();
        if (!confirm('确定删除此文章？')) return;
        try {
            await api.wiki.deleteArticle(id);
            loadCategories();
        } catch (err) {
            alert('删除失败');
        }
    };

    const handleFileUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
        const file = e.target.files?.[0];
        if (!file) return;

        try {
            setLoading(true);
            const res = await api.wiki.uploadFile(file);
            setEditorContent(res.content);
            if (!editorTitle) {
                setEditorTitle(res.filename.replace('.md', ''));
            }
        } catch (err) {
            alert('文件上传解析失败');
        } finally {
            setLoading(false);
            if (fileInputRef.current) fileInputRef.current.value = '';
        }
    };

    // Category Management Locals
    const [isCatModalOpen, setIsCatModalOpen] = useState(false);
    const [catForm, setCatForm] = useState({ title: '', description: '', color: 'indigo', iconName: 'Book' });

    const handleSaveCategory = async () => {
        try {
            await api.wiki.createCategory(catForm);
            setIsCatModalOpen(false);
            loadCategories();
        } catch (err) {
            alert('创建分类失败');
        }
    };

    if (view === 'editor') {
        return (
            <div className="bg-white rounded-2xl shadow-sm border border-slate-200 overflow-hidden flex flex-col h-[calc(100vh-140px)]">
                {/* Editor Header */}
                <div className="p-4 border-b border-slate-200 flex justify-between items-center bg-slate-50">
                    <div className="flex items-center gap-4">
                        <button onClick={() => setView('list')} className="p-2 hover:bg-white rounded-lg transition-colors border border-transparent hover:border-slate-200 text-slate-500">
                            <ChevronLeft size={20} />
                        </button>
                        <h3 className="font-bold text-slate-800">{editingArticle ? '编辑文章' : '新建文章'}</h3>
                    </div>
                    <div className="flex items-center gap-3">
                        <input
                            type="file"
                            accept=".md"
                            ref={fileInputRef}
                            className="hidden"
                            onChange={handleFileUpload}
                        />
                        <button
                            onClick={() => fileInputRef.current?.click()}
                            className="flex items-center gap-2 px-3 py-1.5 rounded-lg border border-slate-300 text-slate-600 text-sm hover:bg-slate-50"
                        >
                            <Upload size={16} /> 导入 MD
                        </button>
                        <button
                            onClick={handleSaveArticle}
                            className="flex items-center gap-2 px-4 py-2 rounded-lg bg-indigo-600 text-white font-bold text-sm hover:bg-indigo-700 shadow-sm"
                        >
                            <Save size={16} /> 保存发布
                        </button>
                    </div>
                </div>

                {/* Editor Body */}
                <div className="flex-1 overflow-hidden flex">
                    {/* Settings Pane */}
                    <div className="w-80 border-r border-slate-200 p-6 bg-slate-50 space-y-6 overflow-y-auto">
                        <div>
                            <label className="block text-xs font-bold text-slate-500 uppercase mb-2">文章标题</label>
                            <input
                                type="text"
                                value={editorTitle}
                                onChange={e => setEditorTitle(e.target.value)}
                                className="w-full px-3 py-2 rounded-lg border border-slate-300 focus:outline-none focus:border-indigo-500"
                                placeholder="输入标题..."
                            />
                        </div>

                        <div>
                            <label className="block text-xs font-bold text-slate-500 uppercase mb-2">所属分类</label>
                            <select
                                value={editorCategoryId}
                                onChange={e => setEditorCategoryId(e.target.value)}
                                className="w-full px-3 py-2 rounded-lg border border-slate-300 focus:outline-none focus:border-indigo-500"
                            >
                                <option value="" disabled>选择分类</option>
                                {categories.map(cat => (
                                    <option key={cat.id} value={cat.id}>{cat.title}</option>
                                ))}
                            </select>
                        </div>

                        <div>
                            <label className="block text-xs font-bold text-slate-500 uppercase mb-2">发布状态</label>
                            <select
                                value={editorStatus}
                                onChange={e => setEditorStatus(e.target.value)}
                                className="w-full px-3 py-2 rounded-lg border border-slate-300 focus:outline-none focus:border-indigo-500"
                            >
                                <option value="draft">草稿</option>
                                <option value="published">发布</option>
                            </select>
                        </div>

                        <div className="pt-6 border-t border-slate-200">
                            <p className="text-xs text-slate-400 leading-relaxed">
                                提示：支持标准 Markdown 语法。
                                上传 .md 文件可快速导入内容。
                            </p>
                        </div>
                    </div>

                    {/* Markdown Input Area */}
                    <div className="flex-1 flex flex-col">
                        <textarea
                            value={editorContent}
                            onChange={e => setEditorContent(e.target.value)}
                            className="flex-1 w-full p-6 focus:outline-none resize-none font-mono text-sm leading-relaxed text-slate-800"
                            placeholder="# 开始编写..."
                        />
                    </div>
                </div>
            </div>
        );
    }

    return (
        <div className="flex bg-white rounded-2xl shadow-sm border border-slate-200 overflow-hidden h-[calc(100vh-140px)]">
            {/* Sidebar: Categories */}
            <div className="w-64 bg-slate-50 border-r border-slate-200 flex flex-col">
                <div className="p-4 border-b border-slate-200 flex justify-between items-center">
                    <h3 className="font-bold text-slate-700">知识分类</h3>
                    <button onClick={() => setIsCatModalOpen(true)} className="p-1 hover:bg-slate-200 rounded text-slate-500"><Plus size={18} /></button>
                </div>
                <div className="flex-1 overflow-y-auto p-2 space-y-1">
                    <button
                        onClick={() => setSelectedCategory(null)}
                        className={`w-full text-left px-3 py-2 rounded-lg text-sm font-medium flex items-center gap-2 ${!selectedCategory ? 'bg-indigo-50 text-indigo-700' : 'text-slate-600 hover:bg-slate-100'}`}
                    >
                        <Layout size={16} /> 全部文章
                    </button>
                    {categories.map(cat => (
                        <button
                            key={cat.id}
                            onClick={() => setSelectedCategory(cat)}
                            className={`w-full text-left px-3 py-2 rounded-lg text-sm font-medium flex items-center gap-2 ${selectedCategory?.id === cat.id ? 'bg-indigo-50 text-indigo-700' : 'text-slate-600 hover:bg-slate-100'}`}
                        >
                            <span className={`w-2 h-2 rounded-full bg-${cat.color}-500`}></span>
                            {cat.title}
                        </button>
                    ))}
                </div>
            </div>

            {/* Main: Article List */}
            <div className="flex-1 flex flex-col">
                <div className="p-6 border-b border-slate-200 flex justify-between items-center">
                    <div>
                        <h2 className="text-xl font-bold text-slate-800">{selectedCategory ? selectedCategory.title : '所有文章'}</h2>
                        <p className="text-sm text-slate-500">{selectedCategory ? selectedCategory.description : '管理所有知识库内容'}</p>
                    </div>
                    <button
                        onClick={handleCreateArticle}
                        className="bg-indigo-600 text-white px-4 py-2 rounded-lg text-sm font-bold flex items-center gap-2 hover:bg-indigo-700"
                    >
                        <Plus size={16} /> 新建文章
                    </button>
                </div>

                <div className="flex-1 overflow-y-auto p-6">
                    <div className="grid grid-cols-1 gap-4">
                        {loading ? <p className="text-slate-500">加载中...</p> :
                            (selectedCategory ? selectedCategory.articles : categories.flatMap(c => c.articles)).map((article: any) => (
                                <div key={article.id} className="bg-white p-4 rounded-xl border border-slate-200 hover:border-indigo-300 transition-all shadow-sm flex items-center justify-between group">
                                    <div className="flex items-center gap-4">
                                        <div className="w-10 h-10 rounded-lg bg-indigo-50 flex items-center justify-center text-indigo-600">
                                            <FileText size={20} />
                                        </div>
                                        <div>
                                            <h4 className="font-bold text-slate-800 mb-1">{article.title}</h4>
                                            <div className="flex items-center gap-3 text-xs text-slate-500">
                                                <span className="bg-slate-100 px-2 py-0.5 rounded text-slate-600">{article.category?.title || '未分类'}</span>
                                                <span>{new Date(article.date || Date.now()).toLocaleDateString()}</span>
                                                <span className={`${article.status === 'published' ? 'text-emerald-600' : 'text-amber-600'}`}>
                                                    {article.status === 'published' ? '已发布' : '草稿'}
                                                </span>
                                            </div>
                                        </div>
                                    </div>
                                    <div className="flex gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
                                        <button onClick={() => handleEditArticle(article)} className="p-2 text-slate-400 hover:text-indigo-600 hover:bg-indigo-50 rounded-lg">
                                            <Edit3 size={18} />
                                        </button>
                                        <button onClick={(e) => handleDeleteArticle(article.id, e)} className="p-2 text-slate-400 hover:text-rose-600 hover:bg-rose-50 rounded-lg">
                                            <Trash2 size={18} />
                                        </button>
                                    </div>
                                </div>
                            ))}
                        {!loading && (selectedCategory ? selectedCategory.articles : categories.flatMap(c => c.articles)).length === 0 && (
                            <div className="text-center py-12 text-slate-400">
                                <Folder size={48} className="mx-auto mb-4 opacity-50" />
                                <p>暂无文章，点击右上角新建</p>
                            </div>
                        )}
                    </div>
                </div>
            </div>

            {/* Category Modal - Simplified */}
            {isCatModalOpen && (
                <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
                    <div className="bg-white p-6 rounded-2xl w-96 max-w-full">
                        <h3 className="font-bold text-lg mb-4">新建分类</h3>
                        <div className="space-y-4">
                            <input
                                className="w-full border border-slate-300 rounded-lg px-4 py-2"
                                placeholder="分类名称"
                                value={catForm.title} onChange={e => setCatForm({ ...catForm, title: e.target.value })}
                            />
                            <input
                                className="w-full border border-slate-300 rounded-lg px-4 py-2"
                                placeholder="描述"
                                value={catForm.description} onChange={e => setCatForm({ ...catForm, description: e.target.value })}
                            />
                            <select
                                className="w-full border border-slate-300 rounded-lg px-4 py-2"
                                value={catForm.color} onChange={e => setCatForm({ ...catForm, color: e.target.value })}
                            >
                                <option value="indigo">Indigo</option>
                                <option value="blue">Blue</option>
                                <option value="emerald">Emerald</option>
                                <option value="amber">Amber</option>
                                <option value="rose">Rose</option>
                            </select>
                            <div className="flex gap-3 mt-6">
                                <button onClick={() => setIsCatModalOpen(false)} className="flex-1 py-2 text-slate-500 font-bold">取消</button>
                                <button onClick={handleSaveCategory} className="flex-1 py-2 bg-indigo-600 text-white rounded-lg font-bold">创建</button>
                            </div>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
};
