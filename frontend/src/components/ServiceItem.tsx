import React from 'react';
import { Activity, Globe, Trash2, FolderOpen, Clock, Lock, Unlock, Terminal, ChevronDown, Monitor, CheckCircle2, Settings2 } from 'lucide-react';
import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';
function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

interface ServiceItemProps {
  service: any;
  task: any;
  isExpanded: boolean;
  onToggleAccordion: (name: string, hasExtraActions: boolean) => void;
  renderIcon: (name: string, size?: number, className?: string) => React.ReactNode;
  handleToggleService: (name: string, status: string) => void;
  handleRemoveFromHome: (name: string) => void;
  handleSwitchVersion: (name: string, version: string) => void;
  handleOpenLocalTerminal: (name: string, type: string) => void;
  handleToggleHttps: (name: string) => void;
  openTerminalDropdown: string | null;
  setOpenTerminalDropdown: (name: string | null) => void;
  setIsModalOpen: (open: boolean) => void;
  apacheHttps: boolean;
  nginxHttps: boolean;
  isOpenSslEnabled: boolean;
  setActiveTab: (tab: string) => void;
  handleOpenPluginFolder: (name: string) => void;
}

const ServiceIdentity: React.FC<any> = ({ service, renderIcon }) => (
  <div className="flex items-center gap-3">
    {renderIcon(service.name, 18, "text-slate-900 dark:text-white")}
    <h3 className="text-base font-black text-slate-900 dark:text-white uppercase italic tracking-tighter">{service.name}</h3>
  </div>
);

const VersionSwitcher: React.FC<any> = ({ service, task, handleSwitchVersion }) => {
  const installedVersions = task?.installedVers || [];
  const show = (service.name === 'PHP' || service.name === 'Node.js' || service.name === 'Python') && installedVersions.length > 0;

  if (!show) return null;

  return (
    <div className="flex items-center gap-1 ml-1" onClick={(e) => e.stopPropagation()} onKeyDown={(e) => e.stopPropagation()}>
      {installedVersions.map((ver: string) => {
        const systemString = (service.activeVersion || "").toString().toLowerCase().trim();
        const cleanVer = ver.toString().replace(/^v/, "").replace(/^[a-z. ]+-/, "").trim();
        const isActive = systemString.includes(cleanVer.toLowerCase());

        return (
          <button
            key={ver}
            type="button"
            onClick={() => handleSwitchVersion(service.name, ver)}
            className={cn(
              "px-1.5 py-0.5 rounded-sm text-[8px] font-black uppercase tracking-widest border transition-all",
              isActive
                ? "bg-blue-600 border-blue-500 text-white shadow-lg"
                : "bg-slate-100 dark:bg-slate-800 border-slate-200 dark:border-white/5 text-slate-400 dark:text-slate-500 hover:border-blue-500/50 hover:text-blue-500"
            )}
          >
            {isActive && <CheckCircle2 size={8} className="inline mr-1" />}
            {cleanVer}
          </button>
        );
      })}
    </div>
  );
};

const ServiceStatus: React.FC<any> = ({ service }) => {
  const isRunning = service.status === 'Running';
  const showStats = isRunning && service.name !== 'OpenSSL';

  return (
    <div className="flex items-center gap-3 flex-wrap">
      {service.remainingDays > 0 && (
        <div className="flex items-center gap-1 px-1.5 py-0.5 bg-indigo-500/10 border border-indigo-500/20 rounded-sm text-[8px] font-bold text-indigo-600 dark:text-indigo-400 uppercase tracking-widest">
          <Clock size={10} />
          {service.remainingDays} Days Left
        </div>
      )}

      <div className={cn(
        "text-[8px] font-black uppercase tracking-[0.2em] px-2 py-0.5 rounded-sm border flex items-center gap-1.5",
        isRunning
          ? "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border-emerald-500/20"
          : "bg-slate-100 dark:bg-slate-900/80 text-slate-400 dark:text-slate-500 border-slate-200 dark:border-white/5"
      )}>
        {isRunning && <div className="w-1 h-1 rounded-full bg-emerald-500 animate-pulse" />}
        {service.status}
      </div>

      {showStats && (
        <div className="flex items-center gap-2 animate-in fade-in slide-in-from-left-2 duration-300">
          {service.pid > 0 && (
            <div className="flex items-center gap-1 px-1.5 py-0.5 bg-blue-500/10 border border-blue-500/20 rounded-sm text-[8px] font-bold text-blue-600 dark:text-blue-400 uppercase tracking-widest">
              <Activity size={10} />
              PID: {service.pid}
            </div>
          )}
          {((service.ports && service.ports.length > 0) || service.port > 0) && (
            <div className="flex items-center gap-1 px-1.5 py-0.5 bg-amber-500/10 border border-amber-500/20 rounded-sm text-[8px] font-bold text-amber-600 dark:text-amber-400 uppercase tracking-widest">
              <Globe size={10} />
              Port: {service.ports && service.ports.length > 0 ? service.ports.join(', ') : service.port}
            </div>
          )}
        </div>
      )}
    </div>
  );
};

