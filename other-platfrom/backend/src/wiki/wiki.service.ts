import { Injectable, NotFoundException } from '@nestjs/common';
import { InjectRepository } from '@nestjs/typeorm';
import { Repository } from 'typeorm';
import { WikiCategory } from './entities/wiki-category.entity';
import { Article } from './entities/article.entity';
import { v4 as uuidv4 } from 'uuid';

@Injectable()
export class WikiService {
  constructor(
    @InjectRepository(WikiCategory) private wikiRepo: Repository<WikiCategory>,
    @InjectRepository(Article) private articleRepo: Repository<Article>,
  ) {}

  async findAll() {
    return this.wikiRepo.find({
      relations: ['articles'],
      order: {
        title: 'ASC',
      },
    });
  }

  // --- Category Management ---
  async createCategory(data: Partial<WikiCategory>) {
    const category = this.wikiRepo.create({
      ...data,
      id: uuidv4(),
    });
    return this.wikiRepo.save(category);
  }

  async updateCategory(id: string, data: Partial<WikiCategory>) {
    await this.wikiRepo.update(id, data);
    return this.wikiRepo.findOne({ where: { id }, relations: ['articles'] });
  }

  async deleteCategory(id: string) {
    return this.wikiRepo.delete(id);
  }

  // --- Article Management ---
  async createArticle(data: Partial<Article> & { categoryId: string }) {
    const category = await this.wikiRepo.findOne({
      where: { id: data.categoryId },
    });
    if (!category) throw new NotFoundException('Category not found');

    const article = this.articleRepo.create({
      id: uuidv4(),
      title: data.title,
      excerpt: data.excerpt || '',
      content: data.content || '',
      author: data.author || 'Admin',
      readTime: data.readTime || '5 min',
      date: new Date().toISOString(),
      tags: data.tags || [],
      status: data.status || 'draft',
      category: category,
    });

    return this.articleRepo.save(article);
  }

  async updateArticle(
    id: string,
    data: Partial<Article> & { categoryId?: string },
  ) {
    const article = await this.articleRepo.findOne({ where: { id } });
    if (!article) throw new NotFoundException('Article not found');

    if (data.categoryId) {
      const category = await this.wikiRepo.findOne({
        where: { id: data.categoryId },
      });
      if (category) article.category = category;
    }

    Object.assign(article, {
      ...data,
      categoryId: undefined, // remove from merge
    });

    return this.articleRepo.save(article);
  }

  async getArticle(id: string) {
    return this.articleRepo.findOne({ where: { id }, relations: ['category'] });
  }

  async deleteArticle(id: string) {
    return this.articleRepo.delete(id);
  }
}
