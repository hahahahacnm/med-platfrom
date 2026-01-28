
import { Product, Subject, WikiCategory } from '../types';

const API_URL = '/api';

const getHeaders = () => {
    const token = localStorage.getItem('token');
    return token ? { 'Authorization': `Bearer ${token}`, 'Content-Type': 'application/json' } : { 'Content-Type': 'application/json' };
};

const handleResponse = async (res: Response) => {
    if (res.status === 401) {
        localStorage.removeItem('token');
    }
    if (!res.ok) {
        let errorMessage = 'Request failed';
        try {
            const errorData = await res.json();
            // NestJS validation errors return an array of messages
            if (Array.isArray(errorData.message)) {
                errorMessage = errorData.message.join(', ');
            } else {
                errorMessage = errorData.message || res.statusText;
            }
        } catch (e) {
            errorMessage = res.statusText;
        }
        throw new Error(errorMessage);
    }
    return res.json();
};

export const api = {
    auth: {
        login: async (email: string, password: string, captchaCode?: string, captchaId?: string) => {
            const res = await fetch(`${API_URL}/auth/login`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ email, password, captchaCode, captchaId })
            });
            const data = await handleResponse(res);
            localStorage.setItem('token', data.access_token);
            return data.user;
        },
        register: async (data: any) => {
            const res = await fetch(`${API_URL}/auth/register`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(data)
            });
            const resData = await handleResponse(res);
            localStorage.setItem('token', resData.access_token);
            return resData.user;
        },
        profile: async () => {
            const res = await fetch(`${API_URL}/auth/profile`, {
                headers: getHeaders()
            });
            return handleResponse(res);
        },
        logout: () => {
            localStorage.removeItem('token');
        },
        getCaptcha: async () => {
            const res = await fetch(`${API_URL}/auth/captcha`);
            return handleResponse(res);
        },
        checkSession: async () => {
            const res = await fetch(`${API_URL}/auth/check`, { headers: getHeaders() });
            return handleResponse(res);
        }
    },
    store: {
        getProducts: async (): Promise<Product[]> => {
            const res = await fetch(`${API_URL}/store/products`, { headers: getHeaders() });
            return handleResponse(res);
        },
        createProduct: async (product: any) => {
            const res = await fetch(`${API_URL}/store/products`, {
                method: 'POST',
                headers: getHeaders(),
                body: JSON.stringify(product)
            });
            return handleResponse(res);
        },
        updateProduct: async (id: string, product: any) => {
            const res = await fetch(`${API_URL}/store/products/${id}`, {
                method: 'PUT',
                headers: getHeaders(),
                body: JSON.stringify(product)
            });
            return handleResponse(res);
        },
        deleteProduct: async (id: string) => {
            const res = await fetch(`${API_URL}/store/products/${id}`, {
                method: 'DELETE',
                headers: getHeaders()
            });
            return handleResponse(res);
        },
        checkout: async (products: any[], amount: number, couponCode?: string, payType?: string) => {
            const res = await fetch(`${API_URL}/store/checkout`, {
                method: 'POST',
                headers: getHeaders(),
                body: JSON.stringify({ products, amount, couponCode, payType })
            });
            return handleResponse(res);
        },
        getTransactions: async () => {
            const res = await fetch(`${API_URL}/store/transactions`, { headers: getHeaders() });
            return handleResponse(res);
        },
        addCoupon: async (productId: string, data: any) => {
            const res = await fetch(`${API_URL}/store/products/${productId}/coupons`, {
                method: 'POST',
                headers: getHeaders(),
                body: JSON.stringify(data)
            });
            return handleResponse(res);
        },
        deleteCoupon: async (couponId: string) => {
            const res = await fetch(`${API_URL}/store/coupons/${couponId}`, {
                method: 'DELETE',
                headers: getHeaders()
            });
            return handleResponse(res);
        },
        validateCoupon: async (code: string, productIds: string[]) => {
            const res = await fetch(`${API_URL}/store/validate-coupon`, {
                method: 'POST',
                headers: getHeaders(),
                body: JSON.stringify({ code, productIds })
            });
            return handleResponse(res);
        }
    },
    quiz: {
        getSubjects: async (): Promise<Subject[]> => {
            const res = await fetch(`${API_URL}/quiz/subjects`, { headers: getHeaders() });
            return handleResponse(res);
        },
        createSubject: async (data: any) => {
            const res = await fetch(`${API_URL}/quiz/subjects`, {
                method: 'POST',
                headers: getHeaders(),
                body: JSON.stringify(data)
            });
            return handleResponse(res);
        },
        updateSubject: async (id: string, data: any) => {
            const res = await fetch(`${API_URL}/quiz/subjects/${id}`, {
                method: 'PATCH',
                headers: getHeaders(),
                body: JSON.stringify(data)
            });
            return handleResponse(res);
        },
        deleteSubject: async (id: string) => {
            const res = await fetch(`${API_URL}/quiz/subjects/${id}`, {
                method: 'DELETE',
                headers: getHeaders()
            });
            return handleResponse(res);
        },
        getChapters: async (subjectId: string) => {
            const res = await fetch(`${API_URL}/quiz/${subjectId}/chapters`, { headers: getHeaders() });
            return handleResponse(res);
        },
        importQuestions: async (subjectId: string, files: File[]) => {
            const formData = new FormData();
            files.forEach(file => formData.append('files', file));
            const res = await fetch(`${API_URL}/quiz/subjects/${subjectId}/import`, {
                method: 'POST',
                headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` },
                body: formData
            });
            // return handleResponse(res); // handleResponse expects json but we might return simple success. Nestjs returns object.
            return handleResponse(res);
        },
        deleteChapter: async (id: string) => {
            const res = await fetch(`${API_URL}/quiz/chapters/${id}`, {
                method: 'DELETE',
                headers: getHeaders()
            });
            return handleResponse(res);
        },
        updateQuestion: async (id: number, data: any) => {
            const res = await fetch(`${API_URL}/quiz/questions/${id}`, {
                method: 'PATCH',
                headers: getHeaders(),
                body: JSON.stringify(data)
            });
            return handleResponse(res);
        },
        addComment: async (questionId: number, content: string) => {
            const res = await fetch(`${API_URL}/quiz/comments`, {
                method: 'POST',
                headers: getHeaders(),
                body: JSON.stringify({ questionId, content })
            });
            return handleResponse(res);
        },
        getComments: async (questionId: number) => {
            const res = await fetch(`${API_URL}/quiz/questions/${questionId}/comments`, { headers: getHeaders() });
            return handleResponse(res);
        }
    },
    wiki: {
        getCategories: async (): Promise<WikiCategory[]> => {
            const res = await fetch(`${API_URL}/wiki/categories`, { headers: getHeaders() });
            return handleResponse(res);
        },
        createCategory: async (data: any) => {
            const res = await fetch(`${API_URL}/wiki/categories`, {
                method: 'POST',
                headers: getHeaders(),
                body: JSON.stringify(data)
            });
            return handleResponse(res);
        },
        updateCategory: async (id: string, data: any) => {
            const res = await fetch(`${API_URL}/wiki/categories/${id}`, {
                method: 'PUT',
                headers: getHeaders(),
                body: JSON.stringify(data)
            });
            return handleResponse(res);
        },
        deleteCategory: async (id: string) => {
            const res = await fetch(`${API_URL}/wiki/categories/${id}`, {
                method: 'DELETE',
                headers: getHeaders()
            });
            return handleResponse(res);
        },
        getArticle: async (id: string) => {
            const res = await fetch(`${API_URL}/wiki/articles/${id}`, { headers: getHeaders() });
            return handleResponse(res);
        },
        createArticle: async (data: any) => {
            const res = await fetch(`${API_URL}/wiki/articles`, {
                method: 'POST',
                headers: getHeaders(),
                body: JSON.stringify(data)
            });
            return handleResponse(res);
        },
        updateArticle: async (id: string, data: any) => {
            const res = await fetch(`${API_URL}/wiki/articles/${id}`, {
                method: 'PUT',
                headers: getHeaders(),
                body: JSON.stringify(data)
            });
            return handleResponse(res);
        },
        deleteArticle: async (id: string) => {
            const res = await fetch(`${API_URL}/wiki/articles/${id}`, {
                method: 'DELETE',
                headers: getHeaders()
            });
            return handleResponse(res);
        },
        uploadFile: async (file: File) => {
            const formData = new FormData();
            formData.append('file', file);
            const res = await fetch(`${API_URL}/wiki/upload`, {
                method: 'POST',
                headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` },
                body: formData
            });
            return handleResponse(res);
        }
    },
    chat: {
        createCompletion: async (data: { model: string, messages: { role: string, content: string }[], stream?: boolean }) => {
            const res = await fetch(`${API_URL}/v1/chat/completions`, {
                method: 'POST',
                headers: getHeaders(),
                body: JSON.stringify(data)
            });
            return handleResponse(res);
        }
    },
    user: {
        updateProgress: async (key: string, data: any) => {
            const res = await fetch(`${API_URL}/users/progress`, {
                method: 'POST',
                headers: getHeaders(),
                body: JSON.stringify({ key, data })
            });
            return handleResponse(res);
        },
        toggleBookmark: async (item: any) => {
            const res = await fetch(`${API_URL}/users/bookmark`, {
                method: 'POST',
                headers: getHeaders(),
                body: JSON.stringify(item)
            });
            return handleResponse(res);
        },
        addQuizResult: async (result: any) => {
            const res = await fetch(`${API_URL}/users/quiz-history`, {
                method: 'POST',
                headers: getHeaders(),
                body: JSON.stringify(result)
            });
            return handleResponse(res);
        },
        updateSubscriptions: async (subscriptions: any[]) => {
            const res = await fetch(`${API_URL}/users/subscriptions`, {
                method: 'POST',
                headers: getHeaders(),
                body: JSON.stringify({ subscriptions })
            });
            return handleResponse(res);
        },
        adminUpdateSubscriptions: async (userId: string, subscriptions: any[]) => {
            const res = await fetch(`${API_URL}/users/${userId}/subscriptions`, {
                method: 'POST',
                headers: getHeaders(),
                body: JSON.stringify({ subscriptions })
            });
            return handleResponse(res);
        },
        createUser: async (data: any) => {
            const res = await fetch(`${API_URL}/users`, {
                method: 'POST',
                headers: getHeaders(),
                body: JSON.stringify(data)
            });
            return handleResponse(res);
        },
        getAllUsers: async () => {
            const res = await fetch(`${API_URL}/users`, { headers: getHeaders() });
            return handleResponse(res);
        },
        async deleteUser(id: string) {
            const res = await fetch(`${API_URL}/users/${id}`, {
                method: 'DELETE',
                headers: getHeaders()
            });
            return handleResponse(res);
        },
        async uploadAvatar(file: File) {
            const formData = new FormData();
            formData.append('file', file);
            const res = await fetch(`${API_URL}/users/avatar`, {
                method: 'POST',
                headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` },
                body: formData
            });
            return handleResponse(res);
        }
    },
    feedback: {
        submit: async (data: { type: string, content: string, contact?: string }) => {
            const res = await fetch(`${API_URL}/feedback`, {
                method: 'POST',
                headers: getHeaders(),
                body: JSON.stringify(data)
            });
            return handleResponse(res);
        },
        getAll: async () => {
            const res = await fetch(`${API_URL}/feedback`, { headers: getHeaders() });
            return handleResponse(res);
        },
        updateStatus: async (id: string, status: string) => {
            // Currently backend only toggles, so no need body if just toggle, 
            // but if we extend to set specific status, we might need body.
            // Our backend service `updateStatus` toggles it. 
            // Ideally it should accept status. 
            // Current Controller: Post(patch) ':id/status' -> service.updateStatus(id) (toggles)
            const res = await fetch(`${API_URL}/feedback/${id}/status`, {
                method: 'PATCH',
                headers: getHeaders(),
                body: JSON.stringify({ status })
            });
            return handleResponse(res);
        },
        delete: async (id: string) => {
            const res = await fetch(`${API_URL}/feedback/${id}`, {
                method: 'DELETE',
                headers: getHeaders()
            });
            return handleResponse(res);
        },

    },
    dashboard: {
        getStats: async () => {
            const res = await fetch(`${API_URL}/dashboard/stats`, { headers: getHeaders() });
            return handleResponse(res);
        },
        getRevenueTrend: async (days: number = 30) => {
            const res = await fetch(`${API_URL}/dashboard/revenue-trend?days=${days}`, { headers: getHeaders() });
            return handleResponse(res);
        },
        getSubjectDistribution: async () => {
            const res = await fetch(`${API_URL}/dashboard/subject-distribution`, { headers: getHeaders() });
            return handleResponse(res);
        }
    },
    settings: {
        getAll: async () => {
            const res = await fetch(`${API_URL}/settings`, { headers: getHeaders() });
            return handleResponse(res);
        },
        update: async (key: string, value: string) => {
            const res = await fetch(`${API_URL}/settings/${key}`, {
                method: 'PUT',
                headers: getHeaders(),
                body: JSON.stringify({ value })
            });
            return handleResponse(res);
        }
    },
    announcements: {
        getAll: async () => {
            const res = await fetch(`${API_URL}/announcements`, { headers: getHeaders() });
            return handleResponse(res);
        },
        getLatest: async () => {
            const res = await fetch(`${API_URL}/announcements/latest`, { headers: getHeaders() });
            return handleResponse(res);
        },
        create: async (data: { title: string, content: string, visible?: boolean }) => {
            const res = await fetch(`${API_URL}/announcements`, {
                method: 'POST',
                headers: getHeaders(),
                body: JSON.stringify(data)
            });
            return handleResponse(res);
        },
        update: async (id: string, data: { title?: string, content?: string, visible?: boolean }) => {
            const res = await fetch(`${API_URL}/announcements/${id}`, {
                method: 'PATCH',
                headers: getHeaders(),
                body: JSON.stringify(data)
            });
            return handleResponse(res);
        },
        delete: async (id: string) => {
            const res = await fetch(`${API_URL}/announcements/${id}`, {
                method: 'DELETE',
                headers: getHeaders()
            });
            return handleResponse(res);
        }
    }
};
