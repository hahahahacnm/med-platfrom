import {
  Entity,
  Column,
  PrimaryGeneratedColumn,
  ManyToOne,
  OneToMany,
} from 'typeorm';
import { Subject } from './subject.entity';
import { Question } from './question.entity';

@Entity()
export class Chapter {
  @PrimaryGeneratedColumn('uuid')
  id: string;

  @Column()
  title: string;

  @ManyToOne(() => Subject, (subject) => subject.chapters, {
    onDelete: 'CASCADE',
  })
  subject: Subject;

  @OneToMany(() => Question, (question) => question.chapter, { cascade: true })
  questions: Question[];
}
