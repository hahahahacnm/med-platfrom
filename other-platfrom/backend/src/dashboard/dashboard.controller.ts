import { Controller, Get, UseGuards, Query } from '@nestjs/common';
import { DashboardService } from './dashboard.service';
import { RolesGuard } from '../auth/roles.guard';
import { AuthGuard } from '@nestjs/passport';

@Controller('dashboard')
@UseGuards(AuthGuard('jwt'), RolesGuard)
export class DashboardController {
  constructor(private readonly dashboardService: DashboardService) {}

  @Get('stats')
  async getStats() {
    return this.dashboardService.getStats();
  }
  @Get('revenue-trend')
  async getRevenueTrend(@Query('days') days?: number) {
    return this.dashboardService.getRevenueTrend(days ? Number(days) : 30);
  }

  @Get('subject-distribution')
  async getSubjectDistribution() {
    return this.dashboardService.getSubjectDistribution();
  }
}
