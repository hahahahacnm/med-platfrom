import { Entity, PrimaryGeneratedColumn, Column, ManyToOne, JoinColumn } from 'typeorm';
import { Product } from './product.entity';

@Entity()
export class Coupon {
    @PrimaryGeneratedColumn('uuid')
    id: string;

    @Column()
    code: string;

    @Column({ type: 'simple-enum', enum: ['amount', 'percent'] })
    type: 'amount' | 'percent';

    @Column('float')
    value: number;

    @Column('int', { nullable: true })
    usageLimit: number; // null means infinite

    @Column('int', { default: 0 })
    usedCount: number;

    @ManyToOne(() => Product, (product) => product.coupons, { onDelete: 'CASCADE' })
    @JoinColumn({ name: 'productId' })
    product: Product;

    @Column()
    productId: string;
}
