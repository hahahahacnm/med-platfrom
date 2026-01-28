import { Injectable } from '@nestjs/common';
import { InjectRepository } from '@nestjs/typeorm';
import { Repository } from 'typeorm';
import { User } from '../users/user.entity';
import { Transaction } from '../store/entities/transaction.entity';
import { Feedback } from '../feedback/entities/feedback.entity';

@Injectable()
export class DashboardService {
  constructor(
    @InjectRepository(User)
    private userRepo: Repository<User>,
    @InjectRepository(Transaction)
    private transactionRepo: Repository<Transaction>,
    @InjectRepository(Feedback)
    private feedbackRepo: Repository<Feedback>,
  ) {}

  async getStats() {
    const now = new Date();
    const sevenDaysAgo = new Date(now);
    sevenDaysAgo.setDate(now.getDate() - 7);

    const totalRevenue = (await this.transactionRepo.sum('amount')) || 0;

    // --- Users Stats ---
    const userCount = await this.userRepo.count();
    // For SQLite, dates are stored as strings or numbers usually, but TypeORM handles Date objects if configured.
    // If issues arise, we might need raw query or string comparison. Assuming TypeORM Map works.
    // Note: In SQLite default, dates are not always perfect with standard operators if not uniform.
    // But let's try standard TypeORM operator first.
    const newUsersLast7Days = await this.userRepo
      .createQueryBuilder('user')
      .where('user.createdAt > :date', { date: sevenDaysAgo })
      .getCount();

    const usersPrevTotal = userCount - newUsersLast7Days;
    const usersTrend =
      usersPrevTotal > 0
        ? (newUsersLast7Days / usersPrevTotal) * 100
        : newUsersLast7Days > 0
          ? 100
          : 0;

    // --- Active Data Calculation ---
    const users = await this.userRepo.find({
      select: ['subscriptions', 'quizHistory'],
    });
    const activeSubsCount = users.filter(
      (u) => u.subscriptions && u.subscriptions.length > 0,
    ).length;

    // Active Subs Trend Proxy: Based on recent transactions
    // We assume new transactions ~ new subs (ignoring renewals for simplicity in 'Trend')
    const newTxLast7Days = await this.transactionRepo
      .createQueryBuilder('tx')
      .where('tx.createdAt > :date', { date: sevenDaysAgo })
      .getCount();

    const activeSubsPrev = Math.max(0, activeSubsCount - newTxLast7Days);
    const activeSubsTrend =
      activeSubsPrev > 0
        ? (newTxLast7Days / activeSubsPrev) * 100
        : newTxLast7Days > 0
          ? 100
          : 0;

    // --- Quiz Stats ---
    let totalQuizCount = 0;
    let quizzesLast7Days = 0;

    users.forEach((u) => {
      if (u.quizHistory && Array.isArray(u.quizHistory)) {
        totalQuizCount += u.quizHistory.length;
        u.quizHistory.forEach((q: any) => {
          const dateStr = q.completedAt || q.date;
          if (dateStr) {
            const d = new Date(dateStr);
            if (d > sevenDaysAgo) {
              quizzesLast7Days++;
            }
          }
        });
      }
    });

    const quizPrevTotal = Math.max(0, totalQuizCount - quizzesLast7Days);
    const quizCountTrend =
      quizPrevTotal > 0
        ? (quizzesLast7Days / quizPrevTotal) * 100
        : quizzesLast7Days > 0
          ? 100
          : 0;

    return {
      revenue: totalRevenue,
      users: userCount,
      usersTrend,
      activeSubs: activeSubsCount,
      activeSubsTrend,
      quizCount: totalQuizCount,
      quizCountTrend,
    };
  }

  async getRevenueTrend(days: number = 30) {
    // Get transactions from the last N days
    const startDate = new Date();
    startDate.setDate(startDate.getDate() - days);

    const transactions = await this.transactionRepo
      .createQueryBuilder('transaction')
      .where('transaction.createdAt >= :date', { date: startDate })
      .getMany();

    // Group by date
    const dailyRevenue = new Map<string, number>();

    // Initialize last N days with 0
    for (let i = 0; i < days; i++) {
      const d = new Date();
      d.setDate(d.getDate() - i);
      const dateStr = d.toISOString().split('T')[0]; // YYYY-MM-DD
      dailyRevenue.set(dateStr, 0);
    }

    transactions.forEach((t) => {
      const dateStr =
        t.createdAt instanceof Date
          ? t.createdAt.toISOString().split('T')[0]
          : new Date(t.createdAt).toISOString().split('T')[0];

      const current = dailyRevenue.get(dateStr) || 0;
      dailyRevenue.set(dateStr, current + t.amount);
    });

    // Convert to array and sort by date
    return Array.from(dailyRevenue.entries())
      .map(([date, amount]) => ({ date, amount }))
      .sort((a, b) => a.date.localeCompare(b.date));
  }

  async getSubjectDistribution() {
    const allUsers = await this.userRepo.find({ select: ['quizHistory'] });
    const distribution = new Map<string, number>();

    allUsers.forEach((user) => {
      if (user.quizHistory && Array.isArray(user.quizHistory)) {
        user.quizHistory.forEach((record: any) => {
          // record.subjectId format: "Subject - Chapter [Mode]"
          // We extract the subject part
          if (record.subjectId && typeof record.subjectId === 'string') {
            const parts = record.subjectId.split(' - ');
            if (parts.length > 0) {
              const subjectName = parts[0].trim();
              const count = Number(record.total) || 0;
              distribution.set(
                subjectName,
                (distribution.get(subjectName) || 0) + count,
              );
            }
          }
        });
      }
    });

    return Array.from(distribution.entries())
      .map(([name, value]) => ({ name, value }))
      .sort((a, b) => b.value - a.value); // Sort by count descending
  }
}
