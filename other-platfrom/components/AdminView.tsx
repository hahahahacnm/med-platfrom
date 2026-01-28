
import React, { useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import {
  LayoutDashboard, Users, FileText, ShoppingBag, Settings,
  LogOut, TrendingUp, DollarSign, Activity, Search,
  MoreVertical, Edit3, Trash2, Plus, ChevronDown, ChevronRight, X, CreditCard,
  Database, Shield, Bell, CheckCircle2, AlertCircle, MessageSquare, Lightbulb, Bug, Upload, Check, Menu
} from 'lucide-react';
import { AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, PieChart, Pie, Cell, Legend } from 'recharts';
import { WIKI_CATEGORIES } from '../data';
import { useAppContext } from '../context';

import { WikiManager } from './WikiManager';

// --- Types for Admin ---
type AdminTab = 'dashboard' | 'users' | 'content' | 'store' | 'settings' | 'feedback';

const AdminView: React.FC<{ onExit: () => void }> = ({ onExit }) => {
  const [activeTab, setActiveTab] = useState<AdminTab>('dashboard');
  const [isSidebarOpen, setSidebarOpen] = useState(true);
  const [isMobile, setIsMobile] = useState(false);

  React.useEffect(() => {
    const handleResize = () => {
      const mobile = window.innerWidth < 768;
      setIsMobile(mobile);
      if (mobile) setSidebarOpen(false);
      else setSidebarOpen(true);
    };
    handleResize();
    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, []);

  // --- Render Content Based on Tab ---
  const renderContent = () => {
    switch (activeTab) {
      case 'dashboard': return <DashboardHome />;
      case 'users': return <UserManagement />;
      case 'content': return <ContentManagement />;
      case 'store': return <StoreManagement />;
      case 'settings': return <SystemSettings />;
      case 'feedback': return <FeedbackManagement />;
      default: return <DashboardHome />;
    }
  };

  return (
    <div className="flex h-screen bg-slate-100 font-sans text-slate-900 overflow-hidden">
      {/* Mobile Overlay */}
      {isMobile && isSidebarOpen && (
        <div
          className="fixed inset-0 bg-black/50 z-30"
          onClick={() => setSidebarOpen(false)}
        />
      )}

      {/* Admin Sidebar */}
      <motion.aside
        initial={{ width: isMobile ? 0 : 280 }}
        animate={{ width: isSidebarOpen ? 280 : (isMobile ? 0 : 80) }}
        transition={{ type: 'spring', stiffness: 300, damping: 30 }}
        className={`bg-slate-900 text-slate-300 flex flex-col shadow-2xl z-40 overflow-hidden ${isMobile ? 'fixed inset-y-0 left-0' : 'relative'}`}
      >
        <div className="h-16 flex items-center gap-3 px-6 border-b border-slate-800">
          <div className="w-8 h-8 rounded-lg bg-blue-600 flex items-center justify-center text-white shrink-0">
            <Shield size={18} />
          </div>
          {isSidebarOpen && (
            <motion.span
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              className="font-bold text-lg text-white tracking-wide"
            >
              Admin<span className="text-indigo-400">Panel</span>
            </motion.span>
          )}
        </div>

        <nav className="flex-1 py-6 space-y-2 px-3">
          <AdminNavItem icon={<LayoutDashboard size={20} />} label="仪表盘" isActive={activeTab === 'dashboard'} onClick={() => setActiveTab('dashboard')} isOpen={isSidebarOpen} />
          <AdminNavItem icon={<Users size={20} />} label="用户管理" isActive={activeTab === 'users'} onClick={() => setActiveTab('users')} isOpen={isSidebarOpen} />
          <AdminNavItem icon={<Database size={20} />} label="内容管理" isActive={activeTab === 'content'} onClick={() => setActiveTab('content')} isOpen={isSidebarOpen} />
          <AdminNavItem icon={<ShoppingBag size={20} />} label="商城运营" isActive={activeTab === 'store'} onClick={() => setActiveTab('store')} isOpen={isSidebarOpen} />
          <AdminNavItem icon={<MessageSquare size={20} />} label="用户反馈" isActive={activeTab === 'feedback'} onClick={() => setActiveTab('feedback')} isOpen={isSidebarOpen} />
          <AdminNavItem icon={<Settings size={20} />} label="系统设置" isActive={activeTab === 'settings'} onClick={() => setActiveTab('settings')} isOpen={isSidebarOpen} />
        </nav>

        <div className="p-4 border-t border-slate-800">
          <button
            onClick={onExit}
            className={`flex items-center gap-3 text-slate-400 hover:text-white hover:bg-slate-800 p-3 rounded-xl transition-all w-full ${!isSidebarOpen && 'justify-center'}`}
          >
            <LogOut size={20} />
            {isSidebarOpen && <span>退出管理后台</span>}
          </button>
        </div>
      </motion.aside>

      {/* Main Content Area */}
      <main className="flex-1 flex flex-col h-full overflow-hidden relative">
        {/* Top Header */}
        <header className="h-16 bg-white border-b border-slate-200 flex items-center justify-between px-4 md:px-8 shadow-sm z-10 shrink-0">
          <div className="flex items-center text-sm breadcrumbs text-slate-500">
            <button
              onClick={() => setSidebarOpen(!isSidebarOpen)}
              className="mr-3 p-2 hover:bg-slate-100 rounded-lg md:hidden text-slate-600"
            >
              <Menu size={20} />
            </button>
            <span className="font-medium text-indigo-600">题酷 Admin</span>
            <ChevronRight size={14} className="mx-2" />
            <span className="capitalize">{activeTab}</span>
          </div>
          <div className="flex items-center gap-6">
            <button className="relative text-slate-400 hover:text-indigo-600 transition-colors">
              <Bell size={20} />
              <span className="absolute top-0 right-0 w-2 h-2 bg-rose-500 rounded-full border border-white"></span>
            </button>
            <div className="flex items-center gap-3 pl-6 border-l border-slate-200">
              <div className="text-right hidden md:block">
                <div className="text-sm font-bold text-slate-800">Administrator</div>
                <div className="text-xs text-slate-400">Super User</div>
              </div>
              <div className="w-9 h-9 rounded-full bg-slate-200 border-2 border-white shadow-sm overflow-hidden">
                <img src="https://ui-avatars.com/api/?name=Admin&background=6366f1&color=fff" alt="Admin" />
              </div>
            </div>
          </div>
        </header>

        {/* Scrollable Content */}
        <div className="flex-1 overflow-y-auto p-8 scroll-smooth bg-slate-50">
          <AnimatePresence mode="wait">
            <motion.div
              key={activeTab}
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -10 }}
              transition={{ duration: 0.2 }}
              className="max-w-7xl mx-auto"
            >
              {renderContent()}
            </motion.div>
          </AnimatePresence>
        </div>
      </main>
    </div>
  );
};

// --- Sub-components ---

const AdminNavItem = ({ icon, label, isActive, onClick, isOpen }: any) => (
  <button
    onClick={onClick}
    className={`flex items-center gap-4 px-4 py-3 rounded-xl transition-all group w-full ${isActive
      ? 'bg-indigo-600 text-white shadow-lg shadow-indigo-900/50'
      : 'text-slate-400 hover:bg-slate-800 hover:text-white'
      } ${!isOpen && 'justify-center px-2'}`}
  >
    <div className={`${isActive ? 'text-white' : 'text-slate-400 group-hover:text-white'}`}>
      {icon}
    </div>
    {isOpen && <span className="font-medium tracking-wide text-sm">{label}</span>}
  </button>
);

