import {
  Injectable,
  UnauthorizedException,
  BadRequestException,
} from '@nestjs/common';
import { UsersService } from '../users/users.service';
import { JwtService } from '@nestjs/jwt';
import * as bcrypt from 'bcrypt';
import * as svgCaptcha from 'svg-captcha';
import { v4 as uuidv4 } from 'uuid';
import { SettingsService } from '../settings/settings.service';

@Injectable()
export class AuthService {
  private captchas = new Map<string, { text: string; expires: number }>();
  private loginAttempts = new Map<string, { count: number; expires: number }>();

  constructor(
    private usersService: UsersService,
    private jwtService: JwtService,
    private settingsService: SettingsService,
  ) { }

  async validateUser(email: string, pass: string): Promise<any> {
    const user = await this.usersService.findOne(email);
    if (user && (await bcrypt.compare(pass, user.password))) {
      const { password, ...result } = user;
      return result;
    }
    return null;
  }

  async validateCaptchaInput(captchaId: string, captchaCode: string) {
    if (!captchaId || !captchaCode) {
      throw new BadRequestException('请输入验证码');
    }

    const captcha = this.captchas.get(captchaId);
    if (!captcha) {
      throw new BadRequestException('验证码已过期，请刷新重试');
    }

    if (captcha.text !== captchaCode.toLowerCase()) {
      throw new BadRequestException('验证码错误');
    }

    // Consume captcha
    this.captchas.delete(captchaId);
  }

  isCaptchaRequired(email: string): boolean {
    const attempt = this.loginAttempts.get(email);
    if (!attempt) return false;
    // 如果过期则重置
    if (Date.now() > attempt.expires) {
      this.loginAttempts.delete(email);
      return false;
    }
    return attempt.count >= 3;
  }

  recordLoginFailure(email: string) {
    const attempt = this.loginAttempts.get(email) || { count: 0, expires: 0 };
    if (Date.now() > attempt.expires) {
      attempt.count = 0;
    }
    attempt.count++;
    attempt.expires = Date.now() + 15 * 60 * 1000; // 锁定计数 15 分钟
    this.loginAttempts.set(email, attempt);
  }

  resetLoginAttempts(email: string) {
    this.loginAttempts.delete(email);
  }

  async login(user: any) {
    const sessionId = uuidv4();
    await this.usersService.update(user.id, { loginSessionId: sessionId });

    const payload = {
      email: user.email,
      sub: user.id,
      name: user.name,
      role: user.role,
      sid: sessionId,
    };
    return {
      access_token: this.jwtService.sign(payload),
      user: {
        ...user,
      },
    };
  }

  async generateCaptcha() {
    const captcha = svgCaptcha.create({
      size: 4,
      noise: 2,
      color: true,
      background: '#f0f0f0',
    });
    const id = uuidv4();
    // Expires in 5 minutes
    this.captchas.set(id, {
      text: captcha.text.toLowerCase(),
      expires: Date.now() + 5 * 60 * 1000,
    });

    // Cleanup old captchas
    if (this.captchas.size > 1000) {
      const now = Date.now();
      for (const [key, val] of this.captchas.entries()) {
        if (val.expires < now) {
          this.captchas.delete(key);
        }
      }
    }

    // Also cleanup login attempts
    if (this.loginAttempts.size > 1000) {
      const now = Date.now();
      for (const [key, val] of this.loginAttempts.entries()) {
        if (val.expires < now) {
          this.loginAttempts.delete(key);
        }
      }
    }

    return {
      id,
      image: captcha.data,
    };
  }

  async register(registerDto: any) {
    const registrationEnabled = await this.settingsService.getBoolean(
      'registration_enabled',
    );
    if (!registrationEnabled) {
      throw new BadRequestException('系统已暂停新用户注册，请联系管理员');
    }

    await this.validateCaptchaInput(
      registerDto.captchaId,
      registerDto.captchaCode,
    );

    const existing = await this.usersService.findOne(registerDto.email);
    if (existing) {
      throw new UnauthorizedException('该邮箱已被注册');
    }
    const hashedPassword = await bcrypt.hash(registerDto.password, 10);
    const user = await this.usersService.create({
      email: registerDto.email,
      password: hashedPassword,
      name: registerDto.name,
    });
    const { password, ...result } = user;
    return result;
  }
}
