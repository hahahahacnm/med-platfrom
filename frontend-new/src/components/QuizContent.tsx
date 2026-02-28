import React from 'react';
import { Layers, BookOpen } from 'lucide-react';
import { CatalogPage } from './CatalogPage';

export function QuizContent() {
    return (
        <CatalogPage
            type="quiz"
            title="题库中心"
            description="深度覆盖医学核心知识，开启高质量练习"
            icon={Layers}
            treeEndpoint="/category-tree"
            skeletonEndpoint="/questions/skeleton"
            colorClass="text-primary"
        />
    );
}
