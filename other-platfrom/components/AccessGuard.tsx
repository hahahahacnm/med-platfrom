
import React from 'react';
import { Lock, ArrowLeft } from 'lucide-react';
import { useAppContext } from '../context';
import { AccessId } from '../types';

interface AccessGuardProps {
  accessId: AccessId;
  children: React.ReactNode;
  title?: string;
  onBack?: () => void;
  backLabel?: string;
}

export const AccessGuard: React.FC<AccessGuardProps> = ({ 
  accessId, 
  children, 
  title = "订阅后查看",
  onBack,
  backLabel = "返回"
}) => {
  const { hasAccess, setView } = useAppContext();

  if (hasAccess(accessId)) {
    return <>{children}</>;
  }

  return (
    <div className="relative min-h-[400px] bg-slate-50 rounded-2xl border border-slate-200 overflow-hidden flex flex-col items-center justify-center text-center p-8">
      <div className="bg-white p-4 rounded-full shadow-lg mb-6">
        <Lock size={32} className="text-slate-400" />
      </div>
      <h3 className="text-xl font-bold text-slate-800 mb-2">{title}</h3>
      <p className="text-slate-500 max-w-md mb-8">
        该内容需要订阅相关课程包后才能解锁。请前往商城购买对应服务。
      </p>
      
      <div className="flex flex-col gap-3 w-full max-w-xs">
        <button 
          onClick={() => setView('store')}
          className="bg-blue-600 text-white px-8 py-3 rounded-xl font-semibold hover:bg-blue-700 transition-colors shadow-lg shadow-blue-500/20 w-full"
        >
          前往订阅
        </button>
        
        {onBack && (
          <button 
            onClick={onBack}
            className="flex items-center justify-center gap-2 text-slate-500 hover:text-slate-800 hover:bg-slate-100 px-8 py-3 rounded-xl font-medium transition-colors w-full"
          >
            <ArrowLeft size={18} />
            {backLabel}
          </button>
        )}
      </div>
      
      {/* Blurred background effect hint */}
      <div className="absolute inset-0 -z-10 opacity-50 blur-sm pointer-events-none select-none overflow-hidden" aria-hidden="true">
        <div className="p-8 text-slate-300">
           Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. 
           (Preview content hidden)
           <br/><br/>
           Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur.
        </div>
      </div>
    </div>
  );
};
