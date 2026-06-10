import React, { useState } from 'react';
import { X, Save, Server, User, Globe, Lock, Key, Hash, RefreshCw } from 'lucide-react';
import { clsx } from 'clsx';
import * as AppBackend from '../../wailsjs/go/main/App';

const SSHSessionForm = ({ session, onClose, onSave, addToast }) => {
  const [formData, setFormData] = useState(session || {
    id: Math.random().toString(36).substr(2, 9),
    name: '',
    host: '',
    port: 22,
    user: 'root',
    authMethod: 'password',
    password: '',
    keyPath: '',
    passphrase: '',
    createdAt: Date.now()
  });

  const [saving, setSaving] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setSaving(true);
    try {
      if (session) {
        await AppBackend.UpdateSSHSession(formData);
      } else {
        await AppBackend.AddSSHSession(formData);
      }
      addToast('Success', `Session ${session ? 'updated' : 'created'} successfully`);
      onSave();
    } catch (err) {
      addToast('Error', 'Failed to save session: ' + err, 'error');
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="w-[380px] bg-white dark:bg-[#1e293b] border-l border-slate-200 dark:border-white/10 flex flex-col h-full animate-in slide-in-from-right duration-200 shrink-0">
      <div className="px-5 py-4 border-b border-slate-100 dark:border-white/5 flex items-center justify-between bg-slate-50 dark:bg-slate-800/50">
        <div>
          <h3 className="text-sm font-bold text-slate-900 dark:text-white uppercase tracking-wider">
            {session ? 'Edit Connection' : 'New Connection'}
          </h3>
        </div>
        <button onClick={onClose} className="p-1.5 hover:bg-slate-200 dark:hover:bg-white/10 rounded text-slate-500 transition-colors">
          <X size={18} />
        </button>
      </div>

      <form onSubmit={handleSubmit} className="flex-1 overflow-y-auto p-5 space-y-6 custom-scrollbar">
        <div className="space-y-4">
          <div>
            <label className="block text-[10px] font-bold text-slate-400 uppercase tracking-widest mb-1.5 ml-0.5">Label</label>
            <input
              required
              type="text"
              placeholder="e.g. Production Web"
              className="w-full px-3 py-2 bg-slate-100 dark:bg-white/5 border border-transparent focus:border-blue-500 focus:bg-white dark:focus:bg-slate-900 rounded-md outline-none text-slate-900 dark:text-white transition-all text-sm"
              value={formData.name}
              onChange={e => setFormData({ ...formData, name: e.target.value })}
            />
          </div>

          <div className="grid grid-cols-4 gap-2">
            <div className="col-span-3">
              <label className="block text-[10px] font-bold text-slate-400 uppercase tracking-widest mb-1.5 ml-0.5">Address</label>
              <input
                required
                type="text"
                placeholder="1.2.3.4 or example.com"
                className="w-full px-3 py-2 bg-slate-100 dark:bg-white/5 border border-transparent focus:border-blue-500 focus:bg-white dark:focus:bg-slate-900 rounded-md outline-none text-slate-900 dark:text-white transition-all text-sm"
                value={formData.host}
                onChange={e => setFormData({ ...formData, host: e.target.value })}
              />
            </div>
            <div className="col-span-1">
              <label className="block text-[10px] font-bold text-slate-400 uppercase tracking-widest mb-1.5 ml-0.5">Port</label>
              <input
                required
                type="number"
                className="w-full px-2 py-2 bg-slate-100 dark:bg-white/5 border border-transparent focus:border-blue-500 focus:bg-white dark:focus:bg-slate-900 rounded-md outline-none text-slate-900 dark:text-white transition-all text-sm text-center"
                value={formData.port}
                onChange={e => setFormData({ ...formData, port: parseInt(e.target.value) })}
              />
            </div>
          </div>

          <div>
            <label className="block text-[10px] font-bold text-slate-400 uppercase tracking-widest mb-1.5 ml-0.5">Username</label>
            <input
              required
              type="text"
              placeholder="root"
              className="w-full px-3 py-2 bg-slate-100 dark:bg-white/5 border border-transparent focus:border-blue-500 focus:bg-white dark:focus:bg-slate-900 rounded-md outline-none text-slate-900 dark:text-white transition-all text-sm"
              value={formData.user}
              onChange={e => setFormData({ ...formData, user: e.target.value })}
            />
          </div>

          <div>
            <label className="block text-[10px] font-bold text-slate-400 uppercase tracking-widest mb-1.5 ml-0.5">Authentication Method</label>
            <div className="flex p-1 bg-slate-100 dark:bg-slate-900 rounded-md">
              <button
                type="button"
                onClick={() => setFormData({ ...formData, authMethod: 'password' })}
                className={clsx(
                  "flex-1 py-1.5 text-[11px] font-bold rounded transition-all",
                  formData.authMethod === 'password' ? "bg-white dark:bg-blue-600 text-blue-600 dark:text-white shadow-sm" : "text-slate-500 hover:text-slate-700 dark:hover:text-slate-300"
                )}
              >
                Password
              </button>
              <button
                type="button"
                onClick={() => setFormData({ ...formData, authMethod: 'key' })}
                className={clsx(
                  "flex-1 py-1.5 text-[11px] font-bold rounded transition-all",
                  formData.authMethod === 'key' ? "bg-white dark:bg-blue-600 text-blue-600 dark:text-white shadow-sm" : "text-slate-500 hover:text-slate-700 dark:hover:text-slate-300"
                )}
              >
                Key File
              </button>
            </div>
          </div>

          {formData.authMethod === 'password' ? (
            <div>
              <label className="block text-[10px] font-bold text-slate-400 uppercase tracking-widest mb-1.5 ml-0.5">Password</label>
              <input
                required
                type="password"
                placeholder="••••••••"
                className="w-full px-3 py-2 bg-slate-100 dark:bg-white/5 border border-transparent focus:border-blue-500 focus:bg-white dark:focus:bg-slate-900 rounded-md outline-none text-slate-900 dark:text-white transition-all text-sm"
                value={formData.password}
                onChange={e => setFormData({ ...formData, password: e.target.value })}
              />
            </div>
          ) : (
            <div className="space-y-4">
              <div>
                <label className="block text-[10px] font-bold text-slate-400 uppercase tracking-widest mb-1.5 ml-0.5">Private Key Path</label>
                <input
                  required
                  type="text"
                  placeholder="e.g. /home/user/.ssh/id_rsa"
                  className="w-full px-3 py-2 bg-slate-100 dark:bg-white/5 border border-transparent focus:border-blue-500 focus:bg-white dark:focus:bg-slate-900 rounded-md outline-none text-slate-900 dark:text-white transition-all text-sm"
                  value={formData.keyPath}
                  onChange={e => setFormData({ ...formData, keyPath: e.target.value })}
                />
              </div>
              <div>
                <label className="block text-[10px] font-bold text-slate-400 uppercase tracking-widest mb-1.5 ml-0.5">Passphrase</label>
                <input
                  type="password"
                  placeholder="optional"
                  className="w-full px-3 py-2 bg-slate-100 dark:bg-white/5 border border-transparent focus:border-blue-500 focus:bg-white dark:focus:bg-slate-900 rounded-md outline-none text-slate-900 dark:text-white transition-all text-sm"
                  value={formData.passphrase}
                  onChange={e => setFormData({ ...formData, passphrase: e.target.value })}
                />
              </div>
            </div>
          )}
        </div>
      </form>

      <div className="p-4 border-t border-slate-100 dark:border-white/5 bg-white dark:bg-[#1e293b] flex gap-2">
        <button
          type="button"
          onClick={onClose}
          className="flex-1 py-2 text-xs font-bold text-slate-500 hover:text-slate-900 dark:hover:text-white transition-colors border border-slate-200 dark:border-white/10 rounded-md"
        >
          Cancel
        </button>
        <button
          onClick={handleSubmit}
          disabled={saving}
          className="flex-1 py-2 bg-blue-600 hover:bg-blue-700 text-white text-xs font-bold rounded-md flex items-center justify-center gap-2 transition-all disabled:opacity-50"
        >
          {saving ? <RefreshCw className="animate-spin" size={14} /> : <Save size={14} />}
          {session ? 'Update' : 'Save'}
        </button>
      </div>
    </div>
  );
};

export default SSHSessionForm;
