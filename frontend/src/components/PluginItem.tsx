import React, { useState } from "react";
import {
  CheckCircle2,
  Trash2,
  ChevronDown,
  ChevronUp,
  Download,
} from "lucide-react";
import VersionDropdown from "@/components/VersionDropdown";
import CircularProgress from "@/components/CircularProgress";
import {
  cn,
  handleActionKey,
  shouldIgnoreCardToggle,
  isCustomVersionAllowed,
  isParentTaskActive,
  isTaskVersionInstalled,
} from "@/utils/utils";


interface PluginModuleItemProps {
  module: any;
  taskName: string;
  taskInstalled: boolean;
  progress: any;
  onInstallModule: (taskName: string, moduleName: string) => void;
  onUninstallModule: (taskName: string, moduleName: string) => void;
}

const PluginModuleItem: React.FC<PluginModuleItemProps> = ({
  module,
  taskName,
  taskInstalled,
  progress,
  onInstallModule,
  onUninstallModule,
}) => {
  const moduleProgress = progress[module.name];
  const isModuleActive =
    Boolean(moduleProgress) &&
    moduleProgress.status !== "Ready" &&
    moduleProgress.status !== "Completed" &&
    !moduleProgress.status?.startsWith("Error");
  const isModuleInstalled =
    Boolean(module.isInstalled) ||
    module.status === "Ready" ||
    module.status === "Completed";

  let buttonClass = "text-blue-500 hover:bg-blue-500/10 hover:scale-110";
  let buttonTitle = `Install ${module.name}`;
  if (isModuleInstalled) {
    buttonClass = "text-emerald-500 cursor-not-allowed";
    buttonTitle = "Installed";
  } else if (!taskInstalled) {
    buttonClass = "text-slate-300 cursor-not-allowed";
    buttonTitle = `Install ${taskName} first`;
  }

  return (
    <div className="flex items-center justify-between p-3 bg-slate-50 dark:bg-white/5 rounded-sm border border-slate-100 dark:border-white/5 group/mod">
      <div className="flex items-center gap-3">
        <div
          className={cn(
            "w-6 h-6 rounded-sm flex items-center justify-center text-[10px] font-black",
            isModuleInstalled
              ? "bg-emerald-500/10 text-emerald-600"
              : "bg-blue-500/10 text-blue-600",
          )}
        >
          {module.name[0]}
        </div>
        <div className="flex items-center gap-3">
          <div className="flex flex-col">
            <span className="text-xs font-bold text-slate-700 dark:text-slate-300">
              {module.name}
            </span>
            <p className="text-[8px] text-slate-400 font-bold uppercase tracking-tighter mt-0.5">
              {module.status}
            </p>
          </div>

          {module.version && (
            <span className="text-[8px] font-bold text-slate-500 uppercase tracking-widest bg-slate-100 dark:bg-white/5 px-1.5 py-0.5 rounded-sm h-fit">
              v{module.version}
            </span>
          )}
        </div>
      </div>

      <div className="flex items-center gap-4">
        {isModuleActive && (
          <div className="flex items-center gap-3">
            <div className="flex flex-col items-end">
              <span className="text-[8px] font-black text-blue-500 uppercase">
                {moduleProgress.status}
              </span>
              <span className="text-[7px] text-slate-400 font-bold">
                {moduleProgress.percentage.toFixed(0)}%
              </span>
            </div>
            <div className="w-8 h-8">
              <CircularProgress
                percentage={moduleProgress.percentage}
                size={32}
              />
            </div>
          </div>
        )}

        {!isModuleActive && (
          <div className="flex items-center gap-2">
            {isModuleInstalled && (
              <button
                type="button"
                onClick={() => onUninstallModule(taskName, module.name)}
                className="p-1.5 bg-slate-100 dark:bg-white/5 hover:bg-rose-500/10 text-slate-400 dark:text-slate-500 hover:text-rose-600 dark:hover:text-rose-400 rounded-sm transition-all border border-slate-200 dark:border-white/5"
                title={`Uninstall ${module.name}`}
              >
                <Trash2 size={12} />
              </button>
            )}

            <button
              type="button"
              disabled={isModuleInstalled || !taskInstalled}
              onClick={() => onInstallModule(taskName, module.name)}
              className={cn("p-2 rounded-sm transition-all", buttonClass)}
              title={buttonTitle}
            >
              {isModuleInstalled ? (
                <CheckCircle2 size={16} />
              ) : (
                <Download size={16} />
              )}
            </button>
          </div>
        )}
      </div>
    </div>
  );
};

