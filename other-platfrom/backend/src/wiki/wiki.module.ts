import { Module } from '@nestjs/common';
import { TypeOrmModule } from '@nestjs/typeorm';
import { WikiService } from './wiki.service';
import { WikiController } from './wiki.controller';
import { WikiCategory } from './entities/wiki-category.entity';
import { Article } from './entities/article.entity';

@Module({
  imports: [TypeOrmModule.forFeature([WikiCategory, Article])],
  controllers: [WikiController],
  providers: [WikiService],
})
export class WikiModule {}
