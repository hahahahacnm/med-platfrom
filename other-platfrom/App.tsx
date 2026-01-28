
import React, { useState, useMemo, useEffect } from 'react';
import { AnimatePresence, motion } from 'framer-motion';
import {
  LayoutDashboard, ShoppingBag, BookOpen, BrainCircuit,
  User, Menu, X, CheckCircle, Star, Eraser, Bot, Shield, MessageSquare, LogOut
} from 'lucide-react';
import { UserState, ViewState, CartItem, Product, QuizResult, AccessId, Bookmark, WikiCategory, Article } from './types';
import { AppContext } from './context';
import { AIChatBot } from './components/AIChatBot';
import { api } from './services/api';

// Views
import StoreView from './components/StoreView';
import QuizView from './components/QuizView';
import WikiView from './components/WikiView';
import ProfileView from './components/ProfileView';
import FavoritesView from './components/FavoritesView';
import MistakesView from './components/MistakesView';
import AdminView from './components/AdminView';
import FeedbackView from './components/FeedbackView';
import AuthView from './components/AuthView';

// --- Main App Component ---
const App: React.FC = () => {
  // State
  const [view, setView] = useState<ViewState>('home');
  const [cart, setCart] = useState<CartItem[]>([]);
  const [user, setUser] = useState<UserState>({
    isAuthenticated: false,
    name: '',
    email: '',
    role: 'user',
    subscriptions: [], // Start with NO access
    balance: 0,
    quizHistory: [],
    chapterProgress: {},
    bookmarks: []
  });

  const [aiEnabled, setAiEnabled] = useState(true);

  const fetchSettings = async () => {
    try {
      const settings = await api.settings.getAll();
      const aiSetting = settings.find((s: any) => s.key === 'ai_enabled');
      if (aiSetting) {
        setAiEnabled(aiSetting.value === 'true');
      }
    } catch (e) {
      console.error('Failed to load settings', e);
    }
  };

  // Wiki Navigation State (Lifted)
  const [wikiState, setWikiState] = useState<{ category: WikiCategory | null; article: Article | null }>({
    category: null,
    article: null
  });

  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);

  // Authentication Actions
  // Authentication Actions
  const login = (userData: any) => {
    setUser(prev => ({
      ...prev,
      ...userData,
      isAuthenticated: true,
      // Ensure arrays are arrays (backend safeguards)
      subscriptions: userData.subscriptions || [],
      quizHistory: userData.quizHistory || [],
      bookmarks: userData.bookmarks || [],
    }));
    setView('home');
  };

  const logout = () => {
    api.auth.logout();
    setUser({
      isAuthenticated: false,
      name: '',
      email: '',
      role: 'user',
      subscriptions: [],
      balance: 0,
      quizHistory: [],
      chapterProgress: {},
      bookmarks: []
    });
    setView('home');
  };

  useEffect(() => {
    const initAuth = async () => {
      try {
        if (localStorage.getItem('token')) {
          const userData = await api.auth.profile();
          login(userData);
        }
      } catch (e) {
        localStorage.removeItem('token');
      }
    };
    initAuth();
    fetchSettings();

    // Periodic session check
    const interval = setInterval(async () => {
      if (user.isAuthenticated) {
        try {
          await api.auth.checkSession();
        } catch (e: any) {
          // If checkSession fails (e.g. 401), handleResponse will clear token
          // Here we just need to ensure UI reflects logout
          if (!localStorage.getItem('token')) {
            logout();
            // alert('您已在其他设备登录，当前会话已失效'); // Optional: notify user
          }
        }
      }
    }, 5000); // Check every 5 seconds

    return () => clearInterval(interval);
  }, [user.isAuthenticated]);

  // Actions
  const addToCart = (product: Product) => {
    if (cart.some(item => item.id === product.id)) return;
    setCart(prev => [...prev, { ...product, cartId: `${product.id}-${Date.now()}` }]);
  };

  const removeFromCart = (cartId: string) => {
    setCart(prev => prev.filter(item => item.cartId !== cartId));
  };

  const checkout = async (couponCode?: string, payType?: string) => {
    const newAccessIds = cart.map(item => item.accessId);
    const amount = cart.reduce((sum, item) => sum + item.price, 0);

    // Backend checkout (Records transaction)
    if (user.isAuthenticated) {
      try {
        const res = await api.store.checkout(cart, amount, couponCode, payType);

        if (res.payUrl) {
          // Redirect to payment
          window.location.href = res.payUrl;
          return;
        }

        // Handle free purchase or immediate success
        // Sync subscriptions from server to get accurate expiration dates.
        const updatedUser = await api.auth.profile();
        setUser(prev => ({ ...prev, subscriptions: updatedUser.subscriptions }));
      } catch (e: any) {
        console.error("Checkout failed", e);
        alert('支付失败: ' + (e.message || '请重试'));
        return;
      }
    } else {
      // Guest logic (Mock)
      const newSubs = cart.map(item => ({
        accessId: item.accessId,
        startDate: new Date().toISOString(),
        expiresAt: null
      }));
      const updatedSubscriptions = [...user.subscriptions, ...newSubs];
      setUser(prev => ({
        ...prev,
        subscriptions: updatedSubscriptions
      }));
    }

    setCart([]);
    alert("支付成功！您已解锁相关课程内容。");
  };

  const addQuizResult = (result: QuizResult) => {
    setUser(prev => ({
      ...prev,
      quizHistory: [result, ...prev.quizHistory]
    }));
    if (user.isAuthenticated) {
      api.user.addQuizResult(result).catch(console.error);
    }
  };

  const hasAccess = (id: AccessId) => {
    // Find subscription
    const sub = user.subscriptions.find(s => {
      if (typeof s === 'string') return s === id;
      return s.accessId === id;
    });

    if (!sub) return false;

    // If string, assume legacy/forever
    if (typeof sub === 'string') return true;

    // Check expiry
    if (!sub.expiresAt) return true; // Forever

    return new Date(sub.expiresAt) > new Date();
  };

  // --- Quiz Progress Logic ---
  const getProgressKey = (subjectId: string, chapterTitle: string) => `${subjectId}_${chapterTitle}`;

  const updateChapterProgress = (subjectId: string, chapterTitle: string, index: number, isCorrect?: boolean) => {
    const key = getProgressKey(subjectId, chapterTitle);
    setUser(prev => {
      const currentProgress = prev.chapterProgress[key] || { lastIndex: 0, history: {} };
      const newHistory = { ...currentProgress.history };

      if (isCorrect !== undefined) {
        newHistory[index] = isCorrect;
      }

      const newProgress = {
        lastIndex: index,
        history: newHistory
      };

      if (prev.isAuthenticated) {
        // Debounce or immediate call? For correctness, call immediately but silently catch error
        api.user.updateProgress(key, newProgress).catch(console.error);
      }

      return {
        ...prev,
        chapterProgress: {
          ...prev.chapterProgress,
          [key]: newProgress
        }
      };
    });
  };

  const resetChapterProgress = (subjectId: string, chapterTitle: string) => {
    const key = getProgressKey(subjectId, chapterTitle);

    // We update local state first
    setUser(prev => {
      const newProgress = { ...prev.chapterProgress };
      delete newProgress[key];
      return {
        ...prev,
        chapterProgress: newProgress
      };
    });

    // Also need to persist the reset (saving empty or null)
    if (user.isAuthenticated) {
      // Sending an empty object or a specific flag to backend to clear it
      // Our backend update logic is a merge: `...user.chapterProgress, [body.key]: body.data`
      // So putting `{ lastIndex: 0, history: {} }` effectively resets it visually,
      // but 'deleting' the key on backend requires different logic.
      // For simplicity, we overwrite with a "blank" progress state.
      api.user.updateProgress(key, { lastIndex: 0, history: {} }).catch(console.error);
    }
  };

  // --- Bookmark Logic ---
  const toggleBookmark = (item: Omit<Bookmark, 'timestamp'>) => {
    setUser(prev => {
      const exists = prev.bookmarks.some(b => b.id === item.id);
      let newBookmarks;

      const newItem = { ...item, timestamp: Date.now() };

      if (exists) {
        newBookmarks = prev.bookmarks.filter(b => b.id !== item.id);
      } else {
        newBookmarks = [newItem, ...prev.bookmarks];
      }

      if (prev.isAuthenticated) {
        // The database stores the whole list? No, API toggles a single item.
        // Let's verify backend logic. UsersController.toggleBookmark adds/removes the ITEM.
        // So we just send the item.
        api.user.toggleBookmark(newItem).catch(console.error);
      }

      return {
        ...prev,
        bookmarks: newBookmarks
      };
    });
  };

  const isBookmarked = (id: string) => user.bookmarks.some(b => b.id === id);

  // Stats calculation
  const stats = useMemo(() => {
    const totalQuestions = user.quizHistory.reduce((acc, curr) => acc + curr.total, 0);
    const totalCorrect = user.quizHistory.reduce((acc, curr) => acc + curr.correct, 0);
    const correctRate = totalQuestions > 0 ? Math.round((totalCorrect / totalQuestions) * 100) : 0;
    return { totalQuestions, correctRate, subsCount: user.subscriptions.length };
  }, [user]);

  // View Routing
  const renderView = () => {
    switch (view) {
      case 'store': return <StoreView />;
      case 'quiz': return <QuizView />;
      case 'wiki': return <WikiView />;
      case 'profile': return <ProfileView />;
      case 'favorites': return <FavoritesView />;
      case 'mistakes': return <MistakesView />;
      case 'assistant': return aiEnabled ? <AIChatBot /> : <HomeDashboard stats={stats} onNavigate={setView} />;
      case 'feedback': return <FeedbackView />;
      // 'admin' is handled at top level
      default: return (
        <HomeDashboard
          stats={stats}
          onNavigate={setView}
        />
      );
    }
  };

  // --- Context Wrapper to provide auth methods ---
  const contextValue = {
    user, cart, view, setView,
    addToCart, removeFromCart, checkout,
    addQuizResult, hasAccess,
    updateChapterProgress, resetChapterProgress,
    toggleBookmark, isBookmarked,
    wikiState, setWikiState,
    login, logout
  };

  return (
    <AppContext.Provider value={contextValue}>
      {!user.isAuthenticated ? (
        <AuthView />
      ) : (view as string) === 'admin' ? (
        <AdminView onExit={() => { fetchSettings(); setView('home'); }} />
      ) : (
        <div className="min-h-screen bg-slate-50 flex flex-col md:flex-row font-sans text-slate-900">

          {/* Sidebar Navigation (Desktop) */}
          <aside className="hidden md:flex w-64 flex-col bg-white border-r border-slate-200 fixed h-full z-20">
            <div className="p-4 flex items-center gap-2 border-b border-slate-100">
              <div className="bg-blue-600 text-white p-1.5 rounded-lg">
                <CheckCircle size={20} />
              </div>
              <span className="text-lg font-bold text-blue-700">
                题酷
              </span>
            </div>

            <nav className="flex-1 px-3 py-4 space-y-1">
              <NavItem icon={<LayoutDashboard size={18} />} label="总览" active={view === 'home'} onClick={() => setView('home')} />
              <NavItem icon={<ShoppingBag size={18} />} label="商城" active={view === 'store'} onClick={() => setView('store')} />
              <NavItem icon={<BrainCircuit size={18} />} label="题库" active={view === 'quiz'} onClick={() => setView('quiz')} />
              <NavItem icon={<Eraser size={18} />} label="错题集" active={view === 'mistakes'} onClick={() => setView('mistakes')} />
              <NavItem icon={<BookOpen size={18} />} label="知识库" active={view === 'wiki'} onClick={() => setView('wiki')} />
              <NavItem icon={<Star size={18} />} label="我的收藏" active={view === 'favorites'} onClick={() => setView('favorites')} />
              <div className="pt-2 mt-2 border-t border-slate-100">
                {aiEnabled && (
                  <NavItem
                    icon={<Bot size={18} />}
                    label="智能助教"
                    active={view === 'assistant'}
                    onClick={() => setView('assistant')}
                  />
                )}
                <NavItem
                  icon={<MessageSquare size={18} />}
                  label="问题反馈"
                  active={view === 'feedback'}
                  onClick={() => setView('feedback')}
                />
              </div>

              {/* Admin Portal Link */}
              {user.role === 'admin' && (
                <div className="pt-2 mt-auto">
                  <button
                    onClick={() => setView('admin')}
                    className="w-full flex items-center gap-3 px-3 py-2.5 rounded-lg transition-all duration-200 font-medium text-sm text-indigo-500 hover:bg-indigo-50 hover:text-indigo-600"
                  >
                    <Shield size={18} />
                    <span>管理员后台</span>
                  </button>
                </div>
              )}
            </nav>

            <div className="p-3 border-t border-slate-100">
              <button
                onClick={() => setView('profile')}
                className={`flex items-center gap-3 p-2 w-full rounded-xl transition-colors mb-2 ${view === 'profile' ? 'bg-slate-100' : 'hover:bg-slate-50'}`}
              >
                <div className={`w-8 h-8 rounded-full flex items-center justify-center shrink-0 overflow-hidden ${user.avatar ? 'bg-white' : 'bg-blue-100 text-blue-600'}`}>
                  {user.avatar ? (
                    <img src={user.avatar} alt={user.name} className="w-full h-full object-cover" />
                  ) : (
                    <User size={16} />
                  )}
                </div>
                <div className="text-left flex-1 min-w-0">
                  <p className="text-sm font-bold text-slate-800 truncate">{user.name}</p>
                  <p className="text-xs text-slate-500 truncate">
                    {user.subscriptions.length > 0 ? <span className="text-amber-500 font-bold">Pro 会员</span> : '普通用户'}
                  </p>
                </div>
              </button>
            </div>
          </aside>

          {/* Mobile Header */}
          <div className="md:hidden fixed top-0 w-full bg-white/80 backdrop-blur-md z-30 border-b border-slate-200 px-4 h-16 flex items-center justify-between">
            <div className="flex items-center gap-2">
              <div className="bg-blue-600 text-white p-1.5 rounded-lg">
                <CheckCircle size={20} />
              </div>
              <span className="font-bold text-lg">题酷</span>
            </div>
            <button onClick={() => setIsMobileMenuOpen(true)} className="p-2 text-slate-600">
              <Menu size={24} />
            </button>
          </div>

          {/* Mobile Menu Drawer */}
          <AnimatePresence>
            {isMobileMenuOpen && (
              <>
                <motion.div
                  initial={{ opacity: 0 }} animate={{ opacity: 1 }} exit={{ opacity: 0 }}
                  className="fixed inset-0 bg-black/50 z-40 md:hidden"
                  onClick={() => setIsMobileMenuOpen(false)}
                />
                <motion.div
                  initial={{ x: '100%' }} animate={{ x: 0 }} exit={{ x: '100%' }}
                  transition={{ type: "spring", damping: 25, stiffness: 200 }}
                  className="fixed inset-y-0 right-0 w-64 bg-white z-50 p-6 shadow-2xl md:hidden flex flex-col"
                >
                  <div className="flex justify-between items-center mb-8">
                    <span className="font-bold text-xl">菜单</span>
                    <button onClick={() => setIsMobileMenuOpen(false)}><X size={24} /></button>
                  </div>
                  <nav className="space-y-4 flex-1">
                    <NavItem icon={<LayoutDashboard size={20} />} label="总览" active={view === 'home'} onClick={() => { setView('home'); setIsMobileMenuOpen(false) }} />
                    <NavItem icon={<ShoppingBag size={20} />} label="商城" active={view === 'store'} onClick={() => { setView('store'); setIsMobileMenuOpen(false) }} />
                    <NavItem icon={<BrainCircuit size={20} />} label="题库" active={view === 'quiz'} onClick={() => { setView('quiz'); setIsMobileMenuOpen(false) }} />
                    <NavItem icon={<Eraser size={20} />} label="错题集" active={view === 'mistakes'} onClick={() => { setView('mistakes'); setIsMobileMenuOpen(false) }} />
                    <NavItem icon={<BookOpen size={20} />} label="知识库" active={view === 'wiki'} onClick={() => { setView('wiki'); setIsMobileMenuOpen(false) }} />
                    <NavItem icon={<Star size={20} />} label="我的收藏" active={view === 'favorites'} onClick={() => { setView('favorites'); setIsMobileMenuOpen(false) }} />
                    {aiEnabled && (
                      <NavItem
                        icon={<Bot size={20} />}
                        label="智能助教"
                        active={view === 'assistant'}
                        onClick={() => { setView('assistant'); setIsMobileMenuOpen(false); }}
                      />
                    )}
                    <NavItem
                      icon={<MessageSquare size={20} />}
                      label="问题反馈"
                      active={view === 'feedback'}
                      onClick={() => { setView('feedback'); setIsMobileMenuOpen(false); }}
                    />
                    <NavItem icon={<User size={20} />} label="个人中心" active={view === 'profile'} onClick={() => { setView('profile'); setIsMobileMenuOpen(false) }} />
                    {user.role === 'admin' && (
                      <div className="pt-4 border-t border-slate-100">
                        <NavItem icon={<Shield size={20} />} label="管理员后台" active={view === 'admin'} onClick={() => { setView('admin'); setIsMobileMenuOpen(false) }} />
                      </div>
                    )}
                  </nav>
                </motion.div>
              </>
            )}
          </AnimatePresence>

          {/* Main Content Area */}
          <main className={`flex-1 md:ml-64 transition-all ${view === 'assistant'
            ? 'fixed inset-0 top-16 md:static md:h-screen md:pt-0 overflow-hidden flex flex-col z-0'
            : `min-h-screen ${view === 'quiz' ? 'pt-16' : 'pt-20'} md:pt-0`
            }`}>
            <AnimatePresence mode="wait">
              <motion.div
                key={view}
                initial={{ opacity: 0, y: view === 'assistant' ? 0 : 10 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, y: view === 'assistant' ? 0 : -10 }}
                transition={{ duration: 0.2 }}
                className={view === 'assistant' ? 'flex-1 flex flex-col h-full w-full overflow-hidden' : `${view === 'quiz' ? 'pt-0 md:pt-0' : 'pt-4 md:pt-8'} p-4 md:p-8 max-w-7xl mx-auto w-full`}
              >
                {renderView()}
              </motion.div>
            </AnimatePresence>
          </main>
        </div>
      )}
    </AppContext.Provider>
  );
};

