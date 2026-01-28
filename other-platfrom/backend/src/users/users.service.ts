import { Injectable, NotFoundException } from '@nestjs/common';
import { InjectRepository } from '@nestjs/typeorm';
import { Repository } from 'typeorm';
import { User } from './user.entity';

@Injectable()
export class UsersService {
  constructor(
    @InjectRepository(User)
    private usersRepository: Repository<User>,
  ) { }

  async findOne(email: string): Promise<User | null> {
    return this.usersRepository.findOne({ where: { email } });
  }

  async findById(id: string): Promise<User | null> {
    return this.usersRepository.findOne({ where: { id } });
  }

  async create(user: Partial<User>): Promise<User> {
    const newUser = this.usersRepository.create(user);
    return this.usersRepository.save(newUser);
  }

  async update(id: string, updateData: Partial<User>): Promise<User> {
    await this.usersRepository.update(id, updateData);
    const user = await this.findById(id);
    if (!user) throw new NotFoundException('User not found');
    return user;
  }

  async findAllWithStats() {
    const users = await this.usersRepository.find();
    return users.map((user) => {
      const quizHistory = user.quizHistory || [];
      let totalQuestions = 0;
      let totalCorrect = 0;

      quizHistory.forEach((record: any) => {
        const total = Number(record.total) || 0;
        const correct = Number(record.correct) || 0;
        totalQuestions += total;
        totalCorrect += correct;
      });

      const accuracy =
        totalQuestions > 0
          ? Math.round((totalCorrect / totalQuestions) * 100)
          : 0;

      return {
        id: user.id,
        name: user.name,
        email: user.email,
        role: user.role,
        createdAt: user.createdAt,
        stats: {
          quizCount: totalQuestions,
          accuracy: `${accuracy}%`,
          subscriptions: user.subscriptions || [],
        },
      };
    });
  }

  async updateUserSubscriptions(userId: string, subscriptions: any[]) {
    const user = await this.findById(userId);
    if (!user) throw new NotFoundException('User not found');
    user.subscriptions = subscriptions;
    return this.usersRepository.save(user);
  }

  async remove(id: string) {
    const user = await this.findById(id);
    if (!user) throw new NotFoundException('User not found');
    return this.usersRepository.remove(user);
  }
}
