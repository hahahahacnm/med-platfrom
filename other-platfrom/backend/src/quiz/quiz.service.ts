
import { Injectable, NotFoundException, ConflictException } from '@nestjs/common';
import { InjectRepository } from '@nestjs/typeorm';
import { Repository } from 'typeorm';
import { Subject } from './entities/subject.entity';
import { Chapter } from './entities/chapter.entity';
import { Question } from './entities/question.entity';
import { Comment } from './entities/comment.entity';
import { CreateSubjectDto, UpdateSubjectDto } from './dto/subject.dto';
import { CreateCommentDto } from './dto/comment.dto';
import { User } from '../users/user.entity';

@Injectable()
export class QuizService {
  constructor(
    @InjectRepository(Subject) private subjectRepo: Repository<Subject>,
    @InjectRepository(Chapter) private chapterRepo: Repository<Chapter>,
    @InjectRepository(Question) private questionRepo: Repository<Question>,
    @InjectRepository(Comment) private commentRepo: Repository<Comment>
  ) { }

  findAll() {
    return this.subjectRepo.find({
      relations: ['chapters', 'chapters.questions'],
      order: { id: 'ASC' }
    });
  }

  async findChapters(subjectId: string) {
    return this.chapterRepo.find({
      where: { subject: { id: subjectId } },
      relations: ['questions'],
      order: { title: 'ASC' }
    });
  }

  async createSubject(dto: CreateSubjectDto): Promise<Subject> {
    const exists = await this.subjectRepo.findOne({ where: { id: dto.id } });
    if (exists) {
      throw new ConflictException(`Subject with ID ${dto.id} already exists`);
    }
    const subject = this.subjectRepo.create(dto);
    return this.subjectRepo.save(subject);
  }

  async updateSubject(id: string, dto: UpdateSubjectDto): Promise<Subject> {
    const subject = await this.subjectRepo.findOne({ where: { id } });
    if (!subject) {
      throw new NotFoundException(`Subject with ID ${id} not found`);
    }
    Object.assign(subject, dto);
    return this.subjectRepo.save(subject);
  }

  async deleteSubject(id: string): Promise<{ success: boolean }> {
    const result = await this.subjectRepo.delete(id);
    if (result.affected === 0) {
      throw new NotFoundException(`Subject with ID ${id} not found`);
    }
    return { success: true };
  }

  async deleteChapter(id: string): Promise<{ success: boolean }> {
    const result = await this.chapterRepo.delete(id);
    if (result.affected === 0) {
      throw new NotFoundException(`Chapter with ID ${id} not found`);
    }
    return { success: true };
  }

  async updateQuestion(id: number, data: any): Promise<Question> {
    const question = await this.questionRepo.findOne({ where: { id } });
    if (!question) throw new NotFoundException('Question not found');

    Object.assign(question, data);
    return this.questionRepo.save(question);
  }

  async importQuestions(subjectId: string, files: any[]) { // Using any[] to avoid Multer type issues if not globally available
    const subject = await this.subjectRepo.findOne({ where: { id: subjectId } });
    if (!subject) throw new NotFoundException('Subject not found');

    const XLSX = require('xlsx');

    let totalQuestions = 0;
    for (const file of files) {
      // Use filename as chapter title (remove extension)
      // Decode filename in case it's URL encoded
      const originalName = Buffer.from(file.originalname, 'latin1').toString('utf8');
      const title = originalName.replace(/\.[^/.]+$/, "");

      let chapter = await this.chapterRepo.findOne({
        where: { title, subject: { id: subjectId } }
      });

      if (!chapter) {
        chapter = this.chapterRepo.create({
          title,
          subject
        });
        chapter = await this.chapterRepo.save(chapter);
      }

      const workbook = XLSX.read(file.buffer, { type: 'buffer' });
      const sheetName = workbook.SheetNames[0];
      const sheet = workbook.Sheets[sheetName];
      const data = XLSX.utils.sheet_to_json(sheet);

      const questions: Question[] = [];
      for (const row of data) {
        if (!row['题干']) continue;

        const options: { id: string, text: string }[] = [];
        const optionKeys = ['A', 'B', 'C', 'D', 'E', 'F'];
        for (const key of optionKeys) {
          const val = row[`选项${key}`];
          if (val) {
            options.push({ id: key, text: val.toString() });
          }
        }

        let correctAnswers: string[] = [];

        if (options.length > 0) {
          // Multiple Choice Logic
          const answerStr = row['正确答案'] ? row['正确答案'].toString().toUpperCase().trim() : '';
          if (answerStr.includes(',')) {
            correctAnswers = answerStr.split(',').map((s: string) => s.trim());
          } else {
            // If it's single char 'A', split('') makes ['A']. 
            // If it's 'ABC', split('') makes ['A','B','C'].
            correctAnswers = answerStr.split('');
          }
        } else {
          // Essay / Q&A Logic - Keep answer as is
          if (row['正确答案']) {
            correctAnswers = [row['正确答案'].toString()];
          }
        }

        // Allow questions with no options if they have an answer or at least are questions
        // if (options.length === 0) continue; // REMOVED

        const question = new Question();
        question.text = row['题干'];
        question.options = options;
        question.correctAnswers = correctAnswers;
        question.explanation = row['解析'] || '';
        question.chapter = chapter;

        questions.push(question);
      }

      if (questions.length > 0) {
        await this.questionRepo.save(questions);
        totalQuestions += questions.length;
      }
    }
    return { success: true, count: totalQuestions };
  }

  async addComment(userId: string, dto: CreateCommentDto): Promise<Comment> {
    const question = await this.questionRepo.findOne({ where: { id: dto.questionId } });
    if (!question) throw new NotFoundException('Question not found');

    const comment = this.commentRepo.create({
      content: dto.content,
      user: { id: userId } as User,
      question: question
    });

    const saved = await this.commentRepo.save(comment);
    return (await this.commentRepo.findOne({
      where: { id: saved.id },
      relations: ['user'],
      select: {
        id: true,
        content: true,
        createdAt: true,
        user: {
          id: true,
          name: true,
          avatar: true
        }
      }
    })) as Comment;
  }

  async getComments(questionId: number): Promise<Comment[]> {
    return this.commentRepo.find({
      where: { question: { id: questionId } },
      relations: ['user'],
      order: { createdAt: 'DESC' },
      select: {
        id: true,
        content: true,
        createdAt: true,
        user: {
          id: true,
          name: true,
          avatar: true
        }
      }
    });
  }
}
