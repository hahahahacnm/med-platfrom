import { Injectable, OnModuleInit } from '@nestjs/common';
import { InjectRepository } from '@nestjs/typeorm';
import { Repository } from 'typeorm';
import { Setting } from './settings.entity';

@Injectable()
export class SettingsService implements OnModuleInit {
  constructor(
    @InjectRepository(Setting)
    private settingsRepository: Repository<Setting>,
  ) { }

  async onModuleInit() {
    // Initialize default settings if they don't exist
    await this.initDefault(
      'registration_enabled',
      'true',
      'Enable new user registration',
    );
    await this.initDefault('ai_enabled', 'true', 'Enable AI parameters');
    await this.initDefault(
      'ai_base_url',
      'https://api.openai.com/v1',
      'AI Provider Base URL',
    );
    await this.initDefault('ai_api_key', '', 'AI Provider API Key');

    // Payment Settings (ZhifuFM)
    await this.initDefault('payment_merchant_num', '607807646955094016', 'ZhifuFM Merchant Number');
    await this.initDefault('payment_merchant_key', 'efb26f97c51755ab036d24d41f6d3a12', 'ZhifuFM Merchant Key');
    await this.initDefault('payment_base_url', 'https://api-4m8ptufugd1c.zhifu.fm.it88168.com/api', 'ZhifuFM API Base URL');
    await this.initDefault('payment_notify_url', 'http://localhost:3000/payment/notify', 'Payment Notify URL');
    await this.initDefault('payment_return_url', 'http://localhost:5173/profile', 'Payment Return URL');
    await this.initDefault('payment_enable_alipay', 'true', 'Enable Alipay');
    await this.initDefault('payment_enable_wechat', 'true', 'Enable WeChat Pay');
    // await this.initDefault('maintenance_mode', 'false', 'Enable maintenance mode');
  }

  private async initDefault(key: string, value: string, description: string) {
    const exists = await this.settingsRepository.findOne({ where: { key } });
    if (!exists) {
      await this.settingsRepository.save({ key, value, description });
    }
  }

  async get(key: string): Promise<string | null> {
    const setting = await this.settingsRepository.findOne({ where: { key } });
    return setting ? setting.value : null;
  }

  async getBoolean(key: string): Promise<boolean> {
    const val = await this.get(key);
    return val === 'true';
  }

  async getAll() {
    return this.settingsRepository.find();
  }

  async update(key: string, value: string) {
    const setting = await this.settingsRepository.findOne({ where: { key } });
    if (setting) {
      setting.value = value;
      await this.settingsRepository.save(setting);
    } else {
      // Create if not exists (though validation usually prevents this slightly)
      await this.settingsRepository.save({ key, value });
    }
    return this.get(key);
  }
}
