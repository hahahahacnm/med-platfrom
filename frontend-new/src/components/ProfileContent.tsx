import React, { useState, useEffect } from 'react';
import { motion } from 'framer-motion';
import { User, Mail, Shield, Award, Settings, LogOut, Camera } from 'lucide-react';
import { cn } from '../lib/utils';
import api from '../lib/api';

export function ProfileContent() {
    const [user, setUser] = useState<any>(null);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        const fetchProfile = async () => {
            try {
                const res: any = await api.get('/user/profile');
                if (res.data) {
                    setUser(res.data);
                    localStorage.setItem('user', JSON.stringify(res.data));
                }
            } catch (err) {
                console.error('Failed to fetch profile', err);
            } finally {
                setLoading(false);
            }
        };
        fetchProfile();
    }, []);

    const logout = () => {
        localStorage.removeItem('token');
        localStorage.removeItem('user');
        window.location.href = '/login';
    };

    if (loading) return <div className="p-10 animate-pulse bg-muted h-96 rounded-3xl" />;

    return (
        <div className="p-6 lg:p-10 max-w-4xl mx-auto space-y-10">
            {/* Profile Hero */}
            <div className="relative bg-card rounded-[40px] border border-border/50 shadow-xl shadow-indigo-500/5 p-10 overflow-hidden group">
                <div className="absolute top-0 right-0 w-64 h-64 bg-primary/5 rounded-full -translate-y-1/2 translate-x-1/2 blur-3xl group-hover:bg-primary/10 transition-colors duration-700" />

                <div className="relative z-10 flex flex-col items-center text-center">
                    <div className="relative mb-6">
                        <div className="w-32 h-32 rounded-[40px] bg-gradient-to-br from-primary to-indigo-600 p-1 group-hover:scale-105 transition-transform duration-500">
                            <div className="w-full h-full rounded-[38px] bg-card p-1 overflow-hidden">
                                <img
                                    src={user?.avatar?.startsWith('http') ? user.avatar : `http://localhost:8080${user?.avatar || '/uploads/default.png'}`}
                                    className="w-full h-full object-cover rounded-[36px]"
                                />
                            </div>
                        </div>
                        <button className="absolute bottom-0 right-0 p-3 bg-card border border-border shadow-lg rounded-2xl text-primary hover:scale-110 active:scale-95 transition-all">
                            <Camera className="w-5 h-5" />
                        </button>
                    </div>

                    <h1 className="text-3xl font-black text-foreground tracking-tight mb-2">{user?.nickname || user?.username}</h1>
                    <p className="text-muted-foreground font-mono text-sm tracking-widest uppercase mb-8">@{user?.username}</p>

                    <div className="flex flex-wrap justify-center gap-4">
                        <div className="px-6 py-2 bg-primary/5 border border-primary/10 rounded-2xl flex items-center gap-2">
                            <Award className="w-4 h-4 text-primary" />
                            <span className="text-sm font-bold text-primary">学习积分: {user?.points || 0}</span>
                        </div>
                        <div className="px-6 py-2 bg-emerald-500/5 border border-emerald-500/10 rounded-2xl flex items-center gap-2">
                            <Shield className="w-4 h-4 text-emerald-500" />
                            <span className="text-sm font-bold text-emerald-500">{user?.role === 'admin' ? '超级管理员' : '正式会员'}</span>
                        </div>
                    </div>
                </div>
            </div>

            {/* Account Details */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div className="bg-card p-8 rounded-[32px] border border-border/50 shadow-sm space-y-6">
                    <h2 className="text-lg font-black text-foreground uppercase tracking-widest flex items-center gap-3">
                        <Settings className="w-5 h-5 text-muted-foreground" />
                        账号设置
                    </h2>

                    <div className="space-y-4">
                        <div className="p-4 rounded-2xl bg-muted/30 border border-border/20 flex items-center justify-between">
                            <div className="flex items-center gap-3">
                                <Mail className="w-5 h-5 text-muted-foreground" />
                                <div>
                                    <p className="text-[10px] font-black text-muted-foreground uppercase opacity-60">绑定邮箱</p>
                                    <p className="text-sm font-bold">{user?.email || '未绑定'}</p>
                                </div>
                            </div>
                            <button className="text-xs font-black text-primary uppercase">修改</button>
                        </div>

                        <div className="p-4 rounded-2xl bg-muted/30 border border-border/20 flex items-center justify-between">
                            <div className="flex items-center gap-3">
                                <Shield className="w-5 h-5 text-muted-foreground" />
                                <div>
                                    <p className="text-[10px] font-black text-muted-foreground uppercase opacity-60">账号状态</p>
                                    <p className="text-sm font-bold">正常使用中</p>
                                </div>
                            </div>
                            <span className="w-2 h-2 bg-emerald-500 rounded-full animate-pulse" />
                        </div>
                    </div>
                </div>

                <div className="bg-card p-8 rounded-[32px] border border-border/50 shadow-sm flex flex-col justify-between">
                    <div>
                        <h2 className="text-lg font-black text-foreground uppercase tracking-widest flex items-center gap-3 mb-6">
                            <User className="w-5 h-5 text-muted-foreground" />
                            更多操作
                        </h2>
                        <div className="space-y-3">
                            <button className="w-full p-4 rounded-2xl border border-border/50 hover:bg-muted font-bold text-sm text-left transition-all flex items-center justify-between group">
                                个人资料编辑
                                <Settings className="w-4 h-4 opacity-0 group-hover:opacity-100 transition-opacity" />
                            </button>
                            <button className="w-full p-4 rounded-2xl border border-border/50 hover:bg-muted font-bold text-sm text-left transition-all">安全隐私</button>
                        </div>
                    </div>

                    <button
                        onClick={logout}
                        className="mt-8 w-full p-4 bg-rose-500/5 text-rose-500 hover:bg-rose-500 hover:text-white rounded-2xl font-black text-sm uppercase tracking-widest transition-all flex items-center justify-center gap-2"
                    >
                        <LogOut className="w-4 h-4" />
                        退出系统登录
                    </button>
                </div>
            </div>
        </div>
    );
}
