import React, { useState } from 'react';
import { Menu, Search, Bell, X } from 'lucide-react';
import { motion, AnimatePresence } from 'framer-motion';
import type { SidebarProps } from './Sidebar';
import { ThemeToggle } from './ThemeToggle';

// Same nav arrays duplicated or shared, but for simplicity duplicated here
const navItems = [
    { name: '仪表盘', path: '/', icon: 'LayoutDashboard' },
    { name: '题库中心', path: '/quiz', icon: 'BookOpenCheck' },
    { name: '错题本', path: '/mistakes', icon: 'FileWarning' },
    { name: '收藏夹', path: '/favorites', icon: 'Star' },
    { name: '我的笔记', path: '/notes', icon: 'BookMarked' },
];

export function Topbar({ currentPath }: SidebarProps) {
    const [mobileMenuOpen, setMobileMenuOpen] = useState(false);

    return (
        <>
            <header className="h-16 flex-shrink-0 bg-card border-b border-border/50 flex items-center justify-between px-4 lg:px-8 z-20 shadow-sm sticky top-0">
                <div className="flex items-center gap-4">
                    <button
                        className="md:hidden p-2 text-muted-foreground hover:text-foreground rounded-lg hover:bg-muted/50 transition-colors"
                        onClick={() => setMobileMenuOpen(true)}
                    >
                        <Menu className="w-5 h-5" />
                    </button>

                    <div className="relative group hidden sm:block">
                        <Search className="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground group-focus-within:text-primary transition-colors" />
                        <input
                            type="text"
                            placeholder="搜索任何内容..."
                            className="pl-10 pr-4 py-2 w-[240px] focus:w-[320px] transition-all duration-300 bg-muted/40 border border-transparent focus:bg-background focus:border-primary/30 rounded-xl text-sm outline-none shadow-sm"
                        />
                    </div>
                </div>

                <div className="flex items-center gap-3">
                    <ThemeToggle />
                    <button className="relative p-2 text-muted-foreground hover:text-foreground rounded-full hover:bg-muted/50 transition-colors">
                        <Bell className="w-5 h-5" />
                        <span className="absolute top-1 right-1 w-2 h-2 bg-destructive rounded-full border-2 border-card"></span>
                    </button>
                    <div className="md:hidden">
                        <img src="https://api.dicebear.com/7.x/notionists/svg?seed=Felix" alt="User" className="w-8 h-8 rounded-full border border-border bg-background" />
                    </div>
                </div>
            </header>

            {/* Mobile Menu Drawer */}
            <AnimatePresence>
                {mobileMenuOpen && (
                    <>
                        <motion.div
                            initial={{ opacity: 0 }}
                            animate={{ opacity: 1 }}
                            exit={{ opacity: 0 }}
                            className="fixed inset-0 bg-background/80 backdrop-blur-sm z-40 md:hidden"
                            onClick={() => setMobileMenuOpen(false)}
                        />
                        <motion.div
                            initial={{ x: '-100%' }}
                            animate={{ x: 0 }}
                            exit={{ x: '-100%' }}
                            transition={{ type: "spring", damping: 25, stiffness: 200 }}
                            className="fixed top-0 left-0 bottom-0 w-[280px] bg-card z-50 shadow-2xl flex flex-col"
                        >
                            <div className="h-16 flex items-center justify-between px-6 border-b border-border/50">
                                <span className="font-bold text-lg text-primary">Med Platform</span>
                                <button onClick={() => setMobileMenuOpen(false)} className="p-2 -mr-2 text-muted-foreground hover:text-foreground">
                                    <X className="w-5 h-5" />
                                </button>
                            </div>
                            <nav className="flex-1 py-6 px-4 space-y-2 overflow-y-auto text-sm">
                                {navItems.map((item) => (
                                    <a
                                        key={item.path}
                                        href={item.path}
                                        className={`block px-4 py-3 rounded-xl transition-colors ${currentPath === item.path ? 'bg-primary/10 text-primary font-medium' : 'text-muted-foreground hover:bg-muted'}`}
                                    >
                                        {item.name}
                                    </a>
                                ))}
                            </nav>
                        </motion.div>
                    </>
                )}
            </AnimatePresence>
        </>
    );
}
