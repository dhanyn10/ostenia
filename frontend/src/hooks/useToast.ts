import { useState, useCallback } from 'react';

export interface ToastItem {
  id: string;
  title: string;
  message: string;
  type: 'info' | 'success' | 'warn' | 'error';
}

export function useToast() {
  const [toasts, setToasts] = useState<ToastItem[]>([]);

  const removeToast = useCallback((id: string) => {
    setToasts(prev => {
      const next: ToastItem[] = [];
      for (const t of prev) {
        if (t.id !== id) {
          next.push(t);
        }
      }
      return next;
    });
  }, []);

  const addToast = useCallback((
    title: string,
    message: string,
    type: 'info' | 'success' | 'warn' | 'error' = 'info'
  ) => {
    const id = crypto.randomUUID();
    setToasts(prev => [...prev, { id, title, message, type }]);
    setTimeout(() => {
      removeToast(id);
    }, 5000);
  }, [removeToast]);

  return { toasts, setToasts, addToast, removeToast };
}
