import {
  Controller,
  Get,
  Post,
  Body,
  Put,
  Param,
  Delete,
  UseGuards,
  Request,
} from '@nestjs/common';
import { StoreService } from './store.service';
import { RolesGuard } from '../auth/roles.guard';
import { AuthGuard } from '@nestjs/passport';

@Controller('store')
export class StoreController {
  constructor(private readonly storeService: StoreService) { }

  @UseGuards(AuthGuard('jwt'))
  @Post('checkout')
  async checkout(
    @Request() req,
    @Body() body: { products: any[]; amount: number; couponCode?: string; payType?: string },
  ) {
    return this.storeService.checkout(
      req.user.userId,
      body.products,
      body.amount,
      body.couponCode,
      body.payType,
    );
  }

  @UseGuards(AuthGuard('jwt'))
  @Post('validate-coupon')
  async validateCoupon(@Body() body: { code: string; productIds: string[] }) {
    return this.storeService.validateCoupon(body.code, body.productIds);
  }

  @UseGuards(AuthGuard('jwt'), RolesGuard)
  @Get('transactions')
  getAllTransactions() {
    return this.storeService.findAllTransactions();
  }

  @Get('products')
  getAllProducts() {
    return this.storeService.findAll();
  }

  @Get('products/:id')
  getProduct(@Param('id') id: string) {
    return this.storeService.findOne(id);
  }

  @UseGuards(AuthGuard('jwt'), RolesGuard)
  @Post('products')
  createProduct(@Body() body: any) {
    return this.storeService.create(body);
  }

  @UseGuards(AuthGuard('jwt'), RolesGuard)
  @Put('products/:id')
  updateProduct(@Param('id') id: string, @Body() body: any) {
    return this.storeService.update(id, body);
  }

  @UseGuards(AuthGuard('jwt'), RolesGuard)
  @Delete('products/:id')
  deleteProduct(@Param('id') id: string) {
    return this.storeService.remove(id);
  }

  @UseGuards(AuthGuard('jwt'), RolesGuard)
  @Post('products/:id/coupons')
  addCoupon(@Param('id') id: string, @Body() body: any) {
    return this.storeService.addCoupon(id, body);
  }

  @UseGuards(AuthGuard('jwt'), RolesGuard)
  @Delete('coupons/:id')
  deleteCoupon(@Param('id') id: string) {
    return this.storeService.deleteCoupon(id);
  }
}
