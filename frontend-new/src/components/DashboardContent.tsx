import React, { useState, useEffect, useCallback } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import {
    ChartSpline, Target, Trophy, FileBadge, ArrowUpRight,
    Bell, CheckCircle2, ChevronRight, Zap, GraduationCap,
    Flame, BookOpen, Star, AlertCircle, Sparkles
} from 'lucide-react';
import { cn } from '../lib/utils';
import api from '../lib/api';

const containerVariants = {
    hidden: { opacity: 0 },
    visible: {
        opacity: 1,
        transition: { staggerChildren: 0.1 }
    }
};

const itemVariants = {
    hidden: { y: 20, opacity: 0 },
    visible: {
        y: 0,
        opacity: 1,
        transition: { type: "spring", stiffness: 100 }
    }
};

export function DashboardContent() {
    const [loading, setLoading] = useState(true);
    const [statsData, setStatsData] = useState({
        total_count: 0,
        today_count: 0,
        accuracy: 0,
        consecutive_days: 0,
        activity_map: [] as any[],
        rank_list: [] as any[],
    });
    const [notifications, setNotifications] = useState<any[]>([]);
    const [unreadCount, setUnreadCount] = useState(0);
    const [nickname, setNickname] = useState('学霸');

    const fetchStats = useCallback(async () => {
        try {
            const res: any = await api.get('/stats');
            if (res.data) setStatsData(res.data);
        } catch (err) { console.error('Stats error:', err); }
    }, []);

    const fetchNotifications = useCallback(async () => {
        try {
            const res: any = await api.get('/notifications');
            if (res.data) {
                setNotifications(res.data || []);
                setUnreadCount(res.unread_count || 0);
            }
        } catch (err) { console.error('Notif error:', err); }
    }, []);

    useEffect(() => {
        const init = async () => {
            setLoading(true);
            const userNickname = localStorage.getItem('nickname') || localStorage.getItem('username') || '学霸';
            setNickname(userNickname);

            await Promise.all([fetchStats(), fetchNotifications()]);
            setLoading(false);
        };
        init();

        // WebSocket for real-time notifications
        const uid = localStorage.getItem('id');
        if (uid) {
            const wsUrl = `ws://${window.location.host.split(':')[0]}:8080/ws?uid=${uid}`;
            const ws = new WebSocket(wsUrl);
            ws.onmessage = (event) => {
                try {
                    const msg = JSON.parse(event.data);
                    if (msg.type === 'new_notification') {
                        setNotifications(prev => [msg.data, ...prev.slice(0, 14)]);
                        setUnreadCount(c => c + 1);
                    }
                } catch (e) { }
            };
            return () => ws.close();
        }
    }, [fetchStats, fetchNotifications]);

    const markAllAsRead = async () => {
        try {
            await api.put('/notifications/read-all');
            setNotifications(prev => prev.map(n => ({ ...n, is_read: true })));
            setUnreadCount(0);
        } catch (e) { }
    };

    const statsCards = [
        { title: "今日进度", value: statsData.today_count.toString(), unit: "题", icon: Target, color: "text-blue-500", bg: "bg-blue-500/10" },
        { title: "当前正确率", value: `${statsData.accuracy.toFixed(0)}`, unit: "%", icon: Trophy, color: "text-emerald-500", bg: "bg-emerald-500/10" },
        { title: "累计刷题", value: statsData.total_count.toString(), unit: "题", icon: FileBadge, color: "text-rose-500", bg: "bg-rose-500/10" },
        { title: "连续打卡", value: statsData.consecutive_days.toString(), unit: "天", icon: Flame, color: "text-orange-500", bg: "bg-orange-500/10" },
    ];

    const actionTiles = [
        { title: "科学刷题", desc: "章节知识点各个击破", icon: BookOpen, color: "bg-blue-600", path: "/quiz" },
        { title: "错题攻克", desc: "消除薄弱环节点对点", icon: Zap, color: "bg-rose-500", path: "/mistakes" },
        { title: "收藏中心", desc: "精选考点重点复习", icon: Star, color: "bg-amber-500", path: "/favorites" },
        { title: "学霸笔记", desc: "记录思维火花精华", icon: FileBadge, color: "bg-indigo-500", path: "/notes" },
    ];

    return (
        <div className="p-6 lg:p-10 pb-20 max-w-[1600px] mx-auto space-y-10">
            {/* Header / Banner */}
            <motion.div
                initial={{ opacity: 0, y: -20 }}
                animate={{ opacity: 1, y: 0 }}
                className="grid grid-cols-1 lg:grid-cols-4 gap-8"
            >
                <div className="lg:col-span-3 relative overflow-hidden rounded-[2.5rem] bg-gradient-to-br from-primary via-blue-500 to-indigo-600 p-10 sm:p-12 text-white shadow-2xl shadow-primary/20 group">
                    <div className="absolute -top-24 -right-24 w-96 h-96 bg-white/10 blur-3xl rounded-full" />
                    <div className="absolute bottom-0 left-0 p-8 opacity-10 pointer-events-none">
                        <GraduationCap size={200} />
                    </div>

                    <div className="relative z-10 flex flex-col sm:flex-row justify-between items-center gap-8 h-full">
                        <div className="space-y-6">
                            <div className="inline-flex items-center gap-2 px-4 py-1.5 bg-white/20 backdrop-blur-md rounded-full text-sm font-bold border border-white/30">
                                <Sparkles size={14} className="animate-pulse" />
                                <span>备考之路，唯有坚持</span>
                            </div>
                            <h1 className="text-4xl sm:text-5xl font-black tracking-tight leading-tight">
                                你好, {nickname}! <span className="inline-block animate-wave text-5xl">👋</span>
                            </h1>
                            <p className="text-white/80 text-lg max-w-xl font-medium leading-relaxed">
                                每一道错题都是进步的基石，每一次坚持都在通向成功。
                            </p>
                            <button
                                onClick={() => window.location.href = '/quiz'}
                                className="mt-4 px-8 py-4 bg-white text-primary text-lg font-black rounded-2xl shadow-lg shadow-black/10 hover:shadow-xl hover:-translate-y-1 active:scale-95 transition-all duration-300 flex items-center gap-3 group/btn"
                            >
                                立即开启今日挑战
                                <ChevronRight size={20} className="group-hover/btn:translate-x-1 transition-transform" />
                            </button>
                        </div>

                        <div className="flex flex-col gap-6 items-center">
                            <div className="bg-white/10 backdrop-blur-xl rounded-3xl p-8 border border-white/20 shadow-inner text-center min-w-[180px]">
                                <div className="text-white/60 font-bold text-xs uppercase tracking-widest mb-2">已连续刷题</div>
                                <div className="text-6xl font-black drop-shadow-md">{statsData.consecutive_days}<span className="text-lg font-normal opacity-60 ml-1">天</span></div>
                            </div>
                        </div>
                    </div>
                </div>

                <div className="bg-card rounded-[2.5rem] p-8 border border-border shadow-xl flex flex-col justify-between overflow-hidden relative group">
                    <div className="absolute top-0 right-0 w-32 h-32 bg-primary/5 rounded-full -mr-10 -mt-10" />
                    <div className="flex justify-between items-center mb-6">
                        <h3 className="text-lg font-black flex items-center gap-2">
                            <Bell className="text-primary" />
                            新动态
                        </h3>
                        {unreadCount > 0 && (
                            <span className="px-2 py-1 bg-destructive text-white text-[10px] font-black rounded-full animate-bounce">
                                {unreadCount}
                            </span>
                        )}
                    </div>
                    <div className="space-y-4 flex-1 overflow-y-auto pr-2 max-h-[200px] scrollbar-hide">
                        {loading ? (
                            Array(3).fill(0).map((_, i) => <div key={i} className="h-12 bg-muted rounded-xl animate-pulse" />)
                        ) : notifications.length > 0 ? (
                            notifications.map((n, i) => (
                                <div key={i} className={cn("p-4 rounded-2xl text-xs transition-all border border-transparent", n.is_read ? "bg-muted/30" : "bg-primary/5 border-primary/20")}>
                                    <p className="font-black text-foreground mb-1 line-clamp-1">{n.sender?.nickname || '系统通知'}</p>
                                    <p className="text-muted-foreground line-clamp-2 leading-relaxed">{n.content}</p>
                                </div>
                            ))
                        ) : (
                            <div className="h-full flex flex-col items-center justify-center opacity-30 gap-2">
                                <CheckCircle2 size={32} />
                                <span className="text-[10px] font-bold uppercase tracking-widest">暂无新消息</span>
                            </div>
                        )}
                    </div>
                    <button
                        onClick={markAllAsRead}
                        className="mt-6 w-full py-3 text-xs font-black text-primary hover:bg-primary/5 rounded-xl transition-colors border border-primary/10"
                    >
                        全部标记为已读
                    </button>
                </div>
            </motion.div>

            {/* Quick Actions */}
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
                {actionTiles.map((tile, i) => (
                    <motion.div
                        key={i}
                        variants={itemVariants as any}
                        onClick={() => window.location.href = tile.path}
                        className="group bg-card transition-all duration-300 p-6 rounded-3xl border border-border/50 hover:shadow-xl hover:shadow-primary/5 hover:border-primary/20 cursor-pointer flex items-center gap-6"
                    >
                        <div className={cn("w-14 h-14 rounded-2xl flex items-center justify-center text-white shadow-lg shrink-0 transition-transform group-hover:scale-110", tile.color)}>
                            <tile.icon size={28} />
                        </div>
                        <div className="space-y-1">
                            <h4 className="font-black text-foreground group-hover:text-primary transition-colors">{tile.title}</h4>
                            <p className="text-xs text-muted-foreground font-medium">{tile.desc}</p>
                        </div>
                        <div className="ml-auto opacity-0 group-hover:opacity-100 -translate-x-2 group-hover:translate-x-0 transition-all">
                            <ChevronRight className="text-primary" size={20} />
                        </div>
                    </motion.div>
                ))}
            </div>

            {/* Grid for Bottom Content */}
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
                {/* Stats and Heatmap */}
                <div className="lg:col-span-2 space-y-8">
                    {/* Stats */}
                    <motion.div
                        variants={containerVariants as any}
                        initial="hidden"
                        animate="visible"
                        className="grid grid-cols-2 md:grid-cols-4 gap-4"
                    >
                        {statsCards.map((stat, i) => (
                            <div key={i} className="bg-card rounded-3xl p-6 border border-border shadow-sm hover:shadow-md transition-shadow group">
                                <div className={cn("w-10 h-10 rounded-xl mb-4 flex items-center justify-center", stat.bg)}>
                                    <stat.icon size={20} className={stat.color} />
                                </div>
                                <div className="text-2xl font-black text-foreground tabular-nums group-hover:scale-105 transition-transform origin-left">{stat.value}<span className="text-xs font-bold text-muted-foreground ml-1 uppercase">{stat.unit}</span></div>
                                <div className="text-[10px] font-black text-muted-foreground uppercase tracking-widest mt-1">{stat.title}</div>
                            </div>
                        ))}
                    </motion.div>

                    {/* Heatmap Area */}
                    <div className="bg-card rounded-[2.5rem] border border-border shadow-xl p-10 overflow-hidden relative">
                        <div className="absolute top-0 left-0 w-full h-1 bg-gradient-to-r from-primary to-emerald-500" />
                        <h3 className="text-xl font-black mb-10 flex items-center gap-3">
                            <ChartSpline size={24} className="text-primary" />
                            学习活力追踪 <span className="text-xs font-bold text-muted-foreground ml-auto uppercase tracking-widest bg-muted px-3 py-1 rounded-full">最近 14 天</span>
                        </h3>
                        <div className="flex gap-1.5 md:gap-3 justify-between items-end h-[160px] group/heatmap">
                            {statsData.activity_map && statsData.activity_map.length > 0 ? (
                                statsData.activity_map.map((day: any, i: number) => {
                                    let levelColor = "bg-muted/40";
                                    if (day.level === 4) levelColor = "bg-primary shadow-[0_0_15px_rgba(59,130,246,0.3)]";
                                    else if (day.level === 3) levelColor = "bg-primary/70";
                                    else if (day.level === 2) levelColor = "bg-primary/40";
                                    else if (day.level === 1) levelColor = "bg-primary/20";

                                    const hPercentage = Math.max(15, (day.level * 20) + 5);

                                    return (
                                        <div key={i} className="flex flex-col items-center gap-3 flex-1 relative group/col">
                                            <div
                                                className={cn("w-full max-w-[28px] rounded-full transition-all duration-700 cursor-pointer origin-bottom hover:scale-110", levelColor)}
                                                style={{ height: `${hPercentage}%` }}
                                            />
                                            <span className="text-[10px] font-black text-muted-foreground font-mono transition-colors group-hover/col:text-primary">
                                                {day.date.split('-')[2]}
                                            </span>

                                            {/* Tooltip */}
                                            <div className="absolute bottom-full mb-3 left-1/2 -translate-x-1/2 opacity-0 group-hover/col:opacity-100 transition-all pointer-events-none z-30 translate-y-2 group-hover/col:translate-y-0">
                                                <div className="bg-foreground text-background text-[10px] font-black px-3 py-1.5 rounded-xl shadow-xl whitespace-nowrap">
                                                    {day.date} • {day.count} 题
                                                </div>
                                                <div className="w-2 h-2 bg-foreground rotate-45 mx-auto -mt-1 shadow-xl" />
                                            </div>
                                        </div>
                                    );
                                })
                            ) : (
                                <div className="w-full flex flex-col items-center justify-center opacity-20 gap-3">
                                    <AlertCircle size={40} />
                                    <span className="text-sm font-black uppercase tracking-widest">暂无活跃数据，快去刷题吧！</span>
                                </div>
                            )}
                        </div>
                    </div>
                </div>

                {/* Sidebar Column: Leaderboard */}
                <div className="space-y-8">
                    <div className="bg-card rounded-[2.5rem] border border-border shadow-xl p-8 h-full flex flex-col relative overflow-hidden">
                        <div className="absolute -bottom-10 -right-10 w-40 h-40 bg-emerald-500/5 rounded-full" />
                        <div className="flex justify-between items-center mb-10">
                            <h3 className="text-lg font-black flex items-center gap-2">
                                <Trophy className="text-amber-500" />
                                今日卷王榜
                            </h3>
                            <button className="text-[10px] font-black text-primary hover:bg-primary/10 px-3 py-1 rounded-full transition-colors uppercase tracking-widest border border-primary/20">查看总榜</button>
                        </div>

                        <div className="flex-1 space-y-2">
                            {statsData.rank_list && statsData.rank_list.length > 0 ? (
                                statsData.rank_list.map((u: any, i: number) => (
                                    <div key={i} className="group flex items-center gap-4 p-4 rounded-[1.5rem] hover:bg-muted/40 transition-all border border-transparent hover:border-border/50">
                                        <div className={cn(
                                            "w-10 h-10 flex items-center justify-center font-black rounded-xl text-sm shrink-0 shadow-sm",
                                            i === 0 ? "bg-amber-100 text-amber-600 shadow-amber-200/50" :
                                                i === 1 ? "bg-slate-100 text-slate-500 shadow-slate-200/50" :
                                                    i === 2 ? "bg-orange-100 text-orange-600 shadow-orange-200/50" :
                                                        "bg-muted text-muted-foreground"
                                        )}>
                                            {i + 1}
                                        </div>
                                        <img
                                            src={u.avatar ? (u.avatar.startsWith('http') ? u.avatar : `http://localhost:8080${u.avatar}`) : `https://api.dicebear.com/7.x/notionists/svg?seed=${u.username}`}
                                            className="w-10 h-10 rounded-xl border-2 border-background shadow-sm shadow-black/5"
                                            alt=""
                                        />
                                        <div className="flex-1 min-w-0">
                                            <p className="font-black text-sm text-foreground truncate group-hover:text-primary transition-colors">{u.nickname || u.username}</p>
                                            <p className="text-[10px] font-bold text-muted-foreground uppercase opacity-60">今日活跃</p>
                                        </div>
                                        <div className="text-right">
                                            <div className="text-lg font-black text-foreground tabular-nums leading-none">{u.count}</div>
                                            <div className="text-[10px] font-bold text-muted-foreground">题</div>
                                        </div>
                                    </div>
                                ))
                            ) : (
                                <div className="h-full flex flex-col items-center justify-center opacity-20 gap-4 py-10">
                                    <Trophy size={64} />
                                    <p className="text-xs font-black uppercase tracking-widest text-center">卷王还没产生<br />速来抢占席位</p>
                                </div>
                            )}
                        </div>

                        <div className="mt-8 pt-8 border-t border-border/50 flex items-center gap-4 px-2">
                            <div className="w-12 h-12 rounded-2xl bg-primary/10 flex items-center justify-center text-primary shrink-0 animate-pulse">
                                <Zap size={24} />
                            </div>
                            <div>
                                <p className="text-xs font-black text-foreground italic">“努力不会背叛你”</p>
                                <p className="text-[10px] font-bold text-muted-foreground uppercase tracking-widest">—— 考研冲刺中</p>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
}