interface PluginModuleListProps {
  modules: any[];
  taskName: string;
  taskInstalled: boolean;
  progress: any;
  onInstallModule: (taskName: string, moduleName: string) => void;
  onUninstallModule: (taskName: string, moduleName: string) => void;
}

const PluginModuleList: React.FC<PluginModuleListProps> = ({
  modules,
  taskName,
  taskInstalled,
  progress,
  onInstallModule,
  onUninstallModule,
}) => (
  <div
    data-no-card-toggle="true"
    className="mt-4 pt-4 border-t border-slate-200 dark:border-white/5 space-y-2 animate-in slide-in-from-top-2 duration-200 cursor-default relative z-10 pointer-events-auto"
  >
    <p className="text-[9px] font-black text-slate-400 uppercase tracking-widest mb-3">
      Available Modules
    </p>
    <div className="grid grid-cols-1 gap-2">
      {modules.map((module) => (
        <PluginModuleItem
          key={module.name}
          module={module}
          taskName={taskName}
          taskInstalled={taskInstalled}
          progress={progress}
          onInstallModule={onInstallModule}
          onUninstallModule={onUninstallModule}
        />
      ))}
    </div>
  </div>
);



function PluginItem({
  task,
  progress,
  isDropdownOpen,
  onDropdownToggle,
  selectedVersion,
  onVersionChange,
  onDeleteVersion,
  onInstall,
  onCancel,
  onOpenFolder,
  renderIcon,
  onInstallModule,
  onUninstallModule,
}) {
  const [isModulesExpanded, setIsModulesExpanded] = useState(false);
  const availableVersions = task.versions || [];
  const installedVersions = task.installedVers || [];
  const dropdownOptions = [
    ...new Set([...availableVersions, ...installedVersions]),
  ];
  const isCustomAllowed = isCustomVersionAllowed(task.name);

  const parentProgress = progress[task.name];
  const isActive = isParentTaskActive(parentProgress);
  const isSelectedInstalled = isTaskVersionInstalled(
    task,
    selectedVersion,
    installedVersions,
  );
  const hasModules = Boolean(task.modules && task.modules.length > 0);

  const toggleModules = () => setIsModulesExpanded((prev) => !prev);
  const cardRef = React.useRef<HTMLDivElement>(null);

  React.useEffect(() => {
    const cardElement = cardRef.current;
    if (!cardElement) return;

    const onCardClick = (event: MouseEvent) => {
      const target = event.target as HTMLElement;
      if (
        !hasModules ||
        target.closest("[data-card-expand]") ||
        shouldIgnoreCardToggle(target)
      ) {
        return;
      }
      toggleModules();
    };

    cardElement.addEventListener("click", onCardClick);
    return () => {
      cardElement.removeEventListener("click", onCardClick);
    };
  }, [hasModules]);

  return (
    <div
      ref={cardRef}
      data-card-root="true"
      className={cn(
        "bg-white/70 dark:bg-slate-900/40 rounded-sm p-4 border border-slate-200 dark:border-white/5 hover:border-slate-300 dark:hover:border-white/10 transition-all group flex flex-col gap-0 relative shadow-sm dark:shadow-lg",
        hasModules && "hover:bg-slate-50/80 dark:hover:bg-slate-900/60 cursor-pointer",
        isDropdownOpen ? "z-[100]" : "z-0",
      )}
    >
      {hasModules && (
        <button
          type="button"
          data-card-expand="true"
          aria-label={
            isModulesExpanded
              ? `Collapse ${task.name} modules`
              : `Expand ${task.name} modules`
          }
          aria-expanded={isModulesExpanded}
          onClick={(event) => {
            event.stopPropagation();
            setIsModulesExpanded((prev) => !prev);
          }}
          className="absolute inset-0 w-full h-full cursor-pointer bg-transparent border-none p-0 z-0 focus:outline-none focus:ring-1 focus:ring-blue-500/40 rounded-sm"
        />
      )}

      <div className="flex items-center gap-6 w-full relative z-10 pointer-events-auto">
        <div
          className={cn(
            "w-10 h-10 rounded-sm flex items-center justify-center shadow-lg transition-transform group-hover:scale-105 shrink-0",
            task.isInstalled
              ? "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400"
              : "bg-blue-500/10 text-blue-600 dark:text-blue-400",
          )}
        >
          {renderIcon(task.name, 18, "text-slate-900 dark:text-white")}
        </div>

        <div className="flex-1 min-w-0 pointer-events-auto">
          <div className="flex items-center gap-3 flex-wrap">
            <h3 className="text-base font-black text-slate-900 dark:text-white uppercase italic tracking-tighter">
              {task.name}
            </h3>

            {availableVersions.length > 0 || isCustomAllowed ? (
              <VersionDropdown
                current={selectedVersion || task.version}
                options={dropdownOptions}
                isOpen={isDropdownOpen}
                onToggle={onDropdownToggle}
                onChange={onVersionChange}
                allowCustom={isCustomAllowed}
                onCustomClick={() => onOpenFolder(task.name)}
              />
            ) : (
              <span className="text-[9px] font-bold text-slate-500 uppercase tracking-widest bg-slate-100 dark:bg-white/5 px-1.5 py-0.5 rounded-sm">
                v{task.version}
              </span>
            )}

            {task.info && (
              <span className="text-[9px] font-bold text-blue-500 dark:text-blue-400 uppercase tracking-widest bg-blue-500/10 px-1.5 py-0.5 rounded-sm border border-blue-500/20">
                {task.info}
              </span>
            )}

            {installedVersions.map((version) => (
              <button
                type="button"
                key={version}
                onClick={(event) => {
                  event.stopPropagation();
                  onDeleteVersion(task.name, version);
                }}
                onKeyDown={handleActionKey(() =>
                  onDeleteVersion(task.name, version),
                )}
                className="group/tag flex items-center gap-1.5 px-2 py-0.5 bg-slate-100 dark:bg-slate-800/80 border border-slate-200 dark:border-white/10 text-slate-500 dark:text-slate-400 text-[8px] font-bold uppercase tracking-widest rounded-sm hover:border-rose-500/30 hover:bg-rose-500/10 hover:text-rose-400 transition-all cursor-pointer shadow-sm outline-none focus:ring-1 focus:ring-rose-500/40"
                title={`Delete v${version}`}
              >
                <Trash2
                  size={10}
                  className="w-0 opacity-0 group-hover/tag:w-2.5 group-hover/tag:opacity-100 transition-all text-rose-500"
                />
                {version}
              </button>
            ))}
          </div>
        </div>

        {hasModules && (
          <button
            type="button"
            data-card-expand="true"
            onClick={(event) => {
              event.stopPropagation();
              setIsModulesExpanded(!isModulesExpanded);
            }}
            aria-label={
              isModulesExpanded ? "Collapse modules" : "Expand modules"
            }
            className="p-2 hover:bg-slate-100 dark:hover:bg-white/5 rounded-sm transition-all text-slate-400 hover:text-slate-900 dark:hover:text-white pointer-events-auto cursor-pointer"
          >
            {isModulesExpanded ? (
              <ChevronUp size={16} />
            ) : (
              <ChevronDown size={16} />
            )}
          </button>
        )}

        <div className="flex items-center gap-6 shrink-0 pointer-events-auto">
          {isActive && (
            <div className="animate-in fade-in slide-in-from-right-2 duration-300">
              <CircularProgress
                percentage={parentProgress.percentage}
                status={parentProgress.status}
                speed={parentProgress.speed}
                downloaded={parentProgress.downloaded}
                onCancel={() => onCancel(task.name)}
              />
            </div>
          )}

          {!isActive && (
            <button
              type="button"
              disabled={isSelectedInstalled}
              onClick={(event) => {
                event.stopPropagation();
                if (!isSelectedInstalled) onInstall(task);
              }}
              className={cn(
                "px-5 py-2 rounded-sm text-[9px] font-black uppercase tracking-widest transition-all flex items-center justify-center gap-2 shadow-lg",
                isSelectedInstalled
                  ? "bg-emerald-500/10 text-emerald-600 dark:text-emerald-500 border border-emerald-500/20 cursor-pointer pointer-events-none"
                  : "bg-blue-600 hover:bg-blue-500 text-white hover:scale-105 cursor-pointer",
              )}
            >
              {isSelectedInstalled && <CheckCircle2 size={14} />}
              {isSelectedInstalled ? "Ready" : "Download"}
            </button>
          )}
        </div>
      </div>

      {isModulesExpanded && hasModules && (
        <PluginModuleList
          modules={task.modules}
          taskName={task.name}
          taskInstalled={task.isInstalled}
          progress={progress}
          onInstallModule={onInstallModule}
          onUninstallModule={onUninstallModule}
        />
      )}
    </div>
  );
}

export default PluginItem;
