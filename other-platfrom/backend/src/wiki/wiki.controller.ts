import {
  Controller,
  Get,
  Post,
  Put,
  Delete,
  Body,
  Param,
  UseInterceptors,
  UploadedFile,
} from '@nestjs/common';
import { FileInterceptor } from '@nestjs/platform-express';
import { WikiService } from './wiki.service';
import { Express } from 'express';

@Controller('wiki')
export class WikiController {
  constructor(private readonly wikiService: WikiService) {}

  @Get('categories')
  findAll() {
    return this.wikiService.findAll();
  }

  @Post('categories')
  createCategory(@Body() body: any) {
    return this.wikiService.createCategory(body);
  }

  @Put('categories/:id')
  updateCategory(@Param('id') id: string, @Body() body: any) {
    return this.wikiService.updateCategory(id, body);
  }

  @Delete('categories/:id')
  deleteCategory(@Param('id') id: string) {
    return this.wikiService.deleteCategory(id);
  }

  @Get('articles/:id')
  getArticle(@Param('id') id: string) {
    return this.wikiService.getArticle(id);
  }

  @Post('articles')
  createArticle(@Body() body: any) {
    return this.wikiService.createArticle(body);
  }

  @Put('articles/:id')
  updateArticle(@Param('id') id: string, @Body() body: any) {
    return this.wikiService.updateArticle(id, body);
  }

  @Delete('articles/:id')
  deleteArticle(@Param('id') id: string) {
    return this.wikiService.deleteArticle(id);
  }

  @Post('upload')
  @UseInterceptors(FileInterceptor('file'))
  uploadFile(@UploadedFile() file: any) {
    const content = file.buffer.toString('utf8');
    return { content, filename: file.originalname };
  }
}
