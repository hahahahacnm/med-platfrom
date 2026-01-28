import { Injectable, NotFoundException, Inject, forwardRef } from '@nestjs/common';
import { InjectRepository } from '@nestjs/typeorm';
import { Repository } from 'typeorm';
import { Product } from './entities/product.entity';
import { Transaction } from './entities/transaction.entity';
import { User } from '../users/user.entity';
import { Coupon } from './entities/coupon.entity';
import { PaymentService } from '../payment/payment.service';

@Injectable()
export class StoreService {
  constructor(
    @InjectRepository(Product)
    private productsRepository: Repository<Product>,
    @InjectRepository(Transaction)
    private transactionRepo: Repository<Transaction>,
    @InjectRepository(User)
    private userRepo: Repository<User>,
    @InjectRepository(Coupon)
    private couponRepo: Repository<Coupon>,
    @Inject(forwardRef(() => PaymentService))
    private paymentService: PaymentService,
  ) { }

  async checkout(userId: string, products: Product[], amount: number, couponCode?: string, payType?: string) {
    let finalAmount = amount;
    const couponDetails: { code: string; discount: number; productId: string }[] = [];

    if (couponCode) {
      // Find coupons matching the code
      const coupons = await this.couponRepo.find({ where: { code: couponCode } });

      for (const coupon of coupons) {
        // Check if this coupon applies to any product in the cart
        const product = products.find(p => p.id === coupon.productId);
        if (product) {
          // Check limits
          if (coupon.usageLimit !== null && coupon.usedCount >= coupon.usageLimit) {
            continue; // Skip exhausted coupons
          }

          let discount = 0;
          if (coupon.type === 'amount') {
            discount = coupon.value;
          } else if (coupon.type === 'percent') {
            discount = product.price * (coupon.value / 100);
          }

          // Ensure discount doesn't exceed product price (optional, but good practice)
          if (discount > product.price) discount = product.price;

          finalAmount -= discount;
          couponDetails.push({ code: coupon.code, discount, productId: product.id });

          // Update usage
          coupon.usedCount += 1;
          await this.couponRepo.save(coupon);
        }
      }
    }

    // Ensure final amount is not negative
    if (finalAmount < 0) finalAmount = 0;

    const transaction = this.transactionRepo.create({
      userId,
      amount: finalAmount,
      products: products.map((p) => p.id),
      couponDetails: couponDetails.length > 0 ? couponDetails : undefined,
      status: finalAmount > 0 ? 'pending' : 'completed',
    });
    await this.transactionRepo.save(transaction);

    if (finalAmount > 0 && payType) {
      const orderNo = transaction.id.replace(/-/g, ''); // Use simple ID for payment
      try {
        const paymentResult = await this.paymentService.createOrder({
          orderNo,
          amount: finalAmount,
          subject: products.map(p => p.title).join(',').slice(0, 100), // Limit length
          payType: payType,
        });
        return { success: true, transactionId: transaction.id, payUrl: (paymentResult as any).data.payUrl, orderNo };
      } catch (error) {
        // Mark transaction as failed?
        transaction.status = 'failed';
        await this.transactionRepo.save(transaction);
        throw error;
      }
    } else if (finalAmount <= 0) {
      await this.completeTransaction(transaction.id);
      return { success: true, transactionId: transaction.id };
    }

    return { success: true, transactionId: transaction.id, message: 'Pending payment' };
  }

