import {
  Controller,
  Post,
  Body,
  UseGuards,
  Get,
  Request,
  UnauthorizedException,
} from '@nestjs/common';
import { AuthService } from './auth.service';
import { AuthGuard } from '@nestjs/passport';
import { UsersService } from '../users/users.service';
import { LoginDto, CreateUserDto } from './dto/auth.dto';

@Controller('auth')
export class AuthController {
  constructor(
    private authService: AuthService,
    private usersService: UsersService,
  ) { }

  @Post('login')
  async login(@Body() body: LoginDto) {
    // 1. 根据之前的失败记录检查是否需要验证码
    if (this.authService.isCaptchaRequired(body.email)) {
      await this.authService.validateCaptchaInput(
        body.captchaId || '',
        body.captchaCode || '',
      );
    }

    const user = await this.authService.validateUser(body.email, body.password);
    if (!user) {
      this.authService.recordLoginFailure(body.email);

      // 检查下一次尝试是否需要验证码
      if (this.authService.isCaptchaRequired(body.email)) {
        // 抛出特定错误信息，以便前端检测
        throw new UnauthorizedException('密码错误次数过多，请输入验证码');
      }
      throw new UnauthorizedException('邮箱或密码错误');
    }

    // 登录成功
    this.authService.resetLoginAttempts(body.email);
    return this.authService.login(user);
  }

  @Get('captcha')
  async getCaptcha() {
    return this.authService.generateCaptcha();
  }

  @Post('register')
  async register(@Body() body: CreateUserDto) {
    const user = await this.authService.register(body);
    return this.authService.login(user); // Auto login
  }

  @UseGuards(AuthGuard('jwt'))
  @Get('check')
  async checkSession() {
    return { status: 'ok' };
  }

  @UseGuards(AuthGuard('jwt'))
  @Get('profile')
  async getProfile(@Request() req) {
    const user = await this.usersService.findById(req.user.userId);
    if (!user) throw new UnauthorizedException('User not found');
    const { password, ...result } = user;
    return result;
  }
}