const MainActions: React.FC<any> = ({ isInstalled, service, handleToggleService, handleRemoveFromHome, setActiveTab }) => (
  <div className="flex items-center gap-3" onClick={(e) => e.stopPropagation()} onKeyDown={(e) => e.stopPropagation()}>
    {!isInstalled && service.name !== 'OpenSSL' ? (
      <button type="button" onClick={() => setActiveTab('plugins')} className="px-4 py-1.5 bg-blue-600/10 hover:bg-blue-600/20 text-blue-600 dark:text-blue-400 rounded-sm text-[9px] font-black uppercase tracking-widest transition-all border border-blue-500/20">Install First</button>
    ) : (
      <button
        type="button"
        onClick={() => handleToggleService(service.name, service.status)}
        className={cn(
          "w-12 h-6 rounded-sm p-0.5 transition-all duration-300 ease-in-out relative ring-1 ring-inset",
          service.status === 'Running'
            ? "bg-emerald-500 ring-emerald-400/50"
            : "bg-slate-200 dark:bg-slate-800 ring-slate-300 dark:ring-white/5"
        )}
      >
        <div className={cn(
          "w-5 h-5 bg-white rounded-sm transition-all duration-300 shadow-lg",
          service.status === 'Running' ? "translate-x-6" : "translate-x-0"
        )} />
      </button>
    )}

    <button
      type="button"
      onClick={() => handleRemoveFromHome(service.name)}
      className="h-6 px-3 bg-slate-100 dark:bg-white/5 hover:bg-rose-500/10 text-slate-400 dark:text-slate-500 hover:text-rose-600 dark:hover:text-rose-400 rounded-sm text-[9px] font-black uppercase tracking-widest transition-all border border-slate-200 dark:border-white/5"
    >
      <Trash2 size={12} />
    </button>
  </div>
);

