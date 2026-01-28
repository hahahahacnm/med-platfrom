import { Injectable, Logger } from '@nestjs/common';
import { CreateChatCompletionDto } from './dto/chat.dto';
import { SettingsService } from '../settings/settings.service';

@Injectable()
export class ChatService {
  private readonly logger = new Logger(ChatService.name);

  constructor(private settingsService: SettingsService) {}

  async create(createChatDto: CreateChatCompletionDto) {
    const aiEnabled = await this.settingsService.getBoolean('ai_enabled');

    if (!aiEnabled) {
      return this.createMockResponse(
        createChatDto,
        'AI 助手功能当前已关闭，如有疑问请联系管理员。',
      );
    }

    const baseUrl = await this.settingsService.get('ai_base_url');
    const apiKey = await this.settingsService.get('ai_api_key');

    if (!baseUrl || !apiKey) {
      // Fallback to mock if not configured
      return this.mockResponseLogic(createChatDto);
    }

    try {
      // Unify base URL format, remove trailing slash
      const cleanBaseUrl = baseUrl.replace(/\/+$/, '');
      const url = `${cleanBaseUrl}/chat/completions`;

      const response = await fetch(url, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${apiKey}`,
        },
        body: JSON.stringify(createChatDto),
      });

      if (!response.ok) {
        const errorText = await response.text();
        this.logger.error(
          `AI API Error: ${response.status} ${response.statusText} - ${errorText}`,
        );
        return this.createMockResponse(
          createChatDto,
          'AI 服务暂时不可用，请稍后再试。',
        );
      }

      const data = await response.json();
      return data;
    } catch (error) {
      this.logger.error('Failed to call AI API', error);
      return this.createMockResponse(
        createChatDto,
        '连接 AI 服务失败，请检查网络或配置。',
      );
    }
  }

  private mockResponseLogic(createChatDto: CreateChatCompletionDto) {
    const lastUserMessage =
      createChatDto.messages.filter((m) => m.role === 'user').pop()?.content ||
      '';

    let responseContent =
      '你好！我是你的智能医学助教。有什么我可以帮你的吗？（当前运行在模拟模式，请在后台配置 AI 参数）';

    if (lastUserMessage.includes('高血压')) {
      responseContent =
        '高血压（Hypertension）是一种常见的慢性病，通常指收缩压≥140mmHg和/或舒张压≥90mmHg。长期高血压可能导致心脏病、脑卒中等并发症。建议定期监测血压，保持健康饮食和适量运动。';
    } else if (lastUserMessage.includes('糖尿病')) {
      responseContent =
        '糖尿病是一组以高血糖为特征的代谢性疾病。主要分为1型糖尿病和2型糖尿病。典型症状包括多饮、多尿、多食和体重减轻。';
    }

    return this.createMockResponse(createChatDto, responseContent);
  }

  private createMockResponse(
    createChatDto: CreateChatCompletionDto,
    content: string,
  ) {
    return {
      id: `chatcmpl-${Date.now()}`,
      object: 'chat.completion',
      created: Math.floor(Date.now() / 1000),
      model: createChatDto.model,
      choices: [
        {
          index: 0,
          message: {
            role: 'assistant',
            content: content,
          },
          finish_reason: 'stop',
        },
      ],
      usage: {
        prompt_tokens: 0,
        completion_tokens: 0,
        total_tokens: 0,
      },
    };
  }
}
