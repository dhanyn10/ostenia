import React from 'react';
import { Upload, Download } from 'lucide-react';
import * as AppBackend from '../../../wailsjs/go/backend/App';
import { handleActionKey } from "../../utils/a11y";

interface ProfileCategoryProps {
  initApp?: () => void;
}

const ProfileCategory: React.FC<ProfileCategoryProps> = ({ initApp }) => {
  const handleExport = async (type: 'all' | 'config' | 'ssh') => {
    try {
      await AppBackend.ExportProfile(type === 'all' || type === 'config', type === 'all' || type === 'ssh');
    } catch (err) { console.error(err); }
  };

  const handleImport = async () => {
    try {
      await AppBackend.ImportProfile();
      if (initApp) initApp();
    } catch (err) { console.error(err); }
  };

  return (
    <div className="space-y-8 animate-in fade-in slide-in-from-right-4 duration-300">
      <div>
        <h3 className="text-xl font-bold mb-1 tracking-tight text-mui-grey-900 dark:text-white">User Profile</h3>
        <p className="text-sm text-mui-grey-500 dark:text-mui-grey-400">Manage your application profile and portable data.</p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <button
          type="button"
          onKeyDown={handleActionKey(handleImport)} onClick={handleImport}
          className="flex items-start gap-4 p-4 rounded-lg border border-mui-grey-200 dark:border-white/10 hover:bg-mui-grey-50 dark:hover:bg-white/5 transition-all group bg-transparent text-left"
        >
          <div className="p-3 rounded-full bg-mui-blue-500/10 text-mui-blue-500 group-hover:scale-110 transition-transform">
            <Upload size={24} />
          </div>
          <div className="text-left">
            <div className="font-bold text-mui-grey-900 dark:text-white">Import Profile</div>
            <p className="text-[11px] text-mui-grey-500 mt-1 uppercase tracking-wider font-bold">Restore settings from JSON</p>
          </div>
        </button>

        <button
          type="button"
          onKeyDown={handleActionKey(() => handleExport('all'))} onClick={() => handleExport('all')}
          className="flex items-start gap-4 p-4 rounded-lg border border-mui-grey-200 dark:border-white/10 hover:bg-mui-grey-50 dark:hover:bg-white/5 transition-all group bg-transparent text-left"
        >
          <div className="p-3 rounded-full bg-emerald-500/10 text-emerald-500 group-hover:scale-110 transition-transform">
            <Download size={24} />
          </div>
          <div className="text-left">
            <div className="font-bold text-mui-grey-900 dark:text-white">Export All</div>
            <p className="text-[11px] text-mui-grey-500 mt-1 uppercase tracking-wider font-bold">Backup config and SSH sessions</p>
          </div>
        </button>
      </div>

      <div className="pt-4 border-t border-mui-grey-100 dark:border-white/5">
        <h4 className="text-xs font-black text-mui-grey-400 uppercase tracking-[0.2em] mb-4">Granular Export</h4>
        <div className="flex gap-3">
          <button onClick={() => handleExport('config')} className="px-4 py-2 rounded border border-mui-grey-200 dark:border-white/10 text-xs font-bold text-mui-grey-700 dark:text-mui-grey-300 hover:bg-mui-grey-50 dark:hover:bg-white/5 transition-colors">Config Only</button>
          <button onClick={() => handleExport('ssh')} className="px-4 py-2 rounded border border-mui-grey-200 dark:border-white/10 text-xs font-bold text-mui-grey-700 dark:text-mui-grey-300 hover:bg-mui-grey-50 dark:hover:bg-white/5 transition-colors">SSH Sessions Only</button>
        </div>
      </div>
    </div>
  );
};

export default ProfileCategory;
