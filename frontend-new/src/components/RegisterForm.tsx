import React, { useState } from 'react';
import { motion } from 'framer-motion';
import { Mail, Lock, User, ShieldCheck, ArrowRight, BookOpenCheck } from 'lucide-react';

import api from '../lib/api';

export function RegisterForm() {
    const [isLoading, setIsLoading] = useState(false);
    const [errorMsg, setErrorMsg] = useState('');
    const [formData, setFormData] = useState({
        username: '',
        password: '',
        nickname: '',
        major: ''
    });

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setIsLoading(true);
        setErrorMsg('');
        try {
            await api.post('/auth/register', formData);
            window.location.href = "/login";
        } catch (err: any) {
            setErrorMsg(err.response?.data?.message || err.response?.data?.error || '注册失败，请换个账号重试');
        } finally {
            setIsLoading(false);
        }
    };

    return (
        <motion.div
            initial={{ opacity: 0, scale: 0.95, y: 20 }}
            animate={{ opacity: 1, scale: 1, y: 0 }}
            transition={{ duration: 0.4, ease: "easeOut" }}
            className="w-full max-w-[480px] bg-card/80 backdrop-blur-xl border border-border shadow-[0_20px_50px_rgba(0,0,0,0.3)] rounded-[2.5rem] p-10 relative overflow-hidden"
        >
            <div className="flex flex-col items-center text-center mb-10">
                <div className="w-16 h-16 bg-primary rounded-2xl flex items-center justify-center mb-6 shadow-xl shadow-primary/20 rotate-3 group hover:rotate-0 transition-transform duration-300">
                    <BookOpenCheck className="w-9 h-9 text-white" />
                </div>
                <h1 className="text-3xl font-black tracking-tight text-foreground bg-gradient-to-br from-foreground to-foreground/60 bg-clip-text text-transparent">开启新旅程</h1>
                <p className="text-muted-foreground mt-2 font-medium">加入 1,000,000+ 医学同仁的学习社区</p>
            </div>

            <form onSubmit={handleSubmit} className="space-y-4">
                {errorMsg && (
                    <motion.div initial={{ opacity: 0, scale: 0.9 }} animate={{ opacity: 1, scale: 1 }} className="bg-rose-500/10 text-rose-500 text-xs px-4 py-3 rounded-xl font-bold text-center">
                        {errorMsg}
                    </motion.div>
                )}
                <div className="grid grid-cols-2 gap-4">
                    <div className="space-y-2">
                        <label className="text-xs font-bold uppercase tracking-widest text-muted-foreground ml-1">昵称</label>
                        <div className="relative group">
                            <User className="w-4 h-4 absolute left-4 top-1/2 -translate-y-1/2 text-muted-foreground group-focus-within:text-primary transition-colors" />
                            <input
                                required
                                type="text"
                                value={formData.nickname}
                                onChange={(e) => setFormData(prev => ({ ...prev, nickname: e.target.value }))}
                                placeholder="张三"
                                className="w-full pl-10 pr-4 py-3 bg-muted/30 border border-transparent focus:bg-background focus:border-primary/30 rounded-2xl text-sm outline-none transition-all duration-300 shadow-inner"
                            />
                        </div>
                    </div>
                    <div className="space-y-2">
                        <label className="text-xs font-bold uppercase tracking-widest text-muted-foreground ml-1">专业</label>
                        <div className="relative group">
                            <ShieldCheck className="w-4 h-4 absolute left-4 top-1/2 -translate-y-1/2 text-muted-foreground group-focus-within:text-primary transition-colors" />
                            <input
                                required
                                type="text"
                                value={formData.major}
                                onChange={(e) => setFormData(prev => ({ ...prev, major: e.target.value }))}
                                placeholder="内科学"
                                className="w-full pl-10 pr-4 py-3 bg-muted/30 border border-transparent focus:bg-background focus:border-primary/30 rounded-2xl text-sm outline-none transition-all duration-300 shadow-inner"
                            />
                        </div>
                    </div>
                </div>

                <div className="space-y-2">
                    <label className="text-xs font-bold uppercase tracking-widest text-muted-foreground ml-1">账号/邮箱</label>
                    <div className="relative group">
                        <Mail className="w-5 h-5 absolute left-4 top-1/2 -translate-y-1/2 text-muted-foreground group-focus-within:text-primary transition-colors" />
                        <input
                            required
                            type="text"
                            value={formData.username}
                            onChange={(e) => setFormData(prev => ({ ...prev, username: e.target.value }))}
                            placeholder="doctor"
                            className="w-full pl-12 pr-4 py-4 bg-muted/30 border border-transparent focus:bg-background focus:border-primary/30 rounded-2xl text-sm outline-none transition-all duration-300 shadow-inner"
                        />
                    </div>
                </div>

                <div className="space-y-2">
                    <label className="text-xs font-bold uppercase tracking-widest text-muted-foreground ml-1">设置密码</label>
                    <div className="relative group">
                        <Lock className="w-5 h-5 absolute left-4 top-1/2 -translate-y-1/2 text-muted-foreground group-focus-within:text-primary transition-colors" />
                        <input
                            required
                            type="password"
                            value={formData.password}
                            onChange={(e) => setFormData(prev => ({ ...prev, password: e.target.value }))}
                            placeholder="至少 8 位字符"
                            className="w-full pl-12 pr-4 py-4 bg-muted/30 border border-transparent focus:bg-background focus:border-primary/30 rounded-2xl text-sm outline-none transition-all duration-300 shadow-inner"
                        />
                    </div>
                </div>

                <div className="pt-2">
                    <label className="flex items-center gap-3 cursor-pointer group">
                        <input type="checkbox" required className="w-4 h-4 rounded border-border text-primary focus:ring-primary/20 transition-all cursor-pointer" />
                        <span className="text-[11px] text-muted-foreground font-medium">我同意 <a href="#" className="text-primary font-bold hover:underline">用户协议</a> 与 <a href="#" className="text-primary font-bold hover:underline">隐私政策</a></span>
                    </label>
                </div>

                <button
                    disabled={isLoading}
                    className="w-full py-4 mt-4 bg-primary text-white font-bold rounded-2xl shadow-lg shadow-primary/20 hover:shadow-primary/40 hover:-translate-y-0.5 active:translate-y-0 transition-all duration-300 flex items-center justify-center gap-2 group disabled:opacity-70"
                >
                    {isLoading ? (
                        <div className="w-5 h-5 border-2 border-white/20 border-t-white rounded-full animate-spin" />
                    ) : (
                        <>
                            创建账号 <ArrowRight className="w-5 h-5 group-hover:translate-x-1 transition-transform" />
                        </>
                    )}
                </button>
            </form>

            <p className="text-center mt-10 text-xs text-muted-foreground font-medium">
                已经有账号? <a href="/login" className="text-primary font-bold hover:underline">立即登录</a>
            </p>
        </motion.div>
    );
}
