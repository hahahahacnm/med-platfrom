import { Injectable, NotFoundException } from '@nestjs/common';
import { InjectRepository } from '@nestjs/typeorm';
import { Repository } from 'typeorm';
import { Feedback } from './entities/feedback.entity';
import { CreateFeedbackDto } from './dto/create-feedback.dto';

@Injectable()
export class FeedbackService {
  constructor(
    @InjectRepository(Feedback)
    private feedbackRepo: Repository<Feedback>,
  ) {}

  async create(
    createFeedbackDto: CreateFeedbackDto,
    userId: string,
    userName: string,
  ) {
    const feedback = this.feedbackRepo.create({
      ...createFeedbackDto,
      userId,
      userName,
    });
    return this.feedbackRepo.save(feedback);
  }

  async findAll() {
    return this.feedbackRepo.find({ order: { createdAt: 'DESC' } });
  }

  async updateStatus(id: string, status: string) {
    const feedback = await this.feedbackRepo.findOne({ where: { id } });
    if (!feedback) throw new NotFoundException('Feedback not found');

    feedback.status = status;
    return this.feedbackRepo.save(feedback);
  }

  async remove(id: string) {
    const feedback = await this.feedbackRepo.findOne({ where: { id } });
    if (!feedback) throw new NotFoundException('Feedback not found');
    return this.feedbackRepo.remove(feedback);
  }
}
