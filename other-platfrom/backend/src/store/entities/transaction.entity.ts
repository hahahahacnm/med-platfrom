import {
  Entity,
  PrimaryGeneratedColumn,
  Column,
  CreateDateColumn,
  ManyToOne,
  JoinColumn,
} from 'typeorm';
import { User } from '../../users/user.entity';

@Entity()
export class Transaction {
  @PrimaryGeneratedColumn('uuid')
  id: string;

  @Column()
  userId: string;

  @ManyToOne(() => User)
  @JoinColumn({ name: 'userId' })
  user: User;

  @Column()
  amount: number;

  @Column('simple-json')
  products: string[]; // List of product IDs

  @Column('simple-json', { nullable: true })
  couponDetails: { code: string; discount: number; productId: string }[];

  @Column({ default: 'pending' })
  status: 'pending' | 'completed' | 'failed';

  @Column('simple-json', { nullable: true })
  paymentData: any;


  @CreateDateColumn()
  createdAt: Date;
}
