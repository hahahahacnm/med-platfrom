import { Module, OnModuleInit, forwardRef } from '@nestjs/common';
import { TypeOrmModule } from '@nestjs/typeorm';
import { StoreController } from './store.controller';
import { StoreService } from './store.service';
import { Product } from './entities/product.entity';
import { Transaction } from './entities/transaction.entity';
import { Coupon } from './entities/coupon.entity';
import { User } from '../users/user.entity';
import { PaymentModule } from '../payment/payment.module';

@Module({
  imports: [
    TypeOrmModule.forFeature([Product, Transaction, User, Coupon]),
    forwardRef(() => PaymentModule)
  ],
  controllers: [StoreController],
  providers: [StoreService],
  exports: [StoreService],
})
export class StoreModule implements OnModuleInit {
  constructor(private readonly storeService: StoreService) { }

  async onModuleInit() {
    const count = await this.storeService.findAll();
    if (count.length === 0) {
      console.log('Seeding products...');
      const products = [
        {
          id: 'prod_internal_quiz',
          title: '内科学海量题库',
          description: '包含5000+道内科真题与模拟题，智能错题本功能。',
          price: 129,
          duration: '1年',
          durationValue: 1,
          durationUnit: 'year',
          imageUrl:
            'https://images.unsplash.com/photo-1576091160399-112ba8d25d1d?auto=format&fit=crop&q=80&w=400',
          tags: ['题库', '内科', '刷题'],
          accessId: 'quiz_internal',
        },
        {
          id: 'prod_internal_wiki',
          title: '内科临床知识库',
          description: '权威内科诊疗指南与专家视频解析，实时更新。',
          price: 99,
          duration: '1年',
          durationValue: 1,
          durationUnit: 'year',
          imageUrl:
            'https://images.unsplash.com/photo-1532938911079-1b06ac7ceec7?auto=format&fit=crop&q=80&w=400',
          tags: ['知识库', '内科', '指南'],
          accessId: 'wiki_internal',
        },
        {
          id: 'prod_pathology_quiz',
          title: '病理学刷题通关包',
          description: '针对期末与考研的病理学专项训练，含名师解析。',
          price: 59,
          duration: '6个月',
          durationValue: 6,
          durationUnit: 'month',
          imageUrl:
            'https://images.unsplash.com/photo-1579154204601-01588f351e67?auto=format&fit=crop&q=80&w=400',
          tags: ['题库', '基础', '病理'],
          accessId: 'quiz_pathology',
        },
        {
          id: 'prod_pathology_wiki',
          title: '病理学高清图谱库',
          description: '超过1000张高清病理切片图解，显微镜下的微观世界。',
          price: 69,
          duration: '1年',
          durationValue: 1,
          durationUnit: 'year',
          imageUrl:
            'https://images.unsplash.com/photo-1530210124550-912dc1381cb8?auto=format&fit=crop&q=80&w=400',
          tags: ['知识库', '图谱', '病理'],
          accessId: 'wiki_pathology',
        },
        {
          id: 'prod_surgery_quiz',
          title: '外科学专项练习',
          description: '涵盖普外、骨科、神经外科等分科试题。',
          price: 89,
          duration: '1年',
          durationValue: 1,
          durationUnit: 'year',
          imageUrl:
            'https://images.unsplash.com/photo-1551076805-e1869033e561?auto=format&fit=crop&q=80&w=400',
          tags: ['题库', '外科', '真题'],
          accessId: 'quiz_surgery',
        },
        {
          id: 'prod_surgery_wiki',
          title: '外科学手术视频库',
          description: '经典手术术式演示与解剖要点全解析。',
          price: 109,
          duration: '1年',
          durationValue: 1,
          durationUnit: 'year',
          imageUrl:
            'https://images.unsplash.com/photo-1516549655169-df83a092dd14?auto=format&fit=crop&q=80&w=400',
          tags: ['知识库', '视频', '外科'],
          accessId: 'wiki_surgery',
        },
      ];
      for (const p of products) {
        await this.storeService.create(p);
      }
    }
  }
}
