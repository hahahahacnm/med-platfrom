import { IsEmail, IsString, MinLength, IsOptional } from 'class-validator';

export class CreateUserDto {
  @IsEmail({}, { message: '请输入有效的邮箱地址' })
  email: string;

  @IsString()
  @MinLength(6, { message: '密码长度至少为 6 位' })
  password: string;

  @IsString()
  @MinLength(2, { message: '姓名至少为 2 个字符' })
  name: string;

  @IsString()
  captchaCode: string;

  @IsString()
  captchaId: string;
}

export class LoginDto {
  @IsEmail({}, { message: '请输入有效的邮箱地址' })
  email: string;

  @IsString()
  @MinLength(6, { message: '密码长度至少为 6 位' })
  password: string;

  @IsString()
  @IsOptional()
  captchaCode?: string;

  @IsString()
  @IsOptional()
  captchaId?: string;
}
