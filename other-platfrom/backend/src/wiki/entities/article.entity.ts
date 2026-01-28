import {
  Entity,
  PrimaryColumn,
  Column,
  ManyToOne,
  CreateDateColumn,
  UpdateDateColumn,
} from 'typeorm';
import { WikiCategory } from './wiki-category.entity';

@Entity()
export class Article {
  @PrimaryColumn()
  id: string;

  @Column()
  title: string;

  @Column()
  excerpt: string;

  @Column('text')
  content: string;

  @Column()
  author: string;

  @Column()
  readTime: string;

  @Column()
  date: string;

  @Column('simple-json')
  tags: string[];

  @ManyToOne(() => WikiCategory, (category) => category.articles)
  category: WikiCategory;

  @Column({ default: 'published' })
  status: string;

  @CreateDateColumn()
  createdAt: Date;

  @UpdateDateColumn()
  updatedAt: Date;
}
