import React from 'react';
import { motion } from 'framer-motion';
import {
    LayoutDashboard,
    BookOpenCheck,
    FileWarning,
    Star,
    User
} from 'lucide-react';
import { cn } from '../lib/utils';

export function MobileNav({ currentPath }: { currentPath: string }) {
    const items = [
        { name: '仪表盘', path: '/', icon: LayoutDashboard },
        { name: '题库', path: '/quiz', icon: BookOpenCheck },
        { name: '错题', path: '/mistakes', icon: FileWarning },
        { name: '收藏', path: '/favorites', icon: Star },
        { name: '我的', path: '/profile', icon: User },
    ];

    return (
        <nav className="md:hidden fixed bottom-0 left-0 right-0 bg-background/80 backdrop-blur-xl border-t border-border/50 px-4 pb-6 pt-3 flex justify-around items-center z-50">
            {items.map((item) => {
                const isActive = currentPath === item.path;
                return (
                    <a
                        key={item.path}
                        href={item.path}
                        className={cn(
                            "flex flex-col items-center gap-1 transition-all duration-300",
                            isActive ? "text-primary scale-110" : "text-muted-foreground"
                        )}
                    >
                        <div className={cn(
                            "p-2 rounded-xl transition-all",
                            isActive ? "bg-primary/10" : ""
                        )}>
                            <item.icon className="w-5 h-5" />
                        </div>
                        <span className="text-[10px] font-bold uppercase tracking-widest">{item.name}</span>
                        {isActive && (
                            <motion.div
                                layoutId="mobile-nav-indicator"
                                className="w-1 h-1 bg-primary rounded-full"
                            />
                        )}
                    </a>
                );
            })}
        </nav>
    );
}
