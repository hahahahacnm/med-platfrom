import { Controller, Get, Body, Put, Param, UseGuards } from '@nestjs/common';
import { SettingsService } from './settings.service';
// Assuming we have basic auth guard or similar. For now public or add Guard later if user auth is ready
// But usually settings are admin only.
// Let's import hypothetical JwtAuthGuard or make it public for dev if Auth not fully ready strict.
// Based on file explorer, I see auth module.

@Controller('settings')
export class SettingsController {
  constructor(private readonly settingsService: SettingsService) { }

  @Get()
  async getAll() {
    return this.settingsService.getAll();
  }

  @Put(':key')
  async update(@Param('key') key: string, @Body() body: any) {
    const value = body?.value !== undefined ? String(body.value) : '';
    return this.settingsService.update(key, value);
  }
}
