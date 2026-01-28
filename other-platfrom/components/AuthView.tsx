
import React, { useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { CheckCircle, Mail, Lock, User, ArrowRight, Eye, EyeOff, ShieldCheck, Check, X } from 'lucide-react';
import { useAppContext } from '../context';
import { api } from '../services/api';

const AuthView: React.FC = () => {
  const { login } = useAppContext();
  const [isLogin, setIsLogin] = useState(true);
  const [showPassword, setShowPassword] = useState(false);
  const [isLoading, setIsLoading] = useState(false);

  // Form States
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [name, setName] = useState('');
  const [error, setError] = useState('');
  const [hasAgreed, setHasAgreed] = useState(false);
  const [showDocument, setShowDocument] = useState<'agreement' | 'privacy' | 'disclaimer' | null>(null);

  const [showLoginCaptcha, setShowLoginCaptcha] = useState(false);

  // Captcha State
  const [captchaCode, setCaptchaCode] = useState('');
  const [captchaId, setCaptchaId] = useState('');
  const [captchaImage, setCaptchaImage] = useState('');

  const fetchCaptcha = async () => {
    try {
      const data = await api.auth.getCaptcha();
      setCaptchaId(data.id);
      setCaptchaImage(data.image);
    } catch (err) {
      console.error('Failed to fetch captcha', err);
    }
  };

  React.useEffect(() => {
    if (!isLogin || showLoginCaptcha) {
      fetchCaptcha();
    }
  }, [isLogin, showLoginCaptcha]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    if (!hasAgreed) {
      setError('请先阅读并同意用户协议等条款');
      return;
    }

    // Basic Validation
    if (!email || !password || (!isLogin && !name)) {
      setError('请填写所有必填项');
      return;
    }

    if ((!isLogin || showLoginCaptcha) && !captchaCode) {
      setError('请输入验证码');
      return;
    }

    if (password.length < 6) {
      setError('密码长度至少为6位');
      return;
    }

    setIsLoading(true);

    // Real API Call
    try {
      let user;
      if (isLogin) {
        user = await api.auth.login(email, password, captchaCode, captchaId);
      } else {
        user = await api.auth.register({ email, password, name, captchaCode, captchaId });
      }
      login(user);
    } catch (err: any) {
      const msg = err.message || '登录/注册失败，请检查账号密码';
      setError(msg);
      setIsLoading(false);

      if (msg.includes('验证码')) {
        setShowLoginCaptcha(true);
        // fetchCaptcha will be triggered by effect, but we can also ensure it refreshes
      }

      if (!isLogin || showLoginCaptcha || msg.includes('验证码')) {
        fetchCaptcha();
        setCaptchaCode('');
      }
    }
  };

  const getDocumentContent = () => {
    switch (showDocument) {
      case 'agreement':
        return (
          <div className="space-y-4">
            <h4 className="font-bold text-slate-800">1. 服务条款的确认和接纳</h4>
            <p>题酷提供的服务完全按照其发布的服务条款和操作规则严格执行。用户必须完全同意所有服务条款并完成注册程序，才能成为题酷的正式用户。</p>
            <h4 className="font-bold text-slate-800">2. 服务说明</h4>
            <p>题酷运用自己的系统通过互联网向用户提供包括医学题库、AI助教等在内的网络服务。</p>
            <h4 className="font-bold text-slate-800">3. 用户的帐号，密码和安全性</h4>
            <p>用户一旦注册成功，成为题酷的合法用户，将得到一个密码和用户名。用户将对用户名和密码安全负全部责任。</p>
          </div>
        );
      case 'privacy':
        return (
          <div className="space-y-4">
            <h4 className="font-bold text-slate-800">1. 信息收集</h4>
            <p>我们在您注册、使用服务时收集您的个人信息，包括但不限于姓名、邮箱、学习记录等。</p>
            <h4 className="font-bold text-slate-800">2. 信息使用</h4>
            <p>我们使用这些信息来为您提供个性化的学习体验、改进我们的服务以及通知您有关产品更新的信息。</p>
            <h4 className="font-bold text-slate-800">3. 信息保护</h4>
            <p>我们将采取合理的安全手段保护您的个人信息，除非法律规定或经您授权，我们不会向第三方公开您的个人信息。</p>
          </div>
        );
      case 'disclaimer':
        return (
          <div className="space-y-4">
            <h4 className="font-bold text-slate-800">1. 内容免责</h4>
            <p>题酷提供的所有医学知识、题目解析仅供学习参考，不能替代专业的医疗建议、诊断或治疗。</p>
            <h4 className="font-bold text-slate-800">2. 服务中断</h4>
            <p>对于因不可抗力或题酷不能控制的原因造成的网络服务中断或其它缺陷，题酷不承担任何责任，但将尽力减少因此而给用户造成的损失和影响。</p>
          </div>
        );
      default:
        return null;
    }
  };

  return (
    <div className="h-screen w-full flex bg-slate-50 font-sans text-slate-900 overflow-hidden">

      {/* Left Side: Brand & Visuals - Hidden on Mobile */}
      <div className="hidden md:flex md:w-5/12 lg:w-2/5 bg-slate-900 relative flex-col justify-between p-10 lg:p-16 text-white h-full shrink-0">
        {/* Background Decoration */}
        <div className="absolute top-0 left-0 w-full h-full opacity-20 pointer-events-none">
          <svg className="h-full w-full" viewBox="0 0 100 100" preserveAspectRatio="none">
            <path d="M0 100 C 20 0 50 0 100 100 Z" fill="url(#grad1)" />
            <defs>
              <linearGradient id="grad1" x1="0%" y1="0%" x2="100%" y2="0%">
                <stop offset="0%" style={{ stopColor: '#3b82f6', stopOpacity: 1 }} />
                <stop offset="100%" style={{ stopColor: '#14b8a6', stopOpacity: 1 }} />
              </linearGradient>
            </defs>
          </svg>
        </div>

        <div className="relative z-10">
          <div className="flex items-center gap-3 mb-10">
            <div className="bg-blue-600 p-2.5 rounded-xl text-white shadow-lg">
              <CheckCircle size={28} />
            </div>
            <span className="text-3xl font-bold tracking-tight">题酷</span>
          </div>

          <div className="space-y-6 max-w-lg">
            <motion.h1
              key={isLogin ? 'login-h' : 'reg-h'}
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              className="text-4xl lg:text-5xl font-bold leading-tight"
            >
              {isLogin ? '欢迎回到您的医学殿堂' : '开启您的医学进阶之旅'}
            </motion.h1>
            <p className="text-slate-400 text-lg leading-relaxed">
              题酷 提供最权威的医学题库、知识库与 AI 助教服务，助您在医学考试与临床实践中游刃有余。
            </p>
          </div>
        </div>

        <div className="relative z-10 mt-10">
          <div className="flex items-center gap-3 text-base font-medium text-slate-300 mb-2">
            <div className="w-8 h-8 rounded-full bg-slate-800 flex items-center justify-center border border-slate-700">
              <ShieldCheck size={16} className="text-emerald-400" />
            </div>
            <span>专业认证内容</span>
          </div>
          <p className="text-xs text-slate-500">我们永远在这里！</p>
        </div>
      </div>

      {/* Right Side: Form - Full Height, Centered */}
      <div className="w-full md:w-7/12 lg:w-3/5 relative h-full flex flex-col justify-center bg-white px-6 py-6 overflow-y-auto md:overflow-hidden">
        {/* Mobile Header Branding */}
        <div className="md:hidden flex items-center justify-center gap-2 mb-6 shrink-0">
          <div className="bg-blue-600 p-2 rounded-lg text-white shadow-md">
            <CheckCircle size={20} />
          </div>
          <span className="text-2xl font-bold tracking-tight text-slate-900">题酷</span>
        </div>

        <div className="max-w-md w-full mx-auto">
          <div className="flex justify-between items-end mb-5">
            <div>
              <h2 className="text-2xl md:text-3xl font-bold text-slate-900 mb-1">{isLogin ? '账号登录' : '创建新账号'}</h2>
              <p className="text-sm text-slate-500">请输入您的认证信息以继续</p>
            </div>
            <button
              onClick={() => { setIsLogin(!isLogin); setError(''); setShowLoginCaptcha(false); }}
              className="text-sm font-bold text-blue-600 hover:text-blue-700 hover:bg-blue-50 px-3 py-1.5 rounded-lg transition-colors"
            >
              {isLogin ? '免费注册' : '已有账号?'}
            </button>
          </div>

          <AnimatePresence mode="wait">
            <motion.form
              key={isLogin ? 'login' : 'register'}
              initial={{ opacity: 0, x: 20 }}
              animate={{ opacity: 1, x: 0 }}
              exit={{ opacity: 0, x: -20 }}
              transition={{ duration: 0.3 }}
              onSubmit={handleSubmit}
              className="space-y-3"
            >
              {!isLogin && (
                <div className="space-y-1">
                  <label className="text-xs font-bold text-slate-700 ml-1">姓名</label>
                  <div className="relative group">
                    <User className="absolute left-3 top-3 text-slate-400 group-focus-within:text-blue-500 transition-colors" size={18} />
                    <input
                      type="text"
                      placeholder="您的称呼"
                      value={name}
                      onChange={e => setName(e.target.value)}
                      className="w-full bg-slate-50 border border-slate-200 rounded-xl py-2.5 pl-10 pr-4 focus:outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 transition-all font-medium text-slate-800 text-sm"
                    />
                  </div>
                </div>
              )}

              <div className="space-y-1">
                <label className="text-xs font-bold text-slate-700 ml-1">邮箱</label>
                <div className="relative group">
                  <Mail className="absolute left-3 top-3 text-slate-400 group-focus-within:text-blue-500 transition-colors" size={18} />
                  <input
                    type="email"
                    placeholder="student@med.edu"
                    value={email}
                    onChange={e => setEmail(e.target.value)}
                    className="w-full bg-slate-50 border border-slate-200 rounded-xl py-2.5 pl-10 pr-4 focus:outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 transition-all font-medium text-slate-800 text-sm"
                  />
                </div>
              </div>

              <div className="space-y-1">
                <div className="flex justify-between items-center ml-1">
                  <label className="text-xs font-bold text-slate-700">密码</label>
                </div>
                <div className="relative group">
                  <Lock className="absolute left-3 top-3 text-slate-400 group-focus-within:text-blue-500 transition-colors" size={18} />
                  <input
                    type={showPassword ? "text" : "password"}
                    placeholder="••••••••"
                    value={password}
                    onChange={e => setPassword(e.target.value)}
                    className="w-full bg-slate-50 border border-slate-200 rounded-xl py-2.5 pl-10 pr-10 focus:outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 transition-all font-medium text-slate-800 text-sm"
                  />
                  <button
                    type="button"
                    onClick={() => setShowPassword(!showPassword)}
                    className="absolute right-3 top-2.5 text-slate-400 hover:text-slate-600 p-1"
                  >
                    {showPassword ? <EyeOff size={18} /> : <Eye size={18} />}
                  </button>
                </div>
              </div>

              {(!isLogin || showLoginCaptcha) && (
                <div className="space-y-1">
                  <label className="text-xs font-bold text-slate-700 ml-1">验证码</label>
                  <div className="flex gap-3">
                    <div className="relative group flex-1">
                      <ShieldCheck className="absolute left-3 top-3 text-slate-400 group-focus-within:text-blue-500 transition-colors" size={18} />
                      <input
                        type="text"
                        placeholder="输入验证码"
                        value={captchaCode}
                        onChange={e => setCaptchaCode(e.target.value)}
                        className="w-full bg-slate-50 border border-slate-200 rounded-xl py-2.5 pl-10 pr-4 focus:outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 transition-all font-medium text-slate-800 text-sm"
                      />
                    </div>
                    <div
                      className="h-[42px] w-[100px] bg-slate-100 rounded-xl overflow-hidden cursor-pointer hover:opacity-80 transition-opacity border border-slate-200 flex items-center justify-center"
                      onClick={fetchCaptcha}
                      title="点击刷新"
                      dangerouslySetInnerHTML={{ __html: captchaImage }}
                    />
                  </div>
                </div>
              )}

              {/* Agreement Checkbox */}
              <div className="flex items-start gap-2 mt-1">
                <div className="relative flex items-center mt-0.5">
                  <input
                    type="checkbox"
                    id="agreement"
                    checked={hasAgreed}
                    onChange={(e) => {
                      setHasAgreed(e.target.checked);
                      if (e.target.checked && error === '请先阅读并同意用户协议等条款') setError('');
                    }}
                    className="peer h-4 w-4 cursor-pointer appearance-none rounded border border-slate-300 transition-all checked:border-blue-600 checked:bg-blue-600 hover:border-blue-400 focus:ring-2 focus:ring-blue-500/20"
                  />
                  <div className="pointer-events-none absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 text-white opacity-0 peer-checked:opacity-100">
                    <Check size={10} strokeWidth={3} />
                  </div>
                </div>
                <label htmlFor="agreement" className="text-xs text-slate-500 cursor-pointer select-none leading-relaxed scale-95 origin-left">
                  我已仔细阅读并同意题酷
                  <button type="button" onClick={() => setShowDocument('agreement')} className="text-blue-600 hover:underline hover:text-blue-700 font-medium ml-0.5 mr-1 focus:outline-none">用户协议</button>、
                  <button type="button" onClick={() => setShowDocument('privacy')} className="text-blue-600 hover:underline hover:text-blue-700 font-medium mx-1 focus:outline-none">隐私政策</button>、
                  <button type="button" onClick={() => setShowDocument('disclaimer')} className="text-blue-600 hover:underline hover:text-blue-700 font-medium mx-1 focus:outline-none">免责声明</button>
                  <span className="text-[10px] text-slate-400 font-medium block mt-0.5 transition-opacity" style={{ opacity: !hasAgreed ? 1 : 0 }}>* 必须勾选同意后才能继续</span>
                </label>
              </div>

              {error && error !== '请先阅读并同意用户协议等条款' && (
                <div className="bg-red-50 text-red-500 text-xs px-3 py-2 rounded-lg border border-red-100 flex items-center gap-2 font-medium">
                  <div className="w-1.5 h-1.5 bg-red-500 rounded-full shrink-0"></div> {error}
                </div>
              )}

              <button
                type="submit"
                disabled={isLoading || !hasAgreed}
                className={`w-full font-bold py-3 rounded-xl shadow-lg transition-all flex items-center justify-center gap-2 text-sm ${isLoading || !hasAgreed
                  ? 'bg-slate-200 text-slate-400 cursor-not-allowed shadow-none'
                  : 'bg-slate-900 text-white hover:shadow-xl hover:bg-blue-600 active:scale-[0.99]'
                  }`}
              >
                {isLoading ? (
                  <div className="w-5 h-5 border-2 border-slate-400 border-t-white rounded-full animate-spin" />
                ) : (
                  <>
                    {isLogin ? '立即登录' : '注册账号'}
                    <ArrowRight size={16} />
                  </>
                )}
              </button>
            </motion.form>
          </AnimatePresence>
        </div>
      </div>

      {/* Document Modal */}
      <AnimatePresence>
        {showDocument && (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4 backdrop-blur-sm"
            onClick={() => setShowDocument(null)}
          >
            <motion.div
              initial={{ scale: 0.95, opacity: 0, y: 10 }}
              animate={{ scale: 1, opacity: 1, y: 0 }}
              exit={{ scale: 0.95, opacity: 0, y: 10 }}
              className="bg-white rounded-2xl w-full max-w-lg max-h-[80vh] flex flex-col overflow-hidden shadow-2xl ring-1 ring-slate-900/5"
              onClick={e => e.stopPropagation()}
            >
              <div className="p-4 border-b border-slate-100 flex justify-between items-center bg-white sticky top-0 z-10">
                <h3 className="text-lg font-bold text-slate-900 flex items-center gap-2">
                  <ShieldCheck className="text-blue-500" size={20} />
                  {showDocument === 'agreement' && '题酷 用户协议'}
                  {showDocument === 'privacy' && '隐私政策'}
                  {showDocument === 'disclaimer' && '免责声明'}
                </h3>
                <button
                  onClick={() => setShowDocument(null)}
                  className="w-8 h-8 rounded-full bg-slate-50 flex items-center justify-center text-slate-400 hover:bg-slate-100 hover:text-slate-600 transition-colors focus:outline-none"
                >
                  <X size={18} />
                </button>
              </div>
              <div className="p-5 overflow-y-auto text-slate-600 text-sm leading-relaxed space-y-4 text-justify">
                {getDocumentContent()}
                <div className="pt-4 mt-6 border-t border-slate-100 text-xs text-slate-400">
                  最后更新日期：2026年1月21日
                </div>
              </div>
              <div className="p-4 border-t border-slate-100 bg-slate-50 flex justify-end">
                <button
                  onClick={() => {
                    setHasAgreed(true);
                    setShowDocument(null);
                    setError('');
                  }}
                  className="bg-slate-900 text-white font-bold py-2 px-6 rounded-lg hover:bg-blue-600 transition-colors shadow-lg shadow-blue-900/10 active:scale-[0.98] text-sm"
                >
                  阅读并同意
                </button>
              </div>
            </motion.div>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
};

export default AuthView;
