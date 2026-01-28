import { Injectable } from '@nestjs/common';
import { InjectRepository } from '@nestjs/typeorm';
import { Repository } from 'typeorm';
import { Announcement } from './entities/announcement.entity';

@Injectable()
export class AnnouncementsService {
  constructor(
    @InjectRepository(Announcement)
    private announcementRepository: Repository<Announcement>,
  ) {}

  async create(createDto: {
    title: string;
    content: string;
    visible?: boolean;
  }) {
    const announcement = this.announcementRepository.create(createDto);
    return this.announcementRepository.save(announcement);
  }

  async findAll() {
    // Admin sees all, sorted by date DESC
    return this.announcementRepository.find({
      order: { createdAt: 'DESC' },
    });
  }

  async findLatest() {
    // User sees latest active
    return this.announcementRepository.find({
      where: { visible: true },
      order: { createdAt: 'DESC' },
      take: 5,
    });
  }

  async findOne(id: string) {
    return this.announcementRepository.findOneBy({ id });
  }

  async update(
    id: string,
    updateDto: { title?: string; content?: string; visible?: boolean },
  ) {
    await this.announcementRepository.update(id, updateDto);
    return this.findOne(id);
  }

  async remove(id: string) {
    await this.announcementRepository.delete(id);
    return { deleted: true };
  }
}
