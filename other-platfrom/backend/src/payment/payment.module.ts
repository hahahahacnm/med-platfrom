import { Module, forwardRef } from '@nestjs/common';
import { TypeOrmModule } from '@nestjs/typeorm';
import { PaymentService } from './payment.service';
import { PaymentController } from './payment.controller';
import { Transaction } from '../store/entities/transaction.entity';
import { SettingsModule } from '../settings/settings.module';
import { StoreModule } from '../store/store.module';

@Module({
    imports: [
        TypeOrmModule.forFeature([Transaction]),
        SettingsModule,
        forwardRef(() => StoreModule),
    ],
    providers: [PaymentService],
    controllers: [PaymentController],
    exports: [PaymentService],
})
export class PaymentModule { }
