
import { Module } from '@nestjs/common';
import { TypeOrmModule } from '@nestjs/typeorm';
import { QuizService } from './quiz.service';
import { QuizController } from './quiz.controller';
import { Subject } from './entities/subject.entity';
import { Chapter } from './entities/chapter.entity';
import { Question } from './entities/question.entity';

import { Comment } from './entities/comment.entity';

@Module({
  imports: [TypeOrmModule.forFeature([Subject, Chapter, Question, Comment])],
  controllers: [QuizController],
  providers: [QuizService],
})
export class QuizModule { }
