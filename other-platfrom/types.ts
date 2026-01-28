
// --- General ---
export type ViewState = 'home' | 'store' | 'quiz' | 'wiki' | 'profile' | 'favorites' | 'mistakes' | 'assistant' | 'admin' | 'feedback';

// --- Access Control ---
export type AccessId = string;

// --- Store Types ---
export interface Product {
  id: string;
  title: string;
  description: string;
  price: number;
  duration: string;
  imageUrl: string;
  tags: string[];
  accessId: AccessId;
  durationValue?: number;
  durationUnit?: 'day' | 'month' | 'year' | 'forever';
  isPublished?: boolean;
  coupons?: Coupon[];
}

export interface Coupon {
  id: string;
  code: string;
  type: 'amount' | 'percent';
  value: number;
  usageLimit: number;
  usedCount: number;
  productId: string;
}

export interface CartItem extends Product {
  cartId: string;
}

// --- Quiz Types ---
export interface Option {
  id: string;
  text: string;
}

export interface Question {
  id: number;
  text: string;
  options: Option[];
  correctAnswers: string[];
  explanation: string;
}

export interface Comment {
  id: number;
  content: string;
  createdAt: string;
  user: {
    id: string;
    name: string;
    avatar?: string;
  };
}

export interface Chapter {
  id: string;
  title: string;
  questions: Question[];
}

export interface Subject {
  id: AccessId;
  title: string;
  description: string;
  icon: string;
  color: string;
  rawContent?: string;
  chapters?: Chapter[];
}

export interface QuizResult {
  subjectId: string;
  total: number;
  correct: number;
  date: string;
}

// New: Quiz Modes
export type QuizMode = 'practice' | 'fast' | 'test' | 'study';

// New: Detailed Progress Tracking
export interface ChapterProgress {
  lastIndex: number; // Index of the last question viewed/answered
  history: Record<number, boolean>; // Question Index -> true(Correct) | false(Wrong)
}

// --- Wiki Types ---
export interface Article {
  id: string;
  title: string;
  excerpt: string;
  content: string;
  author: string;
  readTime: string;
  date: string;
  tags: string[];
  status?: string;
  category?: WikiCategory;
}

export interface WikiCategory {
  id: AccessId;
  title: string;
  description: string;
  iconName: string;
  color: string;
  articles: Article[];
}

// --- Bookmark Types ---
export interface Bookmark {
  id: string; // Unique composite ID
  type: 'question' | 'article';
  title: string;
  path: string; // e.g. "Internal Medicine > Chapter 1"
  data?: any; // Store Question object or Article ID
  timestamp: number;
}

// --- User Context Types ---
export interface Subscription {
  accessId: AccessId;
  startDate: string;
  expiresAt: string | null;
}

export interface UserState {
  isAuthenticated: boolean; // New
  name: string;
  email: string;
  role?: string; // New: admin check
  subscriptions: (AccessId | Subscription)[];
  balance: number;
  quizHistory: QuizResult[]; // Summary results
  chapterProgress: Record<string, ChapterProgress>; // Detailed progress by chapterId
  bookmarks: Bookmark[];
  avatar?: string;
}