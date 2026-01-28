import { Module } from '@nestjs/common';
import { ServeStaticModule } from '@nestjs/serve-static';
import { join } from 'path';
import { TypeOrmModule } from '@nestjs/typeorm';
import { UsersModule } from './users/users.module';
import { AuthModule } from './auth/auth.module';
import { QuizModule } from './quiz/quiz.module';
import { WikiModule } from './wiki/wiki.module';
import { SeedModule } from './seed/seed.module';
import { ChatModule } from './chat/chat.module';
import { FeedbackModule } from './feedback/feedback.module';
import { User } from './users/user.entity';
import { Subject } from './quiz/entities/subject.entity';
import { Chapter } from './quiz/entities/chapter.entity';
import { Question } from './quiz/entities/question.entity';
import { Comment } from './quiz/entities/comment.entity';
import { WikiCategory } from './wiki/entities/wiki-category.entity';
import { Article } from './wiki/entities/article.entity';
import { Feedback } from './feedback/entities/feedback.entity';
import { AppController } from './app.controller';
import { AppService } from './app.service';
import { Product } from './store/entities/product.entity';
import { StoreModule } from './store/store.module';
import { Transaction } from './store/entities/transaction.entity';
import { DashboardModule } from './dashboard/dashboard.module';
import { SettingsModule } from './settings/settings.module';
import { Setting } from './settings/settings.entity';
import { AnnouncementsModule } from './announcements/announcements.module';
import { Announcement } from './announcements/entities/announcement.entity';
import { Coupon } from './store/entities/coupon.entity';
import { PaymentModule } from './payment/payment.module';

@Module({
  imports: [
    TypeOrmModule.forRoot({
      type: 'sqlite',
      database: 'medical_learning.db',
      entities: [
        User,
        Subject,
        Chapter,
        Question,
        Comment,
        WikiCategory,
        Article,
        Feedback,
        Product,
        Transaction,
        Setting,
        Announcement,
        Coupon,
      ],
      synchronize: true, // Dev only
    }),
    UsersModule,
    AuthModule,
    QuizModule,
    WikiModule,
    SeedModule,
    ChatModule,
    FeedbackModule,
    StoreModule,
    DashboardModule,
    SettingsModule,
    AnnouncementsModule,
    PaymentModule,
    ServeStaticModule.forRoot({
      rootPath: join(process.cwd(), 'uploads'),
      serveRoot: '/uploads',
    }),
  ],
  controllers: [AppController],
  providers: [AppService],
})
export class AppModule { }
