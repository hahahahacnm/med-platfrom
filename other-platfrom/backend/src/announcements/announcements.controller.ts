import {
  Controller,
  Get,
  Post,
  Body,
  Patch,
  Param,
  Delete,
  UseGuards,
} from '@nestjs/common';
import { AnnouncementsService } from './announcements.service';
import { AuthGuard } from '@nestjs/passport';
import { RolesGuard } from '../auth/roles.guard';

@Controller('announcements')
export class AnnouncementsController {
  constructor(private readonly announcementsService: AnnouncementsService) {}

  @Post()
  @UseGuards(AuthGuard('jwt'), RolesGuard)
  create(
    @Body()
    createAnnouncementDto: {
      title: string;
      content: string;
      visible?: boolean;
    },
  ) {
    return this.announcementsService.create(createAnnouncementDto);
  }

  @Get()
  findAll() {
    return this.announcementsService.findAll();
  }

  @Get('latest')
  findLatest() {
    return this.announcementsService.findLatest();
  }

  @Get(':id')
  findOne(@Param('id') id: string) {
    return this.announcementsService.findOne(id);
  }

  @Patch(':id')
  @UseGuards(AuthGuard('jwt'), RolesGuard)
  update(
    @Param('id') id: string,
    @Body()
    updateAnnouncementDto: {
      title?: string;
      content?: string;
      visible?: boolean;
    },
  ) {
    return this.announcementsService.update(id, updateAnnouncementDto);
  }

  @Delete(':id')
  @UseGuards(AuthGuard('jwt'), RolesGuard)
  remove(@Param('id') id: string) {
    return this.announcementsService.remove(id);
  }
}
