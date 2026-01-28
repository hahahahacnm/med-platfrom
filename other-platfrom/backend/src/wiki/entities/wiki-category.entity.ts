import { Entity, PrimaryColumn, Column, OneToMany } from 'typeorm';
import { Article } from './article.entity';

@Entity()
export class WikiCategory {
  @PrimaryColumn()
  id: string;

  @Column()
  title: string;

  @Column()
  description: string;

  @Column()
  iconName: string;

  @Column()
  color: string;

  @OneToMany(() => Article, (article) => article.category, { cascade: true })
  articles: Article[];
}
