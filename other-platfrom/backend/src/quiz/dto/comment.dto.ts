
import { IsNotEmpty, IsNumber, IsString } from 'class-validator';

export class CreateCommentDto {
    @IsNumber()
    @IsNotEmpty()
    questionId: number;

    @IsString()
    @IsNotEmpty()
    content: string;
}
