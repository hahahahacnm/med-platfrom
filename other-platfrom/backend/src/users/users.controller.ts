import {
  Controller,
  Post,
  Body,
  UseGuards,
  Request,
  NotFoundException,
  Get,
  Delete,
  Param,
  UseInterceptors,
  UploadedFile,
  BadRequestException,
} from '@nestjs/common';
import { UsersService } from './users.service';
import { AuthGuard } from '@nestjs/passport';
import { RolesGuard } from '../auth/roles.guard';
import * as bcrypt from 'bcrypt';
import { FileInterceptor } from '@nestjs/platform-express';
import { diskStorage } from 'multer';
import { extname } from 'path';

@Controller('users')
@UseGuards(AuthGuard('jwt'))
export class UsersController {
  constructor(private readonly usersService: UsersService) { }

  @Post('avatar')
  @UseInterceptors(
    FileInterceptor('file', {
      storage: diskStorage({
        destination: './uploads/avatars',
        filename: (req, file, callback) => {
          const uniqueSuffix =
            Date.now() + '-' + Math.round(Math.random() * 1e9);
          const ext = extname(file.originalname);
          callback(null, `avatar-${uniqueSuffix}${ext}`);
        },
      }),
      limits: { fileSize: 5 * 1024 * 1024 }, // 5MB
    }),
  )
  async uploadAvatar(
    @Request() req,
    @UploadedFile() file: Express.Multer.File,
  ) {
    if (!file) {
      throw new BadRequestException('File is not an image');
    }
    const user = await this.usersService.findById(req.user.userId);
    if (!user) throw new NotFoundException('User not found');

    const avatarPath = `/uploads/avatars/${file.filename}`;
    await this.usersService.update(user.id, { avatar: avatarPath });

    return { success: true, avatar: avatarPath };
  }

  @Post('progress')
  async updateProgress(
    @Request() req,
    @Body() body: { key: string; data: any },
  ) {
    const user = await this.usersService.findById(req.user.userId);
    if (!user) throw new NotFoundException('User not found');

    user.chapterProgress = {
      ...user.chapterProgress,
      [body.key]: body.data,
    };
    await this.usersService.update(user.id, {
      chapterProgress: user.chapterProgress,
    });
    return { success: true };
  }

  @Post('bookmark')
  async toggleBookmark(@Request() req, @Body() body: any) {
    const user = await this.usersService.findById(req.user.userId);
    if (!user) throw new NotFoundException('User not found');

    const exists = user.bookmarks.some((b) => b.id === body.id);
    let newBookmarks;
    if (exists) {
      newBookmarks = user.bookmarks.filter((b) => b.id !== body.id);
    } else {
      newBookmarks = [body, ...user.bookmarks];
    }
    await this.usersService.update(user.id, { bookmarks: newBookmarks });
    return { success: true, bookmarks: newBookmarks };
  }

  @Post('quiz-history')
  async addQuizResult(@Request() req, @Body() result: any) {
    const user = await this.usersService.findById(req.user.userId);
    if (!user) throw new NotFoundException('User not found');

    // Ensure result has a timestamp
    const processedResult = {
      ...result,
      completedAt:
        result.completedAt || result.date || new Date().toISOString(),
    };
    const newHistory = [processedResult, ...(user.quizHistory || [])];
    await this.usersService.update(user.id, { quizHistory: newHistory });
    return { success: true, quizHistory: newHistory };
  }

  @UseGuards(RolesGuard)
  @Post(':id/subscriptions')
  async adminUpdateSubscriptions(
    @Param('id') id: string,
    @Body() body: { subscriptions: any[] },
  ) {
    await this.usersService.updateUserSubscriptions(id, body.subscriptions);
    return { success: true, subscriptions: body.subscriptions };
  }

  @Post('subscriptions')
  async updateSubscriptions(
    @Request() req,
    @Body() body: { subscriptions: any[] },
  ) {
    const user = await this.usersService.findById(req.user.userId);
    if (!user) throw new NotFoundException('User not found');

    await this.usersService.update(user.id, {
      subscriptions: body.subscriptions,
    });
    return { success: true, subscriptions: body.subscriptions };
  }

  @UseGuards(RolesGuard)
  @Get()
  async getAllUsers() {
    return this.usersService.findAllWithStats();
  }

  @UseGuards(RolesGuard)
  @Delete(':id')
  async deleteUser(@Param('id') id: string) {
    return this.usersService.remove(id);
  }

  @UseGuards(RolesGuard)
  @Post()
  async createUser(@Body() body: any) {
    const existing = await this.usersService.findOne(body.email);
    if (existing) {
      throw new Error('User already exists');
    }
    if (!body.password) {
      body.password = '123456';
    }
    const hashedPassword = await bcrypt.hash(body.password, 10);
    const newUser = await this.usersService.create({
      ...body,
      password: hashedPassword,
    });
    const { password, ...result } = newUser;
    return result;
  }
}
