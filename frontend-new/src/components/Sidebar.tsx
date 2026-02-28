import React, { useState, useEffect } from 'react';
import { motion } from 'framer-motion';
import {
    LayoutDashboard,
    BookOpenCheck,
    FileWarning,
    Star,
    BookMarked,
    User,
    Settings,
    LogOut
} from 'lucide-react';
import { cn } from '../lib/utils';

export interface SidebarProps {
    currentPath: string;
}

const navItems = [
    { name: '仪表盘', path: '/', icon: LayoutDashboard },
    { name: '题库中心', path: '/quiz', icon: BookOpenCheck },
    { name: '错题本', path: '/mistakes', icon: FileWarning },
    { name: '收藏夹', path: '/favorites', icon: Star },
    { name: '我的笔记', path: '/notes', icon: BookMarked },
];

const bottomItems = [
    { name: '个人中心', path: '/profile', icon: User },
    { name: '系统设置', path: '/settings', icon: Settings },
];

export function Sidebar({ currentPath }: SidebarProps) {
    const [user, setUser] = useState({
        nickname: '',
        username: '学霸',
        avatar: ''
    });

    useEffect(() => {
        setUser({
            nickname: localStorage.getItem('nickname') || '',
            username: localStorage.getItem('username') || '学霸',
            avatar: localStorage.getItem('avatar') || ''
        });
    }, []);

    const handleLogout = () => {
        localStorage.clear();
        window.location.href = '/login';
    };

    const getAvatar = (path: string) => {
        if (!path) return `https://api.dicebear.com/7.x/notionists/svg?seed=${user.username}`;
        return path.startsWith('http') ? path : `http://localhost:8080${path}`;
    };
    return (
        <aside className="hidden md:flex flex-col w-[260px] h-screen bg-card border-r shadow-sm">
            {/* Brand */}
            <div className="h-16 flex items-center px-6 border-b border-border/50">
                <div className="w-8 h-8 rounded-lg bg-primary flex items-center justify-center mr-3 shadow-md shadow-primary/20">
                    <BookOpenCheck className="w-5 h-5 text-white" />
                </div>
                <span className="font-bold text-lg tracking-tight bg-gradient-to-r from-primary to-blue-400 bg-clip-text text-transparent drop-shadow-sm">Med Platform</span>
            </div>

            {/* Nav items */}
            <nav className="flex-1 overflow-y-auto py-6 px-4 space-y-1">
                {navItems.map((item) => {
                    const isActive = currentPath === item.path;
                    return (
                        <a key={item.path} href={item.path} className="block relative">
                            {isActive && (
                                <motion.div
                                    layoutId="active-indicator"
                                    className="absolute inset-0 bg-primary/10 rounded-xl"
                                    initial={false}
                                    transition={{ type: "spring", stiffness: 300, damping: 30 }}
                                />
                            )}
                            <div
                                className={cn(
                                    "relative flex items-center gap-3 px-4 py-3 rounded-xl transition-colors duration-200",
                                    isActive ? "text-primary font-medium" : "text-muted-foreground hover:bg-muted/50 hover:text-foreground"
                                )}
                            >
                                <item.icon className={cn("w-5 h-5", isActive && "text-primary drop-shadow-sm")} />
                                <span>{item.name}</span>
                            </div>
                        </a>
                    );
                })}
            </nav>

            {/* Bottom Profile / Settings */}
            <div className="p-4 border-t border-border/50 space-y-1">
                {bottomItems.map((item) => {
                    const isActive = currentPath === item.path;
                    return (
                        <a key={item.path} href={item.path} className="block relative">
                            {isActive && (
                                <motion.div
                                    layoutId="active-indicator"
                                    className="absolute inset-0 bg-primary/10 rounded-xl"
                                    initial={false}
                                    transition={{ type: "spring", stiffness: 300, damping: 30 }}
                                />
                            )}
                            <div
                                className={cn(
                                    "relative flex items-center gap-3 px-4 py-3 rounded-xl transition-colors duration-200",
                                    isActive ? "text-primary font-medium" : "text-muted-foreground hover:bg-muted/50 hover:text-foreground"
                                )}
                            >
                                <item.icon className="w-5 h-5" />
                                <span>{item.name}</span>
                            </div>
                        </a>
                    );
                })}

                {/* User preview */}
                <div
                    onClick={() => window.location.href = '/profile'}
                    className="mt-4 flex items-center gap-3 px-3 py-2 bg-muted/30 rounded-xl border border-border/50 cursor-pointer hover:bg-muted/50 transition-colors"
                >
                    <img src={getAvatar(user.avatar)} alt="User" className="w-10 h-10 rounded-full border border-border/50 bg-background shadow-sm" />
                    <div className="flex-1 overflow-hidden flex flex-col">
                        <span className="text-sm font-medium text-foreground truncate">{user.nickname || user.username}</span>
                        <span className="text-xs text-muted-foreground truncate">内科主治医师</span>
                    </div>
                    <button onClick={(e) => { e.stopPropagation(); handleLogout(); }}>
                        <LogOut className="w-4 h-4 text-muted-foreground hover:text-destructive transition-colors" />
                    </button>
                </div>
            </div>
        </aside>
    );
}
