import {
  Entity,
  PrimaryGeneratedColumn,
  Column,
  CreateDateColumn,
  OneToMany,
} from 'typeorm';
import { Comment } from '../quiz/entities/comment.entity';

@Entity()
export class User {
  @PrimaryGeneratedColumn('uuid')
  id: string;

  @Column({ unique: true })
  email: string;

  @Column()
  password: string;

  @Column()
  name: string;

  @Column({ default: 'user' })
  role: string;

  @Column('float', { default: 0 })
  balance: number;

  @Column({ nullable: true })
  avatar: string;

  @Column('simple-json', { default: '[]' })
  subscriptions: any[];

  @Column('simple-json', { default: '[]' })
  quizHistory: any[];

  @Column('simple-json', { default: '{}' })
  chapterProgress: any;

  @Column('simple-json', { default: '[]' })
  bookmarks: any[];

  @Column({ nullable: true })
  loginSessionId: string;

  @CreateDateColumn()
  createdAt: Date;

  @OneToMany(() => Comment, (comment) => comment.user)
  comments: Comment[];
}