const NavItem = ({ icon, label, active, onClick }: { icon: React.ReactNode, label: string, active: boolean, onClick: () => void }) => (
  <button
    onClick={onClick}
    className={`w-full flex items-center gap-3 px-3 py-2.5 rounded-lg transition-all duration-200 font-medium text-sm ${active
      ? 'bg-blue-600 text-white shadow-md shadow-blue-500/20'
      : 'text-slate-500 hover:bg-slate-50 hover:text-slate-900'
      }`}
  >
    {icon}
    <span>{label}</span>
  </button>
);

const HomeDashboard = ({ stats, onNavigate }: any) => (
  <div className="space-y-8">
    <header>
      <h1 className="text-3xl font-bold text-slate-900">欢迎回来，同学</h1>
      <p className="text-slate-500 mt-2">今天是学习的好日子。您的学习状态如下：</p>
    </header>

    <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
      <StatCard
        label="已刷题目"
        value={stats.totalQuestions}
        icon={<BrainCircuit className="text-blue-500" />}
        bg="bg-blue-50"
      />
      <StatCard
        label="正确率"
        value={`${stats.correctRate}%`}
        icon={<CheckCircle className="text-emerald-500" />}
        bg="bg-emerald-50"
      />
      <StatCard
        label="有效订阅"
        value={stats.subsCount}
        icon={<ShoppingBag className="text-violet-500" />}
        bg="bg-violet-50"
      />
    </div>

    {/* Announcements Area */}
    <AnnouncementBoard />

    <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
      <div className="bg-white p-8 rounded-3xl shadow-sm border border-slate-100 relative overflow-hidden group cursor-pointer" onClick={() => onNavigate('quiz')}>
        <div className="absolute top-0 right-0 w-40 h-40 bg-blue-100 rounded-full blur-3xl -mr-10 -mt-10 opacity-50 group-hover:opacity-100 transition-opacity"></div>
        <div className="relative z-10">
          <h3 className="text-xl font-bold mb-2">每日一练</h3>
          <p className="text-slate-500 mb-6">保持手感，每日随机抽取10道错题或新题进行训练。</p>
          <button className="bg-slate-900 text-white px-6 py-2.5 rounded-full text-sm font-medium hover:bg-blue-600 transition-colors">
            开始练习
          </button>
        </div>
      </div>

      <div className="bg-gradient-to-br from-indigo-500 to-purple-600 p-8 rounded-3xl shadow-lg text-white relative overflow-hidden">
        <div className="relative z-10">
          <h3 className="text-xl font-bold mb-2">解锁更多课程</h3>
          <p className="text-indigo-100 mb-6">订阅专业版，获取病理学、外科学等高清图谱与专家视频。</p>
          <button
            onClick={() => onNavigate('store')}
            className="bg-white text-indigo-600 px-6 py-2.5 rounded-full text-sm font-bold hover:bg-indigo-50 transition-colors"
          >
            去商城逛逛
          </button>
        </div>
        <BookOpen size={120} className="absolute -bottom-4 -right-4 opacity-10 rotate-12" />
      </div>
    </div>
  </div>
);

