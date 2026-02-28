import React from 'react';
import { BookMarked } from 'lucide-react';
import { CatalogPage } from './CatalogPage';

export function NotesContent() {
    return (
        <CatalogPage
            type="notes"
            title="我的笔记"
            description="笔尖记录思考，沉淀每一份感悟与成长"
            icon={BookMarked}
            treeEndpoint="/notes/note-tree"
            skeletonEndpoint="/notes/skeleton"
            colorClass="text-blue-500"
        />
    );
}
