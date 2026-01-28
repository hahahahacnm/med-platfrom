import { Module } from '@nestjs/common';
import { TypeOrmModule } from '@nestjs/typeorm';
import { SeedService } from './seed.service';
import { Subject } from '../quiz/entities/subject.entity';
import { Chapter } from '../quiz/entities/chapter.entity';
import { Question } from '../quiz/entities/question.entity';
import { WikiCategory } from '../wiki/entities/wiki-category.entity';
import { Article } from '../wiki/entities/article.entity';
import { User } from '../users/user.entity';

@Module({
  imports: [
    TypeOrmModule.forFeature([
      Subject,
      Chapter,
      Question,
      WikiCategory,
      Article,
      User,
    ]),
  ],
  providers: [SeedService],
  exports: [SeedService],
})
export class SeedModule {}
