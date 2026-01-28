import { Module } from '@nestjs/common';
import { TypeOrmModule } from '@nestjs/typeorm';
import { DashboardController } from './dashboard.controller';
import { DashboardService } from './dashboard.service';
import { User } from '../users/user.entity';
import { Transaction } from '../store/entities/transaction.entity';
import { Feedback } from '../feedback/entities/feedback.entity';

@Module({
  imports: [TypeOrmModule.forFeature([User, Transaction, Feedback])],
  controllers: [DashboardController],
  providers: [DashboardService],
})
export class DashboardModule {}
