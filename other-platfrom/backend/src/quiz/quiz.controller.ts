
import { Controller, Get, Param, Post, Body, Patch, Delete, UseInterceptors, UploadedFiles, UseGuards, Request } from '@nestjs/common';
import { FilesInterceptor } from '@nestjs/platform-express';
import { QuizService } from './quiz.service';
import { CreateSubjectDto, UpdateSubjectDto } from './dto/subject.dto';
import { CreateCommentDto } from './dto/comment.dto';
import { AuthGuard } from '@nestjs/passport';

@Controller('quiz')
export class QuizController {
  constructor(private readonly quizService: QuizService) { }

  @Get('subjects')
  findAll() {
    return this.quizService.findAll();
  }

  @Get(':id/chapters')
  findChapters(@Param('id') id: string) {
    return this.quizService.findChapters(id);
  }

  @Post('subjects')
  createSubject(@Body() dto: CreateSubjectDto) {
    return this.quizService.createSubject(dto);
  }

  @Patch('subjects/:id')
  updateSubject(@Param('id') id: string, @Body() dto: UpdateSubjectDto) {
    return this.quizService.updateSubject(id, dto);
  }

  @Delete('subjects/:id')
  deleteSubject(@Param('id') id: string) {
    return this.quizService.deleteSubject(id);
  }

  @Delete('chapters/:id')
  deleteChapter(@Param('id') id: string) {
    return this.quizService.deleteChapter(id);
  }

  @Patch('questions/:id')
  updateQuestion(@Param('id') id: number, @Body() data: any) {
    return this.quizService.updateQuestion(id, data);
  }

  @Post('subjects/:id/import')
  @UseInterceptors(FilesInterceptor('files'))
  importQuestions(@Param('id') id: string, @UploadedFiles() files: any[]) {
    return this.quizService.importQuestions(id, files);
  }

  @UseGuards(AuthGuard('jwt'))
  @Post('comments')
  addComment(@Request() req, @Body() dto: CreateCommentDto) {
    return this.quizService.addComment(req.user.userId, dto);
  }

  @Get('questions/:id/comments')
  getComments(@Param('id') id: string) {
    return this.quizService.getComments(+id);
  }
}
