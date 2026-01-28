
import React, { useState, useRef } from 'react';
import { useAppContext } from '../context';
import { api } from '../services/api';
import { User, Award, History, BookOpen, BrainCircuit, LogOut, Camera } from 'lucide-react';
import { motion } from 'framer-motion';

const ProfileView: React.FC = () => {
  const { user, logout, login } = useAppContext();
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [isUploading, setIsUploading] = useState(false);

  const handleFileChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    try {
      setIsUploading(true);
      const res = await api.user.uploadAvatar(file);
      if (res.success && res.avatar) {
        login({ ...user, avatar: res.avatar });
      }
    } catch (error) {
      console.error('Upload failed', error);
      alert('头像上传失败');
    } finally {
      setIsUploading(false);
    }
  };

  // Calculate stats
  const correct = user.quizHistory.reduce((acc, curr) => acc + curr.correct, 0);
  const total = user.quizHistory.reduce((acc, curr) => acc + curr.total, 0);
  const percentage = total > 0 ? correct / total : 0;

  // Helper to format subscription name
  const formatSub = (sub: any) => {
    const accessId = typeof sub === 'string' ? sub : sub.accessId;
    const expiresAt = typeof sub === 'object' ? sub.expiresAt : null;
    let label = accessId;
    let icon = null;
    let color = 'slate';

    if (accessId.startsWith('quiz_')) {
      label = accessId.replace('quiz_', '') + ' 题库';
      icon = <BrainCircuit size={12} />;
      color = 'blue';
    } else if (accessId.startsWith('wiki_')) {
      label = accessId.replace('wiki_', '') + ' 知识库';
      icon = <BookOpen size={12} />;
      color = 'amber';
    }

    if (expiresAt) {
      const date = new Date(expiresAt).toLocaleDateString();
      label += ` (至 ${date})`;
    } else {
      label += ' (永久)';
    }

    return { text: label, icon, color, id: accessId };
  };

  return (
    <div className="space-y-8 animate-in fade-in duration-500">
      {/* Header */}
      <div className="bg-white p-8 rounded-3xl shadow-sm border border-slate-100 flex flex-col md:flex-row items-center gap-8 relative overflow-hidden group">
        {/* Logout Button */}
        <div className="absolute top-6 right-6 z-10">
          <button
            onClick={logout}
            className="flex items-center gap-2 px-4 py-2 bg-slate-50 text-slate-500 rounded-xl font-bold text-sm hover:bg-rose-50 hover:text-rose-600 transition-colors shadow-sm hover:shadow-md"
          >
            <LogOut size={16} />
            <span className="hidden md:inline">退出登录</span>
          </button>
        </div>

        <div className="flex flex-col items-center gap-3">
          <div className="relative group/avatar cursor-pointer" onClick={() => fileInputRef.current?.click()}>
            <div className={`w-24 h-24 rounded-full flex items-center justify-center text-white shadow-lg shrink-0 overflow-hidden ${user.avatar ? 'bg-white' : 'bg-gradient-to-br from-blue-500 to-teal-400'}`}>
              {user.avatar ? (
                <img src={user.avatar} alt="Avatar" className="w-full h-full object-cover" />
              ) : (
                <User size={48} />
              )}
            </div>
            <div className="absolute inset-0 bg-black/30 rounded-full flex items-center justify-center opacity-0 group-hover/avatar:opacity-100 transition-opacity">
              <Camera size={24} className="text-white" />
            </div>
            {isUploading && (
              <div className="absolute inset-0 bg-black/50 rounded-full flex items-center justify-center">
                <div className="w-6 h-6 border-2 border-white border-t-transparent rounded-full animate-spin"></div>
              </div>
            )}
            <input
              type="file"
              ref={fileInputRef}
              className="hidden"
              accept="image/*"
              onChange={handleFileChange}
            />
          </div>
          <button
            onClick={() => fileInputRef.current?.click()}
            className="text-xs font-bold text-slate-500 bg-slate-100 hover:bg-slate-200 px-3 py-1.5 rounded-full transition-colors"
          >
            编辑头像
          </button>
        </div>
        <div className="text-center md:text-left flex-1">
          <h2 className="text-3xl font-bold text-slate-900">{user.name}</h2>
          <p className="text-slate-500 mt-1 font-medium">{user.email}</p>
          <div className="mt-2 text-sm text-slate-400">
            身份：医学生 / {user.subscriptions.length > 0 ? <span className="text-amber-500 font-bold">Pro 会员</span> : '普通用户'}
          </div>

          <div className="mt-4">
            <div className="text-sm text-slate-400 mb-2">我的订阅：</div>
            <div className="flex flex-wrap items-center justify-center md:justify-start gap-2">
              {user.subscriptions.map((sub: any) => {
                const info = formatSub(sub);
                return (
                  <span key={info.id} className={`px-3 py-1 bg-${info.color}-50 text-${info.color}-600 rounded-full text-xs font-bold border border-${info.color}-100 uppercase flex items-center gap-1`}>
                    {info.icon} {info.text}
                  </span>
                )
              })}
              {user.subscriptions.length === 0 && <span className="text-xs text-slate-400 bg-slate-50 px-3 py-1 rounded-full">暂无订阅</span>}
            </div>
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
        {/* Stats Chart */}
        <div className="bg-white p-8 rounded-3xl shadow-sm border border-slate-100">
          <h3 className="text-lg font-bold text-slate-900 mb-6 flex items-center gap-2">
            <Award size={20} className="text-orange-500" /> 刷题统计
          </h3>
          <div className="h-64 flex flex-col items-center justify-center relative">
            <svg width="200" height="200" viewBox="0 0 100 100" className="transform -rotate-90">
              {/* Background Circle */}
              <circle cx="50" cy="50" r="40" stroke="#f1f5f9" strokeWidth="10" fill="none" />
              {/* Progress Circle */}
              <motion.circle
                cx="50"
                cy="50"
                r="40"
                stroke="#10b981"
                strokeWidth="10"
                fill="none"
                strokeLinecap="round"
                initial={{ pathLength: 0 }}
                animate={{ pathLength: percentage }}
                transition={{ duration: 1.5, ease: "easeOut" }}
              />
            </svg>
            <div className="absolute inset-0 flex flex-col items-center justify-center">
              <span className="text-3xl font-bold text-slate-800">{Math.round(percentage * 100)}%</span>
              <span className="text-xs text-slate-400">正确率</span>
            </div>
          </div>

          <div className="flex justify-center gap-8 mt-4">
            <div className="text-center">
              <div className="text-2xl font-bold text-slate-800">{total}</div>
              <div className="text-xs text-slate-400">总题数</div>
            </div>
            <div className="text-center">
              <div className="text-2xl font-bold text-emerald-500">{correct}</div>
              <div className="text-xs text-slate-400">答对</div>
            </div>
          </div>
        </div>

        {/* History List */}
        <div className="bg-white p-8 rounded-3xl shadow-sm border border-slate-100 flex flex-col">
          <h3 className="text-lg font-bold text-slate-900 mb-6 flex items-center gap-2">
            <History size={20} className="text-blue-500" /> 最近练习
          </h3>
          <div className="flex-1 overflow-y-auto space-y-4 max-h-[300px] pr-2 custom-scrollbar">
            {user.quizHistory.length === 0 ? (
              <div className="flex flex-col items-center justify-center h-40 text-slate-400">
                <BrainCircuit size={32} className="mb-2 opacity-20" />
                <p>暂无练习记录</p>
              </div>
            ) : (
              user.quizHistory.map((h, i) => (
                <div key={i} className="flex items-center justify-between p-4 bg-slate-50 rounded-xl hover:bg-slate-100 transition-colors">
                  <div>
                    <div className="font-bold text-slate-700 text-sm">{h.subjectId}</div>
                    <div className="text-xs text-slate-400 mt-1">{h.date}</div>
                  </div>
                  <div className="text-right">
                    <div className="text-sm font-bold bg-white px-2 py-1 rounded shadow-sm border border-slate-100">
                      <span className="text-emerald-600">{h.correct}</span> <span className="text-slate-300">/</span> {h.total}
                    </div>
                  </div>
                </div>
              ))
            )}
          </div>
        </div>
      </div>
    </div>
  );
};

export default ProfileView;