const DashboardHome = () => {
  const [stats, setStats] = useState<{
    revenue: number;
    users: number;
    usersTrend?: number;
    activeSubs: number;
    activeSubsTrend?: number;
    quizCount: number;
    quizCountTrend?: number;
  }>({
    revenue: 0,
    users: 0,
    usersTrend: 0,
    activeSubs: 0,
    activeSubsTrend: 0,
    quizCount: 0,
    quizCountTrend: 0
  });

  const [trend, setTrend] = useState<any[]>([]);
  const [subjectDistribution, setSubjectDistribution] = useState<any[]>([]);
  const [dateRange, setDateRange] = useState(7); // Default to 7 days

  React.useEffect(() => {
    api.dashboard.getStats().then(setStats).catch(console.error);
    api.dashboard.getSubjectDistribution().then(setSubjectDistribution).catch(console.error);
  }, []);

  React.useEffect(() => {
    api.dashboard.getRevenueTrend(dateRange).then(setTrend).catch(console.error);
  }, [dateRange]);

  const [revenueTrendPct, setRevenueTrendPct] = useState(0);



  React.useEffect(() => {
    if (trend.length >= 2) {
      // Calculate trend based on last 7 days vs previous 7 days
      // If less than 14 days data, compare first half vs second half or just last 2 days
      const days = trend.length;
      const mid = Math.floor(days / 2);
      // Simple approach: Last 7 days vs previous 7 days (if available)
      const period = Math.min(7, mid);
      if (period > 0) {
        const currentPeriod = trend.slice(-period).reduce((acc, cur) => acc + cur.amount, 0);
        const prevPeriod = trend.slice(-period * 2, -period).reduce((acc, cur) => acc + cur.amount, 0);

        if (prevPeriod > 0) {
          setRevenueTrendPct(((currentPeriod - prevPeriod) / prevPeriod) * 100);
        } else if (currentPeriod > 0) {
          setRevenueTrendPct(100);
        } else {
          setRevenueTrendPct(0);
        }
      }
    }
  }, [trend]);

  return (
    <div className="space-y-8">
      <div>
        <h2 className="text-2xl font-bold text-slate-800">仪表盘概览</h2>
        <p className="text-slate-500 mt-1">欢迎回来，今日平台数据概况。</p>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <StatCard
          title="总收入"
          value={`¥ ${stats.revenue || 0}`}
          trend={`${revenueTrendPct >= 0 ? '+' : ''}${revenueTrendPct.toFixed(1)}%`}
          isPositive={revenueTrendPct >= 0}
          icon={<DollarSign className="text-emerald-600" size={24} />}
          bg="bg-emerald-50"
          chartColor="#059669"
        />
        <StatCard
          title="注册用户"
          value={stats.users || 0}
          trend={`${stats.usersTrend >= 0 ? '+' : ''}${(stats.usersTrend || 0).toFixed(1)}%`}
          isPositive={stats.usersTrend >= 0}
          icon={<Users className="text-blue-600" size={24} />}
          bg="bg-blue-50"
          chartColor="#2563eb"
        />
        <StatCard
          title="活跃订阅"
          value={stats.activeSubs || 0}
          trend={`${stats.activeSubsTrend >= 0 ? '+' : ''}${(stats.activeSubsTrend || 0).toFixed(1)}%`}
          isPositive={stats.activeSubsTrend >= 0}
          icon={<Activity className="text-indigo-600" size={24} />}
          bg="bg-indigo-50"
          chartColor="#4f46e5"
        />
        <StatCard
          title="累计刷题"
          value={stats.quizCount || 0}
          trend={`${stats.quizCountTrend >= 0 ? '+' : ''}${(stats.quizCountTrend || 0).toFixed(1)}%`}
          isPositive={stats.quizCountTrend >= 0}
          icon={<FileText className="text-amber-600" size={24} />}
          bg="bg-amber-50"
          chartColor="#d97706"
        />
      </div>

      {/* Charts Area */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        <div className="lg:col-span-2 bg-white p-6 rounded-2xl border border-slate-200 shadow-sm">
          <div className="flex justify-between items-center mb-6">
            <h3 className="font-bold text-slate-800 text-lg">收入趋势</h3>
            <select
              value={dateRange}
              onChange={(e) => setDateRange(Number(e.target.value))}
              className="bg-slate-50 border border-slate-200 rounded-lg text-sm px-3 py-1 text-slate-600 focus:outline-none"
            >
              <option value={7}>最近 7 天</option>
              <option value={30}>最近 30 天</option>
              <option value={365}>今年</option>
            </select>
          </div>
          {/* Real Chart Area */}
          <div className="h-72 w-full px-2 mt-4">
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart data={trend} margin={{ top: 10, right: 10, left: 0, bottom: 0 }}>
                <defs>
                  <linearGradient id="colorRevenue" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#6366f1" stopOpacity={0.3} />
                    <stop offset="95%" stopColor="#6366f1" stopOpacity={0} />
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#f1f5f9" />
                <XAxis
                  dataKey="date"
                  axisLine={false}
                  tickLine={false}
                  tick={{ fill: '#94a3b8', fontSize: 12 }}
                  tickMargin={10}
                  tickFormatter={(str) => {
                    if (!str) return '';
                    const d = new Date(str);
                    return `${d.getMonth() + 1}/${d.getDate()}`;
                  }}
                />
                <YAxis
                  axisLine={false}
                  tickLine={false}
                  tick={{ fill: '#94a3b8', fontSize: 12 }}
                  tickFormatter={(val) => `¥${val}`}
                />
                <Tooltip
                  contentStyle={{ borderRadius: '12px', border: 'none', boxShadow: '0 4px 6px -1px rgb(0 0 0 / 0.1)' }}
                  labelFormatter={(label) => new Date(label).toLocaleDateString()}
                  formatter={(value: number) => [`¥${value}`, '收入']}
                />
                <Area
                  type="monotone"
                  dataKey="amount"
                  stroke="#6366f1"
                  strokeWidth={3}
                  fillOpacity={1}
                  fill="url(#colorRevenue)"
                />
              </AreaChart>
            </ResponsiveContainer>
          </div>
        </div>

        <div className="bg-white p-6 rounded-2xl border border-slate-200 shadow-sm">
          <h3 className="font-bold text-slate-800 text-lg mb-6">累计刷题科目分布</h3>
          <div className="h-72 w-full">
            <ResponsiveContainer width="100%" height="100%">
              <PieChart>
                <Pie
                  data={subjectDistribution}
                  cx="50%"
                  cy="50%"
                  innerRadius={60}
                  outerRadius={80}
                  paddingAngle={5}
                  dataKey="value"
                  label={({ name, percent }) => `${name} ${(percent * 100).toFixed(0)}%`}
                >
                  {subjectDistribution.map((entry, index) => (
                    <Cell key={`cell-${index}`} fill={['#6366f1', '#ec4899', '#10b981', '#f59e0b', '#3b82f6', '#8b5cf6'][index % 6]} />
                  ))}
                </Pie>
                <Tooltip
                  contentStyle={{ borderRadius: '12px', border: 'none', boxShadow: '0 4px 6px -1px rgb(0 0 0 / 0.1)' }}
                  formatter={(value: number) => [`${value} 题`, '做题数']}
                />
                <Legend verticalAlign="bottom" height={36} iconType="circle" />
              </PieChart>
            </ResponsiveContainer>
          </div>
        </div>
      </div>
    </div>
  );
};

const StatCard = ({ title, value, trend, isPositive, icon, bg }: any) => (
  <motion.div
    whileHover={{ y: -5 }}
    className="bg-white p-6 rounded-2xl border border-slate-200 shadow-sm flex flex-col justify-between h-36 relative overflow-hidden"
  >
    <div className="flex justify-between items-start z-10">
      <div>
        <p className="text-sm font-medium text-slate-500 mb-1">{title}</p>
        <h4 className="text-2xl font-bold text-slate-900">{value}</h4>
      </div>
      <div className={`p-3 rounded-xl ${bg}`}>{icon}</div>
    </div>
    <div className="flex items-center gap-2 z-10">
      <span className={`text-xs font-bold px-1.5 py-0.5 rounded flex items-center gap-1 ${isPositive ? 'bg-emerald-50 text-emerald-600' : 'bg-rose-50 text-rose-600'}`}>
        {isPositive ? <TrendingUp size={12} /> : <TrendingUp size={12} className="rotate-180" />}
        {trend}
      </span>
      <span className="text-xs text-slate-400">较上周</span>
    </div>
    {/* Decorational Chart Line */}
    <div className="absolute -bottom-2 -left-2 -right-2 h-16 opacity-10 pointer-events-none">
      <svg viewBox="0 0 100 20" preserveAspectRatio="none" className="w-full h-full">
        <path d="M0,10 Q20,15 40,5 T80,12 T100,8 L100,20 L0,20 Z" fill="currentColor" className={isPositive ? 'text-emerald-500' : 'text-indigo-500'} />
      </svg>
    </div>
  </motion.div>
);

import { api } from '../services/api';

// --- User Management Component ---
const UserManagement = () => {
  const [searchTerm, setSearchTerm] = useState('');
  const [users, setUsers] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [isAddModalOpen, setIsAddModalOpen] = useState(false);
  const [selectedUser, setSelectedUser] = useState<any>(null);

  React.useEffect(() => {
    loadUsers();
  }, []);

  const loadUsers = async () => {
    try {
      setLoading(true);
      const data = await api.user.getAllUsers();
      setUsers(data);
    } catch (err) {
      console.error('Failed to load users', err);
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm('确定要删除该用户吗？此操作不可恢复。')) return;
    try {
      await api.user.deleteUser(id);
      setUsers(prev => prev.filter(u => u.id !== id));
    } catch (err) {
      alert('删除失败');
    }
  };

  const filteredUsers = users.filter(user =>
    user.name?.toLowerCase().includes(searchTerm.toLowerCase()) ||
    user.email?.toLowerCase().includes(searchTerm.toLowerCase())
  );

  return (
    <div className="bg-white rounded-2xl shadow-sm border border-slate-200 overflow-hidden">
      <div className="p-6 border-b border-slate-200 flex flex-col md:flex-row justify-between items-center gap-4">
        <h3 className="text-lg font-bold text-slate-800">用户列表</h3>
        <div className="flex gap-3 w-full md:w-auto">
          <div className="relative flex-1 md:w-64">
            <Search className="absolute left-3 top-2.5 text-slate-400" size={18} />
            <input
              type="text"
              placeholder="搜索用户..."
              value={searchTerm}
              onChange={e => setSearchTerm(e.target.value)}
              className="w-full pl-10 pr-4 py-2 bg-slate-50 border border-slate-200 rounded-lg text-sm focus:outline-none focus:border-indigo-500"
            />
          </div>
          <button
            onClick={() => setIsAddModalOpen(true)}
            className="bg-indigo-600 text-white px-4 py-2 rounded-lg text-sm font-bold flex items-center gap-2 hover:bg-indigo-700">
            <Plus size={16} /> 新增
          </button>
        </div>
      </div>
      <div className="overflow-x-auto">
        <table className="w-full text-left border-collapse">
          <thead>
            <tr className="bg-slate-50 text-slate-500 text-xs uppercase tracking-wider">
              <th className="px-6 py-4 font-semibold">用户</th>
              <th className="px-6 py-4 font-semibold">角色</th>
              <th className="px-6 py-4 font-semibold">刷题数 / 正确率</th>
              <th className="px-6 py-4 font-semibold">订阅等级</th>
              <th className="px-6 py-4 font-semibold">注册日期</th>
              <th className="px-6 py-4 font-semibold text-right">操作</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-100 text-sm">
            {loading ? (
              <tr><td colSpan={6} className="text-center py-8 text-slate-500">加载中...</td></tr>
            ) : filteredUsers.length === 0 ? (
              <tr><td colSpan={6} className="text-center py-8 text-slate-500">暂无用户</td></tr>
            ) : filteredUsers.map(user => (
              <tr key={user.id} className="hover:bg-slate-50 transition-colors group">
                <td className="px-6 py-4">
                  <div className="flex items-center gap-3">
                    <div className="w-8 h-8 rounded-full bg-slate-200 flex items-center justify-center text-xs font-bold text-slate-600">
                      {user.name.charAt(0)}
                    </div>
                    <div>
                      <div className="font-bold text-slate-800">{user.name}</div>
                      <div className="text-xs text-slate-400">{user.email}</div>
                    </div>
                  </div>
                </td>
                <td className="px-6 py-4 text-slate-600">
                  <span className={`px-2 py-0.5 rounded text-xs border ${user.role === 'admin' ? 'bg-purple-50 text-purple-600 border-purple-200' : 'bg-slate-50 text-slate-500 border-slate-200'}`}>
                    {user.role}
                  </span>
                </td>
                <td className="px-6 py-4">
                  <div className="text-slate-800 font-bold">{user.stats.quizCount} 题</div>
                  <div className="text-xs text-slate-500">正确率: {user.stats.accuracy}</div>
                </td>
                <td className="px-6 py-4">
                  <div>
                    {user.stats.subscriptions.length > 0 ? (
                      <span className="text-indigo-600 font-bold text-xs bg-indigo-50 px-2 py-0.5 rounded">
                        已订阅 ({user.stats.subscriptions.length})
                      </span>
                    ) : (
                      <span className="text-slate-400 text-xs">普通用户</span>
                    )}
                  </div>
                </td>
                <td className="px-6 py-4 text-slate-500">{new Date(user.createdAt).toLocaleDateString()}</td>
                <td className="px-6 py-4 text-right">
                  <div className="flex gap-2 justify-end">
                    <button
                      onClick={() => setSelectedUser(user)}
                      className="p-2 text-slate-400 hover:text-indigo-600 hover:bg-indigo-50 rounded-lg transition-colors"
                      title="详情/订阅"
                    >
                      <Edit3 size={16} />
                    </button>
                    <button
                      onClick={() => handleDelete(user.id)}
                      className="p-2 text-slate-400 hover:text-rose-600 hover:bg-rose-50 rounded-lg transition-colors"
                      title="删除用户"
                    >
                      <Trash2 size={16} />
                    </button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      <div className="p-4 border-t border-slate-200 flex justify-between items-center text-sm text-slate-500">
        <span>共 {filteredUsers.length} 位用户</span>
      </div>

      {isAddModalOpen && (
        <AddUserModal
          onClose={() => setIsAddModalOpen(false)}
          onSuccess={() => {
            setIsAddModalOpen(false);
            loadUsers();
          }}
        />
      )}

      {selectedUser && (
        <UserDetailModal
          user={selectedUser}
          onClose={() => setSelectedUser(null)}
          onUpdate={() => {
            loadUsers();
            // keep modal open or close? user might want to see update. 
            // Better reload user data and keep open if possible, but loadUsers refreshes whole list. 
            // For now, let's close it or re-fetch specific user.
            // Simple way: close it to simple interaction
            setSelectedUser(null);
          }}
        />
      )}
    </div>
  );
};

// --- Content Management Component ---

const ContentManagement = () => {
  const [contentType, setContentType] = useState<'quiz' | 'wiki' | 'announcement'>('quiz');
  const [contacts, setContacts] = useState<any[]>([]); // If needed
  const [subjects, setSubjects] = useState<any[]>([]);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingSubject, setEditingSubject] = useState<any>(null);
  const [viewingChapter, setViewingChapter] = useState<any>(null); // New state for viewing chapter details
  const fileInputRef = React.useRef<HTMLInputElement>(null);
  const [importingSubjectId, setImportingSubjectId] = useState<string | null>(null);

  React.useEffect(() => {
    if (contentType === 'quiz') {
      loadSubjects();
    }
  }, [contentType]);

  const loadSubjects = async () => {
    try {
      const data = await api.quiz.getSubjects();
      setSubjects(data);
    } catch (e) {
      console.error(e);
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm('确定删除该科目？删除后包含的章节和题目也将被删除。')) return;
    try {
      await api.quiz.deleteSubject(id);
      loadSubjects();
    } catch (e) {
      alert('删除失败');
    }
  };

  const handleSave = async (data: any) => {
    try {
      if (editingSubject) {
        await api.quiz.updateSubject(editingSubject.id, data);
      } else {
        await api.quiz.createSubject(data);
      }
      setIsModalOpen(false);
      setEditingSubject(null);
      loadSubjects();
    } catch (e: any) {
      alert(e.message || '保存失败');
    }
  };

  const handleImportClick = (subjectId: string) => {
    setImportingSubjectId(subjectId);
    fileInputRef.current?.click();
  };

  const handleFileChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    if (!e.target.files?.length || !importingSubjectId) return;
    const files = Array.from(e.target.files);

    try {
      const res: any = await api.quiz.importQuestions(importingSubjectId, files);
      alert(`成功导入 ${files.length} 个文件，共 ${res.count || 0} 道题目`);
      loadSubjects();
    } catch (err: any) {
      alert('导入失败: ' + err.message);
    } finally {
      setImportingSubjectId(null);
      if (fileInputRef.current) fileInputRef.current.value = '';
    }
  };

  const handleDeleteChapter = async (id: string) => {
    if (!confirm('确定要删除该章节吗？删除后章节内的所有题目都会被删除。')) return;
    try {
      await api.quiz.deleteChapter(id);
      loadSubjects();
    } catch (err: any) {
      alert('删除失败: ' + err.message);
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4 bg-white p-2 rounded-xl border border-slate-200 w-fit">
        <button
          onClick={() => setContentType('quiz')}
          className={`px-4 py-2 rounded-lg text-sm font-bold transition-all ${contentType === 'quiz' ? 'bg-indigo-600 text-white shadow-md' : 'text-slate-500 hover:bg-slate-50'}`}
        >
          题库管理
        </button>
        <button
          onClick={() => setContentType('wiki')}
          className={`px-4 py-2 rounded-lg text-sm font-bold transition-all ${contentType === 'wiki' ? 'bg-pink-600 text-white shadow-md' : 'text-slate-500 hover:bg-slate-50'}`}
        >
          知识库管理
        </button>
        <button
          onClick={() => setContentType('announcement')}
          className={`px-4 py-2 rounded-lg text-sm font-bold transition-all ${contentType === 'announcement' ? 'bg-amber-600 text-white shadow-md' : 'text-slate-500 hover:bg-slate-50'}`}
        >
          公告管理
        </button>
      </div>

      <div className="grid gap-6">
        {contentType === 'quiz' ? (
          <>
            <div className="flex justify-end">
              <button
                onClick={() => { setEditingSubject(null); setIsModalOpen(true); }}
                className="bg-indigo-600 text-white px-4 py-2 rounded-lg text-sm font-bold flex items-center gap-2 hover:bg-indigo-700 shadow-lg shadow-indigo-200"
              >
                <Plus size={16} /> 新增科目
              </button>
            </div>

            {subjects.map(subject => (
              <div key={subject.id} className="bg-white border border-slate-200 rounded-2xl overflow-hidden shadow-sm">
                <div className="p-4 bg-slate-50 border-b border-slate-200 flex justify-between items-center">
                  <div className="flex items-center gap-3">
                    <span className="text-2xl">{subject.icon}</span>
                    <div>
                      <h3 className="font-bold text-slate-800">{subject.title}</h3>
                      <p className="text-xs text-slate-500">
                        {subject.chapters?.length || 0} 个章节，
                        共 {subject.chapters?.reduce((a: number, c: any) => a + (c.questions?.length || 0), 0) || 0} 道题
                      </p>
                    </div>
                  </div>
                  <div className="flex gap-2">
                    <button
                      onClick={() => { setEditingSubject(subject); setIsModalOpen(true); }}
                      className="p-2 text-slate-400 hover:text-indigo-600 hover:bg-indigo-50 rounded-lg"
                      title="编辑科目"
                    >
                      <Edit3 size={16} />
                    </button>
                    <button
                      onClick={() => handleDelete(subject.id)}
                      className="p-2 text-slate-400 hover:text-rose-600 hover:bg-rose-50 rounded-lg"
                      title="删除科目"
                    >
                      <Trash2 size={16} />
                    </button>
                    <button
                      onClick={() => handleImportClick(subject.id)}
                      className="p-2 text-slate-400 hover:text-indigo-600 hover:bg-indigo-50 rounded-lg"
                      title="导入题目 (XLSX)"
                    >
                      <Upload size={16} />
                    </button>
                  </div>
                </div>
                {/* Chapter preview disabled for now to keep it clean, or we can list them */}
                <div className="divide-y divide-slate-100 max-h-60 overflow-y-auto">
                  {subject.chapters?.map((chapter: any, idx: number) => (
                    <div key={chapter.id} className="p-4 pl-12 hover:bg-slate-50 transition-colors group">
                      <div className="flex justify-between items-center">
                        <div className="flex items-center gap-3">
                          <div className="w-6 h-6 rounded-full bg-slate-200 text-slate-500 flex items-center justify-center text-xs font-bold">{idx + 1}</div>
                          <button
                            onClick={() => setViewingChapter(chapter)}
                            className="font-medium text-slate-700 hover:text-indigo-600 transition-colors text-left"
                          >
                            {chapter.title}
                          </button>
                          <span className="text-xs text-slate-400 bg-slate-100 px-2 py-0.5 rounded-full">{chapter.questions?.length || 0} 题</span>
                        </div>
                        {/* Future: Edit Chapter button */}
                        <div className="opacity-0 group-hover:opacity-100 transition-opacity flex items-center gap-2">
                          <button
                            onClick={() => handleDeleteChapter(chapter.id)}
                            className="text-xs font-bold text-rose-400 hover:text-rose-600 hover:bg-rose-50 p-1.5 rounded transition-colors"
                            title="删除章节"
                          >
                            <Trash2 size={12} />
                          </button>
                        </div>
                      </div>
                    </div>
                  ))}
                  {(!subject.chapters || subject.chapters.length === 0) && (
                    <div className="p-4 text-center text-slate-400 text-sm">暂无章节</div>
                  )}
                </div>
              </div>
            ))}

            {subjects.length === 0 && (
              <div className="text-center py-12 text-slate-400 bg-white rounded-2xl border border-dashed border-slate-200">
                暂无题库科目，请点击上方按钮添加。
              </div>
            )}
          </>
        ) : contentType === 'wiki' ? (
          <WikiManager />
        ) : (
          <AnnouncementManager />
        )}
      </div>

      {isModalOpen && (
        <SubjectModal
          subject={editingSubject}
          onClose={() => setIsModalOpen(false)}
          onSave={handleSave}
        />
      )}

      <input
        type="file"
        multiple
        ref={fileInputRef}
        className="hidden"
        onChange={handleFileChange}
        accept=".xlsx,.xls"
      />

      {viewingChapter && (
        <ChapterDetailModal
          chapter={viewingChapter}
          onClose={() => setViewingChapter(null)}
          onUpdate={() => {
            loadSubjects();
            // Ideally reload viewChapter data too, but loadSubjects refreshes everything.
            // We can find the updated chapter in new data.
            // For simplicity, close or just rely on subjects update.
            // Actually, pass a re-fetcher or just update local state if possible.
            // Let's close it or implement a smart refresh.
            setViewingChapter(null);
          }}
        />
      )}
    </div>
  );
};

// --- Chapter Detail Modal ---
const ChapterDetailModal = ({ chapter, onClose, onUpdate }: any) => {
  const [questions, setQuestions] = useState<any[]>(chapter.questions || []);
  const [editingQuestion, setEditingQuestion] = useState<any>(null);

  // Filter / Search could be added here

  const handleUpdate = async (updatedQ: any) => {
    try {
      await api.quiz.updateQuestion(updatedQ.id, updatedQ);
      // Update local state
      setQuestions(prev => prev.map(q => q.id === updatedQ.id ? updatedQ : q));
      onUpdate(); // Trigger parent refresh (although we update local state too)
      setEditingQuestion(null);
    } catch (e: any) {
      alert('更新失败: ' + e.message);
    }
  };

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 backdrop-blur-sm">
      <div className="bg-white rounded-2xl w-full max-w-4xl p-6 shadow-2xl h-[90vh] flex flex-col">
        <div className="flex justify-between items-center mb-4 border-b border-slate-100 pb-4">
          <div>
            <h3 className="text-xl font-bold text-slate-800">{chapter.title}</h3>
            <p className="text-sm text-slate-500">共 {questions.length} 道题目</p>
          </div>
          <button onClick={onClose} className="p-2 hover:bg-slate-100 rounded-lg text-slate-500"><X size={20} /></button>
        </div>

        <div className="flex-1 overflow-y-auto space-y-4 p-2">
          {questions.map((q, idx) => (
            <div key={q.id} className="border border-slate-200 rounded-xl p-4 hover:bg-slate-50 transition-colors">
              <div className="flex justify-between items-start gap-4">
                <div className="flex-1">
                  <div className="flex gap-2 mb-2">
                    <span className="bg-slate-100 text-slate-500 text-xs px-2 py-0.5 rounded font-mono">#{idx + 1}</span>
                    <span className="bg-indigo-50 text-indigo-600 text-xs px-2 py-0.5 rounded font-bold">ID: {q.id}</span>
                  </div>
                  <p className="font-medium text-slate-800 mb-3">{q.text}</p>
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-2 text-sm text-slate-600 mb-3">
                    {q.options.map((opt: any) => (
                      <div key={opt.id} className={`flex gap-2 ${q.correctAnswers.includes(opt.id) ? 'text-emerald-600 font-bold' : ''}`}>
                        <span className="w-5 h-5 rounded-full border flex items-center justify-center text-xs shrink-0
                                            ${q.correctAnswers.includes(opt.id) ? 'border-emerald-500 bg-emerald-50' : 'border-slate-300'}">
                          {opt.id}
                        </span>
                        <span>{opt.text}</span>
                      </div>
                    ))}
                  </div>
                  <div className="bg-slate-50 p-3 rounded-lg text-sm text-slate-500">
                    <span className="font-bold text-slate-700">解析：</span>
                    {q.explanation || '暂无解析'}
                  </div>
                </div>
                <button
                  onClick={() => setEditingQuestion(q)}
                  className="p-2 text-slate-400 hover:text-indigo-600 hover:bg-indigo-50 rounded-lg shrink-0"
                >
                  <Edit3 size={18} />
                </button>
              </div>
            </div>
          ))}
        </div>
      </div>

      {editingQuestion && (
        <QuestionEditModal
          question={editingQuestion}
          onClose={() => setEditingQuestion(null)}
          onSave={handleUpdate}
        />
      )}
    </div>
  );
};

const QuestionEditModal = ({ question, onClose, onSave }: any) => {
  const [formData, setFormData] = useState({
    ...question,
    correctAnswersStr: question.correctAnswers.join(',')
  });

  const handleOptionChange = (idx: number, field: string, value: string) => {
    const newOptions = [...formData.options];
    newOptions[idx] = { ...newOptions[idx], [field]: value };
    setFormData({ ...formData, options: newOptions });
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    const correctAnswers = formData.correctAnswersStr.toUpperCase().split(/[,，]/).map((s: string) => s.trim()).filter(Boolean);
    onSave({
      ...formData,
      correctAnswers
    });
  };

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-[60] backdrop-blur-sm">
      <div className="bg-white rounded-2xl w-full max-w-2xl p-6 shadow-2xl animate-in fade-in zoom-in duration-200 max-h-[90vh] overflow-y-auto">
        <div className="flex justify-between items-center mb-4">
          <h3 className="text-xl font-bold text-slate-800">编辑题目</h3>
          <button onClick={onClose}><X size={20} className="text-slate-400" /></button>
        </div>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-sm font-bold text-slate-700 mb-1">题干</label>
            <textarea
              required
              className="w-full px-4 py-2 border border-slate-200 rounded-lg focus:outline-none focus:border-indigo-500 h-24"
              value={formData.text}
              onChange={e => setFormData({ ...formData, text: e.target.value })}
            />
          </div>

          <div>
            <label className="block text-sm font-bold text-slate-700 mb-1">选项</label>
            <div className="space-y-2">
              {formData.options.map((opt: any, idx: number) => (
                <div key={idx} className="flex gap-2 items-center">
                  <span className="w-8 font-bold text-slate-400 text-center">{opt.id}</span>
                  <input
                    type="text"
                    className="flex-1 px-3 py-2 border border-slate-200 rounded-lg focus:outline-none focus:border-indigo-500 text-sm"
                    value={opt.text}
                    onChange={e => handleOptionChange(idx, 'text', e.target.value)}
                  />
                </div>
              ))}
            </div>
          </div>

          <div>
            <label className="block text-sm font-bold text-slate-700 mb-1">正确答案 (多选逗号分隔, 如 A,B)</label>
            <input
              type="text"
              required
              className="w-full px-4 py-2 border border-slate-200 rounded-lg focus:outline-none focus:border-indigo-500"
              value={formData.correctAnswersStr}
              onChange={e => setFormData({ ...formData, correctAnswersStr: e.target.value })}
            />
          </div>

          <div>
            <label className="block text-sm font-bold text-slate-700 mb-1">解析</label>
            <textarea
              className="w-full px-4 py-2 border border-slate-200 rounded-lg focus:outline-none focus:border-indigo-500 h-24"
              value={formData.explanation}
              onChange={e => setFormData({ ...formData, explanation: e.target.value })}
            />
          </div>

          <div className="flex justify-end gap-3 pt-4 border-t border-slate-100">
            <button type="button" onClick={onClose} className="px-4 py-2 text-slate-500 hover:bg-slate-100 rounded-lg font-bold">取消</button>
            <button type="submit" className="px-4 py-2 bg-indigo-600 text-white rounded-lg font-bold hover:bg-indigo-700">保存修改</button>
          </div>
        </form>
      </div>
    </div>
  );
};

// --- Feedback Management ---
const FeedbackManagement = () => {
  const [feedbacks, setFeedbacks] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);

  React.useEffect(() => {
    loadFeedbacks();
  }, []);

  const loadFeedbacks = async () => {
    try {
      // API call to get all feedback
      const data = await api.feedback.getAll();
      setFeedbacks(data);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const toggleStatus = async (id: string, currentStatus: string) => {
    try {
      await api.feedback.updateStatus(id, currentStatus === 'pending' ? 'resolved' : 'pending');
      setFeedbacks(prev => prev.map(f => f.id === id ? { ...f, status: f.status === 'pending' ? 'resolved' : 'pending' } : f));
    } catch (err) {
      alert('Update failed');
    }
  };

  const deleteFeedback = async (id: string) => {
    if (!confirm('确定删除此反馈？')) return;
    try {
      await api.feedback.delete(id);
      setFeedbacks(prev => prev.filter(f => f.id !== id));
    } catch (err) {
      alert('Delete failed');
    }
  };

  const getIcon = (type: string) => {
    switch (type) {
      case 'bug': return <Bug size={16} className="text-rose-500" />;
      case 'suggestion': return <Lightbulb size={16} className="text-amber-500" />;
      case 'content': return <FileText size={16} className="text-indigo-500" />;
      default: return <MessageSquare size={16} className="text-slate-500" />;
    }
  };

  const getLabel = (type: string) => {
    const map: Record<string, string> = { bug: '系统Bug', suggestion: '功能建议', content: '内容错误', other: '其他' };
    return map[type] || type;
  };

  return (
    <div className="bg-white rounded-2xl shadow-sm border border-slate-200 overflow-hidden">
      <div className="p-6 border-b border-slate-200">
        <h3 className="text-lg font-bold text-slate-800">用户反馈列表</h3>
      </div>
      <div className="divide-y divide-slate-100">
        {feedbacks.map(item => (
          <div key={item.id} className="p-6 hover:bg-slate-50 transition-colors">
            <div className="flex justify-between items-start mb-2">
              <div className="flex items-center gap-3">
                <div className="bg-slate-100 p-1.5 rounded-lg border border-slate-200">{getIcon(item.type)}</div>
                <span className="font-bold text-slate-800 text-sm bg-slate-100 px-2 py-0.5 rounded text-slate-600">{getLabel(item.type)}</span>
                <span className="text-xs text-slate-400 mx-1">•</span>
                <span className="text-xs text-slate-500 font-medium">{item.userName || '匿名用户'}</span>
                <span className="text-xs text-slate-400">{new Date(item.createdAt).toLocaleString()}</span>
                {item.contact && <span className="text-xs text-blue-500 ml-2">({item.contact})</span>}
              </div>
              <button
                onClick={() => toggleStatus(item.id, item.status)}
                className={`px-3 py-1 rounded-full text-xs font-bold border transition-all ${item.status === 'resolved'
                  ? 'bg-emerald-50 text-emerald-600 border-emerald-200'
                  : 'bg-amber-50 text-amber-600 border-amber-200 animate-pulse'
                  }`}
              >
                {item.status === 'resolved' ? '已处理' : '待处理'}
              </button>
            </div>
            <p className="text-slate-700 text-sm ml-10 mb-4">{item.content}</p>
            <div className="ml-10 flex gap-4">
              {item.status === 'pending' && (
                <button onClick={() => toggleStatus(item.id, item.status)} className="text-xs font-bold text-emerald-600 hover:underline">
                  标记为已解决
                </button>
              )}
              <button
                onClick={() => deleteFeedback(item.id)}
                className="text-xs font-bold text-slate-400 hover:text-rose-500 hover:underline"
              >
                删除反馈
              </button>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};

// --- Store Management ---
const StoreManagement = () => {
  const [view, setView] = useState<'products' | 'transactions'>('products');
  const [products, setProducts] = useState<any[]>([]);
  const [transactions, setTransactions] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingProduct, setEditingProduct] = useState<any>(null);

  // Data for linking products to content
  const [wikiCategories, setWikiCategories] = useState<any[]>([]);
  const [quizSubjects, setQuizSubjects] = useState<any[]>([]);

  React.useEffect(() => {
    loadProducts();
    // Fetch available content for linking
    api.wiki.getCategories().then(setWikiCategories).catch(console.error);
    api.quiz.getSubjects().then(setQuizSubjects).catch(console.error);
  }, []);

  React.useEffect(() => {
    if (view === 'transactions') {
      loadTransactions();
    }
  }, [view]);

  const loadProducts = async () => {
    try {
      setLoading(true);
      const data = await api.store.getProducts();
      setProducts(data);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const loadTransactions = async () => {
    try {
      setLoading(true);
      const data = await api.store.getTransactions();
      setTransactions(data);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const handleEdit = (product: any) => {
    setEditingProduct(product);
    setIsModalOpen(true);
  };

  const handleToggleStatus = async (product: any) => {
    const newStatus = product.isPublished === false; // Toggle
    try {
      await api.store.updateProduct(product.id, { isPublished: newStatus });
      setProducts(prev => prev.map(p => p.id === product.id ? { ...p, isPublished: newStatus } : p));
    } catch (err: any) {
      alert(err.message || '操作失败');
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm('确定要删除此商品吗？')) return;
    try {
      await api.store.deleteProduct(id);
      setProducts(prev => prev.filter(p => p.id !== id));
    } catch (err) {
      alert('删除失败');
    }
  };

  const handleSave = async (data: any) => {
    try {
      if (editingProduct) {
        await api.store.updateProduct(editingProduct.id, data);
      } else {
        await api.store.createProduct(data);
      }
      setIsModalOpen(false);
      setEditingProduct(null);
      loadProducts();
    } catch (err: any) {
      alert(err.message || '保存失败');
    }
  };

  const [couponFilter, setCouponFilter] = useState('');

  const filteredTransactions = transactions.filter(t => {
    if (!couponFilter) return true;
    if (t.couponDetails && Array.isArray(t.couponDetails)) {
      return t.couponDetails.some((c: any) => c.code.toLowerCase().includes(couponFilter.toLowerCase()));
    }
    return false;
  });

  const couponSalesTotal = filteredTransactions.reduce((acc, t) => acc + t.amount, 0);

  return (
    <div className="space-y-6">
      <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
        <div>
          <h2 className="text-2xl font-bold text-slate-800">商城运营</h2>
          <p className="text-slate-500">管理订阅商品、定价策略与售卖记录</p>
        </div>
        <div className="flex gap-4">
          <div className="flex items-center gap-2 bg-white p-1 rounded-xl border border-slate-200 shadow-sm">
            <button
              onClick={() => setView('products')}
              className={`px-4 py-1.5 rounded-lg text-sm font-bold transition-all ${view === 'products' ? 'bg-indigo-600 text-white shadow-md' : 'text-slate-500 hover:bg-slate-50'}`}
            >
              商品列表
            </button>
            <button
              onClick={() => setView('transactions')}
              className={`px-4 py-1.5 rounded-lg text-sm font-bold transition-all ${view === 'transactions' ? 'bg-indigo-600 text-white shadow-md' : 'text-slate-500 hover:bg-slate-50'}`}
            >
              售卖记录
            </button>
          </div>
          {view === 'products' && (
            <button
              onClick={() => { setEditingProduct(null); setIsModalOpen(true); }}
              className="bg-indigo-600 text-white px-5 py-2.5 rounded-xl font-bold shadow-lg shadow-indigo-200 hover:bg-indigo-700 transition-colors flex items-center gap-2"
            >
              <Plus size={18} /> 新增商品
            </button>
          )}
        </div>
      </div>

      {view === 'products' ? (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {products.map(product => (
            <div key={product.id} className="bg-white rounded-2xl border border-slate-200 overflow-hidden shadow-sm group">
              <div className="h-40 overflow-hidden relative">
                <img src={product.imageUrl} alt={product.title} className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-500" />
                <div className="absolute top-2 right-2 bg-slate-900/70 backdrop-blur text-white text-xs px-2 py-1 rounded">
                  ID: {product.accessId}
                </div>
              </div>
              <div className="p-5">
                <div className="flex justify-between items-start mb-2">
                  <h3 className="font-bold text-slate-800 line-clamp-1">{product.title}</h3>
                  <div className="flex gap-2">
                    {product.isPublished === false && <span className="bg-rose-100 text-rose-600 text-xs px-2 py-0.5 rounded font-bold">已下架</span>}
                    <span className="font-bold text-indigo-600 bg-indigo-50 px-2 py-0.5 rounded text-sm">¥{product.price}</span>
                  </div>
                </div>
                <p className="text-xs text-slate-500 mb-4 line-clamp-2 h-8">{product.description}</p>

                <div className="flex gap-2 pt-4 border-t border-slate-100">
                  <button
                    onClick={() => handleToggleStatus(product)}
                    className={`flex-1 py-2 rounded-lg text-sm font-medium transition-colors flex items-center justify-center gap-2 ${product.isPublished !== false
                      ? 'bg-amber-50 text-amber-600 hover:bg-amber-100'
                      : 'bg-emerald-50 text-emerald-600 hover:bg-emerald-100'}`}
                  >
                    {product.isPublished !== false ? <><X size={14} /> 下架</> : <><Check size={14} /> 上架</>}
                  </button>
                  <button
                    onClick={() => handleEdit(product)}
                    className="flex-1 py-2 rounded-lg bg-slate-50 text-slate-600 text-sm font-medium hover:bg-slate-100 transition-colors flex items-center justify-center gap-2"
                  >
                    <Edit3 size={14} /> 编辑
                  </button>
                  <button
                    onClick={() => handleDelete(product.id)}
                    className="p-2 rounded-lg bg-rose-50 text-rose-600 text-sm font-medium hover:bg-rose-100 transition-colors flex items-center justify-center"
                    title="永久删除"
                  >
                    <Trash2 size={14} />
                  </button>
                </div>
              </div>
            </div>
          ))}
          {loading && products.length === 0 && <div className="col-span-full py-12 text-center text-slate-400 font-medium">加载中...</div>}
          {!loading && products.length === 0 && <div className="col-span-full py-12 text-center text-slate-400 font-medium">暂无商品</div>}
        </div>
      ) : (
        <div className="bg-white rounded-2xl border border-slate-200 overflow-hidden shadow-sm">
          <div className="p-4 border-b border-slate-200 flex justify-between items-center bg-slate-50/50">
            <div className="flex items-center gap-4">
              <div className="relative">
                <Search className="absolute left-3 top-2.5 text-slate-400" size={16} />
                <input
                  type="text"
                  placeholder="优惠码过滤..."
                  className="pl-9 pr-4 py-2 border border-slate-200 rounded-lg text-sm focus:outline-none focus:border-indigo-500 w-64"
                  value={couponFilter}
                  onChange={e => setCouponFilter(e.target.value)}
                />
              </div>
              {couponFilter && (
                <div className="text-sm font-bold text-slate-700">
                  统计金额: <span className="text-indigo-600">¥{couponSalesTotal}</span>
                </div>
              )}
            </div>
            <div className="text-sm text-slate-500">
              共 {filteredTransactions.length} 条记录
            </div>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full text-left border-collapse">
              <thead>
                <tr className="bg-slate-50 text-slate-500 text-xs uppercase tracking-wider">
                  <th className="px-6 py-4 font-semibold">交易 ID</th>
                  <th className="px-6 py-4 font-semibold">用户</th>
                  <th className="px-6 py-4 font-semibold">所购商品</th>
                  <th className="px-6 py-4 font-semibold">优惠码</th>
                  <th className="px-6 py-4 font-semibold">交易金额</th>
                  <th className="px-6 py-4 font-semibold">交易时间</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100 text-sm">
                {loading && transactions.length === 0 ? (
                  <tr><td colSpan={6} className="text-center py-12 text-slate-500">加载中...</td></tr>
                ) : filteredTransactions.length === 0 ? (
                  <tr><td colSpan={6} className="text-center py-12 text-slate-500">暂无售卖记录</td></tr>
                ) : filteredTransactions.map(item => (
                  <tr key={item.id} className="hover:bg-slate-50 transition-colors">
                    <td className="px-6 py-4 font-mono text-xs text-slate-400">{item.id.slice(0, 8)}...</td>
                    <td className="px-6 py-4">
                      <div className="flex items-center gap-2">
                        <div className="w-7 h-7 rounded-sm bg-indigo-50 text-indigo-600 flex items-center justify-center font-bold text-xs uppercase">
                          {item.user?.name?.charAt(0) || 'U'}
                        </div>
                        <div>
                          <div className="font-bold text-slate-800">{item.user?.name || '未知用户'}</div>
                          <div className="text-[10px] text-slate-400">{item.user?.email}</div>
                        </div>
                      </div>
                    </td>
                    <td className="px-6 py-4">
                      <div className="flex flex-wrap gap-1">
                        {item.products.map((pId: string) => {
                          const p = quizSubjects.find(s => s.id === pId) || products.find(prod => prod.id === pId || prod.accessId === pId);
                          return (
                            <span key={pId} className="bg-slate-100 text-slate-600 px-2 py-0.5 rounded text-[10px] font-bold">
                              {p ? p.title : pId}
                            </span>
                          );
                        })}
                      </div>
                    </td>
                    <td className="px-6 py-4">
                      {item.couponDetails && item.couponDetails.length > 0 ? (
                        <div className="flex flex-col gap-1">
                          {item.couponDetails.map((c: any, idx: number) => (
                            <span key={idx} className="text-xs bg-indigo-50 text-indigo-600 px-1.5 py-0.5 rounded border border-indigo-100 font-mono">
                              {c.code} (-¥{c.discount})
                            </span>
                          ))}
                        </div>
                      ) : <span className="text-slate-300">-</span>}
                    </td>
                    <td className="px-6 py-4">
                      <span className="font-bold text-emerald-600">¥{item.amount}</span>
                    </td>
                    <td className="px-6 py-4 text-slate-500 text-xs">
                      {new Date(item.createdAt).toLocaleString()}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {isModalOpen && (
        <ProductModal
          product={editingProduct}
          onClose={() => setIsModalOpen(false)}
          onSave={handleSave}
          wikiCategories={wikiCategories}
          quizSubjects={quizSubjects}
        />
      )}
    </div>
  );
};

// --- Settings ---
const SystemSettings = () => {
  const [settings, setSettings] = useState<any[]>([]);

  React.useEffect(() => {
    loadSettings();
  }, []);

  const loadSettings = async () => {
    try {
      const data = await api.settings.getAll();
      setSettings(data);
    } catch (e) {
      console.error(e);
    }
  };

  const handleToggle = async (key: string, currentValue: string) => {
    const newValue = currentValue === 'true' ? 'false' : 'true';
    try {
      await api.settings.update(key, newValue);
      setSettings(prev => prev.map(s => s.key === key ? { ...s, value: newValue } : s));
    } catch (e) {
      alert('Update failed');
    }
  };

  const handleUpdate = async (key: string, value: string) => {
    try {
      await api.settings.update(key, value);
      setSettings(prev => prev.map(s => s.key === key ? { ...s, value } : s));
    } catch (e: any) {
      alert('Update failed: ' + e.message);
    }
  };

  const getValue = (key: string) => {
    const val = settings.find(s => s.key === key)?.value;
    return val === 'true';
  };

  const getStringValue = (key: string) => {
    return settings.find(s => s.key === key)?.value || '';
  };

  return (
    <div className="max-w-3xl">
      <h2 className="text-2xl font-bold text-slate-800 mb-6">系统设置</h2>

      <div className="bg-white rounded-2xl border border-slate-200 shadow-sm divide-y divide-slate-100 mb-6">
        <h3 className="px-6 py-4 font-bold text-slate-700 bg-slate-50/50">基础功能开关</h3>
        <SettingItem
          title="允许新用户注册"
          desc="关闭后，暂停新用户注册功能。"
          toggle={getValue('registration_enabled')}
          onToggle={() => handleToggle('registration_enabled', String(getValue('registration_enabled')))}
          icon={<Users className="text-blue-500" />}
        />
        <SettingItem
          title="开启 AI 助教功能"
          desc="启用后，用户可使用 AI 助教进行智能问答。"
          toggle={getValue('ai_enabled')}
          onToggle={() => handleToggle('ai_enabled', String(getValue('ai_enabled')))}
          icon={<TrendingUp className="text-indigo-500" />}
        />
      </div>

      {getValue('ai_enabled') && (
        <div className="bg-white rounded-2xl border border-slate-200 shadow-sm overflow-hidden">
          <h3 className="px-6 py-4 font-bold text-slate-700 bg-slate-50/50 border-b border-slate-100">AI 参数配置 (OpenAI 兼容)</h3>
          <div className="divide-y divide-slate-100">
            <ConfigInputItem
              label="API Base URL"
              desc="AI 服务提供商的基础 URL (例如: https://api.openai.com/v1)"
              value={getStringValue('ai_base_url')}
              onSave={(val) => handleUpdate('ai_base_url', val)}
            />
            <ConfigInputItem
              label="API Key"
              desc="调用 AI 接口所需的密钥 (sk-...)"
              value={getStringValue('ai_api_key')}
              onSave={(val) => handleUpdate('ai_api_key', val)}
              isSecret
            />
          </div>
        </div>
      )}

      <div className="bg-white rounded-2xl border border-slate-200 shadow-sm overflow-hidden mt-6">
        <h3 className="px-6 py-4 font-bold text-slate-700 bg-slate-50/50 border-b border-slate-100">支付参数配置 (ZhifuFM)</h3>
        <div className="divide-y divide-slate-100">
          <SettingItem
            title="启用支付宝支付"
            desc="开启后，用户可选择支付宝进行支付。"
            toggle={getValue('payment_enable_alipay')}
            onToggle={() => handleToggle('payment_enable_alipay', String(getValue('payment_enable_alipay')))}
            icon={<CreditCard className="text-blue-500" />}
          />
          <SettingItem
            title="启用微信支付"
            desc="开启后，用户可选择微信进行支付。"
            toggle={getValue('payment_enable_wechat')}
            onToggle={() => handleToggle('payment_enable_wechat', String(getValue('payment_enable_wechat')))}
            icon={<CreditCard className="text-emerald-500" />}
          />
          <ConfigInputItem
            label="商户号 (Merchant Num)"
            desc="您的 ZhifuFM 商户号"
            value={getStringValue('payment_merchant_num')}
            onSave={(val) => handleUpdate('payment_merchant_num', val)}
          />
          <ConfigInputItem
            label="商户密钥 (Merchant Key)"
            desc="您的 ZhifuFM 接口密钥"
            value={getStringValue('payment_merchant_key')}
            onSave={(val) => handleUpdate('payment_merchant_key', val)}
            isSecret
          />
          <ConfigInputItem
            label="接口地址 (Base URL)"
            desc="ZhifuFM 接口根地址"
            value={getStringValue('payment_base_url')}
            onSave={(val) => handleUpdate('payment_base_url', val)}
          />
          <ConfigInputItem
            label="异步通知地址 (Notify URL)"
            desc="支付成功后回调地址 (必须公网可达)"
            value={getStringValue('payment_notify_url')}
            onSave={(val) => handleUpdate('payment_notify_url', val)}
          />
          <ConfigInputItem
            label="同步跳转地址 (Return URL)"
            desc="支付完成后跳转的地址 (前端)"
            value={getStringValue('payment_return_url')}
            onSave={(val) => handleUpdate('payment_return_url', val)}
          />
        </div>
      </div>
    </div>
  );
};

const SettingItem = ({ title, desc, toggle, onToggle, icon }: any) => (
  <div className="p-6 flex items-center justify-between">
    <div className="flex items-start gap-4">
      <div className="p-2 bg-slate-100 rounded-lg shrink-0">{icon}</div>
      <div>
        <h4 className="font-bold text-slate-800">{title}</h4>
        <p className="text-sm text-slate-500">{desc}</p>
      </div>
    </div>
    <div
      onClick={onToggle}
      className={`w-12 h-6 rounded-full p-1 cursor-pointer transition-colors ${toggle ? 'bg-indigo-600' : 'bg-slate-200'}`}
    >
      <div className={`w-4 h-4 bg-white rounded-full shadow-sm transition-transform ${toggle ? 'translate-x-6' : 'translate-x-0'}`}></div>
    </div>
  </div>
);

const ConfigInputItem = ({ label, desc, value, onSave, isSecret }: any) => {
  const [isEditing, setIsEditing] = useState(false);
  const [tempValue, setTempValue] = useState(value);

  React.useEffect(() => {
    setTempValue(value);
  }, [value]);

  const handleSave = () => {
    onSave(tempValue);
    setIsEditing(false);
  }

  return (
    <div className="p-6">
      <div className="flex justify-between items-start mb-2">
        <div>
          <h4 className="font-bold text-slate-800">{label}</h4>
          <p className="text-sm text-slate-500">{desc}</p>
        </div>
        {!isEditing && (
          <button onClick={() => setIsEditing(true)} className="text-indigo-600 font-bold text-sm hover:underline">
            编辑
          </button>
        )}
      </div>

      {isEditing ? (
        <div className="mt-3 flex gap-3">
          <input
            type={isSecret ? "password" : "text"}
            className="flex-1 px-4 py-2 border border-slate-300 rounded-lg focus:outline-none focus:border-indigo-500 text-sm"
            value={tempValue}
            onChange={e => setTempValue(e.target.value)}
            placeholder={`请输入 ${label}`}
          />
          <button onClick={handleSave} className="bg-indigo-600 text-white px-4 py-2 rounded-lg text-sm font-bold hover:bg-indigo-700">保存</button>
          <button onClick={() => { setIsEditing(false); setTempValue(value); }} className="bg-slate-100 text-slate-600 px-4 py-2 rounded-lg text-sm font-bold hover:bg-slate-200">取消</button>
        </div>
      ) : (
        <div className="mt-2 text-sm text-slate-700 font-mono bg-slate-50 p-2 rounded border border-slate-100 break-all">
          {isSecret ? (value ? '••••••••••••••••••••••••' : '未设置') : (value || '未设置')}
        </div>
      )}
    </div>
  )
}



const AddUserModal = ({ onClose, onSuccess }: { onClose: () => void, onSuccess: () => void }) => {
  const [formData, setFormData] = useState({
    name: '',
    email: '',
    password: '',
    role: 'user'
  });
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!formData.name || !formData.email || !formData.password) {
      alert('请填写完整信息');
      return;
    }
    try {
      setLoading(true);
      await api.user.createUser(formData);
      alert('用户创建成功');
      onSuccess();
    } catch (err: any) {
      alert(err.message || '创建失败');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 backdrop-blur-sm">
      <div className="bg-white rounded-2xl w-full max-w-md p-6 shadow-2xl animate-in fade-in zoom-in duration-200">
        <h3 className="text-xl font-bold text-slate-800 mb-4">新增用户</h3>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-sm font-bold text-slate-700 mb-1">用户名</label>
            <input
              type="text"
              className="w-full px-4 py-2 border border-slate-200 rounded-lg focus:outline-none focus:border-indigo-500"
              value={formData.name}
              onChange={e => setFormData({ ...formData, name: e.target.value })}
              placeholder="请输入用户名"
            />
          </div>
          <div>
            <label className="block text-sm font-bold text-slate-700 mb-1">邮箱</label>
            <input
              type="email"
              className="w-full px-4 py-2 border border-slate-200 rounded-lg focus:outline-none focus:border-indigo-500"
              value={formData.email}
              onChange={e => setFormData({ ...formData, email: e.target.value })}
              placeholder="请输入邮箱"
            />
          </div>
          <div>
            <label className="block text-sm font-bold text-slate-700 mb-1">初始密码</label>
            <input
              type="text"
              className="w-full px-4 py-2 border border-slate-200 rounded-lg focus:outline-none focus:border-indigo-500"
              value={formData.password}
              onChange={e => setFormData({ ...formData, password: e.target.value })}
              placeholder="设置初始密码"
            />
          </div>
          <div>
            <label className="block text-sm font-bold text-slate-700 mb-1">角色</label>
            <select
              className="w-full px-4 py-2 border border-slate-200 rounded-lg focus:outline-none focus:border-indigo-500 bg-white"
              value={formData.role}
              onChange={e => setFormData({ ...formData, role: e.target.value })}
            >
              <option value="user">普通用户</option>
              <option value="admin">管理员</option>
            </select>
          </div>
          <div className="flex gap-3 pt-4">
            <button
              type="button"
              onClick={onClose}
              className="flex-1 py-2 rounded-xl text-slate-600 hover:bg-slate-100 font-bold"
            >
              取消
            </button>
            <button
              type="submit"
              disabled={loading}
              className="flex-1 py-2 rounded-xl bg-indigo-600 text-white font-bold hover:bg-indigo-700 disabled:opacity-50"
            >
              {loading ? '创建中...' : '确认创建'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

export default AdminView;

const ProductModal = ({ product, onClose, onSave, wikiCategories = [], quizSubjects = [] }: any) => {
  const [formData, setFormData] = useState(product || {
    id: '',
    title: '',
    description: '',
    price: 0,
    duration: '1年',
    durationValue: 1,
    durationUnit: 'year',
    imageUrl: 'https://images.unsplash.com/photo-1576091160399-112ba8d25d1d?auto=format&fit=crop&q=80&w=400',
    tags: [],
    accessId: ''
  });

  // Determine initial link type
  const getInitialType = (accId: string) => {
    if (accId.startsWith('wiki_')) return 'wiki';
    if (accId.startsWith('quiz_')) return 'quiz';
    return 'custom';
  };

  const [linkType, setLinkType] = useState<'wiki' | 'quiz' | 'custom'>(
    product ? getInitialType(product.accessId) : 'custom'
  );

  const [coupons, setCoupons] = useState<any[]>(product?.coupons || []);
  const [newCoupon, setNewCoupon] = useState({
    code: '',
    type: 'amount', // or 'percent'
    value: 0,
    usageLimit: '' // string for input, parse to number or null
  });

  const handleAddCoupon = async () => {
    if (!product || !product.id) return;
    if (!newCoupon.code || !newCoupon.value) return;

    try {
      const payload = {
        ...newCoupon,
        usageLimit: newCoupon.usageLimit ? Number(newCoupon.usageLimit) : null
      };
      const added = await api.store.addCoupon(product.id, payload);
      setCoupons([...coupons, added]);
      setNewCoupon({ code: '', type: 'amount', value: 0, usageLimit: '' });
    } catch (e: any) {
      alert('添加优惠码失败: ' + e.message);
    }
  };

  const handleDeleteCoupon = async (id: string) => {
    if (!confirm('确定删除此优惠码？')) return;
    try {
      await api.store.deleteCoupon(id);
      setCoupons(coupons.filter(c => c.id !== id));
    } catch (e: any) {
      alert('删除失败');
    }
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSave(formData);
  };

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 backdrop-blur-sm">
      <div className="bg-white rounded-2xl w-full max-w-lg p-6 shadow-2xl animate-in fade-in zoom-in duration-200 max-h-[90vh] overflow-y-auto">
        <h3 className="text-xl font-bold text-slate-800 mb-4">{product ? '编辑商品' : '新增商品'}</h3>
        <form onSubmit={handleSubmit} className="space-y-4">
          {!product && (
            <div>
              <label className="block text-sm font-bold text-slate-700 mb-1">商品ID (可选，默认自动生成)</label>
              <input
                type="text"
                className="w-full px-4 py-2 border border-slate-200 rounded-lg focus:outline-none focus:border-indigo-500"
                value={formData.id || ''}
                onChange={e => setFormData({ ...formData, id: e.target.value })}
                placeholder="prod_xxxx"
              />
            </div>
          )}
          <div>
            <label className="block text-sm font-bold text-slate-700 mb-1">商品名称</label>
            <input
              type="text"
              required
              className="w-full px-4 py-2 border border-slate-200 rounded-lg focus:outline-none focus:border-indigo-500"
              value={formData.title}
              onChange={e => setFormData({ ...formData, title: e.target.value })}
            />
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-bold text-slate-700 mb-1">价格 (¥)</label>
              <input
                type="number"
                required
                className="w-full px-4 py-2 border border-slate-200 rounded-lg focus:outline-none focus:border-indigo-500"
                value={formData.price}
                onChange={e => setFormData({ ...formData, price: Number(e.target.value) })}
              />
            </div>
            <div>
              <label className="block text-sm font-bold text-slate-700 mb-1">时长</label>
              <div className="flex gap-2">
                <input
                  type="number"
                  min="1"
                  className="w-24 px-4 py-2 border border-slate-200 rounded-lg focus:outline-none focus:border-indigo-500"
                  value={formData.durationValue || 1}
                  onChange={e => {
                    const val = Number(e.target.value);
                    const unit = formData.durationUnit || 'month';
                    const map: any = { day: '天', month: '个月', year: '年', forever: '永久' };
                    const dur = unit === 'forever' ? '永久' : `${val}${map[unit]}`;
                    setFormData({ ...formData, durationValue: val, duration: dur });
                  }}
                />
                <select
                  className="flex-1 px-4 py-2 border border-slate-200 rounded-lg focus:outline-none focus:border-indigo-500 bg-white"
                  value={formData.durationUnit || 'month'}
                  onChange={e => {
                    const unit = e.target.value;
                    const val = formData.durationValue || 1;
                    const map: any = { day: '天', month: '个月', year: '年', forever: '永久' };
                    const dur = unit === 'forever' ? '永久' : `${val}${map[unit]}`;
                    setFormData({ ...formData, durationUnit: unit, duration: dur });
                  }}
                >
                  <option value="day">天</option>
                  <option value="month">个月</option>
                  <option value="year">年</option>
                  <option value="forever">永久</option>
                </select>
              </div>
            </div>
          </div>

          {/* Enhanced Access Control Section */}
          <div className="bg-slate-50 p-4 rounded-xl border border-slate-200">
            <label className="block text-sm font-bold text-slate-700 mb-2">权益关联 (Access)</label>

            <div className="flex gap-2 mb-3">
              <button
                type="button"
                onClick={() => setLinkType('wiki')}
                className={`flex-1 py-1.5 text-xs font-bold rounded-lg border ${linkType === 'wiki' ? 'bg-indigo-100 text-indigo-700 border-indigo-200' : 'bg-white text-slate-500 border-slate-200'}`}
              >
                知识库
              </button>
              <button
                type="button"
                onClick={() => setLinkType('quiz')}
                className={`flex-1 py-1.5 text-xs font-bold rounded-lg border ${linkType === 'quiz' ? 'bg-indigo-100 text-indigo-700 border-indigo-200' : 'bg-white text-slate-500 border-slate-200'}`}
              >
                题库
              </button>
              <button
                type="button"
                onClick={() => setLinkType('custom')}
                className={`flex-1 py-1.5 text-xs font-bold rounded-lg border ${linkType === 'custom' ? 'bg-indigo-100 text-indigo-700 border-indigo-200' : 'bg-white text-slate-500 border-slate-200'}`}
              >
                自定义
              </button>
            </div>

            {linkType === 'wiki' && (
              <select
                className="w-full px-4 py-2 border border-slate-200 rounded-lg text-sm bg-white"
                onChange={e => {
                  if (e.target.value) setFormData({ ...formData, accessId: `wiki_${e.target.value}` });
                }}
                value={formData.accessId.replace('wiki_', '')}
              >
                <option value="">-- 选择关联的知识库分类 --</option>
                {wikiCategories.map((c: any) => (
                  <option key={c.id} value={c.id}>{c.title}</option>
                ))}
              </select>
            )}

            {linkType === 'quiz' && (
              <select
                className="w-full px-4 py-2 border border-slate-200 rounded-lg text-sm bg-white"
                onChange={e => {
                  if (e.target.value) setFormData({ ...formData, accessId: `quiz_${e.target.value}` });
                }}
                value={formData.accessId.replace('quiz_', '')}
              >
                <option value="">-- 选择关联的题库科目 --</option>
                {quizSubjects.map((s: any) => (
                  <option key={s.id} value={s.id}>{s.title}</option>
                ))}
              </select>
            )}

            {linkType === 'custom' && (
              <input
                type="text"
                required
                className="w-full px-4 py-2 border border-slate-200 rounded-lg focus:outline-none focus:border-indigo-500"
                value={formData.accessId}
                onChange={e => setFormData({ ...formData, accessId: e.target.value })}
                placeholder="e.g. quiz_internal"
              />
            )}

            <p className="mt-2 text-xs text-slate-400 font-mono">
              Current Access ID: {formData.accessId}
            </p>
          </div>
          <div>
            <label className="block text-sm font-bold text-slate-700 mb-1">描述</label>
            <textarea
              required
              className="w-full px-4 py-2 border border-slate-200 rounded-lg focus:outline-none focus:border-indigo-500 h-24"
              value={formData.description}
              onChange={e => setFormData({ ...formData, description: e.target.value })}
            />
          </div>
          <div>
            <label className="block text-sm font-bold text-slate-700 mb-1">图片URL</label>
            <input
              type="text"
              required
              className="w-full px-4 py-2 border border-slate-200 rounded-lg focus:outline-none focus:border-indigo-500"
              value={formData.imageUrl}
              onChange={e => setFormData({ ...formData, imageUrl: e.target.value })}
            />
          </div>
          <div>
            <label className="block text-sm font-bold text-slate-700 mb-1">标签 (逗号分隔)</label>
            <input
              type="text"
              className="w-full px-4 py-2 border border-slate-200 rounded-lg focus:outline-none focus:border-indigo-500"
              value={Array.isArray(formData.tags) ? formData.tags.join(',') : formData.tags}
              onChange={e => setFormData({ ...formData, tags: e.target.value.split(',').map((t: string) => t.trim()) })}
              placeholder="题库, 内科, 推荐"
            />
          </div>

          {/* Coupon Management Section */}
          {product && (
            <div className="border-t border-slate-100 pt-4">
              <label className="block text-sm font-bold text-slate-700 mb-2">优惠码管理</label>

              <div className="flex gap-2 mb-3 items-end">
                <div className="flex-1">
                  <input
                    type="text"
                    placeholder="代码 (e.g. SALE)"
                    className="w-full px-3 py-2 border border-slate-200 rounded-lg text-sm"
                    value={newCoupon.code}
                    onChange={e => setNewCoupon({ ...newCoupon, code: e.target.value })}
                  />
                </div>
                <div className="w-24">
                  <select
                    className="w-full px-3 py-2 border border-slate-200 rounded-lg text-sm bg-white"
                    value={newCoupon.type}
                    onChange={e => setNewCoupon({ ...newCoupon, type: e.target.value as any })}
                  >
                    <option value="amount">减免(¥)</option>
                    <option value="percent">折扣(%)</option>
                  </select>
                </div>
                <div className="w-20">
                  <input
                    type="number"
                    placeholder="值"
                    className="w-full px-3 py-2 border border-slate-200 rounded-lg text-sm"
                    value={newCoupon.value || ''}
                    onChange={e => setNewCoupon({ ...newCoupon, value: Number(e.target.value) })}
                  />
                </div>
                <div className="w-20">
                  <input
                    type="number"
                    placeholder="次数"
                    className="w-full px-3 py-2 border border-slate-200 rounded-lg text-sm"
                    value={newCoupon.usageLimit}
                    onChange={e => setNewCoupon({ ...newCoupon, usageLimit: e.target.value })}
                  />
                </div>
                <button
                  type="button"
                  onClick={handleAddCoupon}
                  className="px-3 py-2 bg-indigo-600 text-white rounded-lg text-sm font-bold hover:bg-indigo-700"
                >
                  添加
                </button>
              </div>

              <div className="space-y-2 max-h-40 overflow-y-auto">
                {coupons.map((c: any) => (
                  <div key={c.id} className="flex justify-between items-center p-2 bg-slate-50 rounded-lg border border-slate-200 text-sm">
                    <div className="flex items-center gap-2">
                      <span className="font-mono font-bold text-indigo-600">{c.code}</span>
                      <span className="text-slate-500 text-xs">
                        {c.type === 'amount' ? `减¥${c.value}` : `${c.value}% OFF`}
                      </span>
                      <span className="text-slate-400 text-xs">
                        已用: {c.usedCount} / {c.usageLimit === null ? '∞' : c.usageLimit}
                      </span>
                    </div>
                    <button
                      type="button"
                      onClick={() => handleDeleteCoupon(c.id)}
                      className="text-rose-400 hover:text-rose-600 p-1"
                    >
                      <Trash2 size={14} />
                    </button>
                  </div>
                ))}
                {coupons.length === 0 && <div className="text-center text-xs text-slate-400 py-2">暂无优惠码</div>}
              </div>
            </div>
          )}

          <div className="flex gap-3 pt-4 border-t border-slate-100">
            <button
              type="button"
              onClick={onClose}
              className="flex-1 py-2 rounded-xl text-slate-600 hover:bg-slate-100 font-bold"
            >
              取消
            </button>
            <button
              type="submit"
              className="flex-1 py-2 rounded-xl bg-indigo-600 text-white font-bold hover:bg-indigo-700"
            >
              保存
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

const UserDetailModal = ({ user, onClose, onUpdate }: { user: any, onClose: () => void, onUpdate: () => void }) => {
  const [subscriptions, setSubscriptions] = useState<any[]>(user.stats.subscriptions || []);
  const [availableProducts, setAvailableProducts] = useState<any[]>([]);
  const [selectedProduct, setSelectedProduct] = useState('');
  const [loading, setLoading] = useState(false);

  React.useEffect(() => {
    api.store.getProducts().then(setAvailableProducts).catch(console.error);
  }, []);

  const handleUpdate = async (newSubs: any[]) => {
    try {
      setLoading(true);
      await api.user.adminUpdateSubscriptions(user.id, newSubs);
      setSubscriptions(newSubs);
      // Wait a bit to ensure server updates then call refresh
      onUpdate();
    } catch (err) {
      alert('更新失败');
    } finally {
      setLoading(false);
    }
  };

  const addSub = () => {
    if (!selectedProduct) return;
    // Check if already exists (by accessId)
    const exists = subscriptions.some(s => {
      const sId = typeof s === 'string' ? s : s.accessId;
      // Note: selectedProduct here is the ID or AccessId from the value of select
      // But the select value might be product ID, while subscription stores AccessId usually?
      // Actually AdminView ProductModal saves accessId. store service checkouts accessId?
      // Wait, StoreService.checkout saves subscriptions with accessId derived from product.
      // The select options values are p.id.
      // But we should probably store what the system expects. 
      // Existing system seemed to mix product IDs and AccessIDs?
      // Let's see StoreView... addToCart adds Product. checkout uses item.accessId.
      // So subscriptions store AccessID.
      // The Options in UserDetailModal map p.id as value. We should probably use p.accessId if possible, or lookup.
      return sId === selectedProduct;
    });

    // We need to find the product to get the correct accessId if selectedProduct is just an ID.
    const product = availableProducts.find(p => p.id === selectedProduct);
    const accessIdToSave = product ? product.accessId : selectedProduct;

    // Re-check existence with accessId
    const actuallyExists = subscriptions.some(s => {
      const sId = typeof s === 'string' ? s : s.accessId;
      return sId === accessIdToSave;
    });

    if (actuallyExists) {
      alert("已包含该订阅");
      return;
    }

    // Add as object with default duration? Or just string for infinite?
    // Let's add as object for consistency, default to 1 year maybe? Or forever?
    // Let's Default to forever (null expiry) if added manually by admin, or maybe ask?
    // For simplicity: Add as object with null expiry (Forever).

    // If product has duration, we could use it?
    let expiresAt = null;
    if (product && product.durationValue && product.durationUnit) {
      // Calculate expiry
      const date = new Date();
      if (product.durationUnit === 'day') date.setDate(date.getDate() + product.durationValue);
      if (product.durationUnit === 'month') date.setMonth(date.getMonth() + product.durationValue);
      if (product.durationUnit === 'year') date.setFullYear(date.getFullYear() + product.durationValue);
      if (product.durationUnit !== 'forever') expiresAt = date.toISOString();
    }

    handleUpdate([...subscriptions, { accessId: accessIdToSave, startDate: new Date().toISOString(), expiresAt }]);
    setSelectedProduct('');
  };

  return (
    <div className="fixed inset-0 bg-black/50 z-50 flex items-center justify-center p-4">
      <div className="bg-white rounded-2xl w-full max-w-lg p-6 max-h-[90vh] overflow-y-auto shadow-2xl">
        <div className="flex justify-between items-center mb-6 border-b border-slate-100 pb-4">
          <h3 className="text-xl font-bold text-slate-800">用户详情</h3>
          <button onClick={onClose} className="p-2 hover:bg-slate-100 rounded-lg text-slate-500 hover:text-slate-800 transition-colors"><X size={20} /></button>
        </div>

        <div className="space-y-6">
          <div className="bg-slate-50 p-4 rounded-xl space-y-3 border border-slate-100">
            <div className="flex justify-between text-sm"><span className="text-slate-500">用户 ID</span> <span className="font-mono text-xs bg-white px-2 py-0.5 rounded border border-slate-200">{user.id}</span></div>
            <div className="flex justify-between text-sm"><span className="text-slate-500">姓名</span> <span className="font-bold text-slate-800">{user.name}</span></div>
            <div className="flex justify-between text-sm"><span className="text-slate-500">邮箱</span> <span className="text-slate-800">{user.email}</span></div>
            <div className="flex justify-between text-sm"><span className="text-slate-500">注册时间</span> <span className="text-slate-800">{new Date(user.createdAt).toLocaleDateString()}</span></div>
          </div>

          <div>
            <h4 className="font-bold text-slate-800 mb-3 flex items-center gap-2 text-sm uppercase tracking-wide"><CreditCard size={18} className="text-indigo-600" /> 订阅管理</h4>
            <div className="flex gap-2 mb-4">
              <select
                className="flex-1 bg-white border border-slate-200 rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500"
                value={selectedProduct}
                onChange={e => setSelectedProduct(e.target.value)}
              >
                <option value="">选择订阅/商品...</option>
                {availableProducts.map(p => (
                  <option key={p.id} value={p.id}>{p.title}</option>
                ))}
                <option value="unlimited_30">VIP 30天 (Admin Only)</option>
                <option value="unlimited_365">VIP 1年 (Admin Only)</option>
              </select>
              <button
                onClick={addSub}
                disabled={!selectedProduct || loading}
                className="bg-indigo-600 text-white px-4 py-2 rounded-lg font-bold hover:bg-indigo-700 disabled:opacity-50 text-sm transition-colors"
              >
                添加
              </button>
            </div>

            <div className="space-y-2">
              {subscriptions.length === 0 ? (
                <div className="text-center py-8 text-slate-400 bg-slate-50 rounded-xl border border-dashed border-slate-200 text-sm">
                  暂无订阅
                </div>
              ) : (
                subscriptions.map((sub: any, idx) => {
                  const accessId = typeof sub === 'string' ? sub : sub.accessId;
                  const expiresAt = typeof sub === 'object' ? sub.expiresAt : null;

                  // Find product by ID or AccessID
                  const product = availableProducts.find(p => p.id === accessId || p.accessId === accessId);
                  const label = product ? product.title : accessId;

                  // Sort of safe remove: filter out this specific item
                  // But since we don't have unique IDs for subs necessarily, we filter by index or value

                  return (
                    <div key={idx} className="flex justify-between items-center p-3 bg-white border border-slate-200 rounded-xl shadow-sm hover:shadow transition-shadow">
                      <div>
                        <div className="font-medium text-slate-700 text-sm">{label}</div>
                        <div className="text-xs text-slate-400">
                          ID: {accessId}
                          {expiresAt ? ` · 到期: ${new Date(expiresAt).toLocaleDateString()}` : ' · 永久'}
                        </div>
                      </div>
                      <button
                        onClick={() => {
                          if (!confirm('确定删除此订阅?')) return;
                          const newSubs = [...subscriptions];
                          newSubs.splice(idx, 1);
                          handleUpdate(newSubs);
                        }}
                        className="text-slate-400 hover:text-rose-500 hover:bg-rose-50 p-1.5 rounded-lg transition-colors"
                      >
                        <Trash2 size={16} />
                      </button>
                    </div>
                  );
                })
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

const SubjectModal = ({ subject, onClose, onSave }: any) => {
  const [formData, setFormData] = useState(subject || {
    id: '',
    title: '',
    description: '',
    icon: '📚',
    color: '#6366f1'
  });

  const isEditing = !!subject;

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSave(formData);
  };

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 backdrop-blur-sm">
      <div className="bg-white rounded-2xl w-full max-w-md p-6 shadow-2xl animate-in fade-in zoom-in duration-200">
        <h3 className="text-xl font-bold text-slate-800 mb-4">{isEditing ? '编辑科目' : '新增科目'}</h3>
        <form onSubmit={handleSubmit} className="space-y-4">
          {!isEditing && (
            <div>
              <label className="block text-sm font-bold text-slate-700 mb-1">科目 ID (英文)</label>
              <input
                type="text"
                required
                className="w-full px-4 py-2 border border-slate-200 rounded-lg focus:outline-none focus:border-indigo-500"
                value={formData.id}
                onChange={e => setFormData({ ...formData, id: e.target.value })}
                placeholder="e.g. internal-medicine"
              />
              <p className="text-xs text-slate-400 mt-1">ID 设置后不可修改，用于 URL</p>
            </div>
          )}
          <div>
            <label className="block text-sm font-bold text-slate-700 mb-1">科目名称</label>
            <input
              type="text"
              required
              className="w-full px-4 py-2 border border-slate-200 rounded-lg focus:outline-none focus:border-indigo-500"
              value={formData.title}
              onChange={e => setFormData({ ...formData, title: e.target.value })}
              placeholder="e.g. 内科学"
            />
          </div>
          <div>
            <label className="block text-sm font-bold text-slate-700 mb-1">描述</label>
            <textarea
              className="w-full px-4 py-2 border border-slate-200 rounded-lg focus:outline-none focus:border-indigo-500"
              value={formData.description}
              onChange={e => setFormData({ ...formData, description: e.target.value })}
              placeholder="简短描述..."
            />
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-bold text-slate-700 mb-1">图标 (Emoji)</label>
              <input
                type="text"
                className="w-full px-4 py-2 border border-slate-200 rounded-lg focus:outline-none focus:border-indigo-500"
                value={formData.icon}
                onChange={e => setFormData({ ...formData, icon: e.target.value })}
                placeholder="📚"
              />
            </div>
            <div>
              <label className="block text-sm font-bold text-slate-700 mb-1">颜色 (HEX)</label>
              <input
                type="color"
                className="w-full h-10 px-1 py-1 border border-slate-200 rounded-lg cursor-pointer"
                value={formData.color}
                onChange={e => setFormData({ ...formData, color: e.target.value })}
              />
            </div>
          </div>
          <div className="flex gap-3 pt-4">
            <button
              type="button"
              onClick={onClose}
              className="flex-1 py-2 rounded-xl text-slate-600 hover:bg-slate-100 font-bold"
            >
              取消
            </button>
            <button
              type="submit"
              className="flex-1 py-2 rounded-xl bg-indigo-600 text-white font-bold hover:bg-indigo-700"
            >
              保存
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

function AnnouncementManager() {
  const [announcements, setAnnouncements] = useState<any[]>([]);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingItem, setEditingItem] = useState<any>(null);

  React.useEffect(() => {
    loadAnnouncements();
  }, []);

  const loadAnnouncements = async () => {
    try {
      const data = await api.announcements.getAll();
      setAnnouncements(data);
    } catch (e) {
      console.error(e);
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm('确定删除该公告？')) return;
    try {
      await api.announcements.delete(id);
      loadAnnouncements();
    } catch (e) {
      alert('删除失败');
    }
  };

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault();
    const form = e.target as HTMLFormElement;
    const data = {
      title: (form.elements.namedItem('title') as HTMLInputElement).value,
      content: (form.elements.namedItem('content') as HTMLTextAreaElement).value,
      visible: (form.elements.namedItem('visible') as HTMLInputElement).checked
    };

    try {
      if (editingItem) {
        await api.announcements.update(editingItem.id, data);
      } else {
        await api.announcements.create(data);
      }
      setIsModalOpen(false);
      setEditingItem(null);
      loadAnnouncements();
    } catch (err: any) {
      alert('保存失败: ' + err.message);
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex justify-end">
        <button
          onClick={() => { setEditingItem(null); setIsModalOpen(true); }}
          className="bg-amber-600 text-white px-4 py-2 rounded-lg text-sm font-bold flex items-center gap-2 hover:bg-amber-700 shadow-lg shadow-amber-200"
        >
          <Plus size={16} /> 发布公告
        </button>
      </div>

      <div className="bg-white border border-slate-200 rounded-2xl overflow-hidden shadow-sm">
        <table className="w-full text-left border-collapse">
          <thead>
            <tr className="bg-slate-50 text-slate-500 text-xs uppercase tracking-wider">
              <th className="px-6 py-4 font-semibold">标题</th>
              <th className="px-6 py-4 font-semibold">状态</th>
              <th className="px-6 py-4 font-semibold">发布时间</th>
              <th className="px-6 py-4 font-semibold text-right">操作</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-100 text-sm">
            {announcements.length === 0 ? (
              <tr><td colSpan={4} className="text-center py-8 text-slate-500">暂无公告</td></tr>
            ) : announcements.map(item => (
              <tr key={item.id} className="hover:bg-slate-50 transition-colors">
                <td className="px-6 py-4 font-medium text-slate-800">{item.title}</td>
                <td className="px-6 py-4">
                  <span className={`px-2 py-0.5 rounded text-xs ${item.visible ? 'bg-emerald-50 text-emerald-600' : 'bg-slate-100 text-slate-500'}`}>
                    {item.visible ? '已发布' : '隐藏'}
                  </span>
                </td>
                <td className="px-6 py-4 text-slate-500">{new Date(item.createdAt).toLocaleDateString()}</td>
                <td className="px-6 py-4 text-right">
                  <div className="flex gap-2 justify-end">
                    <button
                      onClick={() => { setEditingItem(item); setIsModalOpen(true); }}
                      className="p-2 text-slate-400 hover:text-indigo-600 hover:bg-indigo-50 rounded-lg"
                    >
                      <Edit3 size={16} />
                    </button>
                    <button
                      onClick={() => handleDelete(item.id)}
                      className="p-2 text-slate-400 hover:text-rose-600 hover:bg-rose-50 rounded-lg"
                    >
                      <Trash2 size={16} />
                    </button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {isModalOpen && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 backdrop-blur-sm">
          <div className="bg-white rounded-2xl w-full max-w-lg p-6 shadow-2xl animate-in fade-in zoom-in duration-200">
            <h3 className="text-xl font-bold text-slate-800 mb-4">{editingItem ? '编辑公告' : '发布公告'}</h3>
            <form onSubmit={handleSave} className="space-y-4">
              <div>
                <label className="block text-sm font-bold text-slate-700 mb-1">标题</label>
                <input
                  name="title"
                  type="text"
                  required
                  defaultValue={editingItem?.title}
                  className="w-full px-4 py-2 border border-slate-200 rounded-lg focus:outline-none focus:border-indigo-500"
                  placeholder="输入公告标题..."
                />
              </div>
              <div>
                <label className="block text-sm font-bold text-slate-700 mb-1">内容</label>
                <textarea
                  name="content"
                  required
                  rows={6}
                  defaultValue={editingItem?.content}
                  className="w-full px-4 py-2 border border-slate-200 rounded-lg focus:outline-none focus:border-indigo-500 resize-none"
                  placeholder="输入公告内容..."
                />
              </div>
              <div className="flex items-center gap-2">
                <input
                  name="visible"
                  type="checkbox"
                  id="visible"
                  defaultChecked={editingItem ? editingItem.visible : true}
                  className="w-4 h-4 text-indigo-600 rounded border-gray-300 focus:ring-indigo-500"
                />
                <label htmlFor="visible" className="text-sm text-slate-700">立即发布</label>
              </div>
              <div className="flex gap-3 pt-4">
                <button
                  type="button"
                  onClick={() => setIsModalOpen(false)}
                  className="flex-1 py-2 rounded-xl text-slate-600 hover:bg-slate-100 font-bold"
                >
                  取消
                </button>
                <button
                  type="submit"
                  className="flex-1 py-2 rounded-xl bg-amber-600 text-white font-bold hover:bg-amber-700"
                >
                  保存
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};
