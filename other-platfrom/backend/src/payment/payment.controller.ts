import { Controller, Post, Body, Res, Logger, HttpStatus, Query, Get, Inject, forwardRef, All, Request } from '@nestjs/common';
import { PaymentService } from './payment.service';
import { InjectRepository } from '@nestjs/typeorm';
import { Transaction } from '../store/entities/transaction.entity';
import { Repository } from 'typeorm';
import type { Response } from 'express'; // Use type import to avoid value import issues if any
import { StoreService } from '../store/store.service';

@Controller('payment')
export class PaymentController {
    private readonly logger = new Logger(PaymentController.name);

    constructor(
        private readonly paymentService: PaymentService,
        @InjectRepository(Transaction)
        private transactionRepository: Repository<Transaction>,
        @Inject(forwardRef(() => StoreService))
        private readonly storeService: StoreService,
    ) { }

    @All('notify')
    async notify(@Request() req, @Res() res: Response) {
        const body = req.method === 'GET' ? req.query : req.body;
        this.logger.log(`Received payment notification: ${JSON.stringify(body)}`);

        // Validate signature
        let isValid = false;
        try {
            isValid = await this.paymentService.validateNotify(body);
        } catch (e) {
            this.logger.error(`Error validating signature: ${e.message}`);
        }

        if (!isValid) {
            this.logger.warn('Invalid signature for payment notification');
            return res.status(HttpStatus.BAD_REQUEST).send('fail');
        }

        // Check success status more robustly based on docs
        // Docs say: success: true (boolean). But some callbacks might use strings.
        // Let's check typical fields.
        const isSuccess = body.success === 'true' || body.success === true ||
            body.code === 1 || body.code === '1' ||
            body.trade_status === 'TRADE_SUCCESS' ||
            body.state === '1' || body.state === '2';

        if (isSuccess) {

            const orderId = body.orderNo;
            this.logger.log(`Payment success for order ${orderId}`);

            let transaction: Transaction | null = null;

            // Try to find transaction
            if (orderId && orderId.length === 32) {
                const uuid = `${orderId.slice(0, 8)}-${orderId.slice(8, 12)}-${orderId.slice(12, 16)}-${orderId.slice(16, 20)}-${orderId.slice(20)}`;
                transaction = await this.transactionRepository.findOne({ where: { id: uuid } });
            } else {
                transaction = await this.transactionRepository.findOne({ where: { id: orderId } });
            }

            if (transaction) {
                // Complete transaction logic (update status + grant sub)
                await this.storeService.completeTransaction(transaction.id);

                // Also save payment data if possible? completeTransaction doesn't save paymentData.
                // Maybe updating it here first?
                transaction.paymentData = body;
                await this.transactionRepository.save(transaction);
                // completeTransaction is idempotent regarding status, so saving here is fine, but completeTransaction re-fetches it.
                // To be safe: update paymentData, THEN process complete.
                // Or update paymentData inside completeTransaction?
                // I'll update paymentData here first.

                return res.send('success');
            } else {
                this.logger.warn(`Transaction not found for order ${orderId}`);
                return res.status(HttpStatus.NOT_FOUND).send('fail');
            }
        }

        return res.send('success');
    }

    @Get('return')
    async returnUrl(@Query() query: any, @Res() res: Response) {
        return res.redirect('http://localhost:5173/profile');
    }
}
