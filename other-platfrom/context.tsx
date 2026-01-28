
import { createContext, useContext } from 'react';
import { UserState, CartItem, ViewState, Product, QuizResult, AccessId, Bookmark, WikiCategory, Article } from './types';

export interface AppContextType {
  user: UserState;
  cart: CartItem[];
  view: ViewState;
  setView: (v: ViewState) => void;
  addToCart: (p: Product) => void;
  removeFromCart: (id: string) => void;
  checkout: (couponCode?: string, payType?: string) => void;
  addQuizResult: (r: QuizResult) => void;
  hasAccess: (id: AccessId) => boolean;
  // Quiz Progress
  updateChapterProgress: (subjectId: string, chapterTitle: string, index: number, isCorrect?: boolean) => void;
  resetChapterProgress: (subjectId: string, chapterTitle: string) => void;
  // Bookmarks
  toggleBookmark: (item: Omit<Bookmark, 'timestamp'>) => void;
  isBookmarked: (id: string) => boolean;
  // Wiki Navigation (Lifted State)
  wikiState: { category: WikiCategory | null; article: Article | null };
  setWikiState: (state: { category: WikiCategory | null; article: Article | null }) => void;
  // Authentication
  login: (userData: any) => void;
  logout: () => void;
}

export const AppContext = createContext<AppContextType | null>(null);

export const useAppContext = () => {
  const context = useContext(AppContext);
  if (!context) throw new Error("useAppContext must be used within AppProvider");
  return context;
};