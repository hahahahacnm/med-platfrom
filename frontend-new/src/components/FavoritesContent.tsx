import React from 'react';
import { Star } from 'lucide-react';
import { CatalogPage } from './CatalogPage';

export function FavoritesContent() {
    return (
        <CatalogPage
            type="favorites"
            title="收藏夹"
            description="收录精华题型，打造您的个人专属知识库"
            icon={Star}
            treeEndpoint="/favorite-tree"
            skeletonEndpoint="/favorites/skeleton"
            colorClass="text-amber-500"
        />
    );
}
