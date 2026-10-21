import React from 'react';

export const handleActionKey = (callback: (event?: any) => void) => (event: React.KeyboardEvent) => {
  if (event.key === 'Enter' || event.key === ' ') {
    event.preventDefault();
    callback(event);
  }
};
