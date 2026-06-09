import React, { useState } from 'react';
import { X, Save, Server, User, Globe, Lock, Key, Hash } from 'lucide-react';
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
    <div className="fixed inset-0 bg-black/60 backdrop-blur-sm z-[100] flex items-center justify-center p-4">
      <div className="bg-white dark:bg-[#1e293b] w-full max-w-lg rounded-2xl shadow-2xl overflow-hidden border border-slate-200 dark:border-white/10 animate-in fade-in zoom-in duration-200">
        <div className="px-6 py-4 border-b border-slate-100 dark:border-white/5 flex items-center justify-between bg-slate-50 dark:bg-white/5">
          <h3 className="text-lg font-bold text-slate-900 dark:text-white flex items-center gap-2">
            {session ? 'Edit SSH Connection' : 'New SSH Connection'}
          </h3>
          <button onClick={onClose} className="p-1 hover:bg-slate-200 dark:hover:bg-white/10 rounded-full text-slate-500 transition-colors">
            <X size={20} />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="p-6 space-y-4">
          <div className="space-y-4">
            <div>
              <label className="block text-xs font-semibold text-slate-500 uppercase tracking-wider mb-1.5 ml-1">Session Name</label>
              <div className="relative">
                <div className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400"><Server size={18} /></div>
                <input
                  required
                  type="text"
                  placeholder="Production Web Server"
                  className="w-full pl-10 pr-4 py-2.5 bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-xl focus:ring-2 focus:ring-blue-500 outline-none text-slate-900 dark:text-white transition-all"
                  value={formData.name}
                  onChange={e => setFormData({ ...formData, name: e.target.value })}
                />
              </div>
            </div>

            <div className="grid grid-cols-4 gap-4">
              <div className="col-span-3">
                <label className="block text-xs font-semibold text-slate-500 uppercase tracking-wider mb-1.5 ml-1">Host / IP</label>
                <div className="relative">
                  <div className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400"><Globe size={18} /></div>
                  <input
                    required
                    type="text"
                    placeholder="192.168.1.100"
                    className="w-full pl-10 pr-4 py-2.5 bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-xl focus:ring-2 focus:ring-blue-500 outline-none text-slate-900 dark:text-white transition-all"
                    value={formData.host}
                    onChange={e => setFormData({ ...formData, host: e.target.value })}
                  />
                </div>
              </div>
              <div className="col-span-1">
                <label className="block text-xs font-semibold text-slate-500 uppercase tracking-wider mb-1.5 ml-1">Port</label>
                <div className="relative">
                  <div className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400"><Hash size={16} /></div>
                  <input
                    required
                    type="number"
                    className="w-full pl-9 pr-2 py-2.5 bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-xl focus:ring-2 focus:ring-blue-500 outline-none text-slate-900 dark:text-white transition-all"
                    value={formData.port}
                    onChange={e => setFormData({ ...formData, port: parseInt(e.target.value) })}
                  />
                </div>
              </div>
            </div>

            <div>
              <label className="block text-xs font-semibold text-slate-500 uppercase tracking-wider mb-1.5 ml-1">Username</label>
              <div className="relative">
                <div className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400"><User size={18} /></div>
                <input
                  required
                  type="text"
                  placeholder="root"
                  className="w-full pl-10 pr-4 py-2.5 bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-xl focus:ring-2 focus:ring-blue-500 outline-none text-slate-900 dark:text-white transition-all"
                  value={formData.user}
                  onChange={e => setFormData({ ...formData, user: e.target.value })}
                />
              </div>
            </div>

            <div className="pt-2">
              <label className="block text-xs font-semibold text-slate-500 uppercase tracking-wider mb-3 ml-1">Authentication</label>
              <div className="flex p-1 bg-slate-100 dark:bg-white/5 rounded-xl">
                <button
                  type="button"
                  onClick={() => setFormData({ ...formData, authMethod: 'password' })}
                  className={clsx(
                    "flex-1 py-2 text-sm font-medium rounded-lg transition-all flex items-center justify-center gap-2",
                    formData.authMethod === 'password' ? "bg-white dark:bg-blue-600 text-blue-600 dark:text-white shadow-sm" : "text-slate-500 hover:text-slate-700"
                  )}
                >
                  <Lock size={16} /> Password
                </button>
                <button
                  type="button"
                  onClick={() => setFormData({ ...formData, authMethod: 'key' })}
                  className={clsx(
                    "flex-1 py-2 text-sm font-medium rounded-lg transition-all flex items-center justify-center gap-2",
                    formData.authMethod === 'key' ? "bg-white dark:bg-blue-600 text-blue-600 dark:text-white shadow-sm" : "text-slate-500 hover:text-slate-700"
                  )}
                >
                  <Key size={16} /> SSH Key
                </button>
              </div>
            </div>

            {formData.authMethod === 'password' ? (
              <div>
                <label className="block text-xs font-semibold text-slate-500 uppercase tracking-wider mb-1.5 ml-1">Password</label>
                <div className="relative">
                  <div className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400"><Lock size={18} /></div>
                  <input
                    required
                    type="password"
                    placeholder="••••••••"
                    className="w-full pl-10 pr-4 py-2.5 bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-xl focus:ring-2 focus:ring-blue-500 outline-none text-slate-900 dark:text-white transition-all"
                    value={formData.password}
                    onChange={e => setFormData({ ...formData, password: e.target.value })}
                  />
                </div>
              </div>
            ) : (
              <div className="space-y-4 animate-in slide-in-from-top-2 duration-200">
                <div>
                  <label className="block text-xs font-semibold text-slate-500 uppercase tracking-wider mb-1.5 ml-1">Private Key Path</label>
                  <input
                    required
                    type="text"
                    placeholder="C:\Users\Name\.ssh\id_rsa"
                    className="w-full px-4 py-2.5 bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-xl focus:ring-2 focus:ring-blue-500 outline-none text-slate-900 dark:text-white transition-all text-sm"
                    value={formData.keyPath}
                    onChange={e => setFormData({ ...formData, keyPath: e.target.value })}
                  />
                </div>
                <div>
                  <label className="block text-xs font-semibold text-slate-500 uppercase tracking-wider mb-1.5 ml-1">Passphrase (Optional)</label>
                  <input
                    type="password"
                    placeholder="••••••••"
                    className="w-full px-4 py-2.5 bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-xl focus:ring-2 focus:ring-blue-500 outline-none text-slate-900 dark:text-white transition-all text-sm"
                    value={formData.passphrase}
                    onChange={e => setFormData({ ...formData, passphrase: e.target.value })}
                  />
                </div>
              </div>
            )}
          </div>

          <div className="pt-4 flex gap-3">
            <button
              type="button"
              onClick={onClose}
              className="flex-1 py-3 text-sm font-bold text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-white/5 rounded-xl transition-colors"
            >
              Cancel
            </button>
            <button
              disabled={saving}
              type="submit"
              className="flex-[2] py-3 bg-blue-600 hover:bg-blue-700 text-white text-sm font-bold rounded-xl shadow-lg shadow-blue-500/25 flex items-center justify-center gap-2 transition-all disabled:opacity-50"
            >
              {saving ? <div className="w-5 h-5 border-2 border-white/30 border-t-white rounded-full animate-spin" /> : <Save size={18} />}
              {session ? 'Update Session' : 'Save Connection'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

export default SSHSessionForm;
