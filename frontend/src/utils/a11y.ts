import React from 'react';

export const handleActionKey = (callback: (e?: any) => void) => (e: React.KeyboardEvent) => {
  if (e.key === 'Enter' || e.key === ' ') {
    e.preventDefault();
    callback(e);
  }
};
