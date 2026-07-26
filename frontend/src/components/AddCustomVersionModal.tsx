import React, { useState } from "react";
import {
  X,
  UploadCloud,
  Folder,
  FileArchive,
  Loader2,
  AlertCircle,
  CheckCircle2,
} from "lucide-react";
import * as AppBackend from "../../wailsjs/go/backend/App";
import { handleActionKey } from "../utils/a11y";

interface AddCustomVersionModalProps {
  isOpen: boolean;
  onClose: () => void;
  serviceName: string;
  onSuccess: () => void;
}

const AddCustomVersionModal: React.FC<AddCustomVersionModalProps> = ({
  isOpen,
  onClose,
  serviceName,
  onSuccess,
}) => {
  const [dragActive, setDragActive] = useState(false);
  const [processing, setProcessing] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [selectedPath, setSelectedPath] = useState<string>("");

  if (!isOpen) return null;

  const handleDrag = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    if (e.type === "dragenter" || e.type === "dragover") {
      setDragActive(true);
    } else if (e.type === "dragleave") {
      setDragActive(false);
    }
  };

  const processPath = async (path: string) => {
    setProcessing(true);
    setError(null);
    setSuccess(null);
    try {
      await AppBackend.ProcessCustomVersion(serviceName, path);
      setSuccess(`Successfully processed custom version for ${serviceName}!`);
      onSuccess();
      setTimeout(() => {
        setSuccess(null);
        setSelectedPath("");
        onClose();
      }, 2000);
    } catch (err: any) {
      setError(err.message || err.toString());
    } finally {
      setProcessing(false);
    }
  };

  const handleDrop = async (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setDragActive(false);

    if (processing) return;

    if (e.dataTransfer.files && e.dataTransfer.files[0]) {
      const file = e.dataTransfer.files[0];
      const path = (file as any).path || file.name;
      if (path) {
        setSelectedPath(path);
        await processPath(path);
      }
    }
  };

  const handleSelectFile = async () => {
    if (processing) return;
    setError(null);
    try {
      const res = await AppBackend.OpenFileDialog({
        Title: `Select Custom ${serviceName} ZIP File`,
        Filters: [
          {
            DisplayName: "ZIP/Nuget Archives (*.zip;*.nupkg)",
            Pattern: "*.zip;*.nupkg",
          },
        ],
      });
      if (res) {
        setSelectedPath(res);
        await processPath(res);
      }
    } catch (err: any) {
      setError(err.message || err.toString());
    }
  };

  const handleSelectFolder = async () => {
    if (processing) return;
    setError(null);
    try {
      const res = await AppBackend.OpenDirectoryDialog({
        Title: `Select Custom ${serviceName} Folder`,
      });
      if (res) {
        setSelectedPath(res);
        await processPath(res);
      }
    } catch (err: any) {
      setError(err.message || err.toString());
    }
  };

  return (
    <div className="fixed inset-0 z-[1000] flex items-center justify-center p-4 sm:p-6">
      {/* Backdrop */}
      <button
        type="button"
        className="absolute inset-0 bg-slate-900/60 animate-in fade-in duration-300 w-full h-full border-none p-0 cursor-default focus:outline-none"
        onKeyDown={handleActionKey(onClose)}
        onClick={onClose}
      />

      {/* Modal */}
      <div className="relative w-full max-w-lg bg-white dark:bg-slate-900 rounded-sm shadow-2xl border border-slate-200 dark:border-white/10 overflow-hidden animate-in zoom-in-95 fade-in duration-200">
        {/* Header */}
        <div className="px-6 py-4 border-b border-slate-100 dark:border-white/5 flex items-center justify-between">
          <h3 className="text-sm font-black text-slate-900 dark:text-white uppercase italic tracking-tighter flex items-center gap-2">
            <UploadCloud size={16} className="text-blue-500" />
            Add Custom {serviceName} Version
          </h3>
          <button
            type="button"
            onClick={onClose}
            className="text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 transition-colors"
          >
            <X size={18} />
          </button>
        </div>

        {/* Content */}
        <div className="p-6 space-y-4">
          <p className="text-[10px] font-bold text-slate-500 dark:text-slate-400 leading-relaxed uppercase tracking-widest">
            Drag & drop a ZIP archive or a direct folder of {serviceName} below.
            We will automatically extract and process it as a ready-to-use version.
          </p>

          {/* Dropzone Area */}
          <div
            onDragEnter={handleDrag}
            onDragOver={handleDrag}
            onDragLeave={handleDrag}
            onDrop={handleDrop}
            className={`border-2 border-dashed rounded-sm p-8 flex flex-col items-center justify-center gap-4 transition-all ${
              dragActive
                ? "border-blue-500 bg-blue-500/10 scale-[1.02]"
                : "border-slate-300 dark:border-white/10 hover:border-slate-400 dark:hover:border-white/20 bg-slate-50 dark:bg-white/5"
            } ${processing ? "opacity-50 cursor-not-allowed" : "cursor-pointer"}`}
          >
            <UploadCloud
              size={48}
              className={`transition-colors ${
                dragActive ? "text-blue-500" : "text-slate-400"
              }`}
            />

            <div className="text-center space-y-1">
              <p className="text-xs font-black text-slate-700 dark:text-slate-300 uppercase tracking-wider">
                {dragActive ? "Drop files here!" : "Drag & Drop ZIP / Folder here"}
              </p>
              <p className="text-[9px] font-bold text-slate-400 dark:text-slate-500 uppercase tracking-widest">
                or use the selection buttons below
              </p>
            </div>

            {selectedPath && (
              <div className="flex items-center gap-2 px-3 py-1.5 bg-slate-100 dark:bg-black/20 border border-slate-200 dark:border-white/10 rounded-sm max-w-full">
                {selectedPath.toLowerCase().endsWith(".zip") ||
                selectedPath.toLowerCase().endsWith(".nupkg") ? (
                  <FileArchive size={14} className="text-blue-500 shrink-0" />
                ) : (
                  <Folder size={14} className="text-amber-500 shrink-0" />
                )}
                <span className="text-[9px] font-bold text-slate-600 dark:text-slate-400 truncate max-w-[280px]">
                  {selectedPath}
                </span>
              </div>
            )}
          </div>

          {/* Action Buttons */}
          <div className="flex items-center justify-center gap-4">
            <button
              type="button"
              disabled={processing}
              onClick={handleSelectFile}
              className="flex items-center gap-2 px-4 py-2.5 bg-blue-600 hover:bg-blue-500 text-white rounded-sm text-[10px] font-black uppercase tracking-widest transition-all hover:scale-105 active:scale-95 disabled:opacity-50 disabled:pointer-events-none shadow-lg shadow-blue-500/20"
            >
              <FileArchive size={14} />
              Select ZIP File
            </button>
            <button
              type="button"
              disabled={processing}
              onClick={handleSelectFolder}
              className="flex items-center gap-2 px-4 py-2.5 bg-emerald-600 hover:bg-emerald-500 text-white rounded-sm text-[10px] font-black uppercase tracking-widest transition-all hover:scale-105 active:scale-95 disabled:opacity-50 disabled:pointer-events-none shadow-lg shadow-emerald-500/20"
            >
              <Folder size={14} />
              Select Folder
            </button>
          </div>

          {/* Feedback/Status blocks */}
          {processing && (
            <div className="flex items-center justify-center gap-2 text-blue-500 py-2">
              <Loader2 size={16} className="animate-spin" />
              <span className="text-[10px] font-black uppercase tracking-widest animate-pulse">
                Processing and validating version...
              </span>
            </div>
          )}

          {error && (
            <div className="flex items-start gap-2.5 p-3 bg-rose-500/10 border border-rose-500/20 rounded-sm text-rose-600 dark:text-rose-400 select-text cursor-text">
              <AlertCircle size={16} className="shrink-0 mt-0.5" />
              <div className="flex-1 space-y-1 select-text cursor-text">
                <p className="text-[10px] font-black uppercase tracking-wider select-text cursor-text">
                  Failed to Process Version
                </p>
                <p className="text-[9px] font-bold leading-relaxed select-text cursor-text">{error}</p>
              </div>
            </div>
          )}

          {success && (
            <div className="flex items-center gap-2.5 p-3 bg-emerald-500/10 border border-emerald-500/20 rounded-sm text-emerald-600 dark:text-emerald-400 select-text cursor-text">
              <CheckCircle2 size={16} className="shrink-0" />
              <span className="text-[10px] font-black uppercase tracking-wider select-text cursor-text">
                {success}
              </span>
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="px-6 py-4 bg-slate-50 dark:bg-white/5 flex items-center justify-end">
          <button
            type="button"
            disabled={processing}
            onClick={onClose}
            className="px-4 py-2 rounded-sm text-[10px] font-black uppercase tracking-widest text-slate-500 hover:text-slate-900 dark:text-slate-400 dark:hover:text-white transition-all disabled:opacity-50"
          >
            Close
          </button>
        </div>
      </div>
    </div>
  );
};

export default AddCustomVersionModal;
