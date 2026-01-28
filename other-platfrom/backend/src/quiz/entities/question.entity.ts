import { Entity, Column, PrimaryGeneratedColumn, ManyToOne, OneToMany } from 'typeorm';
import { Chapter } from './chapter.entity';
import { Comment } from './comment.entity';

@Entity()
export class Question {
  @PrimaryGeneratedColumn('increment')
  id: number;

  @Column()
  text: string;

  @Column('simple-json')
  options: { id: string; text: string }[];

  @Column('simple-json')
  correctAnswers: string[];

  @Column({ type: 'text', nullable: true })
  explanation: string;

  @ManyToOne(() => Chapter, (chapter) => chapter.questions, {
    onDelete: 'CASCADE',
  })
  chapter: Chapter;

  @OneToMany(() => Comment, (comment) => comment.question)
  comments: Comment[];
}