const ServiceExtraActions: React.FC<any> = ({
  service,
  isExpanded,
  hasPhpExtManager,
  hasOpenFolder,
  hasHeidiOpen,
  hasTerminal,
  hasHttpsToggle,
  setIsModalOpen,
  handleOpenPluginFolder,
  openTerminalDropdown,
  setOpenTerminalDropdown,
  handleOpenLocalTerminal,
  handleToggleHttps,
  isHttpsEnabled
}) => (
  <div
    className={cn(
      "transition-all duration-300 ease-in-out overflow-visible",
      isExpanded ? "max-h-24 opacity-100 mt-4" : "max-h-0 opacity-0 mt-0 overflow-hidden"
    )}
    onClick={(e) => e.stopPropagation()}
    onKeyDown={(e) => e.stopPropagation()}
  >
    <div className="flex items-center flex-wrap gap-4 px-1 pb-2">
      {hasPhpExtManager && (
        <button
          type="button"
          onClick={() => setIsModalOpen(true)}
          className="flex items-center gap-2 px-3 py-1.5 h-8 bg-slate-100 dark:bg-white/5 hover:bg-blue-600/10 text-slate-500 dark:text-slate-400 hover:text-blue-600 dark:hover:text-blue-400 rounded-sm text-[9px] font-black uppercase tracking-widest transition-all border border-slate-200 dark:border-white/5"
        >
          <Settings2 size={14} /> Extensions
        </button>
      )}

      {hasOpenFolder && (
        <button type="button" onClick={() => handleOpenPluginFolder(service.name)} className="w-8 h-8 flex items-center justify-center bg-slate-100 dark:bg-white/5 hover:bg-blue-600/10 text-slate-500 dark:text-slate-400 hover:text-blue-600 dark:hover:text-blue-400 rounded-sm border border-slate-200 dark:border-white/5 transition-all" title="Open Folder">
          <FolderOpen size={16} />
        </button>
      )}

      {hasHeidiOpen && (
        <button
          type="button"
          onClick={() => (window as any).go.main.App.OpenHeidiSQL()}
          className="flex items-center gap-2 px-3 py-1.5 h-8 bg-blue-600/10 hover:bg-blue-600/20 text-blue-600 dark:text-blue-400 rounded-sm text-[9px] font-black uppercase tracking-widest transition-all border border-blue-500/20"
        >
          <Monitor size={14} /> Open HeidiSQL
        </button>
      )}

      {hasTerminal && (
        <div className="relative">
          <button
            type="button"
            onClick={() => setOpenTerminalDropdown(openTerminalDropdown === service.name ? null : service.name)}
            className={cn(
              "w-12 h-8 flex items-center justify-center gap-1 rounded-sm border border-slate-200 dark:border-white/5 transition-all",
              openTerminalDropdown === service.name ? "bg-slate-200 dark:bg-slate-700 text-slate-900 dark:text-white" : "bg-slate-100 dark:bg-white/5 text-slate-500 dark:text-slate-400"
            )}
            title="Terminal"
          >
            <Terminal size={16} /> <ChevronDown size={10} />
          </button>

          {openTerminalDropdown === service.name && (
            <>
              <button type="button" className="fixed inset-0 z-[150] w-full h-full bg-transparent border-none p-0 cursor-default focus:outline-none" onClick={() => setOpenTerminalDropdown(null)} />
              <div className="absolute top-full left-0 mt-1 w-40 bg-white dark:bg-slate-900 border border-slate-200 dark:border-white/10 rounded-sm shadow-2xl z-[160] animate-in fade-in slide-in-from-top-1 duration-200">
                <div className="p-1">
                  <button type="button" onClick={() => handleOpenLocalTerminal(service.name, 'cmd')} className="w-full flex items-center gap-3 px-3 py-1.5 rounded-sm text-[10px] font-bold text-slate-500 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-white/5 hover:text-slate-900 dark:hover:text-white transition-all text-left"><Monitor size={12} className="text-blue-500" /> CMD</button>
                  <button type="button" onClick={() => handleOpenLocalTerminal(service.name, 'powershell')} className="w-full flex items-center gap-3 px-3 py-1.5 rounded-sm text-[10px] font-bold text-slate-500 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-white/5 hover:text-slate-900 dark:hover:text-white transition-all text-left"><Monitor size={12} className="text-blue-600" /> PowerShell</button>
                </div>
              </div>
            </>
          )}
        </div>
      )}

      {hasHttpsToggle && (
        <button
          type="button"
          onClick={() => handleToggleHttps(service.name)}
          className={cn(
            "w-14 h-8 rounded-sm p-1 transition-all duration-300 ease-in-out relative ring-1 ring-inset",
            isHttpsEnabled
              ? "bg-rose-500 ring-rose-400/50 shadow-[0_0_10px_rgba(244,63,94,0.3)]"
              : "bg-slate-200 dark:bg-slate-800 ring-slate-300 dark:ring-white/5"
          )}
          title={isHttpsEnabled ? "Disable HTTPS" : "Enable HTTPS"}
        >
          <div className={cn(
            "w-6 h-6 bg-white rounded-sm transition-all duration-300 shadow-lg flex items-center justify-center",
            isHttpsEnabled ? "translate-x-6" : "translate-x-0"
          )}>
            {isHttpsEnabled ? <Lock size={14} className="text-rose-600" /> : <Unlock size={14} className="text-slate-400" />}
          </div>
        </button>
      )}
    </div>
  </div>
);

