import { Entity, Column, PrimaryColumn, OneToMany } from 'typeorm';
import { Chapter } from './chapter.entity';

@Entity()
export class Subject {
  @PrimaryColumn()
  id: string;

  @Column()
  title: string;

  @Column()
  description: string;

  @Column()
  icon: string;

  @Column()
  color: string;

  @OneToMany(() => Chapter, (chapter) => chapter.subject, { cascade: true })
  chapters: Chapter[];
}
