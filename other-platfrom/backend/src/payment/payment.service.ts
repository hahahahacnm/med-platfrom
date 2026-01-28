import { Injectable, Logger, BadRequestException } from '@nestjs/common';
import { SettingsService } from '../settings/settings.service';
import { ZhuFuFm, ZhuFuFmPayType, ZhuFuFmConfig } from '../zhifufm';

@Injectable()
export class PaymentService {
    private readonly logger = new Logger(PaymentService.name);

    constructor(private settingsService: SettingsService) { }

    private async getSdk(): Promise<ZhuFuFm> {
        const merchantNum = await this.settingsService.get('payment_merchant_num');
        const merchantKey = await this.settingsService.get('payment_merchant_key');
        const baseUrl = await this.settingsService.get('payment_base_url');
        const notifyUrl = await this.settingsService.get('payment_notify_url');
        const returnUrl = await this.settingsService.get('payment_return_url');

        if (!merchantNum || !merchantKey || !baseUrl) {
            this.logger.warn('Payment settings missing');
            // For development, we might want to throw or return a dummy?
            // But let's throw to ensure user configures it.
            throw new BadRequestException('Payment settings are not configured. Please contact admin.');
        }

        const config: ZhuFuFmConfig = {
            baseUrl,
            merchantNum,
            merchantKey,
            notifyUrl: notifyUrl || 'http://localhost:3000/payment/notify',
            returnUrl: returnUrl || 'http://localhost:5173/profile',
        };

        return new ZhuFuFm(config);
    }

    async createOrder(params: {
        orderNo: string;
        amount: number;
        subject: string;
        payType: string;
    }) {
        try {
            const sdk = await this.getSdk();
            this.logger.log(`Creating order: ${params.orderNo}, Amount: ${params.amount}, Type: ${params.payType}`);
            const result = await sdk.startOrder({
                orderNo: params.orderNo,
                amount: params.amount,
                payType: params.payType as ZhuFuFmPayType,
                subject: params.subject,
                returnType: 'json' as any
            });
            this.logger.log(`Order created: ${JSON.stringify(result)}`);
            return result;
        } catch (error) {
            this.logger.error(`Failed to create order: ${error.message}`, error.stack);
            throw new BadRequestException(`Payment creation failed: ${error.message}`);
        }
    }

    async validateNotify(body: any) {
        const sdk = await this.getSdk();
        return sdk.validateNotifySign(body);
    }

    getPayTypes() {
        return ZhuFuFmPayType;
    }
}