  async completeTransaction(transactionId: string) {
    const transaction = await this.transactionRepo.findOne({ where: { id: transactionId } });
    if (!transaction) return;

    if (transaction.status !== 'completed') {
      transaction.status = 'completed';
      await this.transactionRepo.save(transaction);
    }

    // Grant access logic (previously in checkout)
    const userId = transaction.userId;
    const products: Product[] = [];
    if (transaction.products) {
      // Fetch products
      // We need simple-json to be parsed. TypeOrm should handle it if defined as simple-json.
      // But transaction.products is string[]
      for (const pid of transaction.products) {
        const p = await this.productsRepository.findOne({ where: { id: pid } });
        if (p) products.push(p);
      }
    }

    const user = await this.userRepo.findOne({ where: { id: userId } });
    if (user) {
      let subs: any[] = user.subscriptions || [];

      // Normalize existing subscriptions
      subs = subs.map(s => {
        if (typeof s === 'string') return { accessId: s, startDate: new Date().toISOString(), expiresAt: null };
        return s;
      });

      for (const p of products) {
        // const dbProduct = await this.productsRepository.findOne({ where: { id: p.id } });
        // Already fetched above
        const dbProduct = p;
        if (!dbProduct) continue;

        const { durationValue, durationUnit, accessId } = dbProduct;

        let subIndex = subs.findIndex((s: any) => s.accessId === accessId);
        let sub = subIndex >= 0 ? subs[subIndex] : null;

        if (!sub) {
          sub = { accessId, startDate: new Date().toISOString(), expiresAt: null };
          subs.push(sub);
        }

        if (sub.expiresAt === null && subIndex >= 0) {
          // Already forever
        } else {
          let baseDate = new Date();
          if (sub.expiresAt && new Date(sub.expiresAt) > baseDate) {
            baseDate = new Date(sub.expiresAt);
          }

          if (durationUnit === 'forever') {
            sub.expiresAt = null;
          } else if (durationValue) {
            const newExpiry = new Date(baseDate);
            if (durationUnit === 'day') newExpiry.setDate(newExpiry.getDate() + durationValue);
            if (durationUnit === 'month') newExpiry.setMonth(newExpiry.getMonth() + durationValue);
            if (durationUnit === 'year') newExpiry.setFullYear(newExpiry.getFullYear() + durationValue);
            sub.expiresAt = newExpiry.toISOString();
          }
        }
      }
      user.subscriptions = subs;
      await this.userRepo.save(user);
    }
  }

  async findAllTransactions() {
    return this.transactionRepo.find({
      relations: ['user'],
      order: { createdAt: 'DESC' },
    });
  }

  async findAll(): Promise<Product[]> {
    return this.productsRepository.find({ relations: ['coupons'] });
  }

  async findOne(id: string): Promise<Product> {
    const product = await this.productsRepository.findOne({ where: { id }, relations: ['coupons'] });
    if (!product) throw new NotFoundException('Product not found');
    return product;
  }

  async create(product: Partial<Product>): Promise<Product> {
    // If id is not provided, generate one or let DB handle if using uuid (but we use PrimaryColumn string so user might provide it)
    // If mocking existing data, IDs are provided.
    // If new product from admin, we should generate ID if not present.
    if (!product.id) {
      product.id = 'prod_' + Date.now();
    }
    const newProduct = this.productsRepository.create(product);
    return this.productsRepository.save(newProduct);
  }

  async update(id: string, updateData: Partial<Product>): Promise<Product> {
    await this.productsRepository.update(id, updateData);
    return this.findOne(id);
  }

  async remove(id: string): Promise<void> {
    await this.productsRepository.delete(id);
  }

  // --- Coupon Logic ---

  async addCoupon(productId: string, data: Partial<Coupon>): Promise<Coupon> {
    const coupon = this.couponRepo.create({ ...data, productId });
    return this.couponRepo.save(coupon);
  }

  async deleteCoupon(couponId: string): Promise<void> {
    await this.couponRepo.delete(couponId);
  }

  async validateCoupon(code: string, productIds: string[]): Promise<{ valid: boolean; discount: number; coupons: any[] }> {
    const coupons = await this.couponRepo.find({ where: { code } });
    if (!coupons.length) return { valid: false, discount: 0, coupons: [] };

    let totalDiscount = 0;
    const appliedCoupons: any[] = [];

    for (const coupon of coupons) {
      if (productIds.includes(coupon.productId)) {
        if (coupon.usageLimit !== null && coupon.usedCount >= coupon.usageLimit) continue;

        // We need product price to calc percent discount.
        const product = await this.productsRepository.findOne({ where: { id: coupon.productId } });
        if (!product) continue;

        let discount = 0;
        if (coupon.type === 'amount') {
          discount = coupon.value;
        } else {
          discount = product.price * (coupon.value / 100);
        }
        if (discount > product.price) discount = product.price;

        totalDiscount += discount;
        appliedCoupons.push({ ...coupon, calculatedDiscount: discount });
      }
    }

    if (totalDiscount === 0 && appliedCoupons.length === 0) return { valid: false, discount: 0, coupons: [] };

    return { valid: true, discount: totalDiscount, coupons: appliedCoupons };
  }
}