const ServiceItem: React.FC<ServiceItemProps> = ({
  service,
  task,
  isExpanded,
  onToggleAccordion,
  renderIcon,
  handleToggleService,
  handleRemoveFromHome,
  handleSwitchVersion,
  handleOpenLocalTerminal,
  handleToggleHttps,
  openTerminalDropdown,
  setOpenTerminalDropdown,
  setIsModalOpen,
  apacheHttps,
  nginxHttps,
  isOpenSslEnabled,
  setActiveTab,
  handleOpenPluginFolder
}) => {
  const isInstalled = (task?.installedVers && task.installedVers.length > 0) || service.name === 'OpenSSL';
  const isWebServer = service.name === 'Apache' || service.name === 'Nginx';

  let isHttpsEnabled = false;
  if (service.name === 'Apache') {
    isHttpsEnabled = apacheHttps;
  } else if (service.name === 'Nginx') {
    isHttpsEnabled = nginxHttps;
  }

  const hasTerminalFacility = service.name !== 'HeidiSQL' && service.name !== 'OpenSSL';
  const hasOpenFolder = isInstalled && service.name !== 'OpenSSL' && service.name !== 'HeidiSQL';
  const hasTerminal = isInstalled && hasTerminalFacility;
  const hasHttpsToggle = isWebServer && isOpenSslEnabled;
  const hasPhpExtManager = service.name === 'PHP';
  const hasHeidiOpen = service.name === 'HeidiSQL' && isInstalled;
  const hasExtraActions = hasOpenFolder || hasTerminal || hasHttpsToggle || hasPhpExtManager || hasHeidiOpen;

  return (
    <div
      className={cn(
        "w-full text-left bg-white/70 dark:bg-slate-900/40 rounded-sm p-4 border border-slate-200 dark:border-white/5 hover:border-slate-300 dark:hover:border-white/10 transition-all flex flex-col relative shadow-sm dark:shadow-lg",
        isExpanded ? "z-[100] ring-1 ring-blue-500/20" : "z-10"
      )}
    >
      <div className="flex items-center gap-5">
        <div className="flex-1 min-w-0 flex items-center gap-3 flex-wrap px-2">
          <button
            type="button"
            className={cn(
              "flex items-center gap-3 outline-none focus:ring-1 focus:ring-blue-500/40 rounded-sm py-1 transition-all",
              hasExtraActions ? "cursor-pointer hover:bg-slate-50 dark:hover:bg-white/5" : "cursor-default"
            )}
            onClick={() => onToggleAccordion(service.name, hasExtraActions)}
          >
            <ServiceIdentity service={service} renderIcon={renderIcon} />
            <ServiceStatus service={service} />
          </button>

          <VersionSwitcher service={service} task={task} handleSwitchVersion={handleSwitchVersion} />
        </div>

        <MainActions isInstalled={isInstalled} service={service} handleToggleService={handleToggleService} handleRemoveFromHome={handleRemoveFromHome} setActiveTab={setActiveTab} />
      </div>

      {hasExtraActions && (
        <ServiceExtraActions
          service={service}
          isExpanded={isExpanded}
          hasPhpExtManager={hasPhpExtManager}
          hasOpenFolder={hasOpenFolder}
          hasHeidiOpen={hasHeidiOpen}
          hasTerminal={hasTerminal}
          hasHttpsToggle={hasHttpsToggle}
          setIsModalOpen={setIsModalOpen}
          handleOpenPluginFolder={handleOpenPluginFolder}
          openTerminalDropdown={openTerminalDropdown}
          setOpenTerminalDropdown={setOpenTerminalDropdown}
          handleOpenLocalTerminal={handleOpenLocalTerminal}
          handleToggleHttps={handleToggleHttps}
          isHttpsEnabled={isHttpsEnabled}
        />
      )}
    </div>
  );
}

export default ServiceItem;
