
import React, { useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { Search, ShoppingBag, Plus, Check, Trash2, CreditCard, X, BookOpen, BrainCircuit, ArrowLeft, Star, ShieldCheck, Zap } from 'lucide-react';
import { useAppContext } from '../context';
import { api } from '../services/api';
import { Product } from '../types';

const StoreView: React.FC = () => {
  const { cart, addToCart, removeFromCart, checkout, hasAccess } = useAppContext();
  const [searchTerm, setSearchTerm] = useState('');
  const [isCartOpen, setIsCartOpen] = useState(false);
  const [selectedProduct, setSelectedProduct] = useState<Product | null>(null);
  const [products, setProducts] = useState<Product[]>([]);
  const [couponCode, setCouponCode] = useState('');
  const [discount, setDiscount] = useState(0);
  const [couponMsg, setCouponMsg] = useState('');
  const [isValidating, setIsValidating] = useState(false);
  const [payType, setPayType] = useState('alipay');
  const [enabledMethods, setEnabledMethods] = useState({ alipay: true, wechat: true });

  React.useEffect(() => {
    api.store.getProducts().then(setProducts).catch(console.error);
    api.settings.getAll().then(settings => {
      const alipay = settings.find((s: any) => s.key === 'payment_enable_alipay')?.value !== 'false';
      const wechat = settings.find((s: any) => s.key === 'payment_enable_wechat')?.value !== 'false';
      setEnabledMethods({ alipay, wechat });
      if (!alipay && wechat) setPayType('wechat');
      if (alipay && !wechat) setPayType('alipay');
      if (!alipay && !wechat) setPayType('');
    }).catch(console.error);
  }, []);

  React.useEffect(() => {
    setDiscount(0);
    setCouponMsg('');
    // Optional: Keep code but force re-apply, or clear code too?
    // Clearing code avoids confusion about "why is it not applied".
    // Or we could try to re-validate silently?
    // Let's just clear usage for safety.
  }, [cart]);

  const filteredProducts = products.filter(p =>
    (p.isPublished !== false) && // Default to true if undefined
    (p.title.toLowerCase().includes(searchTerm.toLowerCase()) ||
      p.tags.some(t => t.includes(searchTerm)))
  ).sort((a, b) => {
    // Pin subscribed items to the top
    const isOwnedA = hasAccess(a.accessId) ? 1 : 0;
    const isOwnedB = hasAccess(b.accessId) ? 1 : 0;
    return isOwnedB - isOwnedA;
  });

  const handleValidateCoupon = async () => {
    if (!couponCode) return;
    setIsValidating(true);
    setCouponMsg('');
    try {
      const productIds = cart.map(i => i.id);
      const res = await api.store.validateCoupon(couponCode, productIds);
      if (res.valid) {
        setDiscount(res.discount);
        setCouponMsg('优惠码生效');
      } else {
        setDiscount(0);
        setCouponMsg('无效的优惠码');
      }
    } catch (e) {
      setDiscount(0);
      setCouponMsg('验证失败');
    } finally {
      setIsValidating(false);
    }
  };

  const cartTotal = cart.reduce((sum, item) => sum + item.price, 0);

  // Helper to determine type styling
  const getProductTypeInfo = (accessId: string) => {
    if (accessId.startsWith('quiz_')) return { icon: <BrainCircuit size={14} />, label: '题库', bg: 'bg-indigo-50', text: 'text-indigo-600' };
    if (accessId.startsWith('wiki_')) return { icon: <BookOpen size={14} />, label: '知识库', bg: 'bg-amber-50', text: 'text-amber-600' };
    return { icon: null, label: '综合', bg: 'bg-slate-50', text: 'text-slate-600' };
  };

  // --- Detail View Component ---
  if (selectedProduct) {
    const isOwned = hasAccess(selectedProduct.accessId);
    const isInCart = cart.some(i => i.id === selectedProduct.id);
    const typeInfo = getProductTypeInfo(selectedProduct.accessId);

    return (
      <motion.div
        initial={{ opacity: 0, x: 20 }}
        animate={{ opacity: 1, x: 0 }}
        exit={{ opacity: 0, x: -20 }}
        className="max-w-5xl mx-auto"
      >
        <button
          onClick={() => setSelectedProduct(null)}
          className="flex items-center text-slate-500 hover:text-slate-900 mb-6 group transition-colors"
        >
          <div className="w-8 h-8 rounded-full bg-white border border-slate-200 flex items-center justify-center mr-2 group-hover:bg-slate-50">
            <ArrowLeft size={16} />
          </div>
          返回商城列表
        </button>

        <div className="bg-white rounded-3xl shadow-xl overflow-hidden border border-slate-100 flex flex-col md:flex-row">
          {/* Left: Image & Visuals */}
          <div className="md:w-5/12 relative bg-slate-100 min-h-[300px] md:min-h-full">
            <img
              src={selectedProduct.imageUrl}
              alt={selectedProduct.title}
              className="absolute inset-0 w-full h-full object-cover"
            />
            <div className="absolute inset-0 bg-gradient-to-t from-black/60 to-transparent flex flex-col justify-end p-8 text-white">
              <div className={`inline-flex items-center gap-1 px-3 py-1 rounded-full text-xs font-bold w-fit mb-2 backdrop-blur-md bg-white/20`}>
                {typeInfo.icon} {typeInfo.label}
              </div>
              <h2 className="text-3xl font-bold leading-tight">{selectedProduct.title}</h2>
            </div>
          </div>

          {/* Right: Info & Actions */}
          <div className="md:w-7/12 p-8 md:p-10 flex flex-col">
            <div className="flex gap-2 mb-6">
              {selectedProduct.tags.map(tag => (
                <span key={tag} className="px-3 py-1 bg-slate-100 text-slate-600 rounded-full text-xs font-bold uppercase tracking-wide">
                  {tag}
                </span>
              ))}
            </div>

            <div className="prose prose-slate mb-8">
              <h3 className="text-lg font-bold text-slate-900 mb-2">课程简介</h3>
              <p className="text-slate-600 leading-relaxed">
                {selectedProduct.description}
                本订阅包将为您解锁平台内对应的{typeInfo.label}内容权限。购买后，您可以在个人中心查看订阅状态，并立即开始学习。
              </p>
            </div>

            {/* Feature List Mockup */}
            <div className="space-y-3 mb-8 bg-slate-50 p-6 rounded-2xl border border-slate-100">
              <h4 className="font-bold text-sm text-slate-800 mb-2">订阅权益包含：</h4>
              <div className="flex items-center gap-3 text-sm text-slate-600">
                <Check size={16} className="text-emerald-500" />
                <span>解锁 {selectedProduct.accessId.split('_')[1] === 'internal' ? '内科学' : selectedProduct.accessId.split('_')[1] === 'surgery' ? '外科学' : '病理学'} 核心{typeInfo.label}内容</span>
              </div>
              <div className="flex items-center gap-3 text-sm text-slate-600">
                <ShieldCheck size={16} className="text-blue-500" />
                <span>有效期 {selectedProduct.duration}，到期自动提醒</span>
              </div>
              <div className="flex items-center gap-3 text-sm text-slate-600">
                <Zap size={16} className="text-amber-500" />
                <span>享受 AI 助教优先答疑服务</span>
              </div>
            </div>

            <div className="mt-auto pt-6 border-t border-slate-100 flex items-center justify-between gap-6">
              <div>
                <div className="text-sm text-slate-400 font-medium mb-1">订阅价格</div>
                <div className="flex items-baseline gap-1">
                  <span className="text-sm font-bold text-blue-600">¥</span>
                  <span className="text-4xl font-bold text-slate-900">{selectedProduct.price}</span>
                  <span className="text-sm text-slate-400">/{selectedProduct.duration}</span>
                </div>
              </div>

              <div className="flex-1 flex justify-end">
                {isOwned ? (
                  <button disabled className="bg-slate-100 text-slate-400 px-8 py-4 rounded-xl font-bold flex items-center gap-2 cursor-default w-full justify-center">
                    <Check size={20} /> 您已拥有此订阅
                  </button>
                ) : isInCart ? (
                  <button
                    onClick={() => setIsCartOpen(true)}
                    className="bg-green-50 text-green-600 border border-green-200 px-8 py-4 rounded-xl font-bold flex items-center gap-2 w-full justify-center hover:bg-green-100 transition-colors"
                  >
                    <ShoppingBag size={20} /> 已在购物车
                  </button>
                ) : (
                  <button
                    onClick={() => { addToCart(selectedProduct); setIsCartOpen(true); }}
                    className="bg-blue-600 text-white px-8 py-4 rounded-xl font-bold flex items-center gap-2 shadow-lg shadow-blue-500/30 hover:bg-blue-700 active:scale-95 transition-all w-full justify-center"
                  >
                    <Plus size={20} /> 加入购物车
                  </button>
                )}
              </div>
            </div>
          </div>
        </div>
      </motion.div>
    );
  }

  // --- Main List View ---
  return (
    <div>
      {/* Store Header */}
      <div className="flex flex-col md:flex-row justify-between items-center mb-8 gap-4">
        <div>
          <h2 className="text-2xl font-bold text-slate-900">精品商城</h2>
          <p className="text-slate-500">分别订阅专业题库与知识库，定制你的学习计划</p>
        </div>

        <div className="flex items-center gap-4 w-full md:w-auto">
          <div className="relative flex-1 md:w-64">
            <Search className="absolute left-3 top-2.5 text-slate-400" size={18} />
            <input
              type="text"
              placeholder="搜索课程、题库..."
              value={searchTerm}
              onChange={e => setSearchTerm(e.target.value)}
              className="w-full bg-white border border-slate-200 rounded-full py-2 pl-10 pr-4 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>
          <button
            onClick={() => setIsCartOpen(true)}
            className="group relative px-5 py-2.5 bg-white border border-slate-200 rounded-full hover:bg-slate-50 transition-all flex items-center gap-2 shadow-sm hover:shadow-md"
          >
            <ShoppingBag size={20} className="text-slate-600 group-hover:text-blue-600 transition-colors" />
            <span className="font-bold text-slate-700 text-sm group-hover:text-blue-600 group-hover:scale-110 transition-all duration-300 origin-center">
              购物车
            </span>
            {cart.length > 0 && (
              <span className="bg-rose-500 text-white text-xs font-bold h-5 min-w-[20px] px-1.5 flex items-center justify-center rounded-full ml-1 animate-pulse">
                {cart.length}
              </span>
            )}
          </button>
        </div>
      </div>

      {/* Product Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {filteredProducts.map(product => {
          const isOwned = hasAccess(product.accessId);
          const isInCart = cart.some(i => i.id === product.id);
          const typeInfo = getProductTypeInfo(product.accessId);

          return (
            <motion.div
              layout
              key={product.id}
              initial={{ opacity: 0, scale: 0.95 }}
              animate={{ opacity: 1, scale: 1 }}
              whileHover={{ y: -5 }}
              onClick={() => setSelectedProduct(product)}
              className={`bg-white rounded-2xl border transition-all overflow-hidden flex flex-col cursor-pointer group ${isOwned ? 'border-emerald-200 shadow-md ring-1 ring-emerald-100' : 'border-slate-100 shadow-sm hover:shadow-xl'}`}
            >
              <div className="h-48 overflow-hidden relative">
                <img
                  src={product.imageUrl}
                  alt={product.title}
                  className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-500"
                />
                {isOwned && (
                  <div className="absolute top-2 right-2 bg-emerald-500 text-white text-xs font-bold px-2 py-1 rounded shadow-md flex items-center gap-1">
                    <Check size={12} /> 已订阅
                  </div>
                )}
                {/* Type Badge */}
                <div className={`absolute top-2 left-2 ${typeInfo.bg} ${typeInfo.text} text-xs font-bold px-2 py-1 rounded flex items-center gap-1 shadow-sm`}>
                  {typeInfo.icon} {typeInfo.label}
                </div>
              </div>
              <div className="p-5 flex flex-col flex-1">
                <div className="flex gap-2 mb-2">
                  {product.tags.map(tag => (
                    <span key={tag} className="text-[10px] bg-slate-100 text-slate-600 px-2 py-0.5 rounded-full">#{tag}</span>
                  ))}
                </div>
                <h3 className="font-bold text-lg mb-2 text-slate-900 group-hover:text-blue-600 transition-colors">{product.title}</h3>
                <p className="text-sm text-slate-500 mb-4 flex-1 line-clamp-2">{product.description}</p>
                <div className="flex items-center justify-between mt-auto pt-4 border-t border-slate-50">
                  <div className="flex items-baseline">
                    <span className="text-sm text-blue-600 font-bold">¥</span>
                    <span className="text-2xl font-bold text-slate-900">{product.price}</span>
                    <span className="text-xs text-slate-400 ml-1">/{product.duration}</span>
                  </div>
                  <button
                    onClick={(e) => {
                      e.stopPropagation(); // Prevent opening detail view
                      if (!isOwned && !isInCart) addToCart(product);
                    }}
                    disabled={isOwned || isInCart}
                    className={`
                      flex items-center gap-1 px-4 py-2 rounded-lg text-sm font-medium transition-colors
                      ${isOwned
                        ? 'bg-emerald-50 text-emerald-600 cursor-default'
                        : isInCart
                          ? 'bg-green-100 text-green-700'
                          : 'bg-slate-900 text-white hover:bg-blue-600 shadow-md hover:shadow-lg'}
                    `}
                  >
                    {isOwned ? <><Check size={16} /> 查看</> : isInCart ? <><Check size={16} /> 已添加</> : <><Plus size={16} /> 订阅</>}
                  </button>
                </div>
              </div>
            </motion.div>
          );
        })}
      </div>

      {/* Cart Drawer */}
      <AnimatePresence>
        {isCartOpen && (
          <>
            <motion.div
              initial={{ opacity: 0 }} animate={{ opacity: 1 }} exit={{ opacity: 0 }}
              className="fixed inset-0 bg-slate-900/40 backdrop-blur-sm z-50"
              onClick={() => setIsCartOpen(false)}
            />
            <motion.div
              initial={{ x: '100%' }} animate={{ x: 0 }} exit={{ x: '100%' }}
              transition={{ type: "spring", damping: 30, stiffness: 300 }}
              className="fixed inset-y-0 right-0 w-full max-w-md bg-white shadow-2xl z-50 flex flex-col"
            >
              <div className="p-5 border-b border-slate-100 flex items-center justify-between bg-white">
                <h2 className="text-xl font-bold flex items-center gap-2">
                  <ShoppingBag className="text-blue-600" />
                  购物车
                </h2>
                <button onClick={() => setIsCartOpen(false)} className="p-2 hover:bg-slate-100 rounded-full">
                  <X size={20} />
                </button>
              </div>

              <div className="flex-1 overflow-y-auto p-5 space-y-4 bg-slate-50">
                {cart.length === 0 ? (
                  <div className="h-full flex flex-col items-center justify-center text-slate-400">
                    <ShoppingBag size={48} className="opacity-20 mb-4" />
                    <p>购物车是空的</p>
                  </div>
                ) : (
                  cart.map(item => (
                    <div key={item.cartId} className="bg-white p-4 rounded-xl shadow-sm border border-slate-100 flex gap-4">
                      <img src={item.imageUrl} alt={item.title} className="w-16 h-16 rounded-lg object-cover bg-slate-100" />
                      <div className="flex-1">
                        <h4 className="font-bold text-sm text-slate-800 line-clamp-2">{item.title}</h4>
                        <div className="flex items-center justify-between mt-2">
                          <span className="text-blue-600 font-bold">¥{item.price}</span>
                          <button onClick={() => removeFromCart(item.cartId)} className="text-slate-400 hover:text-red-500">
                            <Trash2 size={16} />
                          </button>
                        </div>
                      </div>
                    </div>
                  ))
                )}
              </div>

              {cart.length > 0 && (
                <div className="p-6 bg-white border-t border-slate-100 space-y-4">

                  {/* Coupon Input */}
                  <div className="bg-slate-50 p-3 rounded-xl">
                    <div className="flex gap-2">
                      <input
                        type="text"
                        placeholder="输入优惠码"
                        className="flex-1 px-3 py-2 border border-slate-200 rounded-lg text-sm focus:outline-none focus:border-indigo-500"
                        value={couponCode}
                        onChange={(e) => setCouponCode(e.target.value)}
                      />
                      <button
                        onClick={handleValidateCoupon}
                        disabled={isValidating || !couponCode}
                        className="px-3 py-2 bg-indigo-600 text-white rounded-lg text-sm font-bold hover:bg-indigo-700 disabled:opacity-50"
                      >
                        {isValidating ? '...' : '应用'}
                      </button>
                    </div>
                    {couponMsg && (
                      <div className={`text-xs mt-2 px-1 font-bold ${discount > 0 ? 'text-emerald-600' : 'text-rose-500'}`}>
                        {couponMsg}
                      </div>
                    )}
                  </div>

                  <div className="space-y-2 text-slate-500 text-sm">
                    <div className="flex justify-between">
                      <span>商品总额</span>
                      <span>¥{cartTotal}</span>
                    </div>
                    {discount > 0 && (
                      <div className="flex justify-between text-emerald-600 font-bold">
                        <span>优惠减免</span>
                        <span>-¥{discount}</span>
                      </div>
                    )}
                    <div className="flex justify-between text-lg font-bold text-slate-900 border-t border-slate-100 pt-2">
                      <span>总计</span>
                      <span>¥{Math.max(0, cartTotal - discount)}</span>
                    </div>
                  </div>

                  <div className="bg-slate-50 p-3 rounded-xl mt-4">
                    <label className="text-sm font-bold text-slate-700 mb-2 block">支付方式</label>
                    <div className="flex gap-2">
                      {enabledMethods.alipay && (
                        <button
                          onClick={() => setPayType('alipay')}
                          className={`flex-1 py-2 rounded-lg border text-sm font-medium flex items-center justify-center gap-2 ${payType === 'alipay' ? 'border-blue-500 bg-blue-50 text-blue-700' : 'border-slate-200 bg-white text-slate-600'}`}
                        >
                          <svg className="w-4 h-4" viewBox="0 0 1024 1024" version="1.1" xmlns="http://www.w3.org/2000/svg"><path d="M783.504 380.128h-190.256c-21.784 122.992-80.464 227.144-162.832 301.992-87.168-51.528-144.112-119.56-180.208-192.512h172.992c6.264-21.904 11.232-44.408 14.896-67.368h-198.88c5.448-18.784 11.64-37.16 18.264-55.112h206.144c13.72-23.864 25.128-49.336 34.256-76.04h-350.416v-67.664h175.056v-64.84h66.72v64.84h170.832v67.664h-100.224c-9.04 29.56-20.944 57.656-35.344 83.824h139.752v55.112h-95.208c-6.288 23.336-13.68 45.92-22.184 67.368h112.504v62.832zM327.952 546.912c32.744 58.736 78.432 112.008 136.368 153.536-107.032 87.2-243.344 114.392-362.592 119.52v-65.016c87.312-3.768 184.448-24.576 261.264-77.424-13.6-11.416-26.28-23.512-38.04-36.216l3.008-94.4zM533.12 739.04c63.672-68.808 107.864-159.264 127.328-261.736h-66.216c-17.616 80.936-54.76 153.4-106.328 210.024l45.216 51.712z m286.912 114.704h-212.44v-316.4h212.44v316.4zM635.08 827.248V564.08h158.424v263.168H635.08z" fill="#027AFF" /></svg>
                          支付宝
                        </button>
                      )}

                      {enabledMethods.wechat && (
                        <button
                          onClick={() => setPayType('wechat')}
                          className={`flex-1 py-2 rounded-lg border text-sm font-medium flex items-center justify-center gap-2 ${payType === 'wechat' ? 'border-emerald-500 bg-emerald-50 text-emerald-700' : 'border-slate-200 bg-white text-slate-600'}`}
                        >
                          <svg className="w-4 h-4" viewBox="0 0 1024 1024" version="1.1" xmlns="http://www.w3.org/2000/svg"><path d="M667.5 561.4c-9.2-8.3-9-22 0-30.2 9.4-8.6 9-22 0-30.5-54.2-52.1-137.9-74.8-197.8-26.9-63.5 50.8-72 135.2-19.1 190.9 44.9 47.3 115.8 54.4 167.3 18.2l7.7 20.9c2.3 6.3 9.4 9.6 15.6 7.2 6.1-2.4 9.3-9.5 7.1-15.6l-8.5-23.3 22.1-4.7c6.5-1.4 10.6-7.8 9.3-14.3-1.4-6.5-7.7-10.6-14.2-9.3l-24.3 5.2c5-5.3 9.9-10.8 14.7-16.5 0.3-0.3 0.6-0.6 0.9-0.9 9.1-8.3 19.3-43.1 19.3-70.3z m-126-11.8c-10.6 0-19.2-8.6-19.2-19.2s8.6-19.2 19.2-19.2 19.2 8.6 19.2 19.2-8.6 19.2-19.2 19.2z m75.9 0c-10.6 0-19.2-8.6-19.2-19.2s8.6-19.2 19.2-19.2 19.2 8.6 19.2 19.2-8.6 19.2-19.2 19.2z" fill="#07C160" /><path d="M414 430.7c-7.3-6.5-7.3-17.1 0-23.7 7.2-6.5 7.4-16.8 0.4-23.7-41.9-41.3-108.3-60.6-156.9-22.3-51 40-58 107.5-15.8 152.1 36.3 38.3 93.6 44.1 135.6 15.1l5.9 16.3c1.8 5.1 7.3 7.7 12.3 5.9 4.9-1.9 7.6-7.5 5.8-12.4l-6.5-18.1 17.5-3.8c5.2-1.1 8.5-6.2 7.4-11.4-1.1-5.2-6.2-8.4-11.4-7.4l-19.4 4.2c3.9-4.2 7.7-8.6 11.4-13.1 0.2-0.2 0.4-0.5 0.7-0.7 7.3 0 13.2-27.4 13.2-49z m-100-8.9c-8.4 0-15.2-6.8-15.2-15.2s6.8-15.2 15.2-15.2 15.2 6.8 15.2 15.2-6.8 15.2-15.2 15.2z m60.4 0c-8.4 0-15.2-6.8-15.2-15.2s6.8-15.2 15.2-15.2 15.2 6.8 15.2 15.2-6.8 15.2-15.2 15.2z" fill="#07C160" /></svg>
                          微信
                        </button>
                      )}

                      {!enabledMethods.alipay && !enabledMethods.wechat && (
                        <div className="flex-1 text-center py-2 text-sm text-slate-400 bg-slate-100 rounded-lg">暂无可用支付方式</div>
                      )}
                    </div>
                  </div>

                  <button
                    onClick={() => { checkout(discount > 0 ? couponCode : undefined, payType); setIsCartOpen(false); }}
                    className="w-full bg-blue-600 text-white py-3.5 rounded-xl font-bold flex items-center justify-center gap-2 hover:bg-blue-700 active:scale-95 transition-all shadow-lg shadow-blue-500/20"
                  >
                    <CreditCard size={20} />
                    立即结算
                  </button>
                </div>
              )}
            </motion.div>
          </>
        )}
      </AnimatePresence>
    </div >
  );
};

export default StoreView;
