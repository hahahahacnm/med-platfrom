import { Injectable, OnModuleInit } from '@nestjs/common';
import { InjectRepository } from '@nestjs/typeorm';
import { Repository } from 'typeorm';
import { Subject } from '../quiz/entities/subject.entity';
import { Chapter } from '../quiz/entities/chapter.entity';
import { Question } from '../quiz/entities/question.entity';
import { WikiCategory } from '../wiki/entities/wiki-category.entity';
import { Article } from '../wiki/entities/article.entity';
import { SUBJECTS, WIKI_CATEGORIES } from './data.constant';

import { User } from '../users/user.entity';
import * as bcrypt from 'bcrypt';

@Injectable()
export class SeedService implements OnModuleInit {
  constructor(
    @InjectRepository(Subject) private subjectRepo: Repository<Subject>,
    @InjectRepository(Chapter) private chapterRepo: Repository<Chapter>,
    @InjectRepository(Question) private questionRepo: Repository<Question>,
    @InjectRepository(WikiCategory) private wikiRepo: Repository<WikiCategory>,
    @InjectRepository(Article) private articleRepo: Repository<Article>,
    @InjectRepository(User) private userRepo: Repository<User>,
  ) {}

  async onModuleInit() {
    await this.seedSubjects();
    await this.seedWiki();
    await this.seedUsers();
  }

  async seedUsers() {
    const adminExists = await this.userRepo.findOne({
      where: { email: 'admin@med.edu' },
    });
    if (adminExists) return;

    console.log('Seeding Users...');
    const hashedPassword = await bcrypt.hash('admin123', 10);

    const admin = this.userRepo.create({
      email: 'admin@med.edu',
      password: hashedPassword,
      name: 'System Admin',
      role: 'admin',
      balance: 99999,
      subscriptions: ['unlimited_pro'],
      createdAt: new Date(),
    });

    await this.userRepo.save(admin);
    console.log('Admin user created: admin@med.edu / admin123');
  }

  async seedSubjects() {
    const count = await this.subjectRepo.count();
    if (count > 0) return;
    console.log('Seeding Data...');

    for (const sub of SUBJECTS) {
      const subject = this.subjectRepo.create({
        id: sub.id,
        title: sub.title,
        description: sub.description,
        icon: sub.icon,
        color: sub.color,
      });
      await this.subjectRepo.save(subject);

      const chapters = this.parseRawData(sub.rawContent);

      for (const ch of chapters) {
        const chapter = this.chapterRepo.create({
          title: ch.title || '综合练习', // Default title fix
          subject: subject,
        });
        await this.chapterRepo.save(chapter);

        for (const q of ch.questions) {
          const question = this.questionRepo.create({
            text: q.text,
            options: q.options,
            correctAnswers: q.correctAnswers,
            explanation: q.explanation || '',
            chapter: chapter,
          });
          await this.questionRepo.save(question);
        }
      }
    }
  }

  async seedWiki() {
    const count = await this.wikiRepo.count();
    if (count > 0) return;

    for (const cat of WIKI_CATEGORIES) {
      const category = this.wikiRepo.create({
        id: cat.id,
        title: cat.title,
        description: cat.description,
        iconName: cat.iconName,
        color: cat.color,
      });
      await this.wikiRepo.save(category);

      for (const art of cat.articles) {
        const article = this.articleRepo.create({
          ...art,
          category: category,
        });
        await this.articleRepo.save(article);
      }
    }
  }

  parseRawData(rawData: string): any[] {
    const chapters: any[] = [];
    const lines = rawData.split('\n').filter((l) => l.trim());

    let currentChapter: any = { title: '综合练习', questions: [] };
    let currentQ: any = null;

    const pushQuestion = () => {
      if (currentQ) {
        currentChapter.questions.push(currentQ);
        currentQ = null;
      }
    };

    const pushChapter = () => {
      pushQuestion();
      if (currentChapter.questions.length > 0) {
        chapters.push({ ...currentChapter });
      }
    };

    lines.forEach((line) => {
      line = line.trim();
      if (line.startsWith('###')) {
        pushChapter();
        const title = line.replace(/^###\s*/, '').trim();
        currentChapter = { title: title, questions: [] };
        return;
      }
      if (/^\d+[\.、]/.test(line)) {
        pushQuestion();
        currentQ = {
          text: line.replace(/^\d+[\.、]/, '').trim(),
          options: [],
          correctAnswers: [],
        };
      } else if (/^[A-E][\.、]/.test(line) && currentQ) {
        currentQ.options.push({ id: line[0], text: line.substring(2).trim() });
      } else if (/^答案[：:]/i.test(line) && currentQ) {
        currentQ.correctAnswers = line
          .replace(/^答案[：:]/i, '')
          .trim()
          .split('');
      } else if (/^解析[：:]/i.test(line) && currentQ) {
        currentQ.explanation = line.replace(/^解析[：:]/i, '').trim();
      }
    });

    pushChapter();
    if (chapters.length === 0 && currentChapter.questions.length > 0) {
      return [currentChapter];
    }
    return chapters;
  }
}