const AnnouncementBoard = () => {
  const [announcements, setAnnouncements] = React.useState<any[]>([]);

  React.useEffect(() => {
    api.announcements.getLatest().then(setAnnouncements).catch(console.error);
  }, []);

  if (announcements.length === 0) return null;

  return (
    <div className="bg-gradient-to-r from-amber-50 to-orange-50 border border-amber-100 rounded-2xl p-6">
      <div className="flex items-center gap-2 mb-4">
        <div className="bg-amber-100 text-amber-600 p-1.5 rounded-lg">
          <Shield size={18} />
        </div>
        <h3 className="font-bold text-slate-800">最新公告</h3>
      </div>
      <div className="space-y-3">
        {announcements.map((item, idx) => (
          <div key={item.id} className="bg-white/60 p-3 rounded-xl border border-white/50">
            <div className="flex justify-between items-start">
              <h4 className="font-medium text-slate-800 text-sm">{item.title}</h4>
              <span className="text-xs text-slate-400 shrink-0 ml-2">{new Date(item.createdAt).toLocaleDateString()}</span>
            </div>
            <p className="text-xs text-slate-500 mt-1 line-clamp-2">{item.content}</p>
          </div>
        ))}
      </div>
    </div>
  );
};

const StatCard = ({ label, value, icon, bg }: any) => (
  <div className="bg-white p-6 rounded-2xl shadow-sm border border-slate-100 flex items-center gap-4">
    <div className={`p-4 rounded-xl ${bg}`}>{icon}</div>
    <div>
      <div className="text-sm text-slate-500">{label}</div>
      <div className="text-2xl font-bold text-slate-900">{value}</div>
    </div>
  </div>
);

export default App;
