import { Body, Controller, Post, UseGuards } from '@nestjs/common';
import { AuthGuard } from '@nestjs/passport';
import { ChatService } from './chat.service';
import { CreateChatCompletionDto } from './dto/chat.dto';

@Controller('v1/chat')
export class ChatController {
  constructor(private readonly chatService: ChatService) {}

  @UseGuards(AuthGuard('jwt'))
  @Post('completions')
  async createCompletion(@Body() createChatDto: CreateChatCompletionDto) {
    return this.chatService.create(createChatDto);
  }
}
