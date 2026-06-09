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
    <div className="fixed inset-0 z-[100] flex justify-end">
      {/* Backdrop */}
      <div
        className="absolute inset-0 bg-black/40 backdrop-blur-sm animate-in fade-in duration-300"
        onClick={onClose}
      />

      {/* Sidebar */}
      <div className="relative w-full max-w-md bg-white dark:bg-[#1e293b] h-full shadow-2xl border-l border-slate-200 dark:border-white/10 animate-in slide-in-from-right duration-300 flex flex-col">
        <div className="px-6 py-5 border-b border-slate-100 dark:border-white/5 flex items-center justify-between bg-slate-50 dark:bg-white/5">
          <div>
            <h3 className="text-xl font-bold text-slate-900 dark:text-white flex items-center gap-2">
              {session ? 'Edit Connection' : 'New Connection'}
            </h3>
            <p className="text-xs text-slate-500 mt-1">Configure your remote SSH access.</p>
          </div>
          <button onClick={onClose} className="p-2 hover:bg-slate-200 dark:hover:bg-white/10 rounded-xl text-slate-500 transition-colors">
            <X size={20} />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="flex-1 overflow-y-auto p-6 space-y-6 custom-scrollbar">
          <div className="space-y-5">
            <div>
              <label className="block text-xs font-bold text-slate-500 uppercase tracking-widest mb-2 ml-1">General Info</label>
              <div className="space-y-4">
                <div className="relative">
                  <div className="absolute left-3.5 top-1/2 -translate-y-1/2 text-slate-400"><Server size={18} /></div>
                  <input
                    required
                    type="text"
                    placeholder="Display Name (e.g. Prod Server)"
                    className="w-full pl-11 pr-4 py-3 bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-xl focus:ring-2 focus:ring-blue-500 outline-none text-slate-900 dark:text-white transition-all text-sm"
                    value={formData.name}
                    onChange={e => setFormData({ ...formData, name: e.target.value })}
                  />
                </div>
              </div>
            </div>

            <div>
              <label className="block text-xs font-bold text-slate-500 uppercase tracking-widest mb-2 ml-1">Network Settings</label>
              <div className="space-y-4">
                <div className="grid grid-cols-4 gap-3">
                  <div className="col-span-3 relative">
                    <div className="absolute left-3.5 top-1/2 -translate-y-1/2 text-slate-400"><Globe size={18} /></div>
                    <input
                      required
                      type="text"
                      placeholder="Hostname or IP Address"
                      className="w-full pl-11 pr-4 py-3 bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-xl focus:ring-2 focus:ring-blue-500 outline-none text-slate-900 dark:text-white transition-all text-sm"
                      value={formData.host}
                      onChange={e => setFormData({ ...formData, host: e.target.value })}
                    />
                  </div>
                  <div className="col-span-1 relative">
                    <input
                      required
                      type="number"
                      className="w-full px-3 py-3 bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-xl focus:ring-2 focus:ring-blue-500 outline-none text-slate-900 dark:text-white transition-all text-sm text-center"
                      value={formData.port}
                      onChange={e => setFormData({ ...formData, port: parseInt(e.target.value) })}
                    />
                  </div>
                </div>
                <div className="relative">
                  <div className="absolute left-3.5 top-1/2 -translate-y-1/2 text-slate-400"><User size={18} /></div>
                  <input
                    required
                    type="text"
                    placeholder="SSH Username"
                    className="w-full pl-11 pr-4 py-3 bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-xl focus:ring-2 focus:ring-blue-500 outline-none text-slate-900 dark:text-white transition-all text-sm"
                    value={formData.user}
                    onChange={e => setFormData({ ...formData, user: e.target.value })}
                  />
                </div>
              </div>
            </div>

            <div>
              <label className="block text-xs font-bold text-slate-500 uppercase tracking-widest mb-2 ml-1">Authentication</label>
              <div className="flex p-1.5 bg-slate-100 dark:bg-white/5 rounded-2xl mb-4">
                <button
                  type="button"
                  onClick={() => setFormData({ ...formData, authMethod: 'password' })}
                  className={clsx(
                    "flex-1 py-2.5 text-xs font-bold rounded-xl transition-all flex items-center justify-center gap-2",
                    formData.authMethod === 'password' ? "bg-white dark:bg-blue-600 text-blue-600 dark:text-white shadow-sm" : "text-slate-500 hover:text-slate-700"
                  )}
                >
                  <Lock size={14} /> Password
                </button>
                <button
                  type="button"
                  onClick={() => setFormData({ ...formData, authMethod: 'key' })}
                  className={clsx(
                    "flex-1 py-2.5 text-xs font-bold rounded-xl transition-all flex items-center justify-center gap-2",
                    formData.authMethod === 'key' ? "bg-white dark:bg-blue-600 text-blue-600 dark:text-white shadow-sm" : "text-slate-500 hover:text-slate-700"
                  )}
                >
                  <Key size={14} /> SSH Key
                </button>
              </div>

              {formData.authMethod === 'password' ? (
                <div className="relative animate-in fade-in slide-in-from-top-1 duration-200">
                  <div className="absolute left-3.5 top-1/2 -translate-y-1/2 text-slate-400"><Lock size={18} /></div>
                  <input
                    required
                    type="password"
                    placeholder="SSH Password"
                    className="w-full pl-11 pr-4 py-3 bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-xl focus:ring-2 focus:ring-blue-500 outline-none text-slate-900 dark:text-white transition-all text-sm"
                    value={formData.password}
                    onChange={e => setFormData({ ...formData, password: e.target.value })}
                  />
                </div>
              ) : (
                <div className="space-y-4 animate-in fade-in slide-in-from-top-1 duration-200">
                  <div className="relative">
                    <input
                      required
                      type="text"
                      placeholder="Private Key Path (e.g. ~/.ssh/id_rsa)"
                      className="w-full px-4 py-3 bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-xl focus:ring-2 focus:ring-blue-500 outline-none text-slate-900 dark:text-white transition-all text-sm"
                      value={formData.keyPath}
                      onChange={e => setFormData({ ...formData, keyPath: e.target.value })}
                    />
                  </div>
                  <div className="relative">
                    <input
                      type="password"
                      placeholder="Key Passphrase (optional)"
                      className="w-full px-4 py-3 bg-slate-50 dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-xl focus:ring-2 focus:ring-blue-500 outline-none text-slate-900 dark:text-white transition-all text-sm"
                      value={formData.passphrase}
                      onChange={e => setFormData({ ...formData, passphrase: e.target.value })}
                    />
                  </div>
                </div>
              )}
            </div>
          </div>
        </form>

        <div className="p-6 border-t border-slate-100 dark:border-white/5 bg-slate-50/50 dark:bg-white/5 flex gap-3">
          <button
            type="button"
            onClick={onClose}
            className="flex-1 py-3.5 text-sm font-bold text-slate-600 dark:text-slate-400 hover:bg-slate-200/50 dark:hover:bg-white/5 rounded-2xl transition-colors"
          >
            Discard
          </button>
          <button
            onClick={handleSubmit}
            disabled={saving}
            className="flex-[2] py-3.5 bg-blue-600 hover:bg-blue-700 text-white text-sm font-bold rounded-2xl shadow-xl shadow-blue-500/20 flex items-center justify-center gap-2 transition-all disabled:opacity-50 active:scale-[0.98]"
          >
            {saving ? <RefreshCw className="animate-spin" size={18} /> : <Save size={18} />}
            {session ? 'Update Connection' : 'Create Connection'}
          </button>
        </div>
      </div>
    </div>
  );
};

export default SSHSessionForm;
