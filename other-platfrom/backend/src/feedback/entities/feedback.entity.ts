import {
  Entity,
  PrimaryGeneratedColumn,
  Column,
  CreateDateColumn,
} from 'typeorm';

@Entity()
export class Feedback {
  @PrimaryGeneratedColumn('uuid')
  id: string;

  @Column({ nullable: true })
  userId: string; // Optional: user might be anonymous or we just store ID string

  @Column()
  userName: string; // Snapshot name for convenience

  @Column()
  type: string; // bug, suggestion, content, other

  @Column('text')
  content: string;

  @Column({ nullable: true })
  contact: string;

  @Column({ default: 'pending' })
  status: string; // pending, resolved

  @CreateDateColumn()
  createdAt: Date;
}
