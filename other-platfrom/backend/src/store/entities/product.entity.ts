import { Entity, Column, PrimaryColumn, OneToMany } from 'typeorm';
import { Coupon } from './coupon.entity';

@Entity()
export class Product {
  @PrimaryColumn()
  id: string;

  @Column()
  title: string;

  @Column()
  description: string;

  @Column('float')
  price: number;

  @Column()
  duration: string;

  @Column()
  imageUrl: string;

  @Column('simple-json')
  tags: string[];

  @Column()
  accessId: string;

  @Column('int', { nullable: true })
  durationValue: number;

  @Column({ nullable: true })
  durationUnit: string;
  @Column({ default: true })
  isPublished: boolean;

  @OneToMany(() => Coupon, (coupon) => coupon.product)
  coupons: Coupon[];
}
