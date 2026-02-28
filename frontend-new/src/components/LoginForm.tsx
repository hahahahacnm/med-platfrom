import React, { useState } from 'react';
import { motion } from 'framer-motion';
import { Mail, Lock, ArrowRight, BookOpenCheck, Github, Chrome } from 'lucide-react';

import api from '../lib/api';

export function LoginForm() {
    const [isLoading, setIsLoading] = useState(false);
    const [errorMsg, setErrorMsg] = useState('');
    const [formData, setFormData] = useState({ username: '', password: '' });

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setIsLoading(true);
        setErrorMsg('');
        try {
            const res: any = await api.post('/auth/login', formData);
            if (res.token) {
                localStorage.setItem('token', res.token);
                localStorage.setItem('username', res.username || formData.username);
                localStorage.setItem('id', String(res.id || ''));
                localStorage.setItem('nickname', res.nickname || '');
                localStorage.setItem('avatar', res.avatar || '');
                window.location.href = "/";
            } else {
                setErrorMsg('登录失败：返回数据格式错误');
            }
        } catch (err: any) {
            console.error('Login Error:', err);
            setErrorMsg(err.response?.data?.message || err.response?.data?.error || '登录失败，请检查账号密码');
        } finally {
            setIsLoading(false);
        }
    };

    return (
        <motion.div
            initial={{ opacity: 0, scale: 0.95, y: 20 }}
            animate={{ opacity: 1, scale: 1, y: 0 }}
            transition={{ duration: 0.4, ease: "easeOut" }}
            className="w-full max-w-[420px] bg-card/80 backdrop-blur-xl border border-border shadow-[0_20px_50px_rgba(0,0,0,0.3)] rounded-[2.5rem] p-10 relative overflow-hidden"
        >
            {/* Decorative element */}
            <div className="absolute top-0 right-0 -mr-16 -mt-16 w-40 h-40 bg-primary/10 rounded-full blur-3xl pointer-events-none" />

            <div className="flex flex-col items-center text-center mb-10">
                <div className="w-16 h-16 bg-primary rounded-2xl flex items-center justify-center mb-6 shadow-xl shadow-primary/20 rotate-3 group hover:rotate-0 transition-transform duration-300">
                    <BookOpenCheck className="w-9 h-9 text-white" />
                </div>
                <h1 className="text-3xl font-black tracking-tight text-foreground bg-gradient-to-br from-foreground to-foreground/60 bg-clip-text text-transparent">欢迎回来</h1>
                <p className="text-muted-foreground mt-2 font-medium">继续您的医学晋升之路</p>
            </div>

            <form onSubmit={handleSubmit} className="space-y-5">
                {errorMsg && (
                    <motion.div initial={{ opacity: 0, scale: 0.9 }} animate={{ opacity: 1, scale: 1 }} className="bg-rose-500/10 text-rose-500 text-xs px-4 py-3 rounded-xl font-bold">
                        {errorMsg}
                    </motion.div>
                )}

                <div className="space-y-2">
                    <label className="text-xs font-bold uppercase tracking-widest text-muted-foreground ml-1">用户名</label>
                    <div className="relative group">
                        <Mail className="w-5 h-5 absolute left-4 top-1/2 -translate-y-1/2 text-muted-foreground group-focus-within:text-primary transition-colors" />
                        <input
                            required
                            type="text"
                            value={formData.username}
                            onChange={(e) => setFormData(prev => ({ ...prev, username: e.target.value }))}
                            placeholder="doctor@example.com"
                            className="w-full pl-12 pr-4 py-4 bg-muted/30 border border-transparent focus:bg-background focus:border-primary/30 rounded-2xl text-sm outline-none transition-all duration-300 shadow-inner"
                        />
                    </div>
                </div>

                <div className="space-y-2">
                    <div className="flex justify-between items-center px-1">
                        <label className="text-xs font-bold uppercase tracking-widest text-muted-foreground">密码</label>
                        <a href="#" className="text-[10px] font-bold text-primary hover:underline">忘记密码?</a>
                    </div>
                    <div className="relative group">
                        <Lock className="w-5 h-5 absolute left-4 top-1/2 -translate-y-1/2 text-muted-foreground group-focus-within:text-primary transition-colors" />
                        <input
                            required
                            type="password"
                            value={formData.password}
                            onChange={(e) => setFormData(prev => ({ ...prev, password: e.target.value }))}
                            placeholder="••••••••"
                            className="w-full pl-12 pr-4 py-4 bg-muted/30 border border-transparent focus:bg-background focus:border-primary/30 rounded-2xl text-sm outline-none transition-all duration-300 shadow-inner"
                        />
                    </div>
                </div>

                <button
                    disabled={isLoading}
                    className="w-full py-4 mt-4 bg-primary text-white font-bold rounded-2xl shadow-lg shadow-primary/20 hover:shadow-primary/40 hover:-translate-y-0.5 active:translate-y-0 transition-all duration-300 flex items-center justify-center gap-2 group disabled:opacity-70"
                >
                    {isLoading ? (
                        <div className="w-5 h-5 border-2 border-white/20 border-t-white rounded-full animate-spin" />
                    ) : (
                        <>
                            立即进入 <ArrowRight className="w-5 h-5 group-hover:translate-x-1 transition-transform" />
                        </>
                    )}
                </button>
            </form>

            <div className="mt-10 flex items-center gap-4">
                <div className="h-px bg-border flex-1 opacity-50" />
                <span className="text-[10px] uppercase font-black tracking-widest text-muted-foreground">快捷登录</span>
                <div className="h-px bg-border flex-1 opacity-50" />
            </div>

            <div className="grid grid-cols-2 gap-4 mt-8">
                <button className="flex items-center justify-center gap-2 py-3 bg-muted/30 border border-transparent hover:border-border rounded-xl transition-all hover:bg-muted/50">
                    <Github className="w-5 h-5" />
                    <span className="text-xs font-bold">GitHub</span>
                </button>
                <button className="flex items-center justify-center gap-2 py-3 bg-muted/30 border border-transparent hover:border-border rounded-xl transition-all hover:bg-muted/50">
                    <Chrome className="w-5 h-5" />
                    <span className="text-xs font-bold">Google</span>
                </button>
            </div>

            <p className="text-center mt-10 text-xs text-muted-foreground font-medium">
                还没有账号? <a href="/register" className="text-primary font-bold hover:underline">立即注册</a>
            </p>
        </motion.div>
    );
}
